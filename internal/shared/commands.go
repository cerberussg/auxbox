package shared

type CommandType string

const (
	CmdStart CommandType = "start"
	CmdExit  CommandType = "exit"

	CmdPlay    CommandType = "play"
	CmdPause   CommandType = "pause"
	CmdStop    CommandType = "stop"
	CmdSkip    CommandType = "skip"
	CmdBack    CommandType = "back"
	CmdVolume  CommandType = "volume"
	CmdShuffle CommandType = "shuffle"
	CmdRepeat  CommandType = "repeat"

	CmdStatus CommandType = "status"
	CmdList   CommandType = "list"

	CmdRate   CommandType = "rate"
	CmdLabel  CommandType = "label"
	CmdGenre  CommandType = "genre"
	CmdTitle  CommandType = "title"
	CmdArtist CommandType = "artist"
	CmdAlbum  CommandType = "album"
	CmdYear   CommandType = "year"
)

type Command struct {
	Type       CommandType `json:"type"`
	Args       []string    `json:"args,omitempty"`
	Count      int         `json:"count,omitempty"`
	Volume     int         `json:"volume,omitempty"`
	Source     SourceType  `json:"source,omitempty"`
	Path       string      `json:"path,omitempty"`
	Shuffle    bool        `json:"shuffle,omitempty"`
	Repeat     bool        `json:"repeat,omitempty"`
	TrackIndex int         `json:"track_index,omitempty"` // Jump to specific track (1-based)
	Rating     int         `json:"rating,omitempty"`      // 1-5 rating
	Label      string      `json:"label,omitempty"`       // Record label name
	Genre      string      `json:"genre,omitempty"`       // Music genre
	Title      string      `json:"title,omitempty"`       // Track title
	Artist     string      `json:"artist,omitempty"`      // Track artist
	Album      string      `json:"album,omitempty"`       // Album name
	Year       string      `json:"year,omitempty"`        // Release year
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
