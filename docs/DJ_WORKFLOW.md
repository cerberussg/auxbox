# DJ Workflow Guide

> **📋 Status:** This document describes planned DJ features. Phase 4 is in planning, Phases 5-6 are vision/future work. Command syntax and implementation details are subject to change.

auxbox doubles as a powerful DJ preparation tool, allowing you to rate, tag, and organize tracks while listening - perfect for preparing your music library without opening heavyweight DJ software.

## Feature Status Legend

- **✅ Implemented** - Feature is complete and available now
- **🚧 In Planning** - Feature is being designed for Phase 4, details may change
- **📋 Planned Vision** - Future phase concept, syntax and approach TBD

## Table of Contents

- [Overview](#overview)
- [Star Rating System](#star-rating-system) 🚧
- [Genre Tagging](#genre-tagging) 📋
- [Label Tracking](#label-tracking) 📋
- [Rekordbox Integration](#rekordbox-integration) 🚧
- [Complete DJ Workflows](#complete-dj-workflows) 📋
- [Energy Level Organization](#energy-level-organization) 📋

## Overview

As a DJ, you often need to:
- Preview new tracks and rate them for energy level
- Categorize tracks by genre/style
- Track which labels produce your favorite sounds
- Organize massive libraries efficiently
- Sync metadata with professional DJ software (rekordbox)

auxbox aims to provide a lightweight CLI interface to accomplish all of this during casual listening sessions, without the overhead of launching rekordbox or other DJ software.

## Star Rating System

**🚧 Status: Phase 4 - In Planning**

> **Note:** Command syntax below is proposed and subject to change during implementation.

Rate tracks on the fly while listening to build your energy-level system:

```bash
# Preview new tracks (✅ Available now)
auxbox play -f ~/new-tracks-pack/

# Rate the current track (1-5 stars) - 🚧 Planned command
auxbox stars 5    # Peak-hour banger
auxbox stars 4    # High energy, main set material
auxbox stars 3    # Solid track, versatile
auxbox stars 2    # Good opener/breakdown track
auxbox stars 1    # Low energy, intro/outro material

# Skip to next track and continue rating
auxbox skip       # ✅ Available now
auxbox stars 4    # 🚧 Planned command
```

### Rating Strategy

**Energy-level system (recommended):**
- ⭐⭐⭐⭐⭐ (5 stars) - Peak hour bangers, maximum energy
- ⭐⭐⭐⭐ (4 stars) - High energy main set tracks
- ⭐⭐⭐ (3 stars) - Versatile, mid-energy tracks
- ⭐⭐ (2 stars) - Warm-up, openers, breakdown tracks
- ⭐ (1 star) - Intro/outro, ambient, low energy

**Alternative strategies:**
- **Personal preference** - How much you like the track
- **Crowd response** - Historical response from crowds
- **Track quality** - Production quality rating
- **Mixing difficulty** - How easy it is to mix

Choose a consistent system and stick with it across your library.

## Genre Tagging

**📋 Status: Phase 5 - Planned Vision**

> **Note:** This feature is planned for a future phase. Command syntax and implementation approach are not yet designed.

Categorize tracks by style during preview sessions:

```bash
# Tag genres while listening - 📋 Future concept
auxbox genre "Deep House"
auxbox genre "Tech House"
auxbox genre "Progressive Trance"
auxbox genre "Melodic Techno"
auxbox genre "Minimal Tech"
```

### Genre Organization Benefits

- **Style-based playlists** - Quickly find tracks matching the vibe
- **Set preparation** - Filter by genre for specific gigs
- **Library discovery** - Understand your collection's diversity
- **Trend tracking** - See which styles you're collecting most

### Common Electronic Music Genres

**House:**
- Deep House, Tech House, Progressive House, Electro House, Future House, Tropical House

**Techno:**
- Peak Time Techno, Melodic Techno, Minimal Techno, Industrial Techno, Acid Techno

**Trance:**
- Progressive Trance, Uplifting Trance, Psytrance, Tech Trance

**Other:**
- Drum & Bass, Dubstep, Trap, Future Bass, Garage, UK Bass

## Label Tracking

**📋 Status: Phase 6 - Planned Vision**

> **Note:** This feature is planned for a future phase. Command syntax and implementation approach are not yet designed.

Track the source/label for discovery and organization:

```bash
# Tag record labels while listening - 📋 Future concept
auxbox label "Defected Records"
auxbox label "Anjunadeep"
auxbox label "Drumcode"
auxbox label "Hot Creations"
auxbox label "Toolroom"
```

### Label Tracking Benefits

- **Source discovery** - Know which labels produce tracks you love
- **Release tracking** - Follow your favorite labels for new music
- **Style consistency** - Labels often have consistent sound aesthetics
- **Networking** - Identify labels to send demos or collaborate with

## Rekordbox Integration

**🚧 Status: Phase 4 Planning**

> **Note:** Integration strategy is being designed. Implementation details below are proposed approaches.

All metadata written by auxbox will use industry-standard ID3v2 tags that rekordbox reads natively.

### Metadata Field Mapping

| auxbox Feature | ID3v2 Tag | rekordbox Field | Phase |
|----------------|-----------|-----------------|-------|
| Star Rating    | POPM      | Rating (stars)  | Phase 4 (In Development) |
| Genre          | TCON      | Genre           | Phase 5 (Planned) |
| Label          | TPUB      | Label           | Phase 6 (Planned) |

### How Integration Works

1. **auxbox writes to audio files** - Metadata is embedded directly in MP3/WAV/AIFF files
2. **rekordbox reads on import** - When you import tracks, rekordbox detects existing metadata
3. **No duplicate work** - Your ratings, genres, and labels appear automatically

### Rekordbox Database Challenge

**Important Technical Note:**

rekordbox stores star ratings in two places:
1. **ID3v2 POPM frame** - In the audio file itself (portable)
2. **rekordbox database** - In rekordbox's internal SQLite database (fast access)

auxbox can write to ID3v2 tags, but directly writing to the rekordbox database poses challenges:
- Database schema is proprietary and undocumented
- Risk of database corruption if schema changes between versions
- Database locking issues when rekordbox is running
- Loss of portability (ratings tied to one rekordbox installation)

### Integration Strategy

**Phase 4 will implement:**
- ID3v2 POPM frame writing (standardized metadata)
- Rekordbox import detection (triggers rekordbox to read POPM tags)
- Compatibility testing with rekordbox 6.x

**Workflow:**
1. Rate tracks in auxbox (writes to ID3v2)
2. Import/re-import tracks in rekordbox
3. rekordbox reads ratings from ID3v2 tags
4. Ratings appear in rekordbox interface

**Future consideration:**
- Research rekordbox XML export/import as alternative sync method
- Investigate rekordbox API if officially documented

## Complete DJ Workflows

**📋 Status: Aspirational workflows showing planned features**

> **Note:** Workflows below combine implemented features (✅) with planned commands (🚧📋). Full workflow will be available once all phases are complete.

### New Promo Pack Evaluation

```bash
# Load new promo pack (✅ Available now)
auxbox play -f ~/promos/december-2024/

# Listen and rate each track
auxbox status                    # ✅ Check current track
auxbox stars 4                   # 🚧 Rate it (Phase 4)
auxbox genre "Deep House"        # 📋 Tag genre (Phase 5)
auxbox label "Hot Creations"     # 📋 Tag label (Phase 6)

auxbox skip                      # ✅ Next track
auxbox stars 2                   # 🚧 Opener material (Phase 4)
auxbox genre "Minimal Tech"      # 📋 (Phase 5)
auxbox label "Percomaniacs"      # 📋 (Phase 6)

auxbox skip 3                    # ✅ Jump ahead to interesting track
auxbox stars 5                   # 🚧 Peak hour material (Phase 4)
auxbox genre "Peak Time Techno"  # 📋 (Phase 5)
auxbox label "Drumcode"          # 📋 (Phase 6)

# When done (✅ Available now)
auxbox exit
```

### Large Library Organization

```bash
# Load entire library with shuffle (✅ Available now)
auxbox play -f ~/Music/complete-library/ -s -r

# Rate tracks as they play randomly - 🚧 Phase 4
# Perfect for background work while organizing
auxbox stars 3
auxbox skip     # ✅ Available now
auxbox stars 5
auxbox skip

# Progress through thousands of tracks over multiple sessions
```

### Style-Specific Crate Preparation

```bash
# Load specific style folder (✅ Available now)
auxbox play -f ~/Music/techno/ -s

# Rate and tag for sub-genre classification - 🚧📋 Phases 4-5
auxbox stars 5
auxbox genre "Peak Time Techno"
auxbox skip     # ✅ Available now

auxbox stars 4
auxbox genre "Melodic Techno"
auxbox skip

# Build up detailed genre metadata
# Creates organized sub-crates in rekordbox
```

### Pre-Gig Track Selection

```bash
# Load recently downloaded tracks (✅ Available now)
auxbox play -f ~/Downloads/new-tracks/

# Quickly rate for tonight's gig - 🚧 Phase 4
auxbox stars 5    # Definitely playing this
auxbox stars 3    # Maybe if vibe is right
auxbox stars 1    # Not for tonight

# Later in rekordbox: filter by 4-5 stars
# Instant shortlist for your set
```

## Energy Level Organization

**🚧 Status: Phase 4 - Conceptual guide for planned feature**

> **Note:** This section describes how to use the star rating system once Phase 4 is implemented.

### Building Your Star Rating System

**Session 1: Initial Pass**
- Listen through collection casually
- Rate instinctively: Would I play this at peak hour? (5 stars)
- Don't overthink, trust your gut

**Session 2: Refinement**
- Filter by unrated tracks
- Use shuffle to randomly discover forgotten gems
- Compare similar tracks, adjust ratings for consistency

**Session 3: Context Rating**
- Think about specific gigs/venues
- Adjust ratings based on:
  - Time of night (warm-up vs. peak hour)
  - Venue type (club vs. festival)
  - Crowd demographics

### Using Ratings in rekordbox

Once synced, you can:
- **Smart playlists** - Auto-generate playlists by star rating
- **Preparation playlists** - Filter 4-5 star tracks for gigs
- **Discovery** - Find high-rated tracks you haven't played recently
- **Crate organization** - Create energy-level based crates

### Example Rating Guidelines

**5-Star Peak Hour Techno:**
- 128-132 BPM
- High energy, driving basslines
- Crowd-tested, guaranteed floor fillers
- Limited quantity (only your absolute best)

**4-Star Main Set Material:**
- Solid production quality
- Versatile, works in multiple contexts
- Good for building energy

**3-Star Utility Tracks:**
- Good for transitions
- Mixing tools, doubles, edits
- Genre-flexible tracks

**2-Star Warm-Up:**
- Lower energy, groovy
- Good for building atmosphere
- Opening sets, early night

**1-Star Ambient/Outro:**
- Low energy, atmospheric
- Closing tracks, comedown material
- Special moments, not for dance floor

## Tips for DJs

**🚧📋 Status: Best practices for planned features**

> **Note:** These tips will apply once Phase 4-6 features are implemented.

### Efficient Metadata Sessions

- **Time-box sessions** - 30-60 minute focused rating sessions
- **Use shuffle** - Discover forgotten tracks randomly (✅ shuffle available now)
- **Background rating** - Rate while working/coding (🚧 Phase 4)
- **Batch processing** - Rate entire genre folders in one go (🚧 Phase 4)

### Consistency is Key

- **Document your system** - Write down what each star level means
- **Regular calibration** - Periodically review and adjust ratings
- **Compare tracks** - When uncertain, compare to already-rated tracks
- **Trust the process** - Initial ratings may be rough, they'll improve

### Integration with Existing Workflow

**📋 Aspirational workflow - shows how phases 4-6 will fit into DJ prep**

auxbox will complement your existing DJ workflow:

1. **Download tracks** (Beatport, Bandcamp, promos) - ✅ Current
2. **Preview in auxbox** (rate, tag, organize) - 🚧📋 Phases 4-6
3. **Import to rekordbox** (metadata appears automatically) - 🚧 Phase 4
4. **Analyze in rekordbox** (beatgrids, waveforms, cue points) - ✅ Current
5. **Create playlists** (use ratings for smart playlists) - 🚧 After Phase 4
6. **Prepare sets** (filter by rating and genre) - 🚧📋 After Phases 4-5
7. **Play gigs** (refined, organized library) - ✅ Current

## Future Features

See [ROADMAP.md](ROADMAP.md) for planned DJ features:
- **BPM detection** - Automatic tempo analysis
- **Key detection** - Harmonic mixing support
- **Crate management** - DJ-style folder organization
- **Smart playlists** - Auto-generated based on metadata
- **rekordbox XML export** - Alternative sync method

## Next Steps

- Start rating your library with Phase 4 (coming soon)
- See [USER_GUIDE.md](USER_GUIDE.md) for basic playback features
- Check [ROADMAP.md](ROADMAP.md) for Phase 4 progress
