package shared

func NewPlayCommand() Command {
	return Command{Type: CmdPlay}
}

func NewPauseCommand() Command {
	return Command{Type: CmdPause}
}

func NewSkipCommand(count int) Command {
	if count <= 0 {
		count = 1
	}
	return Command{Type: CmdSkip, Count: count}
}

func NewBackCommand(count int) Command {
	if count <= 0 {
		count = 1
	}
	return Command{Type: CmdBack, Count: count}
}

func NewStartCommand(sourceType SourceType, path string) Command {
	return Command{
		Type:   CmdStart,
		Source: sourceType,
		Path:   path,
	}
}

func NewStatusCommand() Command {
	return Command{Type: CmdStatus}
}

func NewListCommand() Command {
	return Command{Type: CmdList}
}

func NewStopCommand() Command {
	return Command{Type: CmdStop}
}

func NewSuccessResponse(message string, data interface{}) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Message: message,
	}
}
