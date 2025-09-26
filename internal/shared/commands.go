package shared

type CommandType string

const (
	CmdStart CommandType = "start"
	CmdStop  CommandType = "stop"

	CmdPlay  CommandType = "play"
	CmdPause CommandType = "pause"
	CmdSkip  CommandType = "skip"
	CmdBack  CommandType = "back"

	CmdStatus CommandType = "status"
	CmdList   CommandType = "list"
)

type Command struct {
	Type   CommandType `json:"type"`
	Args   []string    `json:"args,omitempty"`
	Count  int         `json:"count,omitempty"`  // For skip/back amounts (default 1)
	Source SourceType  `json:"source,omitempty"` // For start command
	Path   string      `json:"path,omitempty"`   // Path to folder/playlist/etc
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
