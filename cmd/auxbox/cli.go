package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cerberussg/auxbox/internal/server"
	"github.com/cerberussg/auxbox/internal/shared"
)

const (
	version = "0.1.0"
	usage   = `auxbox - CLI music player for background listening

Usage:
  auxbox play -f <path>            Load folder and play instantly
  auxbox play --folder <path>      Load folder and play instantly
  auxbox play -f <path> -s         Load folder, shuffle, and play
  auxbox play -p <path>            Load playlist and play instantly
  auxbox play --playlist <path>    Load playlist and play instantly
  auxbox play                      Resume playback (if paused)
  auxbox pause                     Pause playback
  auxbox stop                      Stop playback (reset to beginning)
  auxbox skip [n]                  Skip forward n tracks (default: 1)
  auxbox back [n]                  Skip backward n tracks (default: 1)
  auxbox shuffle                   Toggle shuffle on/off
  auxbox volume [0-100]            Show or set volume percentage
  auxbox status                    Show current track info
  auxbox list                      List tracks in current queue
  auxbox exit                      Exit daemon (stop everything)
  auxbox --help, -h                Show this help
  auxbox --version, -v             Show version

Examples:
  auxbox play -f ~/Downloads/new-pack/     # Instant music from folder
  auxbox play -f ~/jazz -s                 # Load folder, shuffle, and play
  auxbox play -p ~/playlists/workout.m3u  # Switch to playlist while playing
  auxbox shuffle                           # Toggle shuffle on current playlist
  auxbox skip 3
  auxbox volume 75
  auxbox pause
  auxbox exit`
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
	case "_daemon":
		// Internal daemon process - parse source type and path
		if len(args) < 4 {
			log.Fatal("Daemon process requires source type and path")
		}
		sourceType := shared.SourceType(args[2])
		sourcePath := args[3]
		c.runDaemonProcess(sourceType, sourcePath)
	case "play":
		c.handlePlayCommand(args)
	case "pause":
		c.sendCommand(shared.NewPauseCommand())
	case "skip":
		c.handleSkipCommand(args)
	case "back":
		c.handleBackCommand(args)
	case "shuffle":
		c.sendCommand(shared.Command{Type: shared.CmdShuffle})
	case "status":
		c.sendCommand(shared.NewStatusCommand())
	case "list":
		c.sendCommand(shared.NewListCommand())
	case "stop":
		c.sendCommand(shared.NewStopCommand())
	case "volume":
		c.handleVolumeCommand(args)
	case "exit":
		c.sendCommand(shared.NewExitCommand())
	default:
		fmt.Printf("Unknown command: %s\nUse 'auxbox --help' for usage.\n", command)
		os.Exit(1)
	}
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

func (c *CLI) handlePlayCommand(args []string) {
	// If no additional args, just play/resume current playlist
	if len(args) <= 2 {
		c.sendCommand(shared.NewPlayCommand())
		return
	}

	// Parse source flags
	sourceFlag := args[2]
	if len(args) < 4 {
		fmt.Printf("Source flag %s requires a path.\n", sourceFlag)
		fmt.Println("Usage: auxbox play -f <folder> | -p <playlist> [-s]")
		os.Exit(1)
	}

	sourcePath := args[3]

	var sourceType shared.SourceType
	switch sourceFlag {
	case "-f", "--folder":
		sourceType = shared.SourceFolder
	case "-p", "--playlist":
		sourceType = shared.SourcePlaylist
	default:
		fmt.Printf("Unknown source flag: %s\n", sourceFlag)
		fmt.Println("Use -f/--folder or -p/--playlist")
		os.Exit(1)
	}

	// Check for shuffle flag
	shuffle := false
	if len(args) >= 5 {
		for i := 4; i < len(args); i++ {
			if args[i] == "-s" || args[i] == "--shuffle" {
				shuffle = true
				break
			}
		}
	}

	// Validate path exists
	if !c.pathExists(sourcePath) {
		fmt.Printf("Path does not exist: %s\n", sourcePath)
		os.Exit(1)
	}

	// Check if daemon is running
	transport := shared.NewUnixSocketTransport()
	if !transport.IsRunning() {
		// Auto-start daemon with the provided source and begin playback
		fmt.Printf("Starting auxbox daemon with %s: %s\n", sourceType, sourcePath)
		c.startDaemonAndPlay(sourceType, sourcePath, shuffle)
		return
	}

	// Daemon is running - send hot-swap command
	cmd := shared.NewPlayCommand()
	cmd.Source = sourceType
	cmd.Path = sourcePath
	cmd.Shuffle = shuffle
	c.sendCommand(cmd)
}

func (c *CLI) sendCommand(cmd shared.Command) {
	// Check if daemon is running
	transport := shared.NewUnixSocketTransport()
	if !transport.IsRunning() {
		fmt.Printf("auxbox daemon is not running.\n")
		fmt.Printf("Start it with: auxbox play -f <path>\n")
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
	case shared.CmdVolume:
		c.printVolumeResponse(resp)
	case shared.CmdStop:
		fmt.Println("Playback stopped.")
	case shared.CmdExit:
		fmt.Println("auxbox daemon exited.")
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

func (c *CLI) printVolumeResponse(resp *shared.Response) {
	if resp.Data == nil {
		// Just print the message (e.g., "Volume set to 75%")
		fmt.Println(resp.Message)
		return
	}

	// Data might contain current volume info
	if dataMap, ok := resp.Data.(map[string]interface{}); ok {
		if volumeInterface, exists := dataMap["volume"]; exists {
			if volumeFloat, ok := volumeInterface.(float64); ok {
				volumePercent := int(volumeFloat * 100)
				fmt.Printf("Volume: %d%%\n", volumePercent)
				return
			}
		}
	}

	// Fallback to message
	fmt.Println(resp.Message)
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

func (c *CLI) handleVolumeCommand(args []string) {
	// If no volume specified, get current volume
	if len(args) <= 2 {
		c.sendCommand(shared.NewVolumeCommand(-1)) // -1 means "get current volume"
		return
	}

	// Parse volume percentage
	volumeStr := args[2]
	volume, err := strconv.Atoi(volumeStr)

	if err != nil {
		fmt.Printf("Invalid volume: %s. Use a number from 0-100.\n", volumeStr)
		return
	}

	if volume < 0 || volume > 100 {
		fmt.Printf("Volume must be between 0-100, got %d.\n", volume)
		return
	}

	c.sendCommand(shared.NewVolumeCommand(volume))
}

// startDaemonAndPlay starts the daemon and immediately begins playback
func (c *CLI) startDaemonAndPlay(sourceType shared.SourceType, sourcePath string, shuffle bool) {
	// Check if we're being called as the daemon itself
	if len(os.Args) >= 3 && os.Args[1] == "_daemon" {
		c.runDaemonProcess(sourceType, sourcePath)
		return
	}

	// Spawn daemon as background process
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Start daemon process with special flag
	cmd := exec.Command(executable, "_daemon", string(sourceType), sourcePath)

	// Redirect stdout/stderr to avoid output mixing
	cmd.Stdout = nil
	cmd.Stderr = nil

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	// Give daemon a moment to start
	time.Sleep(100 * time.Millisecond)

	// Verify daemon is running
	transport := shared.NewUnixSocketTransport()
	if !transport.IsRunning() {
		fmt.Println("Failed to start daemon - not responding")
		os.Exit(1)
	}

	// Send play command with source info to load and play immediately
	playCmd := shared.NewPlayCommand()
	playCmd.Source = sourceType
	playCmd.Path = sourcePath
	playCmd.Shuffle = shuffle
	resp, err := transport.Send(playCmd)
	if err != nil {
		fmt.Printf("Failed to initialize daemon: %v\n", err)
		os.Exit(1)
	}

	if !resp.Success {
		fmt.Printf("Failed to load source and start playback: %s\n", resp.Message)
		os.Exit(1)
	}

	fmt.Printf("✓ %s\n", resp.Message)
}

// runDaemonProcess runs the actual daemon server (called by background process)
func (c *CLI) runDaemonProcess(sourceType shared.SourceType, sourcePath string) {
	server := server.NewServer()

	// Load tracks based on source type and path before starting server
	log.Printf("Loading tracks from %s (type: %s)", sourcePath, sourceType)

	switch sourceType {
	case shared.SourceFolder:
		if err := server.LoadFolder(sourcePath); err != nil {
			log.Printf("Failed to load tracks from folder: %v", err)
			os.Exit(1)
		}
	case shared.SourcePlaylist:
		if err := server.LoadPlaylist(sourcePath); err != nil {
			log.Printf("Failed to load tracks from playlist: %v", err)
			os.Exit(1)
		}
	default:
		log.Printf("Unsupported source type: %s", sourceType)
		os.Exit(1)
	}

	// Start the server (this will block)
	if err := server.Start(); err != nil {
		log.Printf("Daemon error: %v", err)
		os.Exit(1)
	}
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
