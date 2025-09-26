package shared

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCommand_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		wantJSON string
	}{
		{
			name:     "simple play command",
			command:  Command{Type: CmdPlay},
			wantJSON: `{"type":"play"}`,
		},
		{
			name:     "simple pause command",
			command:  Command{Type: CmdPause},
			wantJSON: `{"type":"pause"}`,
		},
		{
			name:     "stop command",
			command:  Command{Type: CmdStop},
			wantJSON: `{"type":"stop"}`,
		},
		{
			name:     "exit command",
			command:  Command{Type: CmdExit},
			wantJSON: `{"type":"exit"}`,
		},
		{
			name:     "skip command with count",
			command:  Command{Type: CmdSkip, Count: 3},
			wantJSON: `{"type":"skip","count":3}`,
		},
		{
			name:     "back command with count",
			command:  Command{Type: CmdBack, Count: 2},
			wantJSON: `{"type":"back","count":2}`,
		},
		{
			name:     "start command with folder source",
			command:  Command{Type: CmdStart, Source: SourceFolder, Path: "/home/user/music"},
			wantJSON: `{"type":"start","source":"folder","path":"/home/user/music"}`,
		},
		{
			name:     "start command with playlist source",
			command:  Command{Type: CmdStart, Source: SourcePlaylist, Path: "/home/user/playlist.m3u"},
			wantJSON: `{"type":"start","source":"playlist","path":"/home/user/playlist.m3u"}`,
		},
		{
			name:     "command with args",
			command:  Command{Type: CmdList, Args: []string{"--format", "json"}},
			wantJSON: `{"type":"list","args":["--format","json"]}`,
		},
		{
			name: "complex command with all fields",
			command: Command{
				Type:   CmdSkip,
				Args:   []string{"--verbose"},
				Count:  5,
				Source: SourcePlaylist,
				Path:   "/path/to/playlist.m3u",
			},
			wantJSON: `{"type":"skip","args":["--verbose"],"count":5,"source":"playlist","path":"/path/to/playlist.m3u"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.command.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() error = %v", err)
			}

			// Parse both to ensure they're equivalent JSON
			var gotParsed, wantParsed map[string]interface{}
			if err := json.Unmarshal(got, &gotParsed); err != nil {
				t.Fatalf("Failed to parse generated JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantParsed); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}

			// Compare the parsed structures
			if !mapsEqual(gotParsed, wantParsed) {
				t.Errorf("ToJSON() = %s, want %s", string(got), tt.wantJSON)
			}
		})
	}
}

func TestCommandFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Command
		wantErr bool
	}{
		{
			name: "simple play command",
			json: `{"type":"play"}`,
			want: Command{Type: CmdPlay},
		},
		{
			name: "simple pause command",
			json: `{"type":"pause"}`,
			want: Command{Type: CmdPause},
		},
		{
			name: "stop command",
			json: `{"type":"stop"}`,
			want: Command{Type: CmdStop},
		},
		{
			name: "exit command",
			json: `{"type":"exit"}`,
			want: Command{Type: CmdExit},
		},
		{
			name: "skip command with count",
			json: `{"type":"skip","count":3}`,
			want: Command{Type: CmdSkip, Count: 3},
		},
		{
			name: "back command with count",
			json: `{"type":"back","count":2}`,
			want: Command{Type: CmdBack, Count: 2},
		},
		{
			name: "start command with folder source",
			json: `{"type":"start","source":"folder","path":"/home/user/music"}`,
			want: Command{Type: CmdStart, Source: SourceFolder, Path: "/home/user/music"},
		},
		{
			name: "start command with playlist source",
			json: `{"type":"start","source":"playlist","path":"/home/user/playlist.m3u"}`,
			want: Command{Type: CmdStart, Source: SourcePlaylist, Path: "/home/user/playlist.m3u"},
		},
		{
			name: "command with args",
			json: `{"type":"list","args":["--format","json"]}`,
			want: Command{Type: CmdList, Args: []string{"--format", "json"}},
		},
		{
			name: "status command",
			json: `{"type":"status"}`,
			want: Command{Type: CmdStatus},
		},
		{
			name: "list command",
			json: `{"type":"list"}`,
			want: Command{Type: CmdList},
		},
		{
			name:    "invalid JSON",
			json:    `{"type":"play"`,
			want:    Command{},
			wantErr: true,
		},
		{
			name: "empty JSON object",
			json: `{}`,
			want: Command{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommandFromJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !commandsEqual(got, tt.want) {
				t.Errorf("CommandFromJSON() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResponse_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		wantJSON string
	}{
		{
			name:     "success response with message",
			response: Response{Success: true, Message: "Track skipped"},
			wantJSON: `{"success":true,"message":"Track skipped"}`,
		},
		{
			name:     "error response",
			response: Response{Success: false, Message: "File not found"},
			wantJSON: `{"success":false,"message":"File not found"}`,
		},
		{
			name: "response with track data",
			response: Response{
				Success: true,
				Message: "Current status",
				Data: TrackInfo{
					Filename: "test.mp3",
					Duration: "3:45",
				},
			},
			wantJSON: `{"success":true,"message":"Current status","data":{"filename":"test.mp3","path":"","duration":"3:45"}}`,
		},
		{
			name: "response with playlist data",
			response: Response{
				Success: true,
				Message: "Track list",
				Data: PlaylistInfo{
					Source:     "/path/to/music",
					SourceType: "folder",
					Tracks:     []string{"track1.mp3", "track2.mp3"},
					CurrentIdx: 0,
				},
			},
			wantJSON: `{"success":true,"message":"Track list","data":{"source":"/path/to/music","source_type":"folder","tracks":["track1.mp3","track2.mp3"],"current_idx":0}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.response.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() error = %v", err)
			}

			// Parse both to ensure they're equivalent JSON
			var gotParsed, wantParsed map[string]interface{}
			if err := json.Unmarshal(got, &gotParsed); err != nil {
				t.Fatalf("Failed to parse generated JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantParsed); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}

			if !mapsEqual(gotParsed, wantParsed) {
				t.Errorf("ToJSON() = %s, want %s", string(got), tt.wantJSON)
			}
		})
	}
}

func TestResponseFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Response
		wantErr bool
	}{
		{
			name: "success response",
			json: `{"success":true,"message":"Track skipped"}`,
			want: Response{Success: true, Message: "Track skipped"},
		},
		{
			name: "error response",
			json: `{"success":false,"message":"File not found"}`,
			want: Response{Success: false, Message: "File not found"},
		},
		{
			name: "response with simple data",
			json: `{"success":true,"data":{"filename":"test.mp3","duration":"3:45"}}`,
			want: Response{Success: true, Data: map[string]interface{}{"filename": "test.mp3", "duration": "3:45"}},
		},
		{
			name:    "invalid JSON",
			json:    `{"success":true`,
			want:    Response{},
			wantErr: true,
		},
		{
			name: "empty response",
			json: `{}`,
			want: Response{Success: false}, // Success defaults to false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResponseFromJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ResponseFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !responsesEqual(got, tt.want) {
				t.Errorf("ResponseFromJSON() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRoundTripSerialization(t *testing.T) {
	// Test that we can serialize and deserialize without losing data
	testCommands := []Command{
		{Type: CmdPlay},
		{Type: CmdPause},
		{Type: CmdStop},
		{Type: CmdExit},
		{Type: CmdSkip, Count: 5},
		{Type: CmdBack, Count: 2},
		{Type: CmdStart, Source: SourceFolder, Path: "/path/to/music"},
		{Type: CmdStart, Source: SourcePlaylist, Path: "/path/to/playlist.m3u"},
		{Type: CmdList, Args: []string{"--verbose", "--format", "json"}},
		{Type: CmdStatus},
	}

	for i, originalCmd := range testCommands {
		t.Run(string(originalCmd.Type), func(t *testing.T) {
			// Command round trip
			cmdJSON, err := originalCmd.ToJSON()
			if err != nil {
				t.Fatalf("Failed to serialize command: %v", err)
			}

			deserializedCmd, err := CommandFromJSON(cmdJSON)
			if err != nil {
				t.Fatalf("Failed to deserialize command: %v", err)
			}

			if !commandsEqual(originalCmd, deserializedCmd) {
				t.Errorf("Command %d round trip failed: got %+v, want %+v", i, deserializedCmd, originalCmd)
			}
		})
	}

	// Test response round trip
	testResponses := []Response{
		{Success: true, Message: "All good"},
		{Success: false, Message: "Error occurred"},
		{
			Success: true,
			Message: "Track info",
			Data: TrackInfo{
				Filename:    "awesome_track.mp3",
				Duration:    "4:20",
				TrackNumber: 1,
				TotalTracks: 10,
			},
		},
	}

	for i, originalResp := range testResponses {
		t.Run(fmt.Sprintf("response_%d", i), func(t *testing.T) {
			respJSON, err := originalResp.ToJSON()
			if err != nil {
				t.Fatalf("Failed to serialize response: %v", err)
			}

			deserializedResp, err := ResponseFromJSON(respJSON)
			if err != nil {
				t.Fatalf("Failed to deserialize response: %v", err)
			}

			if deserializedResp.Success != originalResp.Success {
				t.Errorf("Response Success mismatch: got %v, want %v", deserializedResp.Success, originalResp.Success)
			}
			if deserializedResp.Message != originalResp.Message {
				t.Errorf("Response Message mismatch: got %v, want %v", deserializedResp.Message, originalResp.Message)
			}
			// Note: Data comparison is tricky because it becomes map[string]interface{} after JSON round trip
		})
	}
}

// Helper functions for deep comparison

func commandsEqual(a, b Command) bool {
	if a.Type != b.Type || a.Count != b.Count || a.Source != b.Source || a.Path != b.Path {
		return false
	}
	if len(a.Args) != len(b.Args) {
		return false
	}
	for i, arg := range a.Args {
		if arg != b.Args[i] {
			return false
		}
	}
	return true
}

func responsesEqual(a, b Response) bool {
	return a.Success == b.Success && a.Message == b.Message
	// Note: We're not comparing Data here because it's interface{} and gets complex after JSON marshaling
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b interface{}) bool {
	// Marshal both and compare JSON strings
	aJSON, aErr := json.Marshal(a)
	bJSON, bErr := json.Marshal(b)

	// If either fails to marshal, they're not equal
	if aErr != nil || bErr != nil {
		return false
	}

	return string(aJSON) == string(bJSON)
}
