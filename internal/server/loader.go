package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cerberussg/auxbox/internal/shared"
)

// Loader handles loading tracks from various sources
type Loader struct{}

// NewLoader creates a new loader instance
func NewLoader() *Loader {
	return &Loader{}
}

// LoadFolder loads audio tracks from a folder
func (l *Loader) LoadFolder(folderPath string) ([]*shared.Track, error) {
	log.Printf("LoadFolder: Starting scan of %s", folderPath)

	// Supported audio extensions
	supportedExts := map[string]bool{
		".mp3":  true,
		".aiff": true,
		".aif":  true,
		".wav":  true, // For testing
	}

	var tracks []*shared.Track

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("LoadFolder: Walk error on %s: %v", path, err)
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip hidden files and macOS metadata files
		filename := info.Name()
		if strings.HasPrefix(filename, ".") || strings.HasPrefix(filename, "._") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedExts[ext] {
			track := &shared.Track{
				Filename: filename,
				Path:     path,
			}
			tracks = append(tracks, track)
			if len(tracks) <= 3 { // Log first few tracks for debugging
				log.Printf("LoadFolder: Found track %d: %s", len(tracks), track.Filename)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("LoadFolder: Walk failed: %v", err)
		return nil, err
	}

	log.Printf("LoadFolder: Found %d total tracks", len(tracks))
	return tracks, nil
}

// LoadPlaylist loads audio tracks from a playlist file
func (l *Loader) LoadPlaylist(playlistPath string) ([]*shared.Track, error) {
	// TODO: Implement playlist file parsing (M3U, PLS, etc.)
	// For now, just return an error
	return nil, fmt.Errorf("playlist loading not yet implemented")
}

// ExpandPath expands a file path (e.g., ~ to home directory)
func (l *Loader) ExpandPath(path string) (string, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not get home directory: %v", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path: %v", err)
	}

	return absPath, nil
}