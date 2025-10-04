# auxbox

A lightweight CLI music player for background listening with daemon architecture.

## Features

### Audio Format Support
- **MP3** - MPEG 1 Audio Layer 3 files
- **WAV** - Waveform Audio Files
- **AIFF/AIF** - Audio Interchange File Format

### Playback Controls
- **Play/Pause/Stop** - Full playback control
- **Skip/Back** - Navigate tracks with optional count (e.g., `skip 3`)
- **Volume Control** - Set volume from 0-100%
- **Auto-advance** - Automatically plays next track when current track ends
- **Position Tracking** - Real time playback position updates

### Source Types
- **Folder** - Load all supported audio files from a directory
- **Playlist** - Load tracks from playlist files (future expansion)

### Daemon Architecture
- **Background operation** - Music continues playing after CLI commands
- **Unix socket communication** - Efficient IPC between CLI and daemon
- **Single instance** - One daemon per user session

## Installation

### Build from Source

```bash
# Clone and build
git clone https://github.com/cerberussg/auxbox
cd auxbox
go build -o auxbox cmd/auxbox/main.go
```

### Add to PATH

After building, move the binary to a location in your PATH:

**Linux (including Arch Linux):**
```bash
# Create local bin directory if it doesn't exist
mkdir -p ~/.local/bin

# Move binary
mv auxbox ~/.local/bin/

# Add to PATH (if not already added)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**macOS:**
```bash
# Option 1: User local (recommended)
mkdir -p ~/bin
mv auxbox ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Option 2: System-wide (requires sudo)
sudo mv auxbox /usr/local/bin/
```

**Windows:**
```bash
# Move to a directory in your PATH, for example:
move auxbox.exe C:\Users\%USERNAME%\AppData\Local\Microsoft\WindowsApps\

# Or create a dedicated directory:
mkdir C:\auxbox
move auxbox.exe C:\auxbox\
# Then add C:\auxbox to your system PATH through System Properties
```

## Usage

### Basic Commands

**Start the daemon with a music folder:**
```bash
auxbox start --folder ~/Music/Albums/new-album/
auxbox start --folder /path/to/your/music/collection/
```

**Basic playback controls:**
```bash
auxbox play      # Start/resume playback
auxbox pause     # Pause playback
auxbox stop      # Stop and reset to beginning
```

**Navigation:**
```bash
auxbox skip      # Skip to next track
auxbox skip 3    # Skip forward 3 tracks
auxbox back      # Go back one track
auxbox back 2    # Go back 2 tracks
```

**Information:**
```bash
auxbox status    # Show current track info with position/duration
auxbox list      # Show all tracks in current queue
```

**Volume control:**
```bash
auxbox volume         # Show current volume
auxbox volume 75      # Set volume to 75%
auxbox volume 0       # Mute
auxbox volume 100     # Max volume
```

**Daemon management:**
```bash
auxbox exit      # Stop daemon and exit
```

### Example Workflow

```bash
# Start daemon with your music folder
auxbox start --folder ~/Downloads/new-pack/

# Output:  Loaded 12 tracks from folder
# auxbox daemon started in background. Use 'auxbox play' to start playback.

# Start playing
auxbox play

# Check what's playing
auxbox status
# Output: � song.mp3 | 2:34/4:12 | Track 1/12 | Source: ~/Downloads/new-pack/

# Skip a few tracks
auxbox skip 3

# Adjust volume
auxbox volume 60

# List all tracks (shows current with � marker)
auxbox list
# Output: Tracks (12 total):
#   1. first-song.mp3
#   2. second-song.mp3
# � 4. current-song.mp3
#   5. next-song.mp3
#   ...

# When done, exit the daemon
auxbox exit
```

### Help and Version

```bash
auxbox --help     # Show usage information
auxbox --version  # Show version information
```

## Technical Details

### Dependencies
- **Go 1.25.1+** - Built with Go modules
- **Beep v2** - Audio processing library
- **Unix sockets** - IPC communication (Linux/macOS)

### Architecture
- **Client Server model** - CLI commands communicate with background daemon
- **Modular design** - Separate packages for audio, playlist, server, and transport
- **Thread safe** - Concurrent-safe playback controls and status updates

### Audio System
- **Automatic format detection** - Based on file extension
- **Real time position tracking** - Updates during playback
- **Volume control with fading** - Smooth volume transitions
- **Resource management** - Proper cleanup of audio streams and files

## Development

**Run tests:**
```bash
go test ./...
```

**Build for different platforms:**
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o auxbox-linux cmd/auxbox/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o auxbox-macos cmd/auxbox/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o auxbox.exe cmd/auxbox/main.go
```

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]