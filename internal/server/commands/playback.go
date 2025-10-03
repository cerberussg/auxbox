package commands

import (
	"fmt"
	"log"

	"github.com/cerberussg/auxbox/internal/audio"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/shared"
)

// PlaybackHandler handles play/pause/stop commands
type PlaybackHandler struct {
	player   *audio.Player
	playlist *playlist.Playlist
}

// NewPlaybackHandler creates a new playback command handler
func NewPlaybackHandler(player *audio.Player, playlist *playlist.Playlist) *PlaybackHandler {
	return &PlaybackHandler{
		player:   player,
		playlist: playlist,
	}
}

// HandlePlay handles the play command
func (h *PlaybackHandler) HandlePlay() shared.Response {
	trackCount := h.playlist.TrackCount()
	log.Printf("Play request - Track count: %d", trackCount)

	if trackCount == 0 {
		return shared.NewErrorResponse("No tracks loaded. Use 'auxbox start --folder <path>' first.")
	}

	// Ensure player has current track (in case it got out of sync)
	playerTrack := h.player.GetCurrentTrack()
	log.Printf("Player current track: %v", playerTrack)

	if playerTrack == nil {
		// Try to find a valid track, skipping any that fail to load
		maxAttempts := 10 // Don't try forever
		for attempts := 0; attempts < maxAttempts; attempts++ {
			currentTrack := h.playlist.GetCurrentTrack()
			log.Printf("Playlist current track: %v", currentTrack)

			if currentTrack == nil {
				return shared.NewErrorResponse("No track available from playlist")
			}

			log.Printf("Setting track: %s", currentTrack.Path)
			h.player.SetCurrentTrack(currentTrack)

			// Check if the track loaded successfully
			if h.player.GetCurrentTrack() != nil {
				log.Printf("Successfully loaded track: %s", currentTrack.Filename)
				break
			}

			// Track failed to load, try next one
			log.Printf("Track failed to load, trying next track...")
			if !h.playlist.Next() {
				return shared.NewErrorResponse("No valid tracks found in playlist")
			}
		}

		// If we still don't have a track after trying, give up
		if h.player.GetCurrentTrack() == nil {
			return shared.NewErrorResponse("Failed to load any valid tracks after multiple attempts")
		}
	}

	if err := h.player.Play(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to play: %v", err))
	}

	currentTrack := h.player.GetCurrentTrack()
	if currentTrack != nil {
		log.Printf("Playing: %s", currentTrack.Filename)
		return shared.NewSuccessResponse(fmt.Sprintf("Playing: %s", currentTrack.Filename), nil)
	}

	return shared.NewSuccessResponse("Playback started", nil)
}

// HandlePause handles the pause command
func (h *PlaybackHandler) HandlePause() shared.Response {
	if err := h.player.Pause(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to pause: %v", err))
	}

	log.Println("Playback paused")
	return shared.NewSuccessResponse("Playback paused", nil)
}

// HandleStop handles the stop command
func (h *PlaybackHandler) HandleStop() shared.Response {
	if err := h.player.Stop(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to stop: %v", err))
	}

	log.Println("Playback stopped")
	return shared.NewSuccessResponse("Playback stopped", nil)
}