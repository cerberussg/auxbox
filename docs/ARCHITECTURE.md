# Architecture Documentation

> **📋 Note:** This document describes the current architecture based on code structure and established patterns. Some implementation details (threading models, exact IPC protocol, performance metrics) are documented as designed patterns and may vary from actual implementation. When in doubt, refer to the source code in `internal/` packages.

Technical documentation for auxbox's design, implementation, and development practices.

## Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Communication Flow](#communication-flow)
- [Audio Pipeline](#audio-pipeline)
- [Dependencies](#dependencies)
- [Threading Model](#threading-model)
- [Development](#development)
- [Testing](#testing)
- [Building](#building)

## Overview

auxbox is a daemon-based CLI music player built in Go, designed for background listening with minimal user interaction. The architecture follows a client-server model where:

- **CLI client** sends commands via Unix sockets
- **Daemon server** manages audio playback in the background
- **Audio system** handles format decoding and playback
- **Playlist system** manages track ordering and navigation

## System Architecture

```
┌─────────────────┐
│   CLI Client    │  User runs: auxbox play -f ~/music
│  (cmd/auxbox)   │
└────────┬────────┘
         │ Unix Socket (IPC)
         ▼
┌─────────────────┐
│  Daemon Server  │  Background process
│ (internal/server)│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Audio System   │  Playback engine
│ (internal/audio)│
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  Audio Output   │  System audio (ALSA/CoreAudio/WASAPI)
│   (Beep v2)     │
└─────────────────┘
```

### Key Design Principles

1. **Single Daemon Instance** - One daemon per user session prevents conflicts
2. **Stateless CLI** - CLI commands are lightweight and exit immediately
3. **Background Persistence** - Music continues playing after CLI exits
4. **Non-blocking Commands** - All commands return quickly without waiting for audio
5. **Thread Safety** - Concurrent access to playback state is synchronized

## Project Structure

```
auxbox/
├── cmd/
│   └── auxbox/          # CLI entry point and command parsing
│       ├── main.go      # Application entry
│       ├── cli.go       # CLI command handling
│       └── ...
│
├── internal/
│   ├── audio/           # Audio playback system
│   │   ├── player.go    # Main player implementation
│   │   ├── system.go    # Audio system initialization
│   │   ├── volume.go    # Volume control
│   │   ├── position.go  # Position tracking
│   │   ├── types.go     # Audio types and interfaces
│   │   └── decoders/    # Format-specific decoders
│   │       ├── mp3.go
│   │       ├── wav.go
│   │       ├── aiff.go
│   │       └── registry.go
│   │
│   ├── client/          # Client-side daemon communication
│   │   └── client.go    # Unix socket client
│   │
│   ├── server/          # Daemon server implementation
│   │   ├── daemon.go    # Server lifecycle
│   │   ├── handler.go   # Command handling
│   │   └── commands/    # Individual command implementations
│   │
│   ├── playlist/        # Playlist management
│   │   ├── playlist.go  # Track list operations
│   │   ├── loader.go    # Source loading (folders/playlists)
│   │   └── shuffle.go   # Shuffle algorithm
│   │
│   └── shared/          # Shared utilities
│       ├── transport.go # Message serialization
│       ├── ipc.go       # IPC socket handling
│       ├── types.go     # Shared types
│       └── commands.go  # Command definitions
│
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
└── README.md            # Project documentation
```

## Core Components

### 1. CLI Client (`cmd/auxbox`)

**Purpose:** Parse user commands and send them to the daemon.

**Responsibilities:**
- Parse command-line arguments
- Check if daemon is running
- Start daemon if not running
- Send command via Unix socket
- Display response to user
- Exit immediately

**Key files:**
- `main.go` - Entry point
- `cli.go` - Command parsing and routing

### 2. Daemon Server (`internal/server`)

**Purpose:** Long-running background process managing audio playback.

**Responsibilities:**
- Listen on Unix socket for commands
- Manage audio player lifecycle
- Handle playback state (play/pause/stop)
- Process commands concurrently
- Track playback position
- Auto-advance to next track

**Key files:**
- `daemon.go` - Server initialization and lifecycle
- `handler.go` - Request routing and response generation
- `commands/` - Command implementations

### 3. Audio System (`internal/audio`)

**Purpose:** Handle audio file decoding and playback.

**Responsibilities:**
- Initialize audio output (speaker)
- Load and decode audio files
- Control playback (play/pause/stop)
- Track playback position in real-time
- Manage volume with smooth fading
- Handle multiple audio formats
- Clean up resources properly

**Key files:**
- `player.go` - Main player implementation
- `system.go` - Audio system initialization
- `volume.go` - Volume control and fading
- `position.go` - Position tracking
- `decoders/` - Format-specific decoders

### 4. Playlist System (`internal/playlist`)

**Purpose:** Manage track lists and navigation.

**Responsibilities:**
- Load tracks from folders
- Load tracks from playlist files (.m3u)
- Track current position
- Navigate forward/backward
- Implement shuffle mode
- Implement repeat modes (off/all/one)
- Maintain original track order

**Key files:**
- `playlist.go` - Core playlist operations
- `loader.go` - Source loading
- `shuffle.go` - Shuffle implementation

### 5. Shared Utilities (`internal/shared`)

**Purpose:** Common code used by multiple components.

**Responsibilities:**
- Define shared types and interfaces
- Implement IPC message transport
- Handle Unix socket operations
- Serialize/deserialize commands
- Define command constants

**Key files:**
- `transport.go` - Message serialization
- `ipc.go` - Unix socket helpers
- `types.go` - Shared data structures
- `commands.go` - Command definitions

## Communication Flow

### Command Execution Flow

```
User types: auxbox play -f ~/music

1. CLI parses arguments
   ↓
2. CLI checks if daemon is running (socket exists)
   ↓
3. If not running, spawn daemon process
   ↓
4. CLI connects to Unix socket
   ↓
5. CLI sends command: {"command":"play","args":{"folder":"~/music"}}
   ↓
6. Daemon receives command
   ↓
7. Daemon routes to command handler
   ↓
8. Handler loads playlist from folder
   ↓
9. Handler starts audio playback
   ↓
10. Daemon sends response: {"status":"ok","message":"✓ Loaded 12 tracks"}
    ↓
11. CLI displays message to user
    ↓
12. CLI exits (daemon continues playing)
```

### IPC Protocol

**Transport:** Unix domain sockets (Linux/macOS)

**Message format:** JSON-encoded request/response

**Request structure:**
```json
{
  "command": "play",
  "args": {
    "folder": "~/music"
  }
}
```

**Response structure:**
```json
{
  "status": "ok|error",
  "message": "Human-readable message",
  "data": { /* Optional structured data */ }
}
```

**Socket location:** `/tmp/auxbox-{uid}.sock`

## Audio Pipeline

### Playback Pipeline

```
Audio File (MP3/WAV/AIFF)
    ↓
[Format Detection] (by file extension)
    ↓
[Decoder Selection] (registry lookup)
    ↓
[Stream Decoder] (format-specific decoder)
    ↓
[Volume Control] (gain adjustment)
    ↓
[Speaker Output] (beep.Speaker)
    ↓
[Position Tracking] (real-time updates)
    ↓
System Audio Output
```

### Format Support

| Format | Decoder | Library |
|--------|---------|---------|
| MP3    | `mp3.go` | `hajimehoshi/go-mp3` |
| WAV    | `wav.go` | `gopxl/beep/v2` |
| AIFF   | `aiff.go` | `go-audio/aiff` |

### Position Tracking

Position updates occur in real-time via a background goroutine:

```go
// Simplified position tracking
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        if playing {
            currentPosition += 100 * time.Millisecond
            if currentPosition >= duration {
                handleTrackEnd() // Auto-advance
            }
        }
    }
}()
```

## Dependencies

### Direct Dependencies

**Beep v2** (`github.com/gopxl/beep/v2`)
- Purpose: Audio playback library
- Usage: Audio decoding, mixing, and output
- License: MIT

**go-audio/aiff** (`github.com/go-audio/aiff`)
- Purpose: AIFF file format support
- Usage: AIFF decoding
- License: Apache 2.0

**go-audio/audio** (`github.com/go-audio/audio`)
- Purpose: Audio buffer management
- Usage: Audio data structures
- License: Apache 2.0

### Indirect Dependencies

- `ebitengine/oto/v3` - Low-level audio output
- `hajimehoshi/go-mp3` - MP3 decoding
- Various system libraries for audio output

### Platform-Specific Audio Backends

- **Linux:** ALSA (Advanced Linux Sound Architecture)
- **macOS:** CoreAudio
- **Windows:** WASAPI (future support)

## Threading Model

> **📋 Note:** Threading patterns described below represent typical Go daemon architecture. Actual implementation may differ. Consult source code for precise concurrency details.

### Concurrency Patterns

**Main Thread (Daemon):**
- Listens for incoming IPC connections
- Handles command routing
- Non-blocking command processing

**Audio Thread (Player):**
- Managed by beep/oto libraries
- Handles audio buffer streaming
- Runs independently

**Position Tracker Thread:**
- Updates playback position every 100ms
- Checks for track end
- Triggers auto-advance

**Thread Safety:**
```go
// Example synchronization
type Player struct {
    mu       sync.Mutex
    playing  bool
    streamer beep.Streamer
}

func (p *Player) Play() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.playing = true
    // ... start playback
}
```

All shared state is protected by mutexes to ensure thread-safe access.

## Development

### Prerequisites

- Go 1.25.1 or later
- GCC (for CGO compilation on some platforms)
- System audio libraries (ALSA on Linux)

### Development Workflow

```bash
# Clone repository
git clone https://github.com/cerberussg/auxbox
cd auxbox

# Install dependencies
go mod download

# Build
go build -o auxbox cmd/auxbox/*.go

# Run
./auxbox play -f ~/music

# Run with verbose logging (future feature)
./auxbox --debug play -f ~/music
```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting
- Write godoc comments for exported functions
- Keep functions focused and small
- Prefer explicit error handling over panics

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/audio
go test ./internal/playlist

# Run with verbose output
go test -v ./...
```

### Test Structure

Tests are colocated with implementation:
```
internal/audio/
├── player.go
├── player_test.go
├── volume.go
└── volume_test.go
```

### Key Test Areas

- **Audio decoding** - Format support verification
- **Playlist operations** - Shuffle, repeat, navigation
- **IPC communication** - Message serialization
- **Command handlers** - Response correctness
- **Concurrency** - Thread-safety verification

## Building

### Single Platform Build

```bash
# Build for current platform
go build -o auxbox cmd/auxbox/*.go
```

### Cross-Platform Builds

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o auxbox-linux cmd/auxbox/*.go

# macOS (amd64 - Intel)
GOOS=darwin GOARCH=amd64 go build -o auxbox-macos cmd/auxbox/*.go

# macOS (arm64 - Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o auxbox-macos-arm64 cmd/auxbox/*.go

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o auxbox.exe cmd/auxbox/*.go
```

### Build Flags

```bash
# Optimized release build
go build -ldflags="-s -w" -o auxbox cmd/auxbox/*.go

# Static linking (Linux)
CGO_ENABLED=0 go build -o auxbox cmd/auxbox/*.go
```

### Build Artifacts

- Binary size: ~8-12 MB (varies by platform)
- No external runtime dependencies (except system audio libraries)
- Single self-contained executable

## Performance Considerations

> **📋 Note:** Performance metrics below are estimates based on typical Go audio applications. Actual resource usage may vary based on system, audio format, and playlist size. Profile your specific use case for accurate measurements.

### Memory Usage

- **Idle daemon:** ~10-15 MB (estimated)
- **Playing:** ~20-30 MB (estimated, includes audio buffers)
- **Large playlists:** Minimal increase (tracks stored as file paths)

### CPU Usage

- **Idle:** <1% (estimated)
- **Playing:** 1-3% (estimated, audio decoding and output)
- **Format detection:** Negligible (cached after first access)

### Disk I/O

- **Sequential reading** - Audio files streamed, not loaded entirely
- **Minimal writes** - Only for future metadata features (Phase 4+)

## Future Architecture Considerations

### Phase 4: Metadata Writing

**Challenge:** ID3 tag writing without corrupting files

**Approach:**
- Use established Go ID3 library (e.g., `github.com/bogem/id3v2`)
- Atomic file writes (write to temp, then rename)
- Backup original before modification (optional flag)

### Windows Support

**Challenge:** Unix sockets not available on Windows

**Approach:**
- Abstract IPC layer with interface
- Implement Windows Named Pipes alternative
- Conditional compilation with build tags

### GUI Companion

**Challenge:** CLI-first design, optional GUI

**Approach:**
- GUI as separate binary
- Communicates with daemon via same IPC
- No changes to core architecture

## Troubleshooting Development Issues

### Audio Output Issues

```bash
# Linux: Check ALSA
aplay -l

# macOS: Check CoreAudio
system_profiler SPAudioDataType
```

### Build Issues

```bash
# Update dependencies
go mod tidy
go mod download

# Clear build cache
go clean -cache
go clean -modcache
```

### Debugging

```go
// Add logging to components
import "log"

log.Printf("Player state: playing=%v, position=%v", p.playing, p.position)
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## Next Steps

- See [ROADMAP.md](ROADMAP.md) for planned features
- See [USER_GUIDE.md](USER_GUIDE.md) for usage documentation
- See [DJ_WORKFLOW.md](DJ_WORKFLOW.md) for DJ features
