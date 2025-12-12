# auxbox Development Roadmap

This document outlines the phased development plan for auxbox, tracking completed features and upcoming work.

## Current Status

**✅ Phase 4 Complete!** - All core features are now implemented and stable:
- Playback, shuffle, and repeat modes (Phases 1-3)
- Complete metadata editing with ID3v2 tags (Phase 4)
- DJ workflow integration with Rekordbox and Mixxx

**📋 Phase 5+ Future** - Advanced features under consideration.

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

### Phase 4: Metadata Editing ✅
**Completed: 2024**

**Implemented Features:**
- Complete ID3v2 metadata editing for MP3 files
- 1-5 star rating system (`auxbox rate`)
- Record label tagging (`auxbox label`)
- Genre tagging (`auxbox genre`)
- Track info editing (`auxbox title`, `auxbox artist`, `auxbox album`, `auxbox year`)
- Atomic writes with backup/restore system
- Rekordbox and Mixxx integration via ID3v2 tags

**Technical Implementation:**
- ID3v2 tag writing with POPM (rating), TCON (genre), TPUB (label) frames
- Backup system prevents file corruption
- Format validation (MP3 only, AIFF planned)
- Real-time metadata editing during playback

**DJ Workflow:**
- Rate tracks 1-5 stars for energy-level organization
- Tag genres and labels while listening
- Complete track metadata editing
- Sync with Rekordbox via re-import
- Sync with Mixxx file tag system

## Future Phases

### Phase 5: Advanced Features 📋
**Status: Under Consideration**

Potential features being evaluated:
- AIFF file support for metadata (Phase 4.5)
- Smart playlists based on metadata
- BPM detection and tagging
- Key detection for harmonic mixing
- Waveform analysis
- Cue point management

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
