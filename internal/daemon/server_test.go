package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cerberussg/auxbox/internal/shared"
)

func TestServer_NewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.player == nil {
		t.Error("Server player should not be nil")
	}

	if server.playlist == nil {
		t.Error("Server playlist should not be nil")
	}

	if server.isRunning {
		t.Error("Server should not be running initially")
	}
}

func TestServer_HandleStartCommand_Folder(t *testing.T) {
	server := NewServer()

	// Create a temporary directory with some test files
	tmpDir, err := os.MkdirTemp("", "auxbox_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test audio files
	testFiles := []string{"track1.mp3", "track2.aiff", "track3.wav", "notaudio.txt"}
	for _, filename := range testFiles {
		file, err := os.Create(filepath.Join(tmpDir, filename))
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}

	// Test start command with folder
	cmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	resp := server.HandleCommand(cmd)

	if !resp.Success {
		t.Errorf("Start command failed: %s", resp.Message)
	}

	// Should have loaded 3 audio files (mp3, aiff, wav) but not txt
	if server.playlist.TrackCount() != 3 {
		t.Errorf("Expected 3 tracks, got %d", server.playlist.TrackCount())
	}

	tracks := server.playlist.GetTrackList()
	expectedFiles := map[string]bool{"track1.mp3": false, "track2.aiff": false, "track3.wav": false}

	for _, track := range tracks {
		if _, exists := expectedFiles[track.Filename]; exists {
			expectedFiles[track.Filename] = true
		}
	}

	for filename, found := range expectedFiles {
		if !found {
			t.Errorf("Expected track %s not found", filename)
		}
	}
}

func TestServer_HandleStartCommand_InvalidPath(t *testing.T) {
	server := NewServer()

	cmd := shared.NewStartCommand(shared.SourceFolder, "/path/that/does/not/exist")
	resp := server.HandleCommand(cmd)

	if resp.Success {
		t.Error("Start command should fail for non-existent path")
	}

	if !containsString(resp.Message, "does not exist") {
		t.Errorf("Error message should mention path doesn't exist, got: %s", resp.Message)
	}
}

func TestServer_HandleStartCommand_EmptyFolder(t *testing.T) {
	server := NewServer()

	// Create empty temporary directory
	tmpDir, err := os.MkdirTemp("", "auxbox_test_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	resp := server.HandleCommand(cmd)

	if resp.Success {
		t.Error("Start command should fail for empty folder")
	}

	if !containsString(resp.Message, "No audio files found") {
		t.Errorf("Error message should mention no audio files, got: %s", resp.Message)
	}
}

func TestServer_HandlePlayCommand(t *testing.T) {
	server := NewServer()

	// Test play without tracks loaded
	cmd := shared.NewPlayCommand()
	resp := server.HandleCommand(cmd)

	if resp.Success {
		t.Error("Play command should fail when no tracks are loaded")
	}

	// Load some tracks first
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	startResp := server.HandleCommand(startCmd)
	if !startResp.Success {
		t.Fatal("Failed to load test tracks")
	}

	// Now play should work
	playResp := server.HandleCommand(cmd)
	if !playResp.Success {
		t.Errorf("Play command failed: %s", playResp.Message)
	}

	// Check player state
	if !server.player.IsPlaying() {
		t.Error("Player should be playing after play command")
	}
}

func TestServer_HandlePauseCommand(t *testing.T) {
	server := NewServer()

	// Setup tracks and start playing
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	playCmd := shared.NewPlayCommand()
	server.HandleCommand(playCmd)

	// Now test pause
	pauseCmd := shared.NewPauseCommand()
	resp := server.HandleCommand(pauseCmd)

	if !resp.Success {
		t.Errorf("Pause command failed: %s", resp.Message)
	}

	if server.player.IsPlaying() {
		t.Error("Player should not be playing after pause command")
	}

	if !server.player.IsPaused() {
		t.Error("Player should be paused after pause command")
	}

	// Position should NOT be reset (difference from stop)
	status := server.player.GetStatus()
	// We can't test exact position since it's simulated, but it shouldn't be reset
	if status.Position == "" {
		t.Error("Position should be maintained after pause")
	}
}

func TestServer_HandleSkipCommand(t *testing.T) {
	server := NewServer()

	// Setup multiple tracks
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	// Should start at track 0
	if server.playlist.GetCurrentIndex() != 0 {
		t.Error("Should start at track 0")
	}

	// Skip 1 track
	skipCmd := shared.NewSkipCommand(1)
	resp := server.HandleCommand(skipCmd)

	if !resp.Success {
		t.Errorf("Skip command failed: %s", resp.Message)
	}

	if server.playlist.GetCurrentIndex() != 1 {
		t.Errorf("Should be at track 1 after skip, got %d", server.playlist.GetCurrentIndex())
	}

	// Skip 2 more tracks
	skip2Cmd := shared.NewSkipCommand(2)
	server.HandleCommand(skip2Cmd)

	// Should be at last track now (we have 3 tracks: 0,1,2)
	expectedIndex := 2
	if server.playlist.GetCurrentIndex() != expectedIndex {
		t.Errorf("Should be at track %d after skipping, got %d", expectedIndex, server.playlist.GetCurrentIndex())
	}
}

func TestServer_HandleBackCommand(t *testing.T) {
	server := NewServer()

	// Setup tracks
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	// Skip to last track first
	skipCmd := shared.NewSkipCommand(2)
	server.HandleCommand(skipCmd)

	currentIndex := server.playlist.GetCurrentIndex()

	// Go back 1 track
	backCmd := shared.NewBackCommand(1)
	resp := server.HandleCommand(backCmd)

	if !resp.Success {
		t.Errorf("Back command failed: %s", resp.Message)
	}

	expectedIndex := currentIndex - 1
	if server.playlist.GetCurrentIndex() != expectedIndex {
		t.Errorf("Should be at track %d after going back, got %d", expectedIndex, server.playlist.GetCurrentIndex())
	}
}

func TestServer_HandleStatusCommand(t *testing.T) {
	server := NewServer()

	// Status without tracks
	statusCmd := shared.NewStatusCommand()
	resp := server.HandleCommand(statusCmd)

	if !resp.Success {
		t.Errorf("Status command failed: %s", resp.Message)
	}

	// Load tracks
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	// Status with tracks
	resp2 := server.HandleCommand(statusCmd)
	if !resp2.Success {
		t.Errorf("Status command failed: %s", resp2.Message)
	}

	// Should have track info in response data
	if resp2.Data == nil {
		t.Error("Status response should include track data")
	}
}

func TestServer_HandleListCommand(t *testing.T) {
	server := NewServer()

	// List without tracks
	listCmd := shared.NewListCommand()
	resp := server.HandleCommand(listCmd)

	if !resp.Success {
		t.Errorf("List command failed: %s", resp.Message)
	}

	// Load tracks
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	// List with tracks
	resp2 := server.HandleCommand(listCmd)
	if !resp2.Success {
		t.Errorf("List command failed: %s", resp2.Message)
	}

	// Should have playlist info
	if resp2.Data == nil {
		t.Error("List response should include playlist data")
	}
}

func TestServer_HandleUnknownCommand(t *testing.T) {
	server := NewServer()

	// Create command with unknown type
	cmd := shared.Command{Type: "unknown"}
	resp := server.HandleCommand(cmd)

	if resp.Success {
		t.Error("Unknown command should fail")
	}

	if !containsString(resp.Message, "Unknown command") {
		t.Errorf("Error message should mention unknown command, got: %s", resp.Message)
	}
}

// Helper functions

func createTestDirectory(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "auxbox_test")
	if err != nil {
		t.Fatal(err)
	}

	// Create test audio files
	testFiles := []string{"track1.mp3", "track2.aiff", "track3.wav"}
	for _, filename := range testFiles {
		file, err := os.Create(filepath.Join(tmpDir, filename))
		if err != nil {
			t.Fatal(err)
		}
		file.Close()
	}

	return tmpDir
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestServer_HandleStopCommand(t *testing.T) {
	server := NewServer()

	// Setup tracks and start playing
	tmpDir := createTestDirectory(t)
	defer os.RemoveAll(tmpDir)

	startCmd := shared.NewStartCommand(shared.SourceFolder, tmpDir)
	server.HandleCommand(startCmd)

	playCmd := shared.NewPlayCommand()
	server.HandleCommand(playCmd)

	// Verify playing
	if !server.player.IsPlaying() {
		t.Fatal("Player should be playing before stop")
	}

	// Now test stop
	stopCmd := shared.NewStopCommand()
	resp := server.HandleCommand(stopCmd)

	if !resp.Success {
		t.Errorf("Stop command failed: %s", resp.Message)
	}

	if server.player.IsPlaying() {
		t.Error("Player should not be playing after stop command")
	}

	if server.player.IsPaused() {
		t.Error("Player should not be paused after stop command (should be fully stopped)")
	}

	// Position should be reset
	status := server.player.GetStatus()
	if status.Position != "0:00" {
		t.Errorf("Position should be reset to 0:00 after stop, got %s", status.Position)
	}
}

func TestServer_HandleExitCommand(t *testing.T) {
	server := NewServer()

	// Test exit command
	exitCmd := shared.NewExitCommand()
	resp := server.HandleCommand(exitCmd)

	if !resp.Success {
		t.Errorf("Exit command failed: %s", resp.Message)
	}

	if !containsString(resp.Message, "Exiting") {
		t.Errorf("Exit response should mention exiting, got: %s", resp.Message)
	}

	// Note: We can't easily test the actual os.Exit() call in unit tests
	// The goroutine will attempt to stop the server and exit, but in tests
	// this won't actually terminate the test process
}
