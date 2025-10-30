# Contributing to auxbox

Thank you for considering contributing to auxbox! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Development Setup](#development-setup)

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Respect differing opinions and experiences
- Accept responsibility for mistakes

### Unacceptable Behavior

- Harassment, discrimination, or offensive comments
- Personal attacks or trolling
- Spam or promotional content
- Publishing others' private information

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.25.1+** installed ([download](https://golang.org/dl/))
- **Git** for version control
- **GCC** (for CGO compilation)
- **Linux or macOS** (Windows support coming in future phases)
- Familiarity with Go and audio programming (helpful but not required)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/auxbox
cd auxbox

# Add upstream remote
git remote add upstream https://github.com/cerberussg/auxbox
```

### Build and Test

```bash
# Install dependencies
go mod download

# Build
go build -o auxbox cmd/auxbox/*.go

# Run tests
go test ./...

# Try it out
./auxbox play -f ~/Music
```

## Development Workflow

### Branch Strategy

- `master` - Stable, production-ready code
- `phase_X` - Feature branches for roadmap phases (e.g., `phase_4`)
- `bug_*` - Bug fix branches (e.g., `bug_list`, `bug_volume`)
- `feature_*` - Individual feature branches

### Creating a Branch

```bash
# Update your fork
git checkout master
git pull upstream master

# Create feature branch
git checkout -b feature_your_feature_name

# Or bug fix branch
git checkout -b bug_description_of_bug
```

### Making Changes

```bash
# Make your changes
# ...

# Stage changes
git add .

# Commit with descriptive message
git commit -m "Add feature: brief description"

# Push to your fork
git push origin feature_your_feature_name
```

### Keeping Your Branch Updated

```bash
# Fetch upstream changes
git fetch upstream

# Rebase on master
git rebase upstream/master

# Force push if needed (only on your branch!)
git push origin feature_your_feature_name --force
```

## Coding Standards

### Go Style Guidelines

Follow [Effective Go](https://golang.org/doc/effective_go.html) and these project-specific rules:

#### Formatting

```bash
# Format all code with gofmt
gofmt -w .

# Or use goimports (preferred)
goimports -w .
```

#### Naming Conventions

```go
// Good: Descriptive, exported function names
func LoadPlaylistFromFolder(path string) (*Playlist, error)

// Bad: Vague, unclear names
func load(p string) (*Playlist, error)

// Good: Private functions with clear intent
func parseM3UFile(path string) ([]string, error)

// Good: Descriptive variable names
currentTrackIndex := 0
isShuffleEnabled := false

// Bad: Single-letter or unclear names
i := 0  // What does 'i' represent?
flag := false  // What flag?
```

#### Error Handling

```go
// Good: Explicit error checking
file, err := os.Open(path)
if err != nil {
    return nil, fmt.Errorf("failed to open file: %w", err)
}
defer file.Close()

// Bad: Ignoring errors
file, _ := os.Open(path)

// Good: Wrapped errors with context
return fmt.Errorf("loading playlist from %s: %w", path, err)
```

#### Documentation

```go
// Good: Exported functions have godoc comments
// LoadPlaylistFromFolder loads all supported audio files from the specified
// directory path. It returns a Playlist containing the tracks in alphabetical
// order, or an error if the directory cannot be read.
func LoadPlaylistFromFolder(path string) (*Playlist, error) {
    // ...
}

// Good: Complex logic has inline comments
// Shuffle uses Fisher-Yates algorithm to ensure each track
// is selected exactly once before reshuffling
func (p *Playlist) Shuffle() {
    // ...
}
```

#### Package Organization

- Keep packages focused and cohesive
- Minimize dependencies between packages
- Use internal/ for non-exported packages
- Group related functionality together

### Project-Specific Conventions

#### Command Structure

New commands should follow this pattern:

```go
// In internal/server/commands/
package commands

import "github.com/cerberussg/auxbox/internal/shared"

func HandleNewCommand(req *shared.Request, deps *Dependencies) *shared.Response {
    // 1. Validate input
    if req.Args["param"] == "" {
        return shared.ErrorResponse("param is required")
    }

    // 2. Perform operation
    result, err := doSomething(req.Args["param"])
    if err != nil {
        return shared.ErrorResponse(fmt.Sprintf("failed: %v", err))
    }

    // 3. Return success response
    return shared.SuccessResponse(fmt.Sprintf("✓ Operation completed: %s", result))
}
```

#### Testing Patterns

```go
func TestNewFeature(t *testing.T) {
    // Arrange - Set up test data
    playlist := NewPlaylist()
    playlist.Add("track1.mp3")

    // Act - Execute the operation
    err := playlist.Shuffle()

    // Assert - Verify results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(playlist.Tracks) != 1 {
        t.Errorf("expected 1 track, got %d", len(playlist.Tracks))
    }
}
```

## Testing Requirements

### Test Coverage

- **New features** must include tests
- **Bug fixes** should include regression tests
- **Minimum coverage** - Aim for 70%+ on new code

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/playlist

# Run with verbose output
go test -v ./internal/audio

# Run specific test
go test -run TestShuffle ./internal/playlist
```

### Writing Tests

**Good test characteristics:**
- **Fast** - Tests should run in milliseconds
- **Isolated** - No dependencies on external state
- **Repeatable** - Same input always produces same output
- **Comprehensive** - Cover edge cases and error paths

**Example test:**

```go
func TestPlaylistShuffle(t *testing.T) {
    tests := []struct {
        name     string
        tracks   []string
        expected int
    }{
        {"empty playlist", []string{}, 0},
        {"single track", []string{"track1.mp3"}, 1},
        {"multiple tracks", []string{"a.mp3", "b.mp3", "c.mp3"}, 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewPlaylist()
            for _, track := range tt.tracks {
                p.Add(track)
            }

            p.Shuffle()

            if len(p.Tracks) != tt.expected {
                t.Errorf("expected %d tracks, got %d", tt.expected, len(p.Tracks))
            }
        })
    }
}
```

### Integration Tests

For features that interact with files or audio:

```go
func TestLoadPlaylistFromFolder(t *testing.T) {
    // Create temp directory with test files
    tmpDir := t.TempDir()
    createTestFile(t, tmpDir, "track1.mp3")
    createTestFile(t, tmpDir, "track2.mp3")

    // Test loading
    playlist, err := LoadPlaylistFromFolder(tmpDir)
    if err != nil {
        t.Fatalf("failed to load playlist: %v", err)
    }

    if len(playlist.Tracks) != 2 {
        t.Errorf("expected 2 tracks, got %d", len(playlist.Tracks))
    }
}
```

## Pull Request Process

### Before Submitting

- [ ] Code follows project style guidelines
- [ ] All tests pass: `go test ./...`
- [ ] Code is formatted: `gofmt -w .`
- [ ] New features have tests
- [ ] Documentation is updated (if applicable)
- [ ] Commit messages are clear and descriptive

### Creating a Pull Request

1. **Push your branch** to your fork
2. **Open PR** on GitHub against `cerberussg/auxbox:master`
3. **Fill out PR template** with:
   - Description of changes
   - Related issue (if applicable)
   - Testing performed
   - Screenshots (for UI changes)

### PR Title Format

```
Add: Brief description of new feature
Fix: Brief description of bug fix
Refactor: Brief description of refactoring
Docs: Brief description of documentation changes
Test: Brief description of test additions/changes
```

**Examples:**
- `Add: Star rating command for DJ workflow (Phase 4)`
- `Fix: Volume control not persisting between tracks`
- `Refactor: Simplify playlist shuffle algorithm`
- `Docs: Update installation guide for Arch Linux`
- `Test: Add integration tests for folder loading`

### PR Description Template

```markdown
## Summary
Brief description of what this PR does

## Changes
- Bullet list of specific changes
- Made to the codebase

## Testing
- How you tested these changes
- Manual testing performed
- Automated tests added

## Related Issue
Closes #123 (if applicable)

## Screenshots
(if applicable)
```

### Review Process

1. **Automated checks** run (tests, linting)
2. **Code review** by maintainers
3. **Address feedback** if requested
4. **Approval** from maintainer
5. **Merge** to master

### After Merge

- Delete your feature branch
- Pull latest master: `git pull upstream master`
- Celebrate! 🎉

## Reporting Bugs

### Before Reporting

- Check if the bug is already reported (search issues)
- Ensure you're using the latest version
- Gather detailed information about the bug

### Bug Report Template

```markdown
## Description
Clear description of the bug

## Steps to Reproduce
1. Run command: auxbox play -f ~/music
2. Observe behavior
3. See error

## Expected Behavior
What you expected to happen

## Actual Behavior
What actually happened

## Environment
- OS: Linux / macOS / Windows
- OS Version: Ubuntu 22.04 / macOS 13 / etc
- auxbox Version: git commit hash or version number
- Go Version: go version output

## Additional Context
Any other relevant information, logs, or screenshots
```

### Creating an Issue

1. Go to [Issues](https://github.com/cerberussg/auxbox/issues)
2. Click "New Issue"
3. Select "Bug Report" template
4. Fill out all sections
5. Add appropriate labels (bug, urgent, etc.)

## Suggesting Features

### Feature Request Template

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Who would benefit and how?

## Proposed Implementation
(Optional) Your ideas on how to implement this

## Alternatives Considered
(Optional) Other approaches you've thought about

## Additional Context
Any mockups, examples, or relevant information
```

### Roadmap Alignment

Check [ROADMAP.md](ROADMAP.md) to see if your feature aligns with planned phases:
- **Phase 4:** Star rating system
- **Phase 5:** Genre tagging
- **Phase 6:** Label tracking
- **Future:** Other enhancements

If your feature fits an existing phase, mention it in your proposal.

## Development Setup

### Recommended Tools

**Editor/IDE:**
- Visual Studio Code with Go extension
- GoLand
- Vim/Neovim with vim-go

**Useful Go Tools:**
```bash
# Install helpful tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/godoc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Linting:**
```bash
# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o auxbox cmd/auxbox/*.go

# Run with delve debugger
dlv exec ./auxbox -- play -f ~/Music
```

### Local Development Workflow

```bash
# Terminal 1: Build and run
go build -o auxbox cmd/auxbox/*.go && ./auxbox play -f ~/Music

# Terminal 2: Monitor daemon logs (if logging is enabled)
tail -f /tmp/auxbox.log

# Terminal 3: Test commands
./auxbox status
./auxbox skip
./auxbox volume 75
```

## Questions?

- Open a [Discussion](https://github.com/cerberussg/auxbox/discussions)
- Join our community chat (if available)
- Ask in an issue (for project-specific questions)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

## Thank You!

Your contributions make auxbox better for everyone. We appreciate your time and effort! 🎵
