package daemon

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// AudioSystem manages speaker initialization and audio hardware setup
type AudioSystem struct {
	initialized bool
}

// NewAudioSystem creates a new audio system manager
func NewAudioSystem() *AudioSystem {
	return &AudioSystem{
		initialized: false,
	}
}

// Initialize initializes the speaker with the given format, using fallback strategies if needed
func (a *AudioSystem) Initialize(format beep.Format) error {
	if a.initialized {
		return nil // Already initialized
	}

	if format.SampleRate == 0 {
		return fmt.Errorf("invalid sample rate: %d", format.SampleRate)
	}

	// Strategy 1: Standard initialization with current format
	bufferSize := format.SampleRate.N(time.Second / 10)
	err := speaker.Init(format.SampleRate, bufferSize)
	if err == nil {
		fmt.Printf("Audio initialized successfully with sample rate: %d Hz, buffer size: %d\n",
			format.SampleRate, bufferSize)
		a.initialized = true
		return nil
	}

	fmt.Printf("Primary audio initialization failed: %v\n", err)

	// Strategy 2: Try with larger buffer for stability
	bufferSize = format.SampleRate.N(time.Second / 5)
	err = speaker.Init(format.SampleRate, bufferSize)
	if err == nil {
		fmt.Printf("Audio initialized with larger buffer: %d Hz, buffer size: %d\n",
			format.SampleRate, bufferSize)
		a.initialized = true
		return nil
	}

	fmt.Printf("Large buffer initialization failed: %v\n", err)

	// Strategy 3: Try with common sample rates as fallbacks
	fallbackRates := []beep.SampleRate{44100, 48000, 22050, 16000}
	for _, rate := range fallbackRates {
		if rate == format.SampleRate {
			continue // Already tried this rate
		}

		bufferSize = rate.N(time.Second / 10)
		err = speaker.Init(rate, bufferSize)
		if err == nil {
			fmt.Printf("Audio initialized with fallback sample rate: %d Hz (original: %d Hz)\n",
				rate, format.SampleRate)
			a.initialized = true
			return nil
		}
		fmt.Printf("Fallback rate %d Hz failed: %v\n", rate, err)
	}

	// Strategy 4: Platform-specific troubleshooting hints
	platformHint := a.getPlatformAudioHint()

	return fmt.Errorf("all audio initialization strategies failed. %s. Last error: %w", platformHint, err)
}

// IsInitialized returns whether the audio system has been initialized
func (a *AudioSystem) IsInitialized() bool {
	return a.initialized
}

// getPlatformAudioHint provides platform-specific troubleshooting information
func (a *AudioSystem) getPlatformAudioHint() string {
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