package commands

import (
	"fmt"
	"log"

	"github.com/cerberussg/auxbox/internal/audio"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/shared"
)

// InfoHandler handles status/list/volume commands
type InfoHandler struct {
	player   *audio.Player
	playlist *playlist.Playlist
}

// NewInfoHandler creates a new info command handler
func NewInfoHandler(player *audio.Player, playlist *playlist.Playlist) *InfoHandler {
	return &InfoHandler{
		player:   player,
		playlist: playlist,
	}
}

// HandleStatus handles the status command
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
		TrackNumber: h.playlist.GetCurrentIndex() + 1, // 1-based indexing for display
		TotalTracks: h.playlist.TrackCount(),
		Source:      h.playlist.GetSource(),
	}

	return shared.NewSuccessResponse("Current status", trackInfo)
}

// HandleList handles the list command
func (h *InfoHandler) HandleList() shared.Response {
	tracks := h.playlist.GetTrackList()
	if len(tracks) == 0 {
		return shared.NewSuccessResponse("No tracks loaded", nil)
	}

	// Convert tracks to string slice for JSON serialization
	trackNames := make([]string, len(tracks))
	for i, track := range tracks {
		trackNames[i] = track.Filename
	}

	playlistInfo := shared.PlaylistInfo{
		Source:     h.playlist.GetSource(),
		SourceType: string(h.playlist.GetSourceType()),
		Tracks:     trackNames,
		CurrentIdx: h.playlist.GetCurrentIndex(),
	}

	return shared.NewSuccessResponse(fmt.Sprintf("%d tracks loaded", len(tracks)), playlistInfo)
}

// HandleVolume handles the volume command
func (h *InfoHandler) HandleVolume(cmd shared.Command) shared.Response {
	// If volume is -1, return current volume
	if cmd.Volume == -1 {
		status := h.player.GetStatus()
		volumePercent := int(status.Volume * 100)

		volumeData := map[string]interface{}{
			"volume": status.Volume, // 0.0-1.0 for API compatibility
		}

		return shared.NewSuccessResponse(
			fmt.Sprintf("Volume: %d%%", volumePercent),
			volumeData,
		)
	}

	// Set volume (cmd.Volume is 0-100 percentage)
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