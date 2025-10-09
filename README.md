# auxbox

A lightweight CLI music player for background listening with daemon architecture. Perfect for casual listening and DJ track preparation.

## Features

### Audio Format Support
- **MP3** - MPEG 1 Audio Layer 3 files
- **WAV** - Waveform Audio Files
- **AIFF/AIF** - Audio Interchange File Format

### Streamlined Playback
- **Instant music** - One command from silence to sound: `auxbox play -f ~/music`
- **Hot-swapping** - Switch music sources seamlessly while playing
- **Auto-advance** - Automatically plays next track when current track ends
- **Position tracking** - Real time playback position updates

### Playback Controls
- **Play/Pause/Stop** - Full playback control
- **Skip/Back** - Navigate tracks with optional count (e.g., `skip 3`)
- **Shuffle mode** - Random track selection with toggle support
- **Repeat modes** - Off (default), repeat-all (loop playlist), repeat-one (loop track)
- **Volume control** - Set volume from 0-100%

### Source Types
- **Folder** - Load all supported audio files from a directory
- **Playlist** - Load tracks from playlist files (.m3u support)

### DJ Workflow Integration
- **Track rating** - Rate tracks 1-5 stars while listening
- **Genre tagging** - Categorize tracks by musical style
- **Label tracking** - Tag record labels for organization
- **Rekordbox compatibility** - All metadata syncs with rekordbox

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
go build -o auxbox cmd/auxbox/*.go
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

### Instant Music Playback

**Load and play music instantly:**
```bash
auxbox play -f ~/Music/Albums/new-album/       # Load folder and play instantly
auxbox play --folder ~/Music/collection/       # Same as above (long form)
auxbox play -f ~/Music/jazz -s                 # Load folder, enable shuffle, and play
auxbox play -f ~/Music/chill -r                # Load folder with repeat-all enabled
auxbox play -f ~/Music/workout -s -r           # Load folder, shuffle, and repeat-all
auxbox play -p ~/playlists/favorites.m3u       # Load playlist and play instantly
auxbox play --playlist ~/playlists/rock.m3u    # Same as above (long form)
```

**Hot-swap sources while playing:**
```bash
auxbox play -f ~/different-folder/              # Switch to new folder instantly
auxbox play -p ~/playlists/workout.m3u          # Switch to playlist while playing
```

**Basic playback controls:**
```bash
auxbox play      # Resume playback (if paused)
auxbox pause     # Pause playback
auxbox stop      # Stop and reset to beginning
```

### Navigation and Control

**Track navigation:**
```bash
auxbox skip      # Skip to next track (or random if shuffle is on)
auxbox skip 3    # Skip forward 3 tracks
auxbox back      # Go back one track (or random if shuffle is on)
auxbox back 2    # Go back 2 tracks
```

**Shuffle mode:**
```bash
auxbox shuffle   # Toggle shuffle on (random track selection)
auxbox shuffle   # Toggle shuffle off (sequential playback)
```

**Repeat modes:**
```bash
auxbox repeat    # Cycle: off → repeat-all → repeat-one → off
# First press:   Repeat all enabled (loops playlist)
# Second press:  Repeat one enabled (loops current track)
# Third press:   Repeat off (stops at end)
```

**Information:**
```bash
auxbox status    # Show current track info with position/duration
auxbox list      # Show tracks (windowed view: 15 tracks around current position)
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
# Load music and start playing instantly - one command!
auxbox play -f ~/Downloads/new-pack/

# Output: ✓ Loaded 12 tracks from folder and started playback
# Music starts playing immediately!

# Check what's playing
auxbox status
# Output: ▶ song.mp3 | 2:34/4:12 | Track 1/12 | Source: ~/Downloads/new-pack/

# Enable shuffle mode for random playback
auxbox shuffle
# Output: Playlist shuffled

# Skip to a random track
auxbox skip
# Picks a random track from the playlist

# Switch to a different folder while playing
auxbox play -f ~/Music/jazz-collection/
# Output: ✓ Loaded 8 tracks from folder and started playback
# Seamlessly switches to new music source

# Load and shuffle in one command
auxbox play -f ~/Music/5300-track-library/ -s
# Output: ✓ Loaded 5300 tracks from folder (shuffled) and started playback
# Each track completion picks a random next track

# Load with shuffle and repeat-all for infinite random playback
auxbox play -f ~/Music/study-mix/ -s -r
# Output: ✓ Loaded 42 tracks from folder (shuffled, repeat-all) and started playback
# Plays random tracks forever

# Enable repeat mode on current playlist
auxbox repeat
# Output: Repeat all enabled
# Playlist will loop when it reaches the end

# Cycle to repeat-one to loop current track
auxbox repeat
# Output: Repeat one enabled
# Current track will replay indefinitely

# Cycle back to repeat off
auxbox repeat
# Output: Repeat off
# Playback stops at end of playlist

# Toggle shuffle off for sequential playback
auxbox shuffle
# Output: Shuffle disabled, restored original order

# Adjust volume
auxbox volume 60

# View tracks around current position (windowed for large playlists)
auxbox list
# Output for small playlist (≤15 tracks):
# Tracks (8 total):
#   1. first-jazz-song.mp3
#   2. second-jazz-song.mp3
# ▶ 4. current-jazz-song.mp3
#   5. next-jazz-song.mp3
#   ...
#
# Output for large playlist (shows 15-track window):
# Tracks (5763 total, showing 2429-2443):
#   2429. track-2428.mp3
#   ...
# ▶ 2436. current-track.mp3
#   ...
#   2443. track-2442.mp3

# When done, exit the daemon
auxbox exit
```

## DJ Workflow Integration

auxbox doubles as a powerful DJ preparation tool, allowing you to rate and tag tracks while listening - perfect for organizing your music library without opening heavyweight DJ software.

### Track Rating System
Rate tracks on the fly while listening to build your energy-level system:

```bash
# Preview new tracks and rate them
auxbox play -f ~/new-tracks-pack/

# While listening, rate tracks 1-5 stars
auxbox stars 5          # Peak-hour banger
auxbox stars 2          # Good opener/breakdown track
auxbox stars 4          # High energy, main set material
```

### Genre Tagging
Categorize tracks by style during preview sessions:

```bash
auxbox genre "Deep House"
auxbox genre "Tech House"
auxbox genre "Progressive Trance"
```

### Record Label Organization
Track the source/label for discovery and organization:

```bash
auxbox label "Defected Records"
auxbox label "Anjunadeep"
auxbox label "Drumcode"
```

### Complete DJ Prep Workflow
```bash
# Load new promo pack for evaluation
auxbox play -f ~/promos/december-2024/

# Listen and rate each track
auxbox stars 4
auxbox genre "Deep House"
auxbox label "Hot Creations"

auxbox skip                # Next track
auxbox stars 2             # Opener material
auxbox genre "Minimal Tech"
auxbox label "Percomaniacs"

# Hot-swap to different style pack
auxbox play -f ~/downloads/techno-pack/

# Continue rating and tagging...
auxbox stars 5
auxbox genre "Peak Time Techno"
auxbox label "Drumcode"
```

### Rekordbox Integration
All metadata is written directly to your audio files using industry-standard ID3v2 tags:
- **Stars** → POPM (Popularimeter) frame compatible with rekordbox star ratings
- **Genre** → TCON (Content Type) field
- **Label** → TPUB (Publisher) field

When you open rekordbox later, all your ratings, genres, and labels are already there - no re-work needed!

### Benefits for DJs
- **Preview without rekordbox overhead** - Quick track evaluation
- **Bulk rating sessions** - Rate entire packs in one session
- **Energy-level organization** - 5 stars = peak hour, 2 stars = openers
- **Style categorization** - Genre tagging for playlist creation
- **Source tracking** - Know which labels produce your favorite tracks
- **Seamless integration** - Works with your existing rekordbox workflow

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

## Development Roadmap

### Phase 1: Streamlined UX ✅
- Unified play command with source loading
- Hot-swapping music sources while playing
- One-command-to-music workflow

### Phase 2: Shuffle Feature ✅
- Random track selection mode
- Toggle shuffle on/off during playback
- `-s` flag for instant shuffle on load
- Works with skip/back commands

### Phase 3: Repeat Modes ✅
- Three repeat modes: off (default), repeat-all, repeat-one
- Toggle repeat modes with `auxbox repeat` command
- `-r` flag for instant repeat-all on load
- Auto-loop playlists and tracks
- Seamless track transitions

### Phase 4: DJ Star Rating ⭐
- 1-5 star rating system while listening
- Rekordbox-compatible metadata writing
- Energy-level track organization

### Phase 5: Genre Tagging 🎵
- Real-time genre classification
- Style-based track organization
- DJ workflow integration

### Phase 6: Label Tracking 🏷️
- Record label metadata tracking
- Source discovery and organization
- Complete DJ preparation workflow

## Development

**Run tests:**
```bash
go test ./...
```

**Build for different platforms:**
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o auxbox-linux cmd/auxbox/*.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o auxbox-macos cmd/auxbox/*.go

# Windows
GOOS=windows GOARCH=amd64 go build -o auxbox.exe cmd/auxbox/*.go
```

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]