package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cerberussg/auxbox/internal/shared"
)

// Server represents the auxbox daemon server
type Server struct {
	transport shared.Transport
	player    *Player
	playlist  *Playlist
	isRunning bool
	mu        sync.RWMutex
}

// NewServer creates a new daemon server instance
func NewServer() *Server {
	return &Server{
		transport: shared.NewUnixSocketTransport(),
		player:    NewPlayer(),
		playlist:  NewPlaylist(),
		isRunning: false,
	}
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
		return s.handlePlayCommand()
	case shared.CmdPause:
		return s.handlePauseCommand()
	case shared.CmdSkip:
		return s.handleSkipCommand(cmd)
	case shared.CmdBack:
		return s.handleBackCommand(cmd)
	case shared.CmdStatus:
		return s.handleStatusCommand()
	case shared.CmdList:
		return s.handleListCommand()
	case shared.CmdStop:
		return s.handleStopCommand()
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
	expandedPath, err := s.expandPath(cmd.Path)
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
		if err := s.loadFolder(expandedPath); err != nil {
			return shared.NewErrorResponse(fmt.Sprintf("Failed to load folder: %v", err))
		}
	case shared.SourcePlaylist:
		if err := s.loadPlaylist(expandedPath); err != nil {
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

func (s *Server) handlePlayCommand() shared.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.playlist.TrackCount() == 0 {
		return shared.NewErrorResponse("No tracks loaded. Use 'auxbox start --folder <path>' first.")
	}

	// Ensure player has current track (in case it got out of sync)
	if s.player.GetCurrentTrack() == nil {
		currentTrack := s.playlist.GetCurrentTrack()
		if currentTrack != nil {
			s.player.SetCurrentTrack(currentTrack)
		}
	}

	if err := s.player.Play(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to play: %v", err))
	}

	currentTrack := s.player.GetCurrentTrack()
	if currentTrack != nil {
		log.Printf("Playing: %s", currentTrack.Filename)
		return shared.NewSuccessResponse(fmt.Sprintf("Playing: %s", currentTrack.Filename), nil)
	}

	return shared.NewSuccessResponse("Playback started", nil)
}

func (s *Server) handlePauseCommand() shared.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.player.Pause(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to pause: %v", err))
	}

	log.Println("Playback paused")
	return shared.NewSuccessResponse("Playback paused", nil)
}

func (s *Server) handleSkipCommand(cmd shared.Command) shared.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := cmd.Count
	if count <= 0 {
		count = 1
	}

	skipped := 0
	for i := 0; i < count; i++ {
		if s.playlist.Next() {
			skipped++
		} else {
			break // Reached end of playlist
		}
	}

	if skipped == 0 {
		return shared.NewErrorResponse("Already at the end of playlist")
	}

	// Update player with new current track
	currentTrack := s.playlist.GetCurrentTrack()
	if currentTrack != nil {
		s.player.SetCurrentTrack(currentTrack)
		log.Printf("Skipped %d track(s), now at: %s", skipped, currentTrack.Filename)
		return shared.NewSuccessResponse(
			fmt.Sprintf("Skipped %d track(s), now playing: %s", skipped, currentTrack.Filename),
			nil,
		)
	}

	return shared.NewSuccessResponse(fmt.Sprintf("Skipped %d track(s)", skipped), nil)
}

func (s *Server) handleBackCommand(cmd shared.Command) shared.Response {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := cmd.Count
	if count <= 0 {
		count = 1
	}

	moved := 0
	for i := 0; i < count; i++ {
		if s.playlist.Previous() {
			moved++
		} else {
			break // Reached beginning of playlist
		}
	}

	if moved == 0 {
		return shared.NewErrorResponse("Already at the beginning of playlist")
	}

	// Update player with new current track
	currentTrack := s.playlist.GetCurrentTrack()
	if currentTrack != nil {
		s.player.SetCurrentTrack(currentTrack)
		log.Printf("Moved back %d track(s), now at: %s", moved, currentTrack.Filename)
		return shared.NewSuccessResponse(
			fmt.Sprintf("Moved back %d track(s), now playing: %s", moved, currentTrack.Filename),
			nil,
		)
	}

	return shared.NewSuccessResponse(fmt.Sprintf("Moved back %d track(s)", moved), nil)
}

func (s *Server) handleStatusCommand() shared.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currentTrack := s.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewSuccessResponse("No track loaded", nil)
	}

	status := s.player.GetStatus()
	trackInfo := shared.TrackInfo{
		Filename:    currentTrack.Filename,
		Path:        currentTrack.Path,
		Duration:    status.Duration,
		Position:    status.Position,
		TrackNumber: s.playlist.GetCurrentIndex() + 1, // 1-based indexing for display
		TotalTracks: s.playlist.TrackCount(),
		Source:      s.playlist.GetSource(),
	}

	return shared.NewSuccessResponse("Current status", trackInfo)
}

func (s *Server) handleListCommand() shared.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tracks := s.playlist.GetTrackList()
	if len(tracks) == 0 {
		return shared.NewSuccessResponse("No tracks loaded", nil)
	}

	// Convert tracks to string slice for JSON serialization
	trackNames := make([]string, len(tracks))
	for i, track := range tracks {
		trackNames[i] = track.Filename
	}

	playlistInfo := shared.PlaylistInfo{
		Source:     s.playlist.GetSource(),
		SourceType: string(s.playlist.GetSourceType()),
		Tracks:     trackNames,
		CurrentIdx: s.playlist.GetCurrentIndex(),
	}

	return shared.NewSuccessResponse(fmt.Sprintf("%d tracks loaded", len(tracks)), playlistInfo)
}

func (s *Server) handleStopCommand() shared.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.player.Stop(); err != nil {
		return shared.NewErrorResponse(fmt.Sprintf("Failed to stop: %v", err))
	}

	log.Println("Playback stopped")
	return shared.NewSuccessResponse("Playback stopped", nil)
}

func (s *Server) handleExitCommand() shared.Response {
	// Stop the daemon
	go func() {
		// Give time for response to be sent
		// time.Sleep(100 * time.Millisecond)
		s.Stop()
		os.Exit(0)
	}()

	return shared.NewSuccessResponse("Exiting daemon", nil)
}

// Helper methods

func (s *Server) expandPath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get home directory: %v", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path: %v", err)
	}

	return absPath, nil
}

func (s *Server) loadFolder(folderPath string) error {
	// Supported audio extensions
	supportedExts := map[string]bool{
		".mp3":  true,
		".aiff": true,
		".aif":  true,
		".wav":  true, // For testing
	}

	var tracks []*shared.Track

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedExts[ext] {
			track := &shared.Track{
				Filename: info.Name(),
				Path:     path,
			}
			tracks = append(tracks, track)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return s.playlist.LoadTracks(tracks, folderPath, shared.SourceFolder)
}

func (s *Server) loadPlaylist(playlistPath string) error {
	// TODO: Implement playlist file parsing (M3U, PLS, etc.)
	// For now, just return an error
	return fmt.Errorf("playlist loading not yet implemented")
}
