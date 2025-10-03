package commands

import (
	"fmt"
	"log"

	"github.com/cerberussg/auxbox/internal/audio"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/shared"
)

// NavigationHandler handles skip/back commands
type NavigationHandler struct {
	player   *audio.Player
	playlist *playlist.Playlist
}

// NewNavigationHandler creates a new navigation command handler
func NewNavigationHandler(player *audio.Player, playlist *playlist.Playlist) *NavigationHandler {
	return &NavigationHandler{
		player:   player,
		playlist: playlist,
	}
}

// HandleSkip handles the skip command
func (h *NavigationHandler) HandleSkip(cmd shared.Command) shared.Response {
	count := cmd.Count
	if count <= 0 {
		count = 1
	}

	skipped := 0
	for i := 0; i < count; i++ {
		if h.playlist.Next() {
			skipped++
		} else {
			break // Reached end of playlist
		}
	}

	if skipped == 0 {
		return shared.NewErrorResponse("Already at the end of playlist")
	}

	// Update player with new current track
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack != nil {
		h.player.SetCurrentTrack(currentTrack)
		log.Printf("Skipped %d track(s), now at: %s", skipped, currentTrack.Filename)
		return shared.NewSuccessResponse(
			fmt.Sprintf("Skipped %d track(s), now playing: %s", skipped, currentTrack.Filename),
			nil,
		)
	}

	return shared.NewSuccessResponse(fmt.Sprintf("Skipped %d track(s)", skipped), nil)
}

// HandleBack handles the back command
func (h *NavigationHandler) HandleBack(cmd shared.Command) shared.Response {
	count := cmd.Count
	if count <= 0 {
		count = 1
	}

	moved := 0
	for i := 0; i < count; i++ {
		if h.playlist.Previous() {
			moved++
		} else {
			break // Reached beginning of playlist
		}
	}

	if moved == 0 {
		return shared.NewErrorResponse("Already at the beginning of playlist")
	}

	// Update player with new current track
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack != nil {
		h.player.SetCurrentTrack(currentTrack)
		log.Printf("Moved back %d track(s), now at: %s", moved, currentTrack.Filename)
		return shared.NewSuccessResponse(
			fmt.Sprintf("Moved back %d track(s), now playing: %s", moved, currentTrack.Filename),
			nil,
		)
	}

	return shared.NewSuccessResponse(fmt.Sprintf("Moved back %d track(s)", moved), nil)
}