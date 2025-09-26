package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cerberussg/auxbox/internal/shared"
)

const (
	version = "0.1.0"
	usage   = `auxbox - CLI music player for background listening

Usage:
  auxbox start --folder <path>     Start daemon with folder source
  auxbox start --playlist <path>   Start daemon with playlist source
  auxbox play                      Start/resume playback
  auxbox pause                     Pause playback
  auxbox skip [n]                  Skip forward n tracks (default: 1)
  auxbox back [n]                  Skip backward n tracks (default: 1)
  auxbox status                    Show current track info
  auxbox list                      List tracks in current queue
  auxbox stop                      Stop daemon
  auxbox --help, -h                Show this help
  auxbox --version, -v             Show version

Examples:
  auxbox start --folder ~/Downloads/new-pack/
  auxbox skip 3
  auxbox back
  auxbox status`
)

// CLI handles command line interface logic
type CLI struct {
	transport shared.Transport
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		transport: shared.NewUnixSocketTransport(),
	}
}

// Run processes command line arguments and executes commands
func (c *CLI) Run(args []string) {
	if len(args) < 2 {
		fmt.Println("No command provided. Use 'auxbox --help' for usage.")
		os.Exit(1)
	}

	command := args[1]

	switch command {
	case "--help", "-h", "help":
		fmt.Println(usage)
		os.Exit(0)
	case "--version", "-v", "version":
		fmt.Printf("auxbox %s\n", version)
		os.Exit(0)
	case "start":
		c.handleStartCommand(args)
	case "play":
		c.sendCommand(shared.NewPlayCommand())
	case "pause":
		c.sendCommand(shared.NewPauseCommand())
	case "skip":
		c.handleSkipCommand(args)
	case "back":
		c.handleBackCommand(args)
	case "status":
		c.sendCommand(shared.NewStatusCommand())
	case "list":
		c.sendCommand(shared.NewListCommand())
	case "stop":
		c.sendCommand(shared.NewStopCommand())
	default:
		fmt.Printf("Unknown command: %s\nUse 'auxbox --help' for usage.\n", command)
		os.Exit(1)
	}
}

func (c *CLI) handleStartCommand(args []string) {
	if len(args) < 4 {
		fmt.Println("Start command requires source type and path.")
		fmt.Println("Usage: auxbox start --folder <path> | --playlist <path>")
		os.Exit(1)
	}

	sourceFlag := args[2]
	sourcePath := args[3]

	var sourceType shared.SourceType
	switch sourceFlag {
	case "--folder", "-f":
		sourceType = shared.SourceFolder
	case "--playlist", "-p":
		sourceType = shared.SourcePlaylist
	default:
		fmt.Printf("Unknown source type: %s\n", sourceFlag)
		fmt.Println("Use --folder or --playlist")
		os.Exit(1)
	}

	// Validate path exists
	if !c.pathExists(sourcePath) {
		fmt.Printf("Path does not exist: %s\n", sourcePath)
		os.Exit(1)
	}

	// Check if daemon is already running
	transport := shared.NewUnixSocketTransport()
	if transport.IsRunning() {
		fmt.Printf("auxbox daemon is already running.\n")
		fmt.Printf("Use 'auxbox stop' to stop the current daemon first.\n")
		os.Exit(1)
	}

	// Start daemon
	cmd := shared.NewStartCommand(sourceType, sourcePath)
	fmt.Printf("Starting auxbox daemon with %s: %s\n", sourceType, sourcePath)

	// TODO: This will actually start the daemon process
	// For now, just show what we would do
	fmt.Printf("Would start daemon with command: %+v\n", cmd)
	fmt.Printf("Daemon functionality coming in next phase!\n")
}

func (c *CLI) handleSkipCommand(args []string) {
	count := 1 // default

	if len(args) > 2 {
		if parsed, err := strconv.Atoi(args[2]); err == nil && parsed > 0 {
			count = parsed
		} else {
			fmt.Printf("Invalid skip count: %s (using default: 1)\n", args[2])
		}
	}

	c.sendCommand(shared.NewSkipCommand(count))
}

func (c *CLI) handleBackCommand(args []string) {
	count := 1 // default

	if len(args) > 2 {
		if parsed, err := strconv.Atoi(args[2]); err == nil && parsed > 0 {
			count = parsed
		} else {
			fmt.Printf("Invalid back count: %s (using default: 1)\n", args[2])
		}
	}

	c.sendCommand(shared.NewBackCommand(count))
}

func (c *CLI) sendCommand(cmd shared.Command) {
	// Check if daemon is running
	transport := shared.NewUnixSocketTransport()
	if !transport.IsRunning() {
		fmt.Printf("auxbox daemon is not running.\n")
		fmt.Printf("Start it with: auxbox start --folder <path>\n")
		os.Exit(1)
	}

	// Send command
	resp, err := transport.Send(cmd)
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		os.Exit(1)
	}

	// Handle response
	if !resp.Success {
		fmt.Printf("Command failed: %s\n", resp.Message)
		os.Exit(1)
	}

	// Print response based on command type
	switch cmd.Type {
	case shared.CmdStatus:
		c.printStatusResponse(resp)
	case shared.CmdList:
		c.printListResponse(resp)
	case shared.CmdStop:
		fmt.Println("auxbox daemon stopped.")
	default:
		if resp.Message != "" {
			fmt.Println(resp.Message)
		} else {
			fmt.Printf("Command %s executed successfully.\n", cmd.Type)
		}
	}
}

func (c *CLI) printStatusResponse(resp *shared.Response) {
	if resp.Data == nil {
		fmt.Println("No track currently playing.")
		return
	}

	// Data comes back as map[string]interface{} after JSON round-trip
	if dataMap, ok := resp.Data.(map[string]interface{}); ok {
		filename := c.getStringFromMap(dataMap, "filename", "Unknown")
		duration := c.getStringFromMap(dataMap, "duration", "")
		position := c.getStringFromMap(dataMap, "position", "")
		trackNum := c.getIntFromMap(dataMap, "track_number", 0)
		totalTracks := c.getIntFromMap(dataMap, "total_tracks", 0)
		source := c.getStringFromMap(dataMap, "source", "")

		// Build status line: "▶ filename | position/duration | Track N/total | Source: path"
		status := fmt.Sprintf("▶ %s", filename)

		if position != "" && duration != "" {
			status += fmt.Sprintf(" | %s/%s", position, duration)
		} else if duration != "" {
			status += fmt.Sprintf(" | %s", duration)
		}

		if trackNum > 0 && totalTracks > 0 {
			status += fmt.Sprintf(" | Track %d/%d", trackNum, totalTracks)
		}

		if source != "" {
			status += fmt.Sprintf(" | Source: %s", source)
		}

		fmt.Println(status)
	} else {
		fmt.Printf("Status: %s\n", resp.Message)
	}
}

func (c *CLI) printListResponse(resp *shared.Response) {
	if resp.Data == nil {
		fmt.Println("No tracks available.")
		return
	}

	if dataMap, ok := resp.Data.(map[string]interface{}); ok {
		if tracksInterface, exists := dataMap["tracks"]; exists {
			if tracks, ok := tracksInterface.([]interface{}); ok {
				fmt.Printf("Tracks (%d total):\n", len(tracks))
				for i, trackInterface := range tracks {
					if track, ok := trackInterface.(string); ok {
						marker := "  "
						if currentIdx := c.getIntFromMap(dataMap, "current_idx", -1); currentIdx == i {
							marker = "▶ "
						}
						fmt.Printf("%s%d. %s\n", marker, i+1, track)
					}
				}
			}
		}
	} else {
		fmt.Printf("Tracks: %s\n", resp.Message)
	}
}

// Helper methods

func (c *CLI) pathExists(path string) bool {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	_, err := os.Stat(path)
	return err == nil
}

func (c *CLI) getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *CLI) getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if val, exists := m[key]; exists {
		// JSON numbers come back as float64
		if f, ok := val.(float64); ok {
			return int(f)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultValue
}
