package server

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cerberussg/auxbox/internal/audio"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/server/commands"
	"github.com/cerberussg/auxbox/internal/shared"
)

// Server represents the auxbox daemon server
type Server struct {
	transport shared.Transport
	player    *audio.Player
	playlist  *playlist.Playlist
	isRunning bool
	mu        sync.RWMutex

	// Command handlers
	playbackHandler   *commands.PlaybackHandler
	navigationHandler *commands.NavigationHandler
	infoHandler       *commands.InfoHandler
	loader           *Loader
}

// NewServer creates a new daemon server instance
func NewServer() *Server {
	player := audio.NewPlayer()
	playlistObj := playlist.NewPlaylist()

	server := &Server{
		transport: shared.NewUnixSocketTransport(),
		player:    player,
		playlist:  playlistObj,
		isRunning: false,

		// Initialize command handlers
		playbackHandler:   commands.NewPlaybackHandler(player, playlistObj),
		navigationHandler: commands.NewNavigationHandler(player, playlistObj),
		infoHandler:       commands.NewInfoHandler(player, playlistObj),
		loader:           NewLoader(),
	}

	// Set up the track completion callback
	player.SetOnTrackComplete(server.onTrackComplete)

	return server
}

// LoadTracks loads tracks into the server's playlist
func (s *Server) LoadTracks(tracks []*shared.Track, source string, sourceType shared.SourceType) error {
	return s.playlist.LoadTracks(tracks, source, sourceType)
}

// Start starts the daemon server
func (s *Server) Start() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("daemon is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	log.Println("Starting auxbox daemon...")

	// Start listening for commands
	return s.transport.Listen(s.handleCommand)
}

// Stop stops the daemon server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("daemon is not running")
	}

	log.Println("Stopping auxbox daemon...")

	// Stop the player
	if err := s.player.Stop(); err != nil {
		log.Printf("Error stopping player: %v", err)
	}

	// Close transport
	if err := s.transport.Close(); err != nil {
		log.Printf("Error closing transport: %v", err)
	}

	s.isRunning = false
	return nil
}

// HandleCommand processes a command directly (used for initialization)
func (s *Server) HandleCommand(cmd shared.Command) shared.Response {
	return s.handleCommand(cmd)
}

// handleCommand processes incoming commands from clients
func (s *Server) handleCommand(cmd shared.Command) shared.Response {
	log.Printf("Received command: %s", cmd.Type)

	switch cmd.Type {
	case shared.CmdStart:
		return s.handleStartCommand(cmd)
	case shared.CmdPlay:
		return s.handlePlayCommand(cmd)
	case shared.CmdPause:
		return s.playbackHandler.HandlePause()
	case shared.CmdStop:
		return s.playbackHandler.HandleStop()
	case shared.CmdSkip:
		return s.navigationHandler.HandleSkip(cmd)
	case shared.CmdBack:
		return s.navigationHandler.HandleBack(cmd)
	case shared.CmdStatus:
		return s.infoHandler.HandleStatus()
	case shared.CmdList:
		return s.infoHandler.HandleList()
	case shared.CmdVolume:
		return s.infoHandler.HandleVolume(cmd)
	case shared.CmdExit:
		return s.handleExitCommand()
	default:
		return shared.NewErrorResponse(fmt.Sprintf("Unknown command: %s", cmd.Type))
	}
}

func (s *Server) handleStartCommand(cmd shared.Command) shared.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate source path
	if cmd.Path == "" {
		return shared.NewErrorResponse("No path provided for start command")
	}

	// Expand path
	expandedPath, err := s.loader.ExpandPath(cmd.Path)
	if err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Invalid path: %v", err))
	}

	// Check if path exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return shared.NewErrorResponse(fmt.Sprintf("Path does not exist: %s", expandedPath))
	}

	// Load tracks based on source type
	switch cmd.Source {
	case shared.SourceFolder:
		if err := s.LoadFolder(expandedPath); err != nil {
			return shared.NewErrorResponse(fmt.Sprintf("Failed to load folder: %v", err))
		}
	case shared.SourcePlaylist:
		if err := s.LoadPlaylist(expandedPath); err != nil {
			return shared.NewErrorResponse(fmt.Sprintf("Failed to load playlist: %v", err))
		}
	default:
		return shared.NewErrorResponse(fmt.Sprintf("Unsupported source type: %s", cmd.Source))
	}

	trackCount := s.playlist.TrackCount()
	if trackCount == 0 {
		return shared.NewErrorResponse("No audio files found in the specified location")
	}

	// Set the first track as current in the player
	firstTrack := s.playlist.GetCurrentTrack()
	if firstTrack != nil {
		s.player.SetCurrentTrack(firstTrack)
	}

	log.Printf("Loaded %d tracks from %s: %s", trackCount, cmd.Source, expandedPath)
	return shared.NewSuccessResponse(
		fmt.Sprintf("Loaded %d tracks from %s", trackCount, cmd.Source),
		nil,
	)
}

func (s *Server) handlePlayCommand(cmd shared.Command) shared.Response {
	// Check if source info is provided (play with source loading)
	if cmd.Path != "" && cmd.Source != "" {
		// Hot-swap source and play
		return s.handlePlayWithSource(cmd)
	}

	// Regular play command - just start/resume playback
	return s.playbackHandler.HandlePlay()
}

func (s *Server) handlePlayWithSource(cmd shared.Command) shared.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate and expand path
	expandedPath, err := s.loader.ExpandPath(cmd.Path)
	if err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Invalid path: %v", err))
	}

	// Check if path exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return shared.NewErrorResponse(fmt.Sprintf("Path does not exist: %s", expandedPath))
	}

	// Load tracks based on source type
	switch cmd.Source {
	case shared.SourceFolder:
		if err := s.LoadFolder(expandedPath); err != nil {
			return shared.NewErrorResponse(fmt.Sprintf("Failed to load folder: %v", err))
		}
	case shared.SourcePlaylist:
		if err := s.LoadPlaylist(expandedPath); err != nil {
			return shared.NewErrorResponse(fmt.Sprintf("Failed to load playlist: %v", err))
		}
	default:
		return shared.NewErrorResponse(fmt.Sprintf("Unsupported source type: %s", cmd.Source))
	}

	trackCount := s.playlist.TrackCount()
	if trackCount == 0 {
		return shared.NewErrorResponse("No audio files found in the specified location")
	}

	// Set the first track as current in the player
	firstTrack := s.playlist.GetCurrentTrack()
	if firstTrack != nil {
		s.player.SetCurrentTrack(firstTrack)
	}

	// Start playback immediately
	playResp := s.playbackHandler.HandlePlay()
	if !playResp.Success {
		return playResp
	}

	log.Printf("Loaded %d tracks from %s: %s and started playback", trackCount, cmd.Source, expandedPath)
	return shared.NewSuccessResponse(
		fmt.Sprintf("Loaded %d tracks from %s and started playback", trackCount, cmd.Source),
		nil,
	)
}

func (s *Server) handleExitCommand() shared.Response {
	// Stop the daemon
	go func() {
		// Give time for response to be sent
		// time.Sleep(100 * time.Millisecond)

		// Properly stop the player first
		if err := s.player.Close(); err != nil {
			log.Printf("Error closing player: %v", err)
		}

		s.Stop()
		os.Exit(0)
	}()

	return shared.NewSuccessResponse("Exiting daemon", nil)
}

// onTrackComplete handles auto-advancing to the next track when current track ends
func (s *Server) onTrackComplete() {
	log.Println("Track completed, scheduling auto-advance to next track")

	// Run auto-advance in a separate goroutine to avoid deadlocks
	go func() {
		// Small delay to ensure the track completion callback has fully finished
		time.Sleep(100 * time.Millisecond)

		s.mu.Lock()
		defer s.mu.Unlock()

		log.Println("Executing auto-advance to next track")

		// Try to advance to next track in playlist
		if s.playlist.Next() {
			// Successfully moved to next track
			nextTrack := s.playlist.GetCurrentTrack()
			if nextTrack != nil {
				log.Printf("Auto-advancing to: %s", nextTrack.Filename)

				// Load the next track into the player
				s.player.SetCurrentTrack(nextTrack)

				// Auto-start playing the next track
				if err := s.player.Play(); err != nil {
					log.Printf("Failed to auto-play next track: %v", err)
				} else {
					log.Printf("Now auto-playing: %s", nextTrack.Filename)
				}
			} else {
				log.Println("Next track is nil, cannot advance")
			}
		} else {
			// Reached end of playlist
			log.Println("Reached end of playlist, playback complete")
		}
	}()
}

// Helper methods


// LoadFolder loads tracks from a folder using the loader
func (s *Server) LoadFolder(folderPath string) error {
	tracks, err := s.loader.LoadFolder(folderPath)
	if err != nil {
		return err
	}

	err = s.playlist.LoadTracks(tracks, folderPath, shared.SourceFolder)
	if err != nil {
		log.Printf("LoadFolder: LoadTracks failed: %v", err)
		return err
	}

	log.Printf("LoadFolder: Successfully loaded %d tracks into playlist", len(tracks))
	return nil
}

// LoadPlaylist loads tracks from a playlist file using the loader
func (s *Server) LoadPlaylist(playlistPath string) error {
	tracks, err := s.loader.LoadPlaylist(playlistPath)
	if err != nil {
		return err
	}

	err = s.playlist.LoadTracks(tracks, playlistPath, shared.SourcePlaylist)
	if err != nil {
		log.Printf("LoadPlaylist: LoadTracks failed: %v", err)
		return err
	}

	log.Printf("LoadPlaylist: Successfully loaded %d tracks into playlist", len(tracks))
	return nil
}
