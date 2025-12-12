# auxbox

A lightweight CLI music player for background listening with daemon architecture. Perfect for casual listening and DJ track preparation.

```bash
# One command from silence to sound
auxbox play -f ~/Music

# Hot-swap sources while playing
auxbox play -f ~/different-folder/

# Shuffle and repeat for infinite playback
auxbox play -f ~/Music/study-mix/ -s -r
```

## ✨ Features

- **Instant playback** - One command starts your music
- **Hot-swapping** - Switch sources seamlessly while playing
- **Shuffle & repeat** - Random playback with multiple repeat modes
- **DJ workflow** - Rate and tag tracks for rekordbox integration
- **Daemon architecture** - Background operation, persistent playback
- **Format support** - MP3, WAV, AIFF/AIF

## 🚀 Quick Start

### Install

```bash
git clone https://github.com/cerberussg/auxbox
cd auxbox
go build -o auxbox cmd/auxbox/*.go
mv auxbox ~/.local/bin/  # Or any directory in your PATH
```

See [INSTALLATION.md](docs/INSTALLATION.md) for platform-specific instructions.

### Basic Usage

```bash
# Load and play instantly
auxbox play -f ~/Music/jazz/

# Navigate tracks
auxbox skip      # Next track
auxbox back      # Previous track

# Controls
auxbox pause     # Pause playback
auxbox play      # Resume
auxbox shuffle   # Toggle shuffle
auxbox repeat    # Cycle repeat modes
auxbox volume 75 # Set volume

# Information
auxbox status    # Current track info
auxbox list      # Show all tracks

# Done listening
auxbox exit
```

## 📚 Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Comprehensive usage guide with examples
- **[DJ Workflow](docs/DJ_WORKFLOW.md)** - Track rating, genre tagging, rekordbox integration
- **[Installation](docs/INSTALLATION.md)** - Platform-specific setup instructions
- **[Architecture](docs/ARCHITECTURE.md)** - Technical documentation and design
- **[Roadmap](docs/ROADMAP.md)** - Development phases and future features
- **[Contributing](docs/CONTRIBUTING.md)** - How to contribute to the project

## 🎵 DJ Features

auxbox is perfect for DJs who want to organize tracks without opening heavy DJ software:

```bash
# Rate tracks while listening
auxbox rate 5                    # 1-5 star rating

# Tag complete metadata
auxbox label "Drumcode"          # Record label
auxbox genre "Techno"            # Genre
auxbox title "Peak Hour Tool"    # Track title
auxbox artist "DJ Name"          # Artist
auxbox album "EP Title"          # Album
auxbox year "2024"               # Release year

# View all metadata
auxbox status -d
```

All metadata syncs with Rekordbox and Mixxx using industry-standard ID3v2 tags. Works on MP3 files. See [DJ_WORKFLOW.md](docs/DJ_WORKFLOW.md) for details.

## 🛠️ Development Status

| Phase | Feature | Status |
|-------|---------|--------|
| Phase 1 | Streamlined UX | ✅ Complete |
| Phase 2 | Shuffle Mode | ✅ Complete |
| Phase 3 | Repeat Modes | ✅ Complete |
| Phase 4 | Metadata Editing | ✅ Complete |
| Phase 5 | Advanced Features | 📋 Planned |

**Phase 4** includes complete ID3v2 metadata editing: rating, label, genre, title, artist, album, and year.

See [ROADMAP.md](docs/ROADMAP.md) for the complete development plan.

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## 📝 License

[Add your license information here]

## 🔗 Links

- [GitHub Repository](https://github.com/cerberussg/auxbox)
- [Issues & Bug Reports](https://github.com/cerberussg/auxbox/issues)
- [Discussions](https://github.com/cerberussg/auxbox/discussions)