package daemon

import (
	"testing"

	"github.com/cerberussg/auxbox/internal/shared"
)

func TestPlaylist_NewPlaylist(t *testing.T) {
	playlist := NewPlaylist()

	if playlist == nil {
		t.Fatal("NewPlaylist() returned nil")
	}

	if playlist.TrackCount() != 0 {
		t.Errorf("New playlist should be empty, got %d tracks", playlist.TrackCount())
	}

	if playlist.GetCurrentIndex() != 0 {
		t.Errorf("Current index should be 0, got %d", playlist.GetCurrentIndex())
	}
}

func TestPlaylist_LoadTracks(t *testing.T) {
	playlist := NewPlaylist()

	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
		{Filename: "track3.mp3", Path: "/path/track3.mp3"},
	}

	err := playlist.LoadTracks(tracks, "/path", shared.SourceFolder)
	if err != nil {
		t.Fatalf("LoadTracks() failed: %v", err)
	}

	if playlist.TrackCount() != 3 {
		t.Errorf("Expected 3 tracks, got %d", playlist.TrackCount())
	}

	if playlist.GetCurrentIndex() != 0 {
		t.Errorf("Current index should reset to 0, got %d", playlist.GetCurrentIndex())
	}

	if playlist.GetSource() != "/path" {
		t.Errorf("Source should be '/path', got '%s'", playlist.GetSource())
	}

	if playlist.GetSourceType() != shared.SourceFolder {
		t.Errorf("Source type should be folder, got %s", playlist.GetSourceType())
	}
}

func TestPlaylist_GetCurrentTrack(t *testing.T) {
	playlist := NewPlaylist()

	// Empty playlist
	if track := playlist.GetCurrentTrack(); track != nil {
		t.Error("Empty playlist should return nil for current track")
	}

	// Load tracks
	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Should get first track
	currentTrack := playlist.GetCurrentTrack()
	if currentTrack == nil {
		t.Fatal("Should have a current track")
	}

	if currentTrack.Filename != "track1.mp3" {
		t.Errorf("Current track should be track1.mp3, got %s", currentTrack.Filename)
	}
}

func TestPlaylist_Next(t *testing.T) {
	playlist := NewPlaylist()

	// Empty playlist
	if playlist.Next() {
		t.Error("Next() should return false for empty playlist")
	}

	// Load tracks
	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
		{Filename: "track3.mp3", Path: "/path/track3.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Move to next track
	if !playlist.Next() {
		t.Error("Next() should return true when moving from track 0 to 1")
	}

	if playlist.GetCurrentIndex() != 1 {
		t.Errorf("Current index should be 1, got %d", playlist.GetCurrentIndex())
	}

	// Move to last track
	if !playlist.Next() {
		t.Error("Next() should return true when moving from track 1 to 2")
	}

	if playlist.GetCurrentIndex() != 2 {
		t.Errorf("Current index should be 2, got %d", playlist.GetCurrentIndex())
	}

	// Try to go beyond last track
	if playlist.Next() {
		t.Error("Next() should return false when already at last track")
	}

	if playlist.GetCurrentIndex() != 2 {
		t.Errorf("Current index should stay at 2, got %d", playlist.GetCurrentIndex())
	}
}

func TestPlaylist_Previous(t *testing.T) {
	playlist := NewPlaylist()

	// Empty playlist
	if playlist.Previous() {
		t.Error("Previous() should return false for empty playlist")
	}

	// Load tracks
	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
		{Filename: "track3.mp3", Path: "/path/track3.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Try to go before first track
	if playlist.Previous() {
		t.Error("Previous() should return false when already at first track")
	}

	// Move to last track first
	playlist.SetCurrentIndex(2)

	// Move to previous track
	if !playlist.Previous() {
		t.Error("Previous() should return true when moving from track 2 to 1")
	}

	if playlist.GetCurrentIndex() != 1 {
		t.Errorf("Current index should be 1, got %d", playlist.GetCurrentIndex())
	}

	// Move to first track
	if !playlist.Previous() {
		t.Error("Previous() should return true when moving from track 1 to 0")
	}

	if playlist.GetCurrentIndex() != 0 {
		t.Errorf("Current index should be 0, got %d", playlist.GetCurrentIndex())
	}
}

func TestPlaylist_SetCurrentIndex(t *testing.T) {
	playlist := NewPlaylist()

	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
		{Filename: "track3.mp3", Path: "/path/track3.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Valid index
	if !playlist.SetCurrentIndex(1) {
		t.Error("SetCurrentIndex(1) should return true")
	}

	if playlist.GetCurrentIndex() != 1 {
		t.Errorf("Current index should be 1, got %d", playlist.GetCurrentIndex())
	}

	// Invalid negative index
	if playlist.SetCurrentIndex(-1) {
		t.Error("SetCurrentIndex(-1) should return false")
	}

	// Invalid high index
	if playlist.SetCurrentIndex(3) {
		t.Error("SetCurrentIndex(3) should return false for 3 tracks")
	}

	// Index should remain unchanged
	if playlist.GetCurrentIndex() != 1 {
		t.Errorf("Current index should remain 1, got %d", playlist.GetCurrentIndex())
	}
}

func TestPlaylist_GetTrackList(t *testing.T) {
	playlist := NewPlaylist()

	// Empty playlist
	tracks := playlist.GetTrackList()
	if len(tracks) != 0 {
		t.Error("Empty playlist should return empty track list")
	}

	// Load tracks
	originalTracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
	}
	playlist.LoadTracks(originalTracks, "/path", shared.SourceFolder)

	// Get track list
	retrievedTracks := playlist.GetTrackList()
	if len(retrievedTracks) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(retrievedTracks))
	}

	// Verify tracks match
	for i, track := range retrievedTracks {
		if track.Filename != originalTracks[i].Filename {
			t.Errorf("Track %d filename mismatch: got %s, want %s",
				i, track.Filename, originalTracks[i].Filename)
		}
		if track.Path != originalTracks[i].Path {
			t.Errorf("Track %d path mismatch: got %s, want %s",
				i, track.Path, originalTracks[i].Path)
		}
	}

	// Verify it's a copy (modifying returned slice shouldn't affect playlist)
	retrievedTracks[0].Filename = "modified.mp3"
	originalTrack := playlist.GetCurrentTrack()
	if originalTrack.Filename == "modified.mp3" {
		t.Error("GetTrackList() should return a copy, not reference to internal tracks")
	}
}

func TestPlaylist_Clear(t *testing.T) {
	playlist := NewPlaylist()

	// Load tracks
	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Verify tracks are loaded
	if playlist.TrackCount() != 2 {
		t.Error("Tracks should be loaded before clear")
	}

	// Clear playlist
	playlist.Clear()

	// Verify everything is cleared
	if playlist.TrackCount() != 0 {
		t.Errorf("Playlist should be empty after clear, got %d tracks", playlist.TrackCount())
	}

	if playlist.GetCurrentIndex() != 0 {
		t.Errorf("Current index should be 0 after clear, got %d", playlist.GetCurrentIndex())
	}

	if playlist.GetSource() != "" {
		t.Errorf("Source should be empty after clear, got '%s'", playlist.GetSource())
	}

	if playlist.GetCurrentTrack() != nil {
		t.Error("Current track should be nil after clear")
	}
}

func TestPlaylist_ConcurrentAccess(t *testing.T) {
	playlist := NewPlaylist()

	// Load tracks
	tracks := []*shared.Track{
		{Filename: "track1.mp3", Path: "/path/track1.mp3"},
		{Filename: "track2.mp3", Path: "/path/track2.mp3"},
		{Filename: "track3.mp3", Path: "/path/track3.mp3"},
	}
	playlist.LoadTracks(tracks, "/path", shared.SourceFolder)

	// Test concurrent access (basic smoke test)
	done := make(chan bool, 3)

	// Goroutine 1: Navigate forward
	go func() {
		for i := 0; i < 10; i++ {
			playlist.Next()
			playlist.Previous()
		}
		done <- true
	}()

	// Goroutine 2: Read current track
	go func() {
		for i := 0; i < 10; i++ {
			playlist.GetCurrentTrack()
			playlist.GetCurrentIndex()
		}
		done <- true
	}()

	// Goroutine 3: Read track list
	go func() {
		for i := 0; i < 10; i++ {
			playlist.GetTrackList()
			playlist.TrackCount()
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we get here without panicking, concurrent access is working
	if playlist.TrackCount() != 3 {
		t.Error("Playlist should still have 3 tracks after concurrent access")
	}
}
