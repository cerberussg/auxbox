package shared

type SourceType string

const (
	SourceFolder   SourceType = "folder"
	SourcePlaylist SourceType = "playlist"
	SourceTwitch   SourceType = "twitch"  // Future
	SourceDiscord  SourceType = "discord" // Future
)

type Track struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Duration string `json:"duration,omitempty"`
}

type TrackInfo struct {
	Filename    string `json:"filename"`
	Path        string `json:"path"`
	Duration    string `json:"duration,omitempty"`     // e.g. "4:12"
	Position    string `json:"position,omitempty"`     // e.g. "2:34"
	TrackNumber int    `json:"track_number,omitempty"` // Current track in queue
	TotalTracks int    `json:"total_tracks,omitempty"` // Total tracks in queue
	Source      string `json:"source,omitempty"`       // Source folder/playlist name
}

type PlaylistInfo struct {
	Source     string   `json:"source"`      // Folder path or playlist name
	SourceType string   `json:"source_type"` // "folder", "playlist", etc
	Tracks     []string `json:"tracks"`      // List of track filenames (windowed for large playlists)
	CurrentIdx int      `json:"current_idx"` // Index of current track (0-based)
	StartIdx   int      `json:"start_idx"`   // Index of first track in window (0-based)
	TotalCount int      `json:"total_count"` // Total number of tracks in playlist
}
