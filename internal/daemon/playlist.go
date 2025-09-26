package daemon

import (
	"sync"

	"github.com/cerberussg/auxbox/internal/shared"
)

// Playlist manages a collection of tracks and current playback position
type Playlist struct {
	tracks     []*shared.Track
	currentIdx int
	source     string // Path to source folder/playlist
	sourceType shared.SourceType
	mu         sync.RWMutex
}

// NewPlaylist creates a new empty playlist
func NewPlaylist() *Playlist {
	return &Playlist{
		tracks:     make([]*shared.Track, 0),
		currentIdx: 0,
	}
}

// LoadTracks loads a collection of tracks into the playlist
func (p *Playlist) LoadTracks(tracks []*shared.Track, source string, sourceType shared.SourceType) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tracks = tracks
	p.source = source
	p.sourceType = sourceType
	p.currentIdx = 0

	return nil
}

// GetCurrentTrack returns the currently selected track
func (p *Playlist) GetCurrentTrack() *shared.Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.tracks) == 0 || p.currentIdx < 0 || p.currentIdx >= len(p.tracks) {
		return nil
	}

	return p.tracks[p.currentIdx]
}

// GetCurrentIndex returns the current track index (0-based)
func (p *Playlist) GetCurrentIndex() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentIdx
}

// Next moves to the next track in the playlist
// Returns true if successful, false if already at the end
func (p *Playlist) Next() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.tracks) == 0 {
		return false
	}

	if p.currentIdx < len(p.tracks)-1 {
		p.currentIdx++
		return true
	}

	return false // Already at the end
}

// Previous moves to the previous track in the playlist
// Returns true if successful, false if already at the beginning
func (p *Playlist) Previous() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.tracks) == 0 {
		return false
	}

	if p.currentIdx > 0 {
		p.currentIdx--
		return true
	}

	return false // Already at the beginning
}

// TrackCount returns the total number of tracks in the playlist
func (p *Playlist) TrackCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.tracks)
}

// GetTrackList returns a copy of all tracks in the playlist
func (p *Playlist) GetTrackList() []*shared.Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a deep copy to prevent external modification
	tracks := make([]*shared.Track, len(p.tracks))
	for i, track := range p.tracks {
		// Create a copy of each track
		trackCopy := &shared.Track{
			Filename: track.Filename,
			Path:     track.Path,
			Duration: track.Duration,
		}
		tracks[i] = trackCopy
	}
	return tracks
}

// GetSource returns the source path (folder or playlist file)
func (p *Playlist) GetSource() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.source
}

// GetSourceType returns the source type (folder, playlist, etc.)
func (p *Playlist) GetSourceType() shared.SourceType {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.sourceType
}

// SetCurrentIndex sets the current track index
// Returns true if successful, false if index is out of range
func (p *Playlist) SetCurrentIndex(idx int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if idx < 0 || idx >= len(p.tracks) {
		return false
	}

	p.currentIdx = idx
	return true
}

// Shuffle randomizes the order of tracks in the playlist
// The current track remains the current track, but its index may change
func (p *Playlist) Shuffle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.tracks) <= 1 {
		return
	}

	// TODO: Implement shuffle algorithm
	// For now, just leave as-is
}

// Clear removes all tracks from the playlist
func (p *Playlist) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tracks = make([]*shared.Track, 0)
	p.currentIdx = 0
	p.source = ""
	p.sourceType = ""
}
