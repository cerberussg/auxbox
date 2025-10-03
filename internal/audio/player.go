package audio

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/cerberussg/auxbox/internal/audio/decoders"
	"github.com/cerberussg/auxbox/internal/shared"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// Player handles audio playback using the beep library
type Player struct {
	currentTrack *shared.Track
	status       PlayerStatus
	mu           sync.RWMutex

	// Core audio components
	streamer beep.StreamSeekCloser
	format   beep.Format
	file     io.ReadCloser

	// Callback for track completion
	onTrackComplete func()

	// Extracted components
	audioSystem     *AudioSystem
	volumeControl   *VolumeControl
	positionTracker *PositionTracker
	registry        *decoders.FormatRegistry
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
		audioSystem:     NewAudioSystem(),
		volumeControl:   NewVolumeControl(),
		positionTracker: NewPositionTracker(),
		registry:        decoders.NewFormatRegistry(),
	}
}

// stopAndCleanup stops current playback and cleans up resources
func (p *Player) stopAndCleanup() {
	// Stop position updates
	p.positionTracker.StopTracking()

	// Stop the speaker from playing this stream
	if ctrl := p.volumeControl.GetControl(); ctrl != nil {
		ctrl.Paused = true
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

	ctrl := p.volumeControl.GetControl()
	if ctrl == nil {
		return fmt.Errorf("audio not initialized")
	}

	// Start fresh playback
	ctrl.Paused = false
	speaker.Play(ctrl)
	p.status.IsPlaying = true
	p.status.IsPaused = false

	// Start position updates
	p.positionTracker.StartTracking()

	return nil
}

// Play starts or resumes audio playback
func (p *Player) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentTrack == nil {
		return fmt.Errorf("no track loaded")
	}

	if p.volumeControl.GetControl() == nil {
		return fmt.Errorf("audio not initialized")
	}

	// If already playing, do nothing
	if p.status.IsPlaying {
		return nil
	}

	// If paused, just resume
	if p.status.IsPaused {
		if ctrl := p.volumeControl.GetControl(); ctrl != nil {
			ctrl.Paused = false
		}
		p.status.IsPlaying = true
		p.status.IsPaused = false
		// Restart position updates
		p.positionTracker.StartTracking()
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

	ctrl := p.volumeControl.GetControl()
	if ctrl == nil {
		return fmt.Errorf("audio not initialized")
	}

	// Pause the audio
	ctrl.Paused = true
	p.status.IsPlaying = false
	p.status.IsPaused = true

	// Stop position updates
	p.positionTracker.StopTracking()

	return nil
}

// Stop stops audio playback completely
func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if ctrl := p.volumeControl.GetControl(); ctrl != nil {
		ctrl.Paused = true
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

	// Stop position updates
	p.positionTracker.StopTracking()

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

	// Update position and duration from tracker
	status := p.status
	status.Position = p.positionTracker.GetPositionString()
	status.Duration = p.positionTracker.GetDurationString()

	return status
}

// SetOnTrackComplete sets the callback function to be called when a track finishes
func (p *Player) SetOnTrackComplete(callback func()) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onTrackComplete = callback
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (p *Player) SetVolume(volume float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.status.Volume = volume
	return p.volumeControl.SetVolume(volume)
}

// GetPosition returns the current playback position
func (p *Player) GetPosition() time.Duration {
	return p.positionTracker.GetPosition()
}

// GetDuration returns the total duration of the current track
func (p *Player) GetDuration() time.Duration {
	return p.positionTracker.GetDuration()
}

// UpdatePosition updates the current position in the status
func (p *Player) UpdatePosition() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status.Position = p.positionTracker.GetPositionString()
}

// Close cleans up all audio resources
func (p *Player) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop position updates
	p.positionTracker.StopTracking()

	// Clean up all components
	p.positionTracker.Cleanup()
	p.volumeControl.Cleanup()
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

	// Use the format registry to decode the file
	streamer, format, err := p.registry.Decode(filePath, file)
	if err != nil {
		file.Close()
		return err
	}

	p.streamer = streamer
	p.format = format
	p.file = file

	// Initialize speaker if not already done
	if !p.audioSystem.IsInitialized() {
		err := p.audioSystem.Initialize(format)
		if err != nil {
			p.cleanup()
			return fmt.Errorf("failed to initialize speaker: %w", err)
		}
	}

	// Set up position tracking
	p.positionTracker.SetStreamer(streamer, format)

	// Set up volume control
	fmt.Printf("Setting up volume control...\n")
	ctrl := p.volumeControl.SetupWithStreamer(streamer, p.onTrackComplete)

	// Verify control was set up properly
	if ctrl == nil {
		return fmt.Errorf("volume control setup failed - ctrl is nil")
	}
	fmt.Printf("Volume control setup complete\n")

	return nil
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
}
