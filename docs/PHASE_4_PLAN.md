# Phase 4 Implementation Plan: DJ Rating & Metadata

**Status:** Planning → Implementation Ready
**Target:** MP3 files with ID3v2 tags
**Future:** AIFF support via custom ID3 chunk library (Phase 4.5+)

## Executive Summary

Phase 4 adds real-time track rating and label tagging to auxbox, enabling DJ workflow integration with Rekordbox and Mixxx. This phase focuses on **MP3 files only** with proven ID3v2 tag writing, leaving AIFF support for future development.

### User Stories

```bash
# While listening to a new promo pack
auxbox play -f ~/Downloads/december-promos/
auxbox status              # "Now playing: Track 01.mp3"
auxbox rate 5              # "Rated 5/5: Track 01.mp3"
auxbox label "Drumcode"    # "🏷️  Label set to 'Drumcode': Track 01.mp3"
auxbox skip
auxbox rate 3              # Rate next track
auxbox skip
auxbox rate 4
auxbox label "Anjunadeep"

# Later in Rekordbox/Mixxx
# Re-import tracks → Ratings appear as stars (⭐⭐⭐⭐⭐)
```

### Success Criteria

- ✅ Rate MP3 tracks 1-5 during playback (displayed as stars in DJ software)
- ✅ Tag record labels on MP3 tracks
- ✅ Metadata syncs to Rekordbox (re-import workflow)
- ✅ Metadata syncs to Mixxx (file tags enabled)
- ✅ Non-destructive writes (backup on error)
- ✅ Works while track is playing
- ✅ Fast response (<100ms for metadata write)

---

## Architecture Overview

### Component Structure

```
auxbox/
├── cmd/auxbox/
│   ├── cli.go              [MODIFY] Add stars, label commands
│   └── cli_test.go         [MODIFY] Add command tests
│
├── internal/
│   ├── shared/
│   │   ├── commands.go     [MODIFY] Add CmdStars, CmdLabel
│   │   └── types.go        [MODIFY] Add Rating, Label fields to Track
│   │
│   ├── server/
│   │   ├── server.go       [MODIFY] Add metadata handler routing
│   │   └── commands/
│   │       ├── metadata.go [NEW] MetadataHandler for stars/label
│   │       └── metadata_test.go [NEW]
│   │
│   └── metadata/           [NEW PACKAGE]
│       ├── writer.go       [NEW] ID3v2 tag writer
│       ├── reader.go       [NEW] ID3v2 tag reader
│       ├── formats.go      [NEW] Format-specific handlers
│       ├── backup.go       [NEW] Atomic write with backup
│       └── metadata_test.go [NEW]
│
└── docs/
    ├── DJ_WORKFLOW.md      [UPDATE] Mark Phase 4 as ✅ Implemented
    └── ROADMAP.md          [UPDATE] Phase 4 status
```

---

## Technical Design

### 1. Dependency: github.com/bogem/id3v2

**Rationale:**
- Pure Go implementation (no CGo)
- Full ID3v2.3 and ID3v2.4 support
- POPM (Popularimeter) frame support
- Active maintenance (2024 commits)
- Used by popular Go music tools

**Installation:**
```bash
go get github.com/bogem/id3v2/v2
```

**go.mod addition:**
```go
require (
    github.com/bogem/id3v2/v2 v2.1.4
    // ... existing deps
)
```

---

### 2. Metadata Package (`internal/metadata/`)

#### 2.1 Writer Interface (`writer.go`)

```go
package metadata

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/bogem/id3v2/v2"
)

const (
    // auxbox identifier for POPM frames
    AuxboxPOPMEmail = "auxbox@auxbox.org"

    // Rating scale (1-5 rating → 1-255 POPM values → displayed as stars in DJ software)
    RatingScale1 = 1
    RatingScale2 = 64
    RatingScale3 = 128
    RatingScale4 = 196
    RatingScale5 = 255
)

type Writer interface {
    WriteRating(filePath string, rating int) error
    WriteLabel(filePath string, label string) error
    WriteMetadata(filePath string, rating int, label string) error
}

type ID3Writer struct {
    backupEnabled bool
}

func NewID3Writer() *ID3Writer {
    return &ID3Writer{
        backupEnabled: true,
    }
}

// WriteRating writes rating to MP3 file (1-5 rating)
func (w *ID3Writer) WriteRating(filePath string, rating int) error {
    if rating < 1 || rating > 5 {
        return fmt.Errorf("invalid rating: %d (must be 1-5)", rating)
    }

    // Open ID3v2 tag (or create if not exists)
    tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
    if err != nil {
        return fmt.Errorf("failed to open ID3 tag: %w", err)
    }
    defer tag.Close()

    // Convert rating to POPM rating value
    ratingValue := ratingToPOPM(rating)

    // Create POPM frame
    popm := id3v2.PopularimeterFrame{
        Email:  AuxboxPOPMEmail,
        Rating: ratingValue,
        Counter: 0,
    }

    // Remove existing POPM frames (avoid duplicates)
    tag.DeleteFrames(tag.CommonID("Popularimeter"))

    // Add new POPM frame
    tag.AddFrame(tag.CommonID("Popularimeter"), popm)

    // Atomic write with backup
    if w.backupEnabled {
        if err := w.backupFile(filePath); err != nil {
            return fmt.Errorf("backup failed: %w", err)
        }
    }

    // Save tag
    if err := tag.Save(); err != nil {
        w.restoreBackup(filePath) // Attempt restore on failure
        return fmt.Errorf("failed to save tag: %w", err)
    }

    return nil
}

// WriteLabel writes record label to TPUB frame
func (w *ID3Writer) WriteLabel(filePath string, label string) error {
    tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
    if err != nil {
        return fmt.Errorf("failed to open ID3 tag: %w", err)
    }
    defer tag.Close()

    // TPUB = Publisher (used for record label)
    tag.AddTextFrame(tag.CommonID("Publisher"), id3v2.EncodingUTF8, label)

    if w.backupEnabled {
        if err := w.backupFile(filePath); err != nil {
            return fmt.Errorf("backup failed: %w", err)
        }
    }

    if err := tag.Save(); err != nil {
        w.restoreBackup(filePath)
        return fmt.Errorf("failed to save tag: %w", err)
    }

    return nil
}

// WriteMetadata writes both rating and label atomically
func (w *ID3Writer) WriteMetadata(filePath string, rating int, label string) error {
    tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
    if err != nil {
        return fmt.Errorf("failed to open ID3 tag: %w", err)
    }
    defer tag.Close()

    // Write rating if provided
    if rating > 0 {
        if rating < 1 || rating > 5 {
            return fmt.Errorf("invalid rating: %d", rating)
        }

        ratingValue := starsToPOPM(rating)
        popm := id3v2.PopularimeterFrame{
            Email:  AuxboxPOPMEmail,
            Rating: ratingValue,
            Counter: 0,
        }
        tag.DeleteFrames(tag.CommonID("Popularimeter"))
        tag.AddFrame(tag.CommonID("Popularimeter"), popm)
    }

    // Write label if provided
    if label != "" {
        tag.AddTextFrame(tag.CommonID("Publisher"), id3v2.EncodingUTF8, label)
    }

    if w.backupEnabled {
        if err := w.backupFile(filePath); err != nil {
            return fmt.Errorf("backup failed: %w", err)
        }
    }

    if err := tag.Save(); err != nil {
        w.restoreBackup(filePath)
        return fmt.Errorf("failed to save tag: %w", err)
    }

    return nil
}

// ratingToPOPM converts 1-5 rating to POPM rating value
func ratingToPOPM(rating int) byte {
    switch rating {
    case 1:
        return RatingScale1
    case 2:
        return RatingScale2
    case 3:
        return RatingScale3
    case 4:
        return RatingScale4
    case 5:
        return RatingScale5
    default:
        return 0
    }
}
```

#### 2.2 Reader Interface (`reader.go`)

```go
package metadata

import (
    "fmt"

    "github.com/bogem/id3v2/v2"
)

type Reader interface {
    ReadRating(filePath string) (int, error)
    ReadLabel(filePath string) (string, error)
}

type ID3Reader struct{}

func NewID3Reader() *ID3Reader {
    return &ID3Reader{}
}

// ReadRating reads POPM frame and converts to 1-5 rating
func (r *ID3Reader) ReadRating(filePath string) (int, error) {
    tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
    if err != nil {
        return 0, fmt.Errorf("failed to open tag: %w", err)
    }
    defer tag.Close()

    // Get all POPM frames
    frames := tag.GetFrames(tag.CommonID("Popularimeter"))
    if len(frames) == 0 {
        return 0, nil // No rating set
    }

    // Look for auxbox POPM frame first
    for _, frame := range frames {
        popm, ok := frame.(id3v2.PopularimeterFrame)
        if ok && popm.Email == AuxboxPOPMEmail {
            return popmToRating(popm.Rating), nil
        }
    }

    // Fallback: use first POPM frame
    if popm, ok := frames[0].(id3v2.PopularimeterFrame); ok {
        return popmToRating(popm.Rating), nil
    }

    return 0, nil
}

// ReadLabel reads TPUB (Publisher) frame
func (r *ID3Reader) ReadLabel(filePath string) (string, error) {
    tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
    if err != nil {
        return "", fmt.Errorf("failed to open tag: %w", err)
    }
    defer tag.Close()

    frames := tag.GetFrames(tag.CommonID("Publisher"))
    if len(frames) == 0 {
        return "", nil
    }

    if textFrame, ok := frames[0].(id3v2.TextFrame); ok {
        return textFrame.Text, nil
    }

    return "", nil
}

// popmToRating converts POPM rating value to 1-5 rating
func popmToRating(rating byte) int {
    switch {
    case rating == 0:
        return 0
    case rating < 32:
        return 1
    case rating < 96:
        return 2
    case rating < 160:
        return 3
    case rating < 224:
        return 4
    default:
        return 5
    }
}
```

#### 2.3 Backup System (`backup.go`)

```go
package metadata

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// backupFile creates a temporary backup before metadata write
func (w *ID3Writer) backupFile(filePath string) error {
    backupPath := filePath + ".auxbox.backup"

    // Remove old backup if exists
    os.Remove(backupPath)

    src, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer src.Close()

    dst, err := os.Create(backupPath)
    if err != nil {
        return err
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    return err
}

// restoreBackup restores from backup on write failure
func (w *ID3Writer) restoreBackup(filePath string) error {
    backupPath := filePath + ".auxbox.backup"

    if _, err := os.Stat(backupPath); os.IsNotExist(err) {
        return fmt.Errorf("backup not found")
    }

    // Remove corrupted file
    os.Remove(filePath)

    // Restore backup
    return os.Rename(backupPath, filePath)
}

// cleanupBackup removes backup file after successful write
func (w *ID3Writer) cleanupBackup(filePath string) {
    backupPath := filePath + ".auxbox.backup"
    os.Remove(backupPath)
}
```

#### 2.4 Format Detection (`formats.go`)

```go
package metadata

import (
    "path/filepath"
    "strings"
)

type AudioFormat int

const (
    FormatUnknown AudioFormat = iota
    FormatMP3
    FormatAIFF
    FormatWAV
)

// DetectFormat returns the audio format based on file extension
func DetectFormat(filePath string) AudioFormat {
    ext := strings.ToLower(filepath.Ext(filePath))

    switch ext {
    case ".mp3":
        return FormatMP3
    case ".aiff", ".aif":
        return FormatAIFF
    case ".wav":
        return FormatWAV
    default:
        return FormatUnknown
    }
}

// IsMetadataSupported returns whether metadata writing is supported
func IsMetadataSupported(filePath string) bool {
    format := DetectFormat(filePath)
    // Phase 4: Only MP3 supported
    return format == FormatMP3
}
```

---

### 3. Command Handler (`internal/server/commands/metadata.go`)

```go
package commands

import (
    "fmt"
    "log"
    "path/filepath"

    "github.com/cerberussg/auxbox/internal/metadata"
    "github.com/cerberussg/auxbox/internal/playlist"
    "github.com/cerberussg/auxbox/internal/shared"
)

type MetadataHandler struct {
    playlist *playlist.Playlist
    writer   metadata.Writer
    reader   metadata.Reader
}

func NewMetadataHandler(playlist *playlist.Playlist) *MetadataHandler {
    return &MetadataHandler{
        playlist: playlist,
        writer:   metadata.NewID3Writer(),
        reader:   metadata.NewID3Reader(),
    }
}

// HandleRate rates the current track (1-5)
func (h *MetadataHandler) HandleRate(cmd shared.Command) shared.Response {
    currentTrack := h.playlist.GetCurrentTrack()
    if currentTrack == nil {
        return shared.NewErrorResponse("No track is currently playing")
    }

    rating := cmd.Rating
    if rating < 1 || rating > 5 {
        return shared.NewErrorResponse(fmt.Sprintf("Invalid rating: %d (must be 1-5)", rating))
    }

    // Check format support
    if !metadata.IsMetadataSupported(currentTrack.Path) {
        ext := filepath.Ext(currentTrack.Path)
        return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
    }

    // Write rating
    if err := h.writer.WriteRating(currentTrack.Path, rating); err != nil {
        log.Printf("Failed to write rating: %v", err)
        return shared.NewErrorResponse(fmt.Sprintf("Failed to rate track: %v", err))
    }

    log.Printf("Rated '%s' with %d/5", currentTrack.Filename, rating)

    return shared.NewSuccessResponse(
        fmt.Sprintf("Rated %d/5: %s", rating, currentTrack.Filename),
        nil,
    )
}

// HandleLabel tags the current track with a record label
func (h *MetadataHandler) HandleLabel(cmd shared.Command) shared.Response {
    currentTrack := h.playlist.GetCurrentTrack()
    if currentTrack == nil {
        return shared.NewErrorResponse("No track is currently playing")
    }

    label := cmd.Label
    if label == "" {
        return shared.NewErrorResponse("Label cannot be empty")
    }

    // Check format support
    if !metadata.IsMetadataSupported(currentTrack.Path) {
        ext := filepath.Ext(currentTrack.Path)
        return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
    }

    // Write label
    if err := h.writer.WriteLabel(currentTrack.Path, label); err != nil {
        log.Printf("Failed to write label: %v", err)
        return shared.NewErrorResponse(fmt.Sprintf("Failed to tag label: %v", err))
    }

    log.Printf("Tagged '%s' with label: %s", currentTrack.Filename, label)

    return shared.NewSuccessResponse(
        fmt.Sprintf("🏷️  Label set to '%s': %s", label, currentTrack.Filename),
        nil,
    )
}
```

---

### 4. Shared Types Updates

#### `internal/shared/commands.go`

```go
// Add to CommandType constants
const (
    // ... existing commands
    CmdRate  CommandType = "rate"
    CmdLabel CommandType = "label"
)

// Add to Command struct
type Command struct {
    // ... existing fields
    Rating int    `json:"rating,omitempty"` // 1-5 rating
    Label  string `json:"label,omitempty"`  // Record label name
}

// Add constructor functions
func NewRateCommand(rating int) Command {
    return Command{
        Type:   CmdRate,
        Rating: rating,
    }
}

func NewLabelCommand(label string) Command {
    return Command{
        Type:  CmdLabel,
        Label: label,
    }
}
```

---

### 5. CLI Integration

#### `cmd/auxbox/cli.go`

**Add to usage string:**
```go
const usage = `auxbox - CLI music player for background listening

Usage:
  // ... existing commands ...
  auxbox rate <1-5>                    Rate current track (1-5)
  auxbox label "<label>"               Tag current track with record label
  // ... rest of commands ...

Examples:
  auxbox rate 5                        # Peak hour banger!
  auxbox rate 3                        # Solid track
  auxbox label "Drumcode"              # Tag with label
  auxbox label "Anjunadeep"
`
```

**Add to CLI.Run() switch:**
```go
func (c *CLI) Run(args []string) {
    // ... existing code ...

    switch command {
    // ... existing cases ...

    case "rate":
        c.handleRateCommand(args)
    case "label":
        c.handleLabelCommand(args)

    // ... rest of cases ...
    }
}

func (c *CLI) handleRateCommand(args []string) {
    if len(args) < 3 {
        fmt.Println("Usage: auxbox rate <1-5>")
        os.Exit(1)
    }

    rating, err := strconv.Atoi(args[2])
    if err != nil || rating < 1 || rating > 5 {
        fmt.Printf("Invalid rating: %s (must be 1-5)\n", args[2])
        os.Exit(1)
    }

    c.sendCommand(shared.NewRateCommand(rating))
}

func (c *CLI) handleLabelCommand(args []string) {
    if len(args) < 3 {
        fmt.Println("Usage: auxbox label \"<label name>\"")
        os.Exit(1)
    }

    label := args[2]
    if label == "" {
        fmt.Println("Label cannot be empty")
        os.Exit(1)
    }

    c.sendCommand(shared.NewLabelCommand(label))
}
```

---

### 6. Server Integration

#### `internal/server/server.go`

**Add metadata handler:**
```go
type Server struct {
    // ... existing fields ...
    metadataHandler *commands.MetadataHandler
}

func NewServer() *Server {
    // ... existing code ...

    server := &Server{
        // ... existing fields ...
        metadataHandler: commands.NewMetadataHandler(playlistObj),
    }

    return server
}

func (s *Server) handleCommand(cmd shared.Command) shared.Response {
    // ... existing cases ...

    case shared.CmdRate:
        return s.metadataHandler.HandleRate(cmd)
    case shared.CmdLabel:
        return s.metadataHandler.HandleLabel(cmd)

    // ... rest
}
```

---

## Testing Strategy

### Unit Tests

**`internal/metadata/metadata_test.go`:**
```go
func TestWriteRating(t *testing.T) {
    // Create test MP3 file
    // Write rating
    // Read back rating
    // Verify POPM frame value
}

func TestWriteLabel(t *testing.T) {
    // Create test MP3 file
    // Write label
    // Read back TPUB frame
    // Verify label text
}

func TestBackupRestore(t *testing.T) {
    // Simulate write failure
    // Verify backup restored
}

func TestFormatDetection(t *testing.T) {
    // Test MP3, AIFF, WAV detection
    // Verify support flags
}
```

**`internal/server/commands/metadata_test.go`:**
```go
func TestHandleRate(t *testing.T) {
    // Mock playlist with test track
    // Send rate command
    // Verify response
    // Verify file written
}
```

### Integration Tests

**Manual workflow test:**
```bash
# 1. Setup test environment
mkdir -p /tmp/auxbox-test
cp test-track.mp3 /tmp/auxbox-test/

# 2. Start auxbox with test folder
auxbox play -f /tmp/auxbox-test/

# 3. Rate track
auxbox rate 5
auxbox label "Test Label"

# 4. Verify with external tool
id3v2 -l /tmp/auxbox-test/test-track.mp3
# Should show: POPM frame with rating 255
# Should show: TPUB frame with "Test Label"

# 5. Import to Rekordbox/Mixxx
# Verify rating appears as 5 stars
```

### Compatibility Testing

**Rekordbox workflow:**
```
1. Rate 5 tracks in auxbox (auxbox rate 1-5)
2. Tag each with different labels
3. Open Rekordbox
4. File → Import Tracks → Select test folder
5. Verify:
   - Ratings appear correctly as stars (⭐ to ⭐⭐⭐⭐⭐)
   - Labels appear in "Label" column
   - Re-importing doesn't duplicate entries
```

**Mixxx workflow:**
```
1. Enable Preferences → Library → Sync track metadata to file tags
2. Rate tracks in auxbox (auxbox rate 1-5)
3. Right-click in Mixxx → Reload from File Tags
4. Verify ratings appear as stars
```

---

## Rollout Plan

### Phase 4.0: Core Implementation (Week 1-2)

- [ ] Add `github.com/bogem/id3v2` dependency
- [ ] Implement `internal/metadata/` package
  - [ ] writer.go (rating + label)
  - [ ] reader.go
  - [ ] backup.go
  - [ ] formats.go
- [ ] Add unit tests (target: >80% coverage)
- [ ] Update shared types (commands.go, types.go)

### Phase 4.1: Command Integration (Week 2)

- [ ] Implement MetadataHandler
- [ ] Add CLI commands (stars, label)
- [ ] Wire handlers in server
- [ ] Add command tests

### Phase 4.2: Testing & Validation (Week 3)

- [ ] Integration tests
- [ ] Rekordbox compatibility testing
- [ ] Mixxx compatibility testing
- [ ] Error handling validation
- [ ] Backup/restore testing

### Phase 4.3: Documentation (Week 3)

- [ ] Update DJ_WORKFLOW.md (mark Phase 4 complete)
- [ ] Update ROADMAP.md
- [ ] Add metadata examples to README
- [ ] Write Phase 4 release notes

### Phase 4.4: Release

- [ ] Merge to master
- [ ] Tag version v0.2.0
- [ ] Announce Phase 4 completion

---

## Future: AIFF Support (Phase 4.5+)

**Approach: Custom ID3 chunk library**

```go
// Future: internal/metadata/aiff_id3/
package aiff_id3

// WriteID3ChunkToAIFF writes ID3v2 data to AIFF file
func WriteID3ChunkToAIFF(filePath string, id3Data []byte) error {
    // 1. Read AIFF file structure
    // 2. Locate or create ID3  chunk (note space!)
    // 3. Write ID3v2 tag data to chunk
    // 4. Update FORM chunk size
    // 5. Write back to file atomically
}
```

**Research needed:**
- AIFF chunk structure (FORM, COMM, SSND, ID3 )
- ID3v2 tag serialization format
- Rekordbox AIFF+ID3 behavior
- Existing Go AIFF libraries for chunk manipulation

**Timeline:** After Phase 4 completion + community demand

---

## Compatibility Matrix

| Feature | MP3 | AIFF | WAV |
|---------|-----|------|-----|
| Track Rating | ✅ Phase 4 | 📋 Future | ❌ Not supported |
| Label Tag | ✅ Phase 4 | 📋 Future | ❌ Not supported |
| Rekordbox Sync | ✅ Phase 4 | 📋 Future | ❌ Not supported |
| Mixxx Sync | ✅ Phase 4 | 📋 Future | ❌ Not supported |

**Software compatibility:**
- ✅ Rekordbox 6.x, 7.x (MP3 ID3v2)
- ✅ Mixxx 2.3+ (MP3 ID3v2)
- ✅ MediaMonkey, MusicBee, foobar2000 (POPM standard)
- ⚠️ iTunes/Music.app (may use different rating scale)

---

## Open Questions

1. **Should we write multiple POPM frames?**
   - Option A: Only `auxbox@auxbox.org` (current plan)
   - Option B: Also write blank email for broader compatibility

2. **Backup cleanup strategy?**
   - Option A: Delete backup immediately after successful write
   - Option B: Keep backup until next write (allows manual recovery)

3. **Rating display in `auxbox status`?**
   - Should we read and display current rating in status output?
   - Performance impact of reading ID3 tags on every status call?

4. **Batch operations?**
   - Future: `auxbox stars 4 --tracks 1-5` (rate multiple tracks)?

---

## Success Metrics

**Phase 4 is complete when:**
- ✅ MP3 rating workflow works end-to-end
- ✅ Rekordbox import shows correct ratings
- ✅ Mixxx file tag sync works
- ✅ Unit test coverage >80%
- ✅ Zero data corruption in 100+ test writes
- ✅ Documentation updated
- ✅ User feedback positive (GitHub issues/discussions)

---

## References

- [ID3v2.3 Specification](http://id3.org/id3v2.3.0)
- [ID3v2.4 Specification](http://id3.org/id3v2.4.0-frames)
- [POPM Frame Format](http://id3.org/id3v2.3.0#Popularimeter)
- [bogem/id3v2 Library](https://github.com/bogem/id3v2)
- [Rekordbox Metadata Guide](docs/DJ_WORKFLOW.md#rekordbox-integration)
- [Mixxx TagLib Integration](https://github.com/mixxxdj/mixxx/wiki/Library-Metadata-Rewrite-Using-Taglib)
