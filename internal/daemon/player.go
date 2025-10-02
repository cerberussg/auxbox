package daemon

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cerberussg/auxbox/internal/shared"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// Player handles audio playback using the beep library
type Player struct {
	currentTrack *shared.Track
	status       PlayerStatus
	mu           sync.RWMutex

	// Beep-related fields
	streamer       beep.StreamSeekCloser
	format         beep.Format
	volumeStreamer *effects.Volume
	ctrl           *beep.Ctrl
	file           io.ReadCloser
	speakerInit    bool
}

// NewPlayer creates a new audio player instance
func NewPlayer() *Player {
	return &Player{
		status: PlayerStatus{
			IsPlaying: false,
			IsPaused:  false,
			Position:  "0:00",
			Duration:  "0:00",
			Volume:    1.0,
		},
		speakerInit: false,
	}
}

// stopAndCleanup stops current playback and cleans up resources
func (p *Player) stopAndCleanup() {
	// Stop the speaker from playing this stream
	if p.ctrl != nil {
		p.ctrl.Paused = true
		// Clear the speaker queue to stop any current playback
		speaker.Clear()
	}

	// Clean up resources
	p.cleanup()

	// Reset status
	p.status.IsPlaying = false
	p.status.IsPaused = false
	p.status.Position = "0:00"
}

// playInternal starts playback without locking (internal use)
func (p *Player) playInternal() error {
	if p.currentTrack == nil {
		return fmt.Errorf("no track loaded")
	}

	if p.ctrl == nil {
		return fmt.Errorf("audio not initialized")
	}

	// Start fresh playback
	p.ctrl.Paused = false
	speaker.Play(p.ctrl)
	p.status.IsPlaying = true
	p.status.IsPaused = false

	return nil
}

// Play starts or resumes audio playback
func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentTrack == nil {
		return fmt.Errorf("no track loaded")
	}

	if p.ctrl == nil {
		return fmt.Errorf("audio not initialized")
	}

	// If already playing, do nothing
	if p.status.IsPlaying {
		return nil
	}

	// If paused, just resume
	if p.status.IsPaused {
		p.ctrl.Paused = false
		p.status.IsPlaying = true
		p.status.IsPaused = false
	} else {
		// Start fresh playback
		return p.playInternal()
	}

	return nil
}

// Pause pauses audio playback
func (p *Player) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.status.IsPlaying {
		return fmt.Errorf("playback is not active")
	}

	if p.ctrl == nil {
		return fmt.Errorf("audio not initialized")
	}

	// Pause the audio
	p.ctrl.Paused = true
	p.status.IsPlaying = false
	p.status.IsPaused = true

	return nil
}

// Stop stops audio playback completely
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl != nil {
		p.ctrl.Paused = true
	}

	// Reset to beginning if possible
	if p.streamer != nil {
		if seeker, ok := p.streamer.(beep.StreamSeeker); ok {
			seeker.Seek(0)
		}
	}

	p.status.IsPlaying = false
	p.status.IsPaused = false
	p.status.Position = "0:00"

	return nil
}

// SetCurrentTrack sets the track to be played
func (p *Player) SetCurrentTrack(track *shared.Track) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Remember if we were playing
	wasPlaying := p.status.IsPlaying

	// Stop and clean up previous track
	p.stopAndCleanup()

	p.currentTrack = track
	p.status.Position = "0:00"
	p.status.IsPlaying = false
	p.status.IsPaused = false

	if track == nil {
		p.status.Duration = "0:00"
		return
	}

	// Load the new audio file
	fmt.Printf("Loading audio file: %s\n", track.Path)
	if err := p.loadAudioFile(track.Path); err != nil {
		fmt.Printf("Error loading audio file %s: %v\n", track.Path, err)
		p.status.Duration = "0:00"
		p.currentTrack = nil // Reset track on failure
		return
	}
	fmt.Printf("Audio file loaded successfully: %s\n", track.Path)

	// If we were playing before, start playing the new track
	if wasPlaying {
		p.playInternal()
	}
}

// GetCurrentTrack returns the currently loaded track
func (p *Player) GetCurrentTrack() *shared.Track {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentTrack
}

// GetStatus returns the current player status
func (p *Player) GetStatus() PlayerStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (p *Player) SetVolume(volume float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	p.status.Volume = volume

	// Apply volume to beep if available
	if p.volumeStreamer != nil {
		if volume == 0.0 {
			p.volumeStreamer.Silent = true
		} else {
			p.volumeStreamer.Silent = false
			// Convert 0.0-1.0 to beep's volume scale
			// beep uses logarithmic scale where 0 = no change
			// We'll use a simple linear conversion for now
			beepVolume := (volume - 1.0) * 2.0 // -2.0 to 0.0 range
			p.volumeStreamer.Volume = beepVolume
		}
	}

	return nil
}

// GetPosition returns the current playback position
func (p *Player) GetPosition() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.streamer == nil {
		return time.Duration(0)
	}

	if seeker, ok := p.streamer.(beep.StreamSeeker); ok {
		position := seeker.Position()
		return p.format.SampleRate.D(position)
	}

	return time.Duration(0)
}

// GetDuration returns the total duration of the current track
func (p *Player) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.streamer == nil {
		return time.Duration(0)
	}

	if seeker, ok := p.streamer.(beep.StreamSeeker); ok {
		length := seeker.Len()
		return p.format.SampleRate.D(length)
	}

	return time.Duration(0)
}

// UpdatePosition updates the current position in the status (should be called periodically)
func (p *Player) UpdatePosition() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status.IsPlaying && p.streamer != nil {
		if seeker, ok := p.streamer.(beep.StreamSeeker); ok {
			position := seeker.Position()
			duration := p.format.SampleRate.D(position)
			p.status.Position = formatDuration(duration)
		}
	}
}

// Close cleans up all audio resources
func (p *Player) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cleanup()
	return nil
}

// IsPlaying returns true if audio is currently playing
func (p *Player) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status.IsPlaying
}

// IsPaused returns true if playback is paused
func (p *Player) IsPaused() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status.IsPaused
}

// loadAudioFile loads an audio file and prepares it for playback
func (p *Player) loadAudioFile(filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Determine file type and decode
	ext := getFileExtension(filePath)
	switch ext {
	case ".mp3":
		streamer, format, err := mp3.Decode(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to decode MP3: %w", err)
		}
		p.streamer = streamer
		p.format = format
		p.file = file

	case ".wav":
		streamer, format, err := wav.Decode(file)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to decode WAV: %w", err)
		}
		p.streamer = streamer
		p.format = format
		p.file = file

	default:
		file.Close()
		return fmt.Errorf("unsupported audio format: %s (supported: .mp3, .wav)", ext)
	}

	// Initialize speaker if not already done
	if !p.speakerInit {
		err := p.initializeSpeakerWithFallbacks()
		if err != nil {
			p.cleanup()
			return fmt.Errorf("failed to initialize speaker: %w", err)
		}
		p.speakerInit = true
	}

	// Calculate duration
	p.calculateDuration()

	// Set up volume control
	fmt.Printf("Setting up volume control...\n")
	p.setupVolumeControl()

	// Verify control was set up properly
	if p.ctrl == nil {
		return fmt.Errorf("volume control setup failed - ctrl is nil")
	}
	fmt.Printf("Volume control setup complete\n")

	return nil
}

// getFileExtension returns the lowercase file extension
func getFileExtension(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return strings.ToLower(filePath[i:])
		}
		if filePath[i] == '/' || filePath[i] == '\\' {
			break
		}
	}
	return ""
}

// calculateDuration calculates and sets the track duration
func (p *Player) calculateDuration() {
	if p.streamer == nil {
		p.status.Duration = "0:00"
		return
	}

	// Get the length of the streamer if possible
	if seeker, ok := p.streamer.(beep.StreamSeeker); ok {
		currentPos := seeker.Position()
		length := seeker.Len()

		// Seek back to the beginning
		seeker.Seek(currentPos)

		// Calculate duration
		duration := p.format.SampleRate.D(length)
		p.status.Duration = formatDuration(duration)
	} else {
		p.status.Duration = "Unknown"
	}
}

// setupVolumeControl creates the volume control streamer
func (p *Player) setupVolumeControl() {
	if p.streamer == nil {
		return
	}

	// Create volume control wrapper
	p.volumeStreamer = &effects.Volume{
		Streamer: p.streamer,
		Base:     2.0,
		Volume:   (p.status.Volume - 1.0) * 2.0, // Apply current volume setting
		Silent:   p.status.Volume == 0.0,
	}

	// Add a callback to detect when track ends
	trackEndCallback := beep.Callback(func() {
		p.mu.Lock()
		p.status.IsPlaying = false
		p.status.IsPaused = false
		p.mu.Unlock()
	})

	// Create a sequence that plays the track then calls the callback
	sequence := beep.Seq(p.volumeStreamer, trackEndCallback)

	// Create control wrapper for pause/resume
	p.ctrl = &beep.Ctrl{
		Streamer: sequence,
		Paused:   true, // Start paused
	}
}

// cleanup cleans up audio resources
func (p *Player) cleanup() {
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	if p.file != nil {
		p.file.Close()
		p.file = nil
	}
	p.volumeStreamer = nil
	p.ctrl = nil
}

// initializeSpeakerWithFallbacks attempts to initialize the speaker with multiple fallback strategies
func (p *Player) initializeSpeakerWithFallbacks() error {
	if p.format.SampleRate == 0 {
		return fmt.Errorf("invalid sample rate: %d", p.format.SampleRate)
	}

	// Strategy 1: Standard initialization with current format
	bufferSize := p.format.SampleRate.N(time.Second / 10)
	err := speaker.Init(p.format.SampleRate, bufferSize)
	if err == nil {
		fmt.Printf("Audio initialized successfully with sample rate: %d Hz, buffer size: %d\n",
			p.format.SampleRate, bufferSize)
		return nil
	}

	fmt.Printf("Primary audio initialization failed: %v\n", err)

	// Strategy 2: Try with larger buffer for stability
	bufferSize = p.format.SampleRate.N(time.Second / 5)
	err = speaker.Init(p.format.SampleRate, bufferSize)
	if err == nil {
		fmt.Printf("Audio initialized with larger buffer: %d Hz, buffer size: %d\n",
			p.format.SampleRate, bufferSize)
		return nil
	}

	fmt.Printf("Large buffer initialization failed: %v\n", err)

	// Strategy 3: Try with common sample rates as fallbacks
	fallbackRates := []beep.SampleRate{44100, 48000, 22050, 16000}
	for _, rate := range fallbackRates {
		if rate == p.format.SampleRate {
			continue // Already tried this rate
		}

		bufferSize = rate.N(time.Second / 10)
		err = speaker.Init(rate, bufferSize)
		if err == nil {
			fmt.Printf("Audio initialized with fallback sample rate: %d Hz (original: %d Hz)\n",
				rate, p.format.SampleRate)
			// Update format to match what was actually initialized
			p.format.SampleRate = rate
			return nil
		}
		fmt.Printf("Fallback rate %d Hz failed: %v\n", rate, err)
	}

	// Strategy 4: Platform-specific troubleshooting hints
	platformHint := p.getPlatformAudioHint()

	return fmt.Errorf("all audio initialization strategies failed. %s. Last error: %w", platformHint, err)
}

// getPlatformAudioHint provides platform-specific troubleshooting information
func (p *Player) getPlatformAudioHint() string {
	switch runtime.GOOS {
	case "linux":
		return "On Linux, ensure ALSA is installed and configured. Try: 'sudo apt install libasound2-dev' (Ubuntu/Debian) or 'sudo pacman -S alsa-lib' (Arch). Check if your user is in the 'audio' group"
	case "darwin":
		return "On macOS, ensure Xcode command line tools are installed: 'xcode-select --install'. AudioToolbox.framework should be available"
	case "windows":
		return "On Windows, ensure audio drivers are properly installed and no other application is exclusively using the audio device"
	default:
		return "Check that audio drivers and development libraries are installed for your platform"
	}
}

// formatDuration converts time.Duration to MM:SS format
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
