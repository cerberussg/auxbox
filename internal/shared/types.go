package shared

// SourceType represents different audio sources
type SourceType string

const (
	SourceFolder   SourceType = "folder"
	SourcePlaylist SourceType = "playlist"
	SourceTwitch   SourceType = "twitch"  // Future
	SourceDiscord  SourceType = "discord" // Future
)

// Track represents a single audio track
type Track struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Duration string `json:"duration,omitempty"` // Will be populated when track is loaded
}

// TrackInfo represents information about the currently playing track
type TrackInfo struct {
	Filename    string `json:"filename"`
	Path        string `json:"path"`
	Duration    string `json:"duration,omitempty"`     // e.g. "4:12"
	Position    string `json:"position,omitempty"`     // e.g. "2:34"
	TrackNumber int    `json:"track_number,omitempty"` // Current track in queue
	TotalTracks int    `json:"total_tracks,omitempty"` // Total tracks in queue
	Source      string `json:"source,omitempty"`       // Source folder/playlist name
}

// PlaylistInfo represents the current playlist/queue
type PlaylistInfo struct {
	Source     string   `json:"source"`      // Folder path or playlist name
	SourceType string   `json:"source_type"` // "folder", "playlist", etc
	Tracks     []string `json:"tracks"`      // List of track filenames
	CurrentIdx int      `json:"current_idx"` // Index of current track
}
