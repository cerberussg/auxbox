package daemon

import (
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
)

// PositionTracker manages audio playback position tracking and duration calculation
type PositionTracker struct {
	streamer   beep.StreamSeekCloser
	format     beep.Format
	position   string
	duration   string
	isPlaying  bool

	// Ticker management
	ticker     *time.Ticker
	stopChan   chan bool
	mu         sync.RWMutex
}

// NewPositionTracker creates a new position tracker
func NewPositionTracker() *PositionTracker {
	return &PositionTracker{
		position:  "0:00",
		duration:  "0:00",
		stopChan:  make(chan bool),
	}
}

// SetStreamer sets the audio streamer and calculates duration
func (pt *PositionTracker) SetStreamer(streamer beep.StreamSeekCloser, format beep.Format) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.streamer = streamer
	pt.format = format
	pt.position = "0:00"

	pt.calculateDuration()
}

// StartTracking starts periodic position updates
func (pt *PositionTracker) StartTracking() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.isPlaying = true
	pt.stopTracking() // Stop any existing ticker

	pt.ticker = time.NewTicker(500 * time.Millisecond) // Update twice per second
	go func() {
		for {
			select {
			case <-pt.ticker.C:
				pt.updatePosition()
			case <-pt.stopChan:
				return
			}
		}
	}()
}

// StopTracking stops periodic position updates
func (pt *PositionTracker) StopTracking() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.isPlaying = false
	pt.stopTracking()
}

// stopTracking internal method to stop ticker (assumes lock is held)
func (pt *PositionTracker) stopTracking() {
	if pt.ticker != nil {
		pt.ticker.Stop()
		pt.ticker = nil

		// Signal the ticker goroutine to stop
		select {
		case pt.stopChan <- true:
		default:
		}
	}
}

// ResetPosition resets position to start
func (pt *PositionTracker) ResetPosition() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.position = "0:00"
}

// GetPosition returns the current playback position
func (pt *PositionTracker) GetPosition() time.Duration {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if pt.streamer == nil {
		return time.Duration(0)
	}

	if seeker, ok := pt.streamer.(beep.StreamSeeker); ok {
		position := seeker.Position()
		return pt.format.SampleRate.D(position)
	}

	return time.Duration(0)
}

// GetDuration returns the total duration of the current track
func (pt *PositionTracker) GetDuration() time.Duration {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if pt.streamer == nil {
		return time.Duration(0)
	}

	if seeker, ok := pt.streamer.(beep.StreamSeeker); ok {
		length := seeker.Len()
		return pt.format.SampleRate.D(length)
	}

	return time.Duration(0)
}

// GetPositionString returns the current position as formatted string
func (pt *PositionTracker) GetPositionString() string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.position
}

// GetDurationString returns the duration as formatted string
func (pt *PositionTracker) GetDurationString() string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.duration
}

// updatePosition updates the current position (internal method)
func (pt *PositionTracker) updatePosition() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if pt.isPlaying && pt.streamer != nil {
		if seeker, ok := pt.streamer.(beep.StreamSeeker); ok {
			position := seeker.Position()
			duration := pt.format.SampleRate.D(position)
			pt.position = formatDuration(duration)
		}
	}
}

// calculateDuration calculates and sets the track duration (assumes lock is held)
func (pt *PositionTracker) calculateDuration() {
	if pt.streamer == nil {
		pt.duration = "0:00"
		return
	}

	// Get the length of the streamer if possible
	if seeker, ok := pt.streamer.(beep.StreamSeeker); ok {
		currentPos := seeker.Position()
		length := seeker.Len()

		// Seek back to the beginning
		seeker.Seek(currentPos)

		// Calculate duration
		duration := pt.format.SampleRate.D(length)
		pt.duration = formatDuration(duration)
	} else {
		pt.duration = "Unknown"
	}
}

// Cleanup cleans up position tracker resources
func (pt *PositionTracker) Cleanup() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.stopTracking()
	pt.streamer = nil
	pt.position = "0:00"
	pt.duration = "0:00"
	pt.isPlaying = false
}