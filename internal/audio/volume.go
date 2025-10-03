package audio

import (
	"fmt"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
)

// VolumeControl manages volume control and audio stream processing
type VolumeControl struct {
	volumeStreamer *effects.Volume
	ctrl           *beep.Ctrl
	volume         float64
	mu             sync.RWMutex
}

// NewVolumeControl creates a new volume control manager
func NewVolumeControl() *VolumeControl {
	return &VolumeControl{
		volume: 1.0, // Default full volume
	}
}

// SetupWithStreamer creates the volume control wrapper around a streamer
func (v *VolumeControl) SetupWithStreamer(streamer beep.StreamSeekCloser, onTrackComplete func()) *beep.Ctrl {
	v.mu.Lock()
	defer v.mu.Unlock()

	if streamer == nil {
		return nil
	}

	// Create volume control wrapper
	v.volumeStreamer = &effects.Volume{
		Streamer: streamer,
		Base:     2.0,
		Volume:   (v.volume - 1.0) * 2.0, // Apply current volume setting
		Silent:   v.volume == 0.0,
	}

	// Add a callback to detect when track ends
	trackEndCallback := beep.Callback(func() {
		// Call the completion callback if set (outside of lock to avoid deadlock)
		if onTrackComplete != nil {
			onTrackComplete()
		}
	})

	// Create a sequence that plays the track then calls the callback
	sequence := beep.Seq(v.volumeStreamer, trackEndCallback)

	// Create control wrapper for pause/resume
	v.ctrl = &beep.Ctrl{
		Streamer: sequence,
		Paused:   true, // Start paused
	}

	return v.ctrl
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (v *VolumeControl) SetVolume(volume float64) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if volume < 0.0 || volume > 1.0 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}

	v.volume = volume

	// Apply volume to beep if available
	if v.volumeStreamer != nil {
		if volume == 0.0 {
			v.volumeStreamer.Silent = true
		} else {
			v.volumeStreamer.Silent = false
			// Convert 0.0-1.0 to beep's volume scale
			// beep uses logarithmic scale where 0 = no change
			// We'll use a simple linear conversion for now
			beepVolume := (volume - 1.0) * 2.0 // -2.0 to 0.0 range
			v.volumeStreamer.Volume = beepVolume
		}
	}

	return nil
}

// GetVolume returns the current volume
func (v *VolumeControl) GetVolume() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.volume
}

// GetControl returns the beep control wrapper
func (v *VolumeControl) GetControl() *beep.Ctrl {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.ctrl
}

// Cleanup cleans up volume control resources
func (v *VolumeControl) Cleanup() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.volumeStreamer = nil
	v.ctrl = nil
}