package shared

// CommandType represents the type of command being sent
type CommandType string

const (
	// Daemon management
	CmdStart CommandType = "start"
	CmdExit  CommandType = "exit"

	// Playback control
	CmdPlay  CommandType = "play"
	CmdPause CommandType = "pause"
	CmdStop  CommandType = "stop"
	CmdSkip  CommandType = "skip"
	CmdBack  CommandType = "back"

	// Info commands
	CmdStatus CommandType = "status"
	CmdList   CommandType = "list"
)

// Command represents a command sent to the daemon
type Command struct {
	Type   CommandType `json:"type"`
	Args   []string    `json:"args,omitempty"`
	Count  int         `json:"count,omitempty"`  // For skip/back amounts (default 1)
	Source SourceType  `json:"source,omitempty"` // For start command
	Path   string      `json:"path,omitempty"`   // Path to folder/playlist/etc
}

// Response represents a response from the daemon
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
