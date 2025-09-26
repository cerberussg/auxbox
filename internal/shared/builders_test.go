package shared

import "testing"

func TestNewPlayCommand(t *testing.T) {
	cmd := NewPlayCommand()

	if cmd.Type != CmdPlay {
		t.Errorf("NewPlayCommand() Type = %v, want %v", cmd.Type, CmdPlay)
	}

	// Should have no other fields set
	if len(cmd.Args) != 0 {
		t.Errorf("NewPlayCommand() Args should be empty, got %v", cmd.Args)
	}
	if cmd.Count != 0 {
		t.Errorf("NewPlayCommand() Count should be 0, got %d", cmd.Count)
	}
	if cmd.Source != "" {
		t.Errorf("NewPlayCommand() Source should be empty, got %v", cmd.Source)
	}
	if cmd.Path != "" {
		t.Errorf("NewPlayCommand() Path should be empty, got %v", cmd.Path)
	}
}

func TestNewPauseCommand(t *testing.T) {
	cmd := NewPauseCommand()

	if cmd.Type != CmdPause {
		t.Errorf("NewPauseCommand() Type = %v, want %v", cmd.Type, CmdPause)
	}
}

func TestNewSkipCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		wantCount int
	}{
		{"positive count", 3, 3},
		{"count of 1", 1, 1},
		{"zero count defaults to 1", 0, 1},
		{"negative count defaults to 1", -5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewSkipCommand(tt.input)

			if cmd.Type != CmdSkip {
				t.Errorf("NewSkipCommand() Type = %v, want %v", cmd.Type, CmdSkip)
			}
			if cmd.Count != tt.wantCount {
				t.Errorf("NewSkipCommand(%d) Count = %d, want %d", tt.input, cmd.Count, tt.wantCount)
			}
		})
	}
}

func TestNewBackCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		wantCount int
	}{
		{"positive count", 2, 2},
		{"count of 1", 1, 1},
		{"zero count defaults to 1", 0, 1},
		{"negative count defaults to 1", -3, 1},
		{"large count", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewBackCommand(tt.input)

			if cmd.Type != CmdBack {
				t.Errorf("NewBackCommand() Type = %v, want %v", cmd.Type, CmdBack)
			}
			if cmd.Count != tt.wantCount {
				t.Errorf("NewBackCommand(%d) Count = %d, want %d", tt.input, cmd.Count, tt.wantCount)
			}
		})
	}
}

func TestNewStartCommand(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		path       string
	}{
		{"folder source", SourceFolder, "/home/user/music"},
		{"playlist source", SourcePlaylist, "/home/user/playlists/chill.m3u"},
		{"twitch source", SourceTwitch, "some-dj-stream"},
		{"discord source", SourceDiscord, "https://discord.com/channels/123/456"},
		{"empty path", SourceFolder, ""},
		{"relative path", SourceFolder, "./music"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewStartCommand(tt.sourceType, tt.path)

			if cmd.Type != CmdStart {
				t.Errorf("NewStartCommand() Type = %v, want %v", cmd.Type, CmdStart)
			}
			if cmd.Source != tt.sourceType {
				t.Errorf("NewStartCommand() Source = %v, want %v", cmd.Source, tt.sourceType)
			}
			if cmd.Path != tt.path {
				t.Errorf("NewStartCommand() Path = %v, want %v", cmd.Path, tt.path)
			}
		})
	}
}

func TestNewStatusCommand(t *testing.T) {
	cmd := NewStatusCommand()

	if cmd.Type != CmdStatus {
		t.Errorf("NewStatusCommand() Type = %v, want %v", cmd.Type, CmdStatus)
	}
}

func TestNewListCommand(t *testing.T) {
	cmd := NewListCommand()

	if cmd.Type != CmdList {
		t.Errorf("NewListCommand() Type = %v, want %v", cmd.Type, CmdList)
	}
}

func TestNewStopCommand(t *testing.T) {
	cmd := NewStopCommand()

	if cmd.Type != CmdStop {
		t.Errorf("NewStopCommand() Type = %v, want %v", cmd.Type, CmdStop)
	}
}

func TestNewSuccessResponse(t *testing.T) {
	tests := []struct {
		name    string
		message string
		data    interface{}
	}{
		{"simple success", "Operation completed", nil},
		{"success with string data", "Track info", "current_track.mp3"},
		{"success with struct data", "Status", TrackInfo{Filename: "test.mp3", Duration: "3:45"}},
		{"success with empty message", "", map[string]string{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewSuccessResponse(tt.message, tt.data)

			if !resp.Success {
				t.Errorf("NewSuccessResponse() Success = %v, want true", resp.Success)
			}
			if resp.Message != tt.message {
				t.Errorf("NewSuccessResponse() Message = %v, want %v", resp.Message, tt.message)
			}
			// Note: We're not deeply comparing Data since it's interface{} -
			// that's what the transport tests handle
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"simple error", "File not found"},
		{"detailed error", "Failed to connect to daemon: connection refused"},
		{"empty error message", ""},
		{"error with special chars", "Path contains invalid characters: /path/with/ünicode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.message)

			if resp.Success {
				t.Errorf("NewErrorResponse() Success = %v, want false", resp.Success)
			}
			if resp.Message != tt.message {
				t.Errorf("NewErrorResponse() Message = %v, want %v", resp.Message, tt.message)
			}
			if resp.Data != nil {
				t.Errorf("NewErrorResponse() Data should be nil, got %v", resp.Data)
			}
		})
	}
}

func TestBuilderConsistency(t *testing.T) {
	// Test that builders create commands that can be properly serialized/deserialized
	commands := []Command{
		NewPlayCommand(),
		NewPauseCommand(),
		NewSkipCommand(3),
		NewBackCommand(2),
		NewStartCommand(SourceFolder, "/test/path"),
		NewStatusCommand(),
		NewListCommand(),
		NewStopCommand(),
	}

	for i, cmd := range commands {
		t.Run(string(cmd.Type), func(t *testing.T) {
			// Test that the command can be serialized
			jsonData, err := cmd.ToJSON()
			if err != nil {
				t.Fatalf("Command %d failed to serialize: %v", i, err)
			}

			// Test that it can be deserialized back
			deserializedCmd, err := CommandFromJSON(jsonData)
			if err != nil {
				t.Fatalf("Command %d failed to deserialize: %v", i, err)
			}

			// Test that the type is preserved
			if deserializedCmd.Type != cmd.Type {
				t.Errorf("Command %d type changed during round-trip: got %v, want %v",
					i, deserializedCmd.Type, cmd.Type)
			}
		})
	}
}

func TestResponseBuilderConsistency(t *testing.T) {
	// Test that response builders create responses that can be properly serialized/deserialized
	responses := []Response{
		NewSuccessResponse("test message", nil),
		NewSuccessResponse("with data", TrackInfo{Filename: "test.mp3"}),
		NewErrorResponse("error message"),
		NewErrorResponse(""),
	}

	for i, resp := range responses {
		t.Run(resp.Message, func(t *testing.T) {
			// Test that the response can be serialized
			jsonData, err := resp.ToJSON()
			if err != nil {
				t.Fatalf("Response %d failed to serialize: %v", i, err)
			}

			// Test that it can be deserialized back
			deserializedResp, err := ResponseFromJSON(jsonData)
			if err != nil {
				t.Fatalf("Response %d failed to deserialize: %v", i, err)
			}

			// Test that success flag is preserved
			if deserializedResp.Success != resp.Success {
				t.Errorf("Response %d success flag changed during round-trip: got %v, want %v",
					i, deserializedResp.Success, resp.Success)
			}

			// Test that message is preserved
			if deserializedResp.Message != resp.Message {
				t.Errorf("Response %d message changed during round-trip: got %v, want %v",
					i, deserializedResp.Message, resp.Message)
			}
		})
	}
}
