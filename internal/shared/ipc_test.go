package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestUnixSocketTransport_NewTransport(t *testing.T) {
	transport := NewUnixSocketTransport()

	if transport == nil {
		t.Fatal("NewUnixSocketTransport() returned nil")
	}

	socketPath := transport.GetSocketPath()
	if socketPath == "" {
		t.Error("Socket path should not be empty")
	}

	// Should contain auxbox.sock somewhere
	if !contains(socketPath, "auxbox.sock") {
		t.Errorf("Socket path should contain 'auxbox.sock', got %s", socketPath)
	}
}

func TestUnixSocketTransport_IsRunning(t *testing.T) {
	transport := NewUnixSocketTransport()

	// Should not be running initially
	if transport.IsRunning() {
		t.Error("Transport should not be running initially")
	}

	// Start a daemon in background
	done := make(chan error)
	go func() {
		err := transport.Listen(func(cmd Command) Response {
			return NewSuccessResponse("test response", nil)
		})
		done <- err
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Should be running now
	if !transport.IsRunning() {
		t.Error("Transport should be running after Listen() starts")
	}

	// Clean shutdown
	transport.Close()

	// Wait for daemon to finish
	select {
	case err := <-done:
		// Accept "use of closed network connection" as expected during shutdown
		if err != nil && !contains(err.Error(), "use of closed network connection") {
			t.Errorf("Unexpected daemon error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Daemon didn't shut down within timeout")
	}

	// Should not be running after close
	time.Sleep(10 * time.Millisecond) // Brief pause for cleanup
	if transport.IsRunning() {
		t.Error("Transport should not be running after Close()")
	}
}

func TestUnixSocketTransport_BasicCommandFlow(t *testing.T) {
	transport := NewUnixSocketTransport()
	defer transport.Close()

	// Start daemon that echoes commands back
	daemonDone := make(chan error)

	go func() {
		err := transport.Listen(func(cmd Command) Response {
			return NewSuccessResponse(fmt.Sprintf("received %s command", cmd.Type), cmd)
		})
		daemonDone <- err
	}()

	// Give daemon time to start listening
	time.Sleep(100 * time.Millisecond)

	// Verify daemon is running
	if !transport.IsRunning() {
		t.Fatal("Daemon should be running")
	}

	// Create client transport
	clientTransport := NewUnixSocketTransport()

	// Test sending a command
	cmd := NewPlayCommand()
	resp, err := clientTransport.Send(cmd)

	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected successful response, got %v", resp.Success)
	}

	expectedMessage := "received play command"
	if resp.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, resp.Message)
	}

	// Clean shutdown
	transport.Close()

	// Wait for daemon to finish
	select {
	case err := <-daemonDone:
		// Accept "use of closed network connection" as expected during shutdown
		if err != nil && !contains(err.Error(), "use of closed network connection") {
			t.Errorf("Unexpected daemon error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Daemon didn't shut down within timeout")
	}
}

func TestUnixSocketTransport_MultipleCommands(t *testing.T) {
	transport := NewUnixSocketTransport()
	defer transport.Close()

	// Track received commands
	var receivedCommands []CommandType
	var mu sync.Mutex

	// Start daemon
	daemonReady := make(chan bool)
	go func() {
		time.Sleep(50 * time.Millisecond)
		daemonReady <- true
	}()

	go func() {
		transport.Listen(func(cmd Command) Response {
			mu.Lock()
			receivedCommands = append(receivedCommands, cmd.Type)
			mu.Unlock()

			return NewSuccessResponse(fmt.Sprintf("processed %s", cmd.Type), nil)
		})
	}()

	<-daemonReady

	// Create client and send multiple commands
	client := NewUnixSocketTransport()

	commands := []Command{
		NewPlayCommand(),
		NewSkipCommand(3),
		NewPauseCommand(),
		NewStatusCommand(),
	}

	for _, cmd := range commands {
		resp, err := client.Send(cmd)
		if err != nil {
			t.Fatalf("Send(%s) error = %v", cmd.Type, err)
		}
		if !resp.Success {
			t.Errorf("Send(%s) failed: %s", cmd.Type, resp.Message)
		}
	}

	// Verify all commands were received
	mu.Lock()
	if len(receivedCommands) != len(commands) {
		t.Errorf("Expected %d commands, received %d", len(commands), len(receivedCommands))
	}

	for i, expected := range commands {
		if i >= len(receivedCommands) {
			t.Errorf("Missing command %d: %s", i, expected.Type)
			continue
		}
		if receivedCommands[i] != expected.Type {
			t.Errorf("Command %d: expected %s, got %s", i, expected.Type, receivedCommands[i])
		}
	}
	mu.Unlock()
}

func TestUnixSocketTransport_ConcurrentClients(t *testing.T) {
	transport := NewUnixSocketTransport()
	defer transport.Close()

	// Track client connections
	clientCount := 0
	var mu sync.Mutex

	// Start daemon
	daemonReady := make(chan bool)
	go func() {
		time.Sleep(50 * time.Millisecond)
		daemonReady <- true
	}()

	go func() {
		transport.Listen(func(cmd Command) Response {
			mu.Lock()
			clientCount++
			count := clientCount
			mu.Unlock()

			// Simulate some work
			time.Sleep(10 * time.Millisecond)

			return NewSuccessResponse(fmt.Sprintf("client %d processed %s", count, cmd.Type), nil)
		})
	}()

	<-daemonReady

	// Launch multiple concurrent clients
	numClients := 5
	var wg sync.WaitGroup
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			client := NewUnixSocketTransport()
			cmd := NewSkipCommand(clientID + 1)

			resp, err := client.Send(cmd)
			if err != nil {
				errors <- fmt.Errorf("client %d error: %v", clientID, err)
				return
			}

			if !resp.Success {
				errors <- fmt.Errorf("client %d failed: %s", clientID, resp.Message)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}

	// Verify all clients were handled
	mu.Lock()
	if clientCount != numClients {
		t.Errorf("Expected %d clients, handled %d", numClients, clientCount)
	}
	mu.Unlock()
}

func TestUnixSocketTransport_SendToNonExistentDaemon(t *testing.T) {
	// Create client without starting daemon
	client := NewUnixSocketTransport()

	cmd := NewPlayCommand()
	_, err := client.Send(cmd)

	if err == nil {
		t.Error("Expected error when sending to non-existent daemon")
	}

	// Error should mention daemon not running
	if !contains(err.Error(), "daemon running") {
		t.Errorf("Error message should mention daemon not running, got: %s", err.Error())
	}
}

func TestUnixSocketTransport_InvalidJSON(t *testing.T) {
	transport := NewUnixSocketTransport()
	defer transport.Close()

	// Start daemon
	daemonReady := make(chan bool)
	go func() {
		time.Sleep(50 * time.Millisecond)
		daemonReady <- true
	}()

	go func() {
		transport.Listen(func(cmd Command) Response {
			return NewSuccessResponse("should not reach here", nil)
		})
	}()

	<-daemonReady

	// We can't easily send invalid JSON through our Send() method,
	// but we can test the error handling by using a custom connection
	// This would be more of an integration test in practice

	// For now, just verify the daemon is running
	client := NewUnixSocketTransport()
	if !client.IsRunning() {
		t.Error("Daemon should be running")
	}
}

func TestUnixSocketTransport_SocketCleanup(t *testing.T) {
	transport := NewUnixSocketTransport()
	socketPath := transport.GetSocketPath()

	// Start and immediately stop daemon
	done := make(chan bool)
	go func() {
		defer func() { done <- true }()

		// This should create the socket file
		transport.Listen(func(cmd Command) Response {
			return NewSuccessResponse("test", nil)
		})
	}()

	// Give it time to create socket
	time.Sleep(50 * time.Millisecond)

	// Socket file should exist
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("Socket file should exist after Listen() starts")
	}

	// Close daemon
	transport.Close()

	// Wait for daemon to finish
	select {
	case <-done:
		// Good
	case <-time.After(1 * time.Second):
		t.Error("Daemon didn't shut down within timeout")
	}

	// Socket file should be cleaned up
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("Socket file should be cleaned up after Close()")
	}
}

func TestUnixSocketTransport_ComplexData(t *testing.T) {
	transport := NewUnixSocketTransport()
	defer transport.Close()

	// Start daemon that returns complex data
	daemonReady := make(chan bool)
	go func() {
		time.Sleep(50 * time.Millisecond)
		daemonReady <- true
	}()

	go func() {
		transport.Listen(func(cmd Command) Response {
			trackInfo := TrackInfo{
				Filename:    "complex_track.mp3",
				Path:        "/path/to/music",
				Duration:    "4:20",
				Position:    "2:10",
				TrackNumber: 3,
				TotalTracks: 15,
				Source:      "test_playlist",
			}

			return NewSuccessResponse("status info", trackInfo)
		})
	}()

	<-daemonReady

	// Send status command
	client := NewUnixSocketTransport()
	cmd := NewStatusCommand()

	resp, err := client.Send(cmd)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected successful response, got %v", resp.Success)
	}

	// Data will be a map[string]interface{} after JSON round-trip
	if resp.Data == nil {
		t.Error("Expected data in response")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMain can be used for setup/teardown if needed
func TestMain(m *testing.M) {
	// Clean up any leftover socket files before running tests
	if homeDir, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(homeDir, ".config", "auxbox")
		socketPath := filepath.Join(configDir, "auxbox.sock")
		os.Remove(socketPath)
	}

	// Run tests
	code := m.Run()

	// Clean up after tests
	if homeDir, err := os.UserHomeDir(); err == nil {
		configDir := filepath.Join(homeDir, ".config", "auxbox")
		socketPath := filepath.Join(configDir, "auxbox.sock")
		os.Remove(socketPath)
	}

	os.Exit(code)
}
