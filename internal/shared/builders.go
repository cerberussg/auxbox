package shared

// Command builders - helper methods for creating common commands

func NewPlayCommand() Command {
	return Command{Type: CmdPlay}
}

func NewPlayTrackCommand(trackIndex int) Command {
	return Command{
		Type:       CmdPlay,
		TrackIndex: trackIndex,
	}
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

func NewVolumeCommand(volume int) Command {
	return Command{Type: CmdVolume, Volume: volume}
}

func NewExitCommand() Command {
	return Command{Type: CmdExit}
}

func NewRateCommand(rating int) Command {
	return Command{
		Type:   CmdRate,
		Rating: rating,
	}
}

func NewLabelCommand(label string) Command {
	return Command{
		Type:  CmdLabel,
		Label: label,
	}
}

func NewGenreCommand(genre string) Command {
	return Command{
		Type:  CmdGenre,
		Genre: genre,
	}
}

func NewTitleCommand(title string) Command {
	return Command{
		Type:  CmdTitle,
		Title: title,
	}
}

func NewArtistCommand(artist string) Command {
	return Command{
		Type:   CmdArtist,
		Artist: artist,
	}
}

func NewAlbumCommand(album string) Command {
	return Command{
		Type:  CmdAlbum,
		Album: album,
	}
}

func NewYearCommand(year string) Command {
	return Command{
		Type: CmdYear,
		Year: year,
	}
}

// Response builders - helper methods for creating common responses

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
