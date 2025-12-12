package metadata

import (
	"fmt"

	"github.com/bogem/id3v2/v2"
)

const (
	// auxbox identifier for POPM frames
	AuxboxPOPMEmail = "auxbox@auxbox.org"
)

// Reader interface for reading metadata from audio files
type Reader interface {
	Read(filePath string) (*Metadata, error)
	ReadRating(filePath string) (int, error)
	ReadLabel(filePath string) (string, error)
}

// ID3Reader reads ID3v2 tags from MP3 files
type ID3Reader struct{}

// NewReader creates a new metadata reader
func NewReader() Reader {
	return &ID3Reader{}
}

// Read extracts all metadata from an audio file
func (r *ID3Reader) Read(filePath string) (*Metadata, error) {
	// Check if format is supported
	if !IsMetadataSupported(filePath) {
		return &Metadata{}, nil // Return empty metadata for unsupported formats
	}

	// Open ID3v2 tag
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		// If file has no ID3 tags, return empty metadata (not an error)
		return &Metadata{}, nil
	}
	defer tag.Close()

	meta := &Metadata{
		Title:   tag.Title(),
		Artist:  tag.Artist(),
		Album:   tag.Album(),
		Year:    tag.Year(),
		Genre:   tag.Genre(),
		Comment: r.readComment(tag),
		Label:   r.readLabel(tag),
		Rating:  r.readRating(tag),
	}

	return meta, nil
}

// ReadRating reads POPM frame and converts to 1-5 rating
func (r *ID3Reader) ReadRating(filePath string) (int, error) {
	if !IsMetadataSupported(filePath) {
		return 0, fmt.Errorf("metadata not supported for this file format")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return 0, nil // No tags = no rating
	}
	defer tag.Close()

	return r.readRating(tag), nil
}

// ReadLabel reads TPUB (Publisher) frame
func (r *ID3Reader) ReadLabel(filePath string) (string, error) {
	if !IsMetadataSupported(filePath) {
		return "", fmt.Errorf("metadata not supported for this file format")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return "", nil // No tags = no label
	}
	defer tag.Close()

	return r.readLabel(tag), nil
}

// readRating extracts rating from POPM frames
func (r *ID3Reader) readRating(tag *id3v2.Tag) int {
	// Get all POPM frames
	frames := tag.GetFrames(tag.CommonID("Popularimeter"))
	if len(frames) == 0 {
		return 0 // No rating set
	}

	// Look for auxbox POPM frame first
	for _, frame := range frames {
		popm, ok := frame.(id3v2.PopularimeterFrame)
		if ok && popm.Email == AuxboxPOPMEmail {
			return popmToRating(popm.Rating)
		}
	}

	// Fallback: use first POPM frame (from other software)
	if popm, ok := frames[0].(id3v2.PopularimeterFrame); ok {
		return popmToRating(popm.Rating)
	}

	return 0
}

// readLabel extracts label from TPUB (Publisher) frame
func (r *ID3Reader) readLabel(tag *id3v2.Tag) string {
	frames := tag.GetFrames(tag.CommonID("Publisher"))
	if len(frames) == 0 {
		return ""
	}

	if textFrame, ok := frames[0].(id3v2.TextFrame); ok {
		return textFrame.Text
	}

	return ""
}

// readComment extracts comment text
func (r *ID3Reader) readComment(tag *id3v2.Tag) string {
	frames := tag.GetFrames(tag.CommonID("Comments"))
	if len(frames) == 0 {
		return ""
	}

	if commentFrame, ok := frames[0].(id3v2.CommentFrame); ok {
		return commentFrame.Text
	}

	return ""
}

// popmToRating converts POPM rating value (0-255) to 1-5 rating
// Based on common mapping used by DJ software
func popmToRating(rating byte) int {
	switch {
	case rating == 0:
		return 0 // Unrated
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
