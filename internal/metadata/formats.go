package metadata

import (
	"path/filepath"
	"strings"
)

// AudioFormat represents supported audio formats
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

// IsMetadataSupported returns whether metadata reading is supported for this file
func IsMetadataSupported(filePath string) bool {
	format := DetectFormat(filePath)
	// Phase 4: Only MP3 supported
	return format == FormatMP3
}
