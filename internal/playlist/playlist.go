package playlist

import (
	"math/rand"
	"sync"
	"time"

	"github.com/cerberussg/auxbox/internal/shared"
)

type Playlist struct {
	tracks     []*shared.Track
	currentIdx int
	source     string
	sourceType shared.SourceType
	isShuffled bool
	mu         sync.RWMutex
}

func NewPlaylist() *Playlist {
	return &Playlist{
		tracks:     make([]*shared.Track, 0),
		currentIdx: 0,
	}
}

func (p *Playlist) LoadTracks(tracks []*shared.Track, source string, sourceType shared.SourceType) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tracks = tracks
	p.source = source
	p.sourceType = sourceType
	p.currentIdx = 0
	p.isShuffled = false

	return nil
}

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

func (p *Playlist) Next() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.tracks) == 0 {
		return false
	}

	if p.isShuffled {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		p.currentIdx = rng.Intn(len(p.tracks))
		return true
	}

	if p.currentIdx < len(p.tracks)-1 {
		p.currentIdx++
		return true
	}

	return false // Already at the end
}

func (p *Playlist) Previous() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.tracks) == 0 {
		return false
	}

	if p.isShuffled {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		p.currentIdx = rng.Intn(len(p.tracks))
		return true
	}

	if p.currentIdx > 0 {
		p.currentIdx--
		return true
	}

	return false // Already at the beginning
}

func (p *Playlist) TrackCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.tracks)
}

func (p *Playlist) GetTrackList() []*shared.Track {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a deep copy to prevent external modification
	tracks := make([]*shared.Track, len(p.tracks))
	for i, track := range p.tracks {
		trackCopy := &shared.Track{
			Filename: track.Filename,
			Path:     track.Path,
			Duration: track.Duration,
		}
		tracks[i] = trackCopy
	}
	return tracks
}

func (p *Playlist) GetSource() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.source
}

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

func (p *Playlist) Shuffle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isShuffled = true
}

func (p *Playlist) Unshuffle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isShuffled = false
}

func (p *Playlist) ToggleShuffle() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isShuffled = !p.isShuffled
	return p.isShuffled
}

func (p *Playlist) IsShuffled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isShuffled
}

func (p *Playlist) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tracks = make([]*shared.Track, 0)
	p.currentIdx = 0
	p.source = ""
	p.sourceType = ""
	p.isShuffled = false
}
