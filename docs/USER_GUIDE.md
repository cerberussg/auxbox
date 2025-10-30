# User Guide

Comprehensive guide to using auxbox for everyday music listening.

## Table of Contents

- [Quick Start](#quick-start)
- [Music Sources](#music-sources)
- [Playback Commands](#playback-commands)
- [Navigation](#navigation)
- [Shuffle Mode](#shuffle-mode)
- [Repeat Modes](#repeat-modes)
- [Volume Control](#volume-control)
- [Information Commands](#information-commands)
- [Daemon Management](#daemon-management)
- [Complete Workflows](#complete-workflows)

## Quick Start

The fastest way to start listening:

```bash
# Load a folder and play instantly
auxbox play -f ~/Music/Albums/your-album/

# Or load a playlist
auxbox play -p ~/playlists/favorites.m3u
```

That's it! Music starts playing immediately.

## Music Sources

auxbox supports two types of music sources:

### Folders

Load all supported audio files from a directory:

```bash
# Load folder
auxbox play -f ~/Music/jazz-collection/
auxbox play --folder ~/Downloads/new-album/

# Load with shuffle enabled
auxbox play -f ~/Music/study-mix/ -s

# Load with repeat-all enabled
auxbox play -f ~/Music/workout/ -r

# Load with both shuffle and repeat
auxbox play -f ~/Music/background/ -s -r
```

**Supported formats:**
- MP3 (MPEG 1 Audio Layer 3)
- WAV (Waveform Audio Files)
- AIFF/AIF (Audio Interchange File Format)

### Playlists

Load tracks from playlist files:

```bash
# Load .m3u playlist
auxbox play -p ~/playlists/chill-vibes.m3u
auxbox play --playlist ~/playlists/rock.m3u

# Load with shuffle
auxbox play -p ~/playlists/favorites.m3u -s

# Load with repeat-all
auxbox play -p ~/playlists/workout.m3u -r
```

**Supported playlist formats:**
- M3U (.m3u)

## Playback Commands

### Play

Start or resume playback:

```bash
# Resume if paused
auxbox play

# Load new source and play
auxbox play -f ~/Music/new-folder/
auxbox play -p ~/playlists/favorites.m3u
```

### Pause

Pause playback without stopping:

```bash
auxbox pause
```

Resume with `auxbox play`.

### Stop

Stop playback and reset to beginning:

```bash
auxbox stop
```

Next play command starts from the first track.

### Hot-Swapping

Switch music sources seamlessly while playing:

```bash
# Currently playing from folder A
auxbox play -f ~/Music/folder-A/

# Switch to folder B instantly
auxbox play -f ~/Music/folder-B/

# Switch to playlist
auxbox play -p ~/playlists/chill.m3u
```

No need to stop first - auxbox handles the transition automatically.

## Navigation

### Skip Forward

Move to the next track:

```bash
# Skip one track
auxbox skip

# Skip multiple tracks
auxbox skip 3    # Skip forward 3 tracks
auxbox skip 10   # Skip forward 10 tracks
```

**Behavior with shuffle:**
- When shuffle is ON: selects a random unplayed track
- When shuffle is OFF: moves sequentially forward

### Skip Backward

Return to previous tracks:

```bash
# Go back one track
auxbox back

# Go back multiple tracks
auxbox back 2    # Back 2 tracks
auxbox back 5    # Back 5 tracks
```

**Behavior with shuffle:**
- When shuffle is ON: selects a random unplayed track (not truly "backward")
- When shuffle is OFF: moves sequentially backward

## Shuffle Mode

Randomize your listening experience:

```bash
# Toggle shuffle on
auxbox shuffle
# Output: Playlist shuffled

# Toggle shuffle off (restores original order)
auxbox shuffle
# Output: Shuffle disabled, restored original order
```

**How shuffle works:**
- Tracks are played in random order
- No track repeats until all tracks have been played
- `skip` and `back` select random unplayed tracks
- Original order is restored when toggled off

**Instant shuffle on load:**

```bash
# Load and shuffle in one command
auxbox play -f ~/Music/library/ -s
```

## Repeat Modes

Control what happens when playback reaches the end:

```bash
# Toggle through repeat modes
auxbox repeat
```

### Three Repeat Modes:

1. **Off (default)** - Playback stops at the end of the playlist
2. **Repeat All** - Loops the entire playlist indefinitely
3. **Repeat One** - Loops the current track indefinitely

### Mode Cycling:

```bash
# First press
auxbox repeat
# Output: Repeat all enabled

# Second press
auxbox repeat
# Output: Repeat one enabled

# Third press
auxbox repeat
# Output: Repeat off
```

**Instant repeat-all on load:**

```bash
# Load with repeat-all enabled
auxbox play -f ~/Music/study-mix/ -r

# Combine with shuffle for infinite random playback
auxbox play -f ~/Music/background/ -s -r
```

## Volume Control

Adjust playback volume:

```bash
# Show current volume
auxbox volume
# Output: Volume: 85%

# Set volume (0-100)
auxbox volume 75     # Set to 75%
auxbox volume 0      # Mute
auxbox volume 100    # Maximum volume
```

Volume changes are applied with smooth fading for a better listening experience.

## Information Commands

### Status

View current playback information:

```bash
auxbox status
```

**Example output:**
```
▶ song-title.mp3 | 2:34/4:12 | Track 5/12 | Source: ~/Music/jazz/
```

**Status indicators:**
- `▶` - Playing
- `⏸` - Paused
- `■` - Stopped

**Information shown:**
- Current track filename
- Position / Duration (mm:ss format)
- Track number / Total tracks
- Source path (folder or playlist)

### List Tracks

View tracks in the current playlist:

```bash
auxbox list
```

**Small playlists (≤15 tracks):**
```
Tracks (8 total):
  1. first-song.mp3
  2. second-song.mp3
  3. third-song.mp3
▶ 4. current-song.mp3
  5. next-song.mp3
  6. another-song.mp3
  7. penultimate.mp3
  8. last-song.mp3
```

**Large playlists (>15 tracks):**

Shows a 15-track window centered around the current track:

```
Tracks (5763 total, showing 2429-2443):
  2429. track-2428.mp3
  2430. track-2429.mp3
  ...
▶ 2436. current-track.mp3
  ...
  2442. track-2441.mp3
  2443. track-2442.mp3
```

This windowed view makes it easy to navigate large music libraries without overwhelming output.

## Daemon Management

auxbox runs as a background daemon that persists between commands.

### Exit

Stop the daemon and exit:

```bash
auxbox exit
```

This stops playback, cleans up resources, and shuts down the daemon.

### Automatic Daemon Start

The daemon starts automatically when you issue your first command. You never need to manually start it.

## Complete Workflows

### Casual Listening Session

```bash
# Start with a folder
auxbox play -f ~/Music/chill/
# Output: ✓ Loaded 24 tracks from folder and started playback

# Check what's playing
auxbox status
# Output: ▶ smooth-vibes.mp3 | 1:23/3:45 | Track 1/24

# Skip a track
auxbox skip

# Adjust volume
auxbox volume 60

# When done
auxbox exit
```

### Shuffled Background Music

```bash
# Load large library with shuffle and repeat
auxbox play -f ~/Music/all-music/ -s -r
# Output: ✓ Loaded 5300 tracks from folder (shuffled, repeat-all) and started playback

# Music plays randomly and loops forever
# Perfect for background listening while working

# Pause when needed
auxbox pause

# Resume later
auxbox play
```

### DJ Track Preview Session

```bash
# Load new promo pack
auxbox play -f ~/Downloads/promos-december/
# Output: ✓ Loaded 15 tracks from folder and started playback

# Listen and skip through tracks
auxbox status    # Check current track
auxbox skip      # Next track
auxbox skip 3    # Jump ahead 3 tracks

# Rate tracks as you go (see DJ Workflow guide)
# ...

# When done previewing
auxbox exit
```

### Playlist Hot-Swapping

```bash
# Start with a morning playlist
auxbox play -p ~/playlists/morning-coffee.m3u

# Switch to work focus playlist
auxbox play -p ~/playlists/work-focus.m3u
# Seamless transition while playing

# Switch to workout music for afternoon
auxbox play -p ~/playlists/workout.m3u

# End with evening chill
auxbox play -p ~/playlists/evening-chill.m3u
```

### Large Library Navigation

```bash
# Load massive collection
auxbox play -f ~/Music/complete-library/
# Output: ✓ Loaded 8420 tracks from folder and started playback

# Use list to see where you are
auxbox list
# Output: Tracks (8420 total, showing 1-15):

# Jump forward quickly
auxbox skip 50

# Check position
auxbox list
# Output: Tracks (8420 total, showing 44-58):

# Enable shuffle for variety
auxbox shuffle

# Enable repeat to never run out
auxbox repeat
# Output: Repeat all enabled
```

## Tips and Tricks

### Quick Volume Adjustments

```bash
# Save your preferred volumes as shell aliases
alias volquiet='auxbox volume 30'
alias volnormal='auxbox volume 70'
alias volloud='auxbox volume 95'
```

### Instant Music Aliases

```bash
# Create shortcuts for your favorite music
alias jazz='auxbox play -f ~/Music/jazz/ -s -r'
alias focus='auxbox play -p ~/playlists/deep-focus.m3u -r'
alias workout='auxbox play -p ~/playlists/high-energy.m3u -s -r'
```

### Check Status While Working

```bash
# Quick status check (add to shell prompt or tmux status bar)
watch -n 5 auxbox status
```

### Waybar Integration (Hyprland)

If you use Hyprland with Waybar, you can add auxbox to your status bar with interactive controls:

**Add to `~/.config/waybar/config.jsonc`:**

```jsonc
{
  // Add "custom/auxbox" to your modules array
  "modules-center": ["clock", "custom/auxbox"],

  // Define the auxbox module
  "custom/auxbox": {
    "interval": 5,
    "exec": "auxbox status",
    "on-click": "auxbox pause",
    "on-double-click": "auxbox play",
    "on-click-right": "auxbox skip"
  }
}
```

**Features:**
- **Auto-refresh** - Updates status every 5 seconds
- **Left click** - Pause playback
- **Double click** - Resume playback
- **Right click** - Skip to next track

The status bar will display your current track information directly in Waybar!

### Background Daemon

Remember: auxbox runs in the background. You can:
- Close your terminal and music keeps playing
- Run commands from any terminal window
- Issue commands while music plays without interruption

## Getting Help

```bash
# Show usage information
auxbox --help

# Show version
auxbox --version
```

## Next Steps

- Learn DJ-specific features in [DJ_WORKFLOW.md](DJ_WORKFLOW.md)
- Understand technical details in [ARCHITECTURE.md](ARCHITECTURE.md)
- See development plans in [ROADMAP.md](ROADMAP.md)
