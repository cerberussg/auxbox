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

	CmdStatus CommandType = "status"
	CmdList   CommandType = "list"
)

type Command struct {
	Type    CommandType `json:"type"`
	Args    []string    `json:"args,omitempty"`
	Count   int         `json:"count,omitempty"`
	Volume  int         `json:"volume,omitempty"`
	Source  SourceType  `json:"source,omitempty"`
	Path    string      `json:"path,omitempty"`
	Shuffle bool        `json:"shuffle,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
