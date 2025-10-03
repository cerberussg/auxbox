package daemon

import (
	"fmt"
	"strings"
	"time"
)

// formatDuration converts time.Duration to MM:SS format
func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// getFileExtension returns the lowercase file extension
func getFileExtension(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return strings.ToLower(filePath[i:])
		}
		if filePath[i] == '/' || filePath[i] == '\\' {
			break
		}
	}
	return ""
}