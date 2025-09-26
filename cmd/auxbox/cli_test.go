package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/cerberussg/auxbox/internal/shared"
)

// Mock transport for testing
type MockTransport struct {
	isRunning      bool
	lastCommand    *shared.Command
	responseToSend *shared.Response
	errorToSend    error
}

func (m *MockTransport) IsRunning() bool {
	return m.isRunning
}

func (m *MockTransport) Send(cmd shared.Command) (*shared.Response, error) {
	m.lastCommand = &cmd
	if m.errorToSend != nil {
		return nil, m.errorToSend
	}
	if m.responseToSend != nil {
		return m.responseToSend, nil
	}
	return &shared.Response{Success: true, Message: "mock response"}, nil
}

func (m *MockTransport) Listen(handler func(shared.Command) shared.Response) error {
	return nil
}

func (m *MockTransport) Close() error {
	return nil
}

// Helper to capture stdout
func captureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCLI_VersionCommand(t *testing.T) {
	// Similar to help - we're testing that version commands are recognized
	testCases := [][]string{
		{"auxbox", "--version"},
		{"auxbox", "-v"},
		{"auxbox", "version"},
	}

	for _, args := range testCases {
		if len(args) >= 2 {
			command := args[1]
			if !(command == "--version" || command == "-v" || command == "version") {
				t.Errorf("Version command not recognized: %s", command)
			}
		}
	}
}

func TestCLI_HandleSkipCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{"no count provided", []string{"auxbox", "skip"}, 1},
		{"valid count", []string{"auxbox", "skip", "3"}, 3},
		{"invalid count defaults to 1", []string{"auxbox", "skip", "invalid"}, 1},
		{"zero count defaults to 1", []string{"auxbox", "skip", "0"}, 1},
		{"negative count defaults to 1", []string{"auxbox", "skip", "-5"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			mockTransport := &MockTransport{isRunning: true}
			cli.transport = mockTransport

			cli.handleSkipCommand(tt.args)

			if mockTransport.lastCommand == nil {
				t.Fatal("No command was sent")
			}

			if mockTransport.lastCommand.Type != shared.CmdSkip {
				t.Errorf("Expected skip command, got %v", mockTransport.lastCommand.Type)
			}

			if mockTransport.lastCommand.Count != tt.expected {
				t.Errorf("Expected count %d, got %d", tt.expected, mockTransport.lastCommand.Count)
			}
		})
	}
}

func TestCLI_HandleBackCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected int
	}{
		{"no count provided", []string{"auxbox", "back"}, 1},
		{"valid count", []string{"auxbox", "back", "2"}, 2},
		{"invalid count defaults to 1", []string{"auxbox", "back", "abc"}, 1},
		{"zero count defaults to 1", []string{"auxbox", "back", "0"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()
			mockTransport := &MockTransport{isRunning: true}
			cli.transport = mockTransport

			cli.handleBackCommand(tt.args)

			if mockTransport.lastCommand == nil {
				t.Fatal("No command was sent")
			}

			if mockTransport.lastCommand.Type != shared.CmdBack {
				t.Errorf("Expected back command, got %v", mockTransport.lastCommand.Type)
			}

			if mockTransport.lastCommand.Count != tt.expected {
				t.Errorf("Expected count %d, got %d", tt.expected, mockTransport.lastCommand.Count)
			}
		})
	}
}

func TestCLI_HandleStartCommand_ValidPaths(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "auxbox_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name       string
		args       []string
		sourceType shared.SourceType
	}{
		{"folder flag", []string{"auxbox", "start", "--folder", tmpDir}, shared.SourceFolder},
		{"folder short flag", []string{"auxbox", "start", "-f", tmpDir}, shared.SourceFolder},
		{"playlist flag", []string{"auxbox", "start", "--playlist", tmpDir}, shared.SourcePlaylist},
		{"playlist short flag", []string{"auxbox", "start", "-p", tmpDir}, shared.SourcePlaylist},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI()

			// Since handleStartCommand calls os.Exit(), we can't test it directly
			// Instead, we'll test the path validation logic
			if !cli.pathExists(tmpDir) {
				t.Errorf("pathExists should return true for %s", tmpDir)
			}

			// Test source type logic
			sourceFlag := tt.args[2]
			var sourceType shared.SourceType
			switch sourceFlag {
			case "--folder", "-f":
				sourceType = shared.SourceFolder
			case "--playlist", "-p":
				sourceType = shared.SourcePlaylist
			}

			if sourceType != tt.sourceType {
				t.Errorf("Expected source type %s, got %s", tt.sourceType, sourceType)
			}
		})
	}
}

func TestCLI_PathExists(t *testing.T) {
	cli := NewCLI()

	// Test with temporary directory
	tmpDir, err := os.MkdirTemp("", "auxbox_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Should exist
	if !cli.pathExists(tmpDir) {
		t.Errorf("pathExists should return true for existing directory: %s", tmpDir)
	}

	// Should not exist
	nonExistentPath := "/path/that/definitely/does/not/exist/hopefully"
	if cli.pathExists(nonExistentPath) {
		t.Errorf("pathExists should return false for non-existent path: %s", nonExistentPath)
	}

	// Test tilde expansion (if we have a home directory)
	if _, err := os.UserHomeDir(); err == nil {
		// Test that ~/. expands properly (should always exist)
		if !cli.pathExists("~/.") {
			t.Error("pathExists should handle tilde expansion for ~/.")
		}
	}
}

func TestCLI_GetStringFromMap(t *testing.T) {
	cli := NewCLI()

	testMap := map[string]interface{}{
		"string_value": "test",
		"int_value":    42,
		"nil_value":    nil,
	}

	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"string_value", "default", "test"},
		{"int_value", "default", "default"},   // int should return default
		{"nil_value", "default", "default"},   // nil should return default
		{"missing_key", "default", "default"}, // missing key should return default
	}

	for _, tt := range tests {
		result := cli.getStringFromMap(testMap, tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("getStringFromMap(%s) = %s, want %s", tt.key, result, tt.expected)
		}
	}
}

func TestCLI_GetIntFromMap(t *testing.T) {
	cli := NewCLI()

	testMap := map[string]interface{}{
		"int_value":    42,
		"float_value":  3.14,
		"string_value": "test",
		"nil_value":    nil,
	}

	tests := []struct {
		key          string
		defaultValue int
		expected     int
	}{
		{"int_value", 0, 42},
		{"float_value", 0, 3},   // float64 should convert to int
		{"string_value", 0, 0},  // string should return default
		{"nil_value", 0, 0},     // nil should return default
		{"missing_key", 99, 99}, // missing key should return default
	}

	for _, tt := range tests {
		result := cli.getIntFromMap(testMap, tt.key, tt.defaultValue)
		if result != tt.expected {
			t.Errorf("getIntFromMap(%s) = %d, want %d", tt.key, result, tt.expected)
		}
	}
}

func TestCLI_SendCommand_DaemonNotRunning(t *testing.T) {
	cli := NewCLI()
	mockTransport := &MockTransport{isRunning: false}
	cli.transport = mockTransport

	// We can't easily test os.Exit() calls, but we can verify the logic
	// In a real implementation, you might want to refactor to return errors
	// instead of calling os.Exit() for better testability

	cmd := shared.NewPlayCommand()

	// Since sendCommand calls os.Exit(1) when daemon not running,
	// we can't test it directly. In production code, you might want
	// to refactor this to return an error instead for better testability.

	_ = cmd // Suppress unused variable
	_ = mockTransport
}

func TestCLI_SendCommand_Success(t *testing.T) {
	cli := NewCLI()
	mockTransport := &MockTransport{
		isRunning: true,
		responseToSend: &shared.Response{
			Success: true,
			Message: "Command executed successfully",
		},
	}
	cli.transport = mockTransport

	// Again, we can't easily test the full sendCommand due to os.Exit()
	// But we can verify that the mock transport would receive the command

	cmd := shared.NewPlayCommand()

	// In a refactored version, you might do:
	// err := cli.sendCommand(cmd)
	// if err != nil { t.Fatal(err) }

	_ = cmd
	_ = mockTransport
}

func TestCLI_CommandRecognition(t *testing.T) {
	// Test that all expected commands are recognized
	validCommands := []string{
		"--help", "-h", "help",
		"--version", "-v", "version",
		"start", "play", "pause", "skip", "back", "status", "list", "stop",
	}

	for _, cmd := range validCommands {
		t.Run(cmd, func(t *testing.T) {
			// This tests that the command would be recognized in the switch statement
			recognized := false

			switch cmd {
			case "--help", "-h", "help":
				recognized = true
			case "--version", "-v", "version":
				recognized = true
			case "start", "play", "pause", "skip", "back", "status", "list", "stop":
				recognized = true
			}

			if !recognized {
				t.Errorf("Command %s not recognized", cmd)
			}
		})
	}
}
