package daemon

import (
	"testing"

	"github.com/cerberussg/auxbox/internal/shared"
)

func TestPlayer_NewPlayer(t *testing.T) {
	player := NewPlayer()

	if player == nil {
		t.Fatal("NewPlayer() returned nil")
	}

	if player.IsPlaying() {
		t.Error("New player should not be playing")
	}

	if player.IsPaused() {
		t.Error("New player should not be paused")
	}

	status := player.GetStatus()
	if status.Volume != 1.0 {
		t.Errorf("Default volume should be 1.0, got %f", status.Volume)
	}

	if status.Position != "0:00" {
		t.Errorf("Default position should be '0:00', got '%s'", status.Position)
	}

	if status.Duration != "0:00" {
		t.Errorf("Default duration should be '0:00', got '%s'", status.Duration)
	}
}

func TestPlayer_SetCurrentTrack(t *testing.T) {
	player := NewPlayer()

	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}

	player.SetCurrentTrack(track)

	currentTrack := player.GetCurrentTrack()
	if currentTrack == nil {
		t.Fatal("Current track should not be nil")
	}

	if currentTrack.Filename != "test.mp3" {
		t.Errorf("Current track filename should be 'test.mp3', got '%s'", currentTrack.Filename)
	}

	if currentTrack.Path != "/path/to/test.mp3" {
		t.Errorf("Current track path should be '/path/to/test.mp3', got '%s'", currentTrack.Path)
	}

	// Position should reset when setting new track
	status := player.GetStatus()
	if status.Position != "0:00" {
		t.Errorf("Position should reset to '0:00' when setting new track, got '%s'", status.Position)
	}
}

func TestPlayer_Play(t *testing.T) {
	player := NewPlayer()

	// Try to play without track
	err := player.Play()
	if err == nil {
		t.Error("Play() should return error when no track is loaded")
	}

	// Set a track and try again
	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)

	err = player.Play()
	if err != nil {
		t.Errorf("Play() should not return error when track is loaded: %v", err)
	}

	if !player.IsPlaying() {
		t.Error("Player should be playing after Play()")
	}

	if player.IsPaused() {
		t.Error("Player should not be paused after Play()")
	}

	status := player.GetStatus()
	if !status.IsPlaying {
		t.Error("Status should show playing")
	}

	if status.IsPaused {
		t.Error("Status should not show paused")
	}
}

func TestPlayer_Pause(t *testing.T) {
	player := NewPlayer()

	// Try to pause without playing
	err := player.Pause()
	if err == nil {
		t.Error("Pause() should return error when not playing")
	}

	// Start playing first
	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)
	player.Play()

	// Now pause should work
	err = player.Pause()
	if err != nil {
		t.Errorf("Pause() should not return error when playing: %v", err)
	}

	if player.IsPlaying() {
		t.Error("Player should not be playing after Pause()")
	}

	if !player.IsPaused() {
		t.Error("Player should be paused after Pause()")
	}

	status := player.GetStatus()
	if status.IsPlaying {
		t.Error("Status should not show playing after pause")
	}

	if !status.IsPaused {
		t.Error("Status should show paused after pause")
	}
}

func TestPlayer_Stop(t *testing.T) {
	player := NewPlayer()

	// Set track and start playing
	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)
	player.Play()

	// Stop playback
	err := player.Stop()
	if err != nil {
		t.Errorf("Stop() should not return error: %v", err)
	}

	if player.IsPlaying() {
		t.Error("Player should not be playing after Stop()")
	}

	if player.IsPaused() {
		t.Error("Player should not be paused after Stop()")
	}

	status := player.GetStatus()
	if status.IsPlaying {
		t.Error("Status should not show playing after stop")
	}

	if status.IsPaused {
		t.Error("Status should not show paused after stop")
	}

	if status.Position != "0:00" {
		t.Errorf("Position should reset to '0:00' after stop, got '%s'", status.Position)
	}
}

func TestPlayer_SetVolume(t *testing.T) {
	player := NewPlayer()

	// Test valid volumes
	validVolumes := []float64{0.0, 0.5, 1.0}
	for _, vol := range validVolumes {
		err := player.SetVolume(vol)
		if err != nil {
			t.Errorf("SetVolume(%f) should not return error: %v", vol, err)
		}

		status := player.GetStatus()
		if status.Volume != vol {
			t.Errorf("Volume should be %f, got %f", vol, status.Volume)
		}
	}

	// Test invalid volumes
	invalidVolumes := []float64{-0.1, 1.1, -1.0, 2.0}
	for _, vol := range invalidVolumes {
		err := player.SetVolume(vol)
		if err == nil {
			t.Errorf("SetVolume(%f) should return error for invalid volume", vol)
		}
	}
}

func TestPlayer_PlayPauseResume(t *testing.T) {
	player := NewPlayer()

	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)

	// Play
	player.Play()
	if !player.IsPlaying() || player.IsPaused() {
		t.Error("Should be playing after Play()")
	}

	// Pause
	player.Pause()
	if player.IsPlaying() || !player.IsPaused() {
		t.Error("Should be paused after Pause()")
	}

	// Resume (play again)
	player.Play()
	if !player.IsPlaying() || player.IsPaused() {
		t.Error("Should be playing after resume")
	}
}

func TestPlayer_GetStatus(t *testing.T) {
	player := NewPlayer()

	// Test initial status
	status := player.GetStatus()
	if status.IsPlaying {
		t.Error("Initial status should not show playing")
	}
	if status.IsPaused {
		t.Error("Initial status should not show paused")
	}
	if status.Position != "0:00" {
		t.Error("Initial position should be '0:00'")
	}
	if status.Duration != "0:00" {
		t.Error("Initial duration should be '0:00'")
	}
	if status.Volume != 1.0 {
		t.Error("Initial volume should be 1.0")
	}

	// Set track and play
	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)
	player.Play()

	// Test status during playback
	status = player.GetStatus()
	if !status.IsPlaying {
		t.Error("Status should show playing")
	}
	if status.IsPaused {
		t.Error("Status should not show paused during playback")
	}
	// Duration gets set to placeholder when playing
	if status.Duration == "0:00" {
		t.Error("Duration should be set during playback")
	}
}

func TestPlayer_ConcurrentAccess(t *testing.T) {
	player := NewPlayer()

	track := &shared.Track{
		Filename: "test.mp3",
		Path:     "/path/to/test.mp3",
	}
	player.SetCurrentTrack(track)

	// Test concurrent access (basic smoke test)
	done := make(chan bool, 3)

	// Goroutine 1: Play/pause repeatedly
	go func() {
		for i := 0; i < 10; i++ {
			player.Play()
			player.Pause()
		}
		done <- true
	}()

	// Goroutine 2: Read status repeatedly
	go func() {
		for i := 0; i < 10; i++ {
			player.GetStatus()
			player.IsPlaying()
			player.IsPaused()
		}
		done <- true
	}()

	// Goroutine 3: Set volume repeatedly
	go func() {
		for i := 0; i < 10; i++ {
			player.SetVolume(0.5)
			player.SetVolume(1.0)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// If we get here without panicking, concurrent access is working
	status := player.GetStatus()
	if status.Volume != 1.0 && status.Volume != 0.5 {
		t.Errorf("Unexpected volume after concurrent access: %f", status.Volume)
	}
}
