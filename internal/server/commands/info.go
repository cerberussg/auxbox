package commands

import (
	"fmt"
	"log"

	"github.com/cerberussg/auxbox/internal/audio"
	"github.com/cerberussg/auxbox/internal/metadata"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/shared"
)

type InfoHandler struct {
	player         *audio.Player
	playlist       *playlist.Playlist
	metadataReader metadata.Reader
}

func NewInfoHandler(player *audio.Player, playlist *playlist.Playlist) *InfoHandler {
	return &InfoHandler{
		player:         player,
		playlist:       playlist,
		metadataReader: metadata.NewReader(),
	}
}

func (h *InfoHandler) HandleStatus() shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewSuccessResponse("No track loaded", nil)
	}

	status := h.player.GetStatus()
	trackInfo := shared.TrackInfo{
		Filename:    currentTrack.Filename,
		Path:        currentTrack.Path,
		Duration:    status.Duration,
		Position:    status.Position,
		TrackNumber: h.playlist.GetCurrentIndex() + 1,
		TotalTracks: h.playlist.TrackCount(),
		Source:      h.playlist.GetSource(),
	}

	// Read metadata if available (MP3 files)
	if metadata.IsMetadataSupported(currentTrack.Path) {
		meta, err := h.metadataReader.Read(currentTrack.Path)
		if err == nil && meta != nil {
			trackInfo.Title = meta.Title
			trackInfo.Artist = meta.Artist
			trackInfo.Album = meta.Album
			trackInfo.Year = meta.Year
			trackInfo.Label = meta.Label
			trackInfo.Rating = meta.Rating
			trackInfo.Genre = meta.Genre
		}
	}

	// Format status message with key | value structure
	message := h.formatStatusMessage(trackInfo)

	return shared.NewSuccessResponse(message, trackInfo)
}

// formatStatusMessage creates a structured key | value display
func (h *InfoHandler) formatStatusMessage(info shared.TrackInfo) string {
	msg := "\n"
	msg += fmt.Sprintf("Filename      | %s\n", info.Filename)
	msg += fmt.Sprintf("Full Path     | %s\n", info.Path)
	msg += fmt.Sprintf("Track         | %d/%d\n", info.TrackNumber, info.TotalTracks)
	msg += fmt.Sprintf("Position      | %s / %s\n", info.Position, info.Duration)
	msg += fmt.Sprintf("Source        | %s\n", info.Source)

	// Add metadata if available
	if info.Title != "" {
		msg += "\n--- ID3 Metadata ---\n"
		msg += fmt.Sprintf("Title         | %s\n", info.Title)
	}
	if info.Artist != "" {
		msg += fmt.Sprintf("Artist        | %s\n", info.Artist)
	}
	if info.Album != "" {
		msg += fmt.Sprintf("Album         | %s\n", info.Album)
	}
	if info.Year != "" {
		msg += fmt.Sprintf("Year          | %s\n", info.Year)
	}
	if info.Label != "" {
		msg += fmt.Sprintf("Label         | %s\n", info.Label)
	}
	if info.Rating > 0 {
		stars := ""
		for i := 0; i < info.Rating; i++ {
			stars += "⭐"
		}
		msg += fmt.Sprintf("Rating        | %d/5 %s\n", info.Rating, stars)
	}
	if info.Genre != "" {
		msg += fmt.Sprintf("Genre         | %s\n", info.Genre)
	}

	return msg
}

func (h *InfoHandler) HandleList() shared.Response {
	// Use windowed approach to avoid loading all tracks
	const windowSize = 15
	tracks, startIdx, totalCount := h.playlist.GetTrackWindow(windowSize)

	if totalCount == 0 {
		return shared.NewSuccessResponse("No tracks loaded", nil)
	}

	trackNames := make([]string, len(tracks))
	for i, track := range tracks {
		trackNames[i] = track.Filename
	}

	playlistInfo := shared.PlaylistInfo{
		Source:     h.playlist.GetSource(),
		SourceType: string(h.playlist.GetSourceType()),
		Tracks:     trackNames,
		CurrentIdx: h.playlist.GetCurrentIndex(),
		StartIdx:   startIdx,
		TotalCount: totalCount,
	}

	return shared.NewSuccessResponse(fmt.Sprintf("%d tracks loaded", totalCount), playlistInfo)
}

func (h *InfoHandler) HandleVolume(cmd shared.Command) shared.Response {
	// If volume is -1, return current volume
	if cmd.Volume == -1 {
		status := h.player.GetStatus()
		volumePercent := int(status.Volume * 100)

		volumeData := map[string]interface{}{
			"volume": status.Volume,
		}

		return shared.NewSuccessResponse(
			fmt.Sprintf("Volume: %d%%", volumePercent),
			volumeData,
		)
	}

	volumeFloat := float64(cmd.Volume) / 100.0

	if err := h.player.SetVolume(volumeFloat); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set volume: %v", err))
	}

	log.Printf("Volume set to %d%%", cmd.Volume)

	if cmd.Volume == 0 {
		return shared.NewSuccessResponse("Volume set to 0% (muted)", nil)
	}

	return shared.NewSuccessResponse(fmt.Sprintf("Volume set to %d%%", cmd.Volume), nil)
}
