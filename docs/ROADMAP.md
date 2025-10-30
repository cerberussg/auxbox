# auxbox Development Roadmap

This document outlines the phased development plan for auxbox, tracking completed features and upcoming work.

## Current Status

**✅ Phases 1-3 Complete** - Core playback, shuffle, and repeat modes are fully implemented and stable.

**🚧 Phase 4 In Planning** - Star rating feature is being designed. Implementation has not started.

**📋 Phases 5-6 Future Vision** - Genre tagging and label tracking are planned concepts. Design and implementation timeline TBD.

## Completed Phases

### Phase 1: Streamlined UX ✅
**Completed: 2024**

- Unified play command with source loading
- Hot-swapping music sources while playing
- One-command-to-music workflow
- Instant playback from folder or playlist

### Phase 2: Shuffle Feature ✅
**Completed: 2024**

- Random track selection mode
- Toggle shuffle on/off during playback
- `-s` flag for instant shuffle on load
- Works seamlessly with skip/back commands
- Maintains original playlist order when toggled off

### Phase 3: Repeat Modes ✅
**Completed: 2024**

- Three repeat modes: off (default), repeat-all, repeat-one
- Toggle repeat modes with `auxbox repeat` command
- `-r` flag for instant repeat-all on load
- Auto-loop playlists and individual tracks
- Seamless track transitions on repeat

## In Progress

### Phase 4: DJ Star Rating ⭐
**Status: Planning**

**Goals:**
- 1-5 star rating system while listening
- Rekordbox-compatible metadata writing
- Energy-level track organization
- Real-time rating during playback sessions

**Technical Challenges:**
- Rekordbox database integration (star ratings stored in DB)
- ID3v2 POPM (Popularimeter) frame compatibility
- Metadata synchronization strategy

**Implementation Plan:**
- TBD - requires research on Rekordbox DB format

## Future Phases

### Phase 5: Genre Tagging 🎵
**Status: Planned**

- Real-time genre classification during listening
- Style-based track organization
- DJ workflow integration for genre-based playlists
- ID3v2 TCON (Content Type) field writing

### Phase 6: Label Tracking 🏷️
**Status: Planned**

- Record label metadata tracking
- Source discovery and organization
- Complete DJ preparation workflow
- ID3v2 TPUB (Publisher) field writing

## Future Considerations

### Potential Features
- Playlist management (save/load custom playlists)
- Crate organization (DJ-style folder management)
- BPM detection and storage
- Key detection for harmonic mixing
- Waveform display integration
- Multiple playlist queue support
- Search and filter commands

### Platform Enhancements
- Windows Named Pipes support (currently Unix sockets only)
- GUI companion app (optional visual interface)
- Mobile remote control
- Web interface for remote management

## Philosophy

auxbox follows an incremental development approach:
1. Each phase delivers a complete, tested feature
2. Features are merged only when fully functional
3. UX simplicity is prioritized over feature complexity
4. DJ workflow integration drives feature priorities
5. Rekordbox compatibility is maintained throughout

## Contributing to the Roadmap

Have ideas for future phases? See [CONTRIBUTING.md](CONTRIBUTING.md) for how to propose new features.
