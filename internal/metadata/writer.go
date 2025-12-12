package metadata

import (
	"fmt"
	"io"
	"math/big"
	"os"

	"github.com/bogem/id3v2/v2"
)

const (
	// Rating scale (1-5 rating → 1-255 POPM values → displayed as stars in DJ software)
	RatingScale1 = 1
	RatingScale2 = 64
	RatingScale3 = 128
	RatingScale4 = 196
	RatingScale5 = 255
)

// Writer interface for writing metadata to audio files
type Writer interface {
	WriteRating(filePath string, rating int) error
	WriteLabel(filePath string, label string) error
	WriteGenre(filePath string, genre string) error
	WriteTitle(filePath string, title string) error
	WriteArtist(filePath string, artist string) error
	WriteAlbum(filePath string, album string) error
	WriteYear(filePath string, year string) error
	WriteMetadata(filePath string, rating int, label string) error
}

// ID3Writer writes ID3v2 tags to MP3 files
type ID3Writer struct {
	backupEnabled bool
}

// NewWriter creates a new metadata writer
func NewWriter() Writer {
	return &ID3Writer{
		backupEnabled: true,
	}
}

// WriteRating writes rating to MP3 file (1-5 rating)
func (w *ID3Writer) WriteRating(filePath string, rating int) error {
	if rating < 1 || rating > 5 {
		return fmt.Errorf("invalid rating: %d (must be 1-5)", rating)
	}

	// Check format support
	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
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
		Email:   AuxboxPOPMEmail,
		Rating:  ratingValue,
		Counter: big.NewInt(0),
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

	// Cleanup backup on success
	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteLabel writes record label to TPUB frame
func (w *ID3Writer) WriteLabel(filePath string, label string) error {
	if label == "" {
		return fmt.Errorf("label cannot be empty")
	}

	// Check format support
	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

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

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteGenre writes genre to TCON frame
func (w *ID3Writer) WriteGenre(filePath string, genre string) error {
	if genre == "" {
		return fmt.Errorf("genre cannot be empty")
	}

	// Check format support
	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	// TCON = Content type (genre)
	tag.SetGenre(genre)

	if w.backupEnabled {
		if err := w.backupFile(filePath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if err := tag.Save(); err != nil {
		w.restoreBackup(filePath)
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteTitle writes track title to TIT2 frame
func (w *ID3Writer) WriteTitle(filePath string, title string) error {
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	tag.SetTitle(title)

	if w.backupEnabled {
		if err := w.backupFile(filePath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if err := tag.Save(); err != nil {
		w.restoreBackup(filePath)
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteArtist writes artist to TPE1 frame
func (w *ID3Writer) WriteArtist(filePath string, artist string) error {
	if artist == "" {
		return fmt.Errorf("artist cannot be empty")
	}

	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	tag.SetArtist(artist)

	if w.backupEnabled {
		if err := w.backupFile(filePath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if err := tag.Save(); err != nil {
		w.restoreBackup(filePath)
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteAlbum writes album to TALB frame
func (w *ID3Writer) WriteAlbum(filePath string, album string) error {
	if album == "" {
		return fmt.Errorf("album cannot be empty")
	}

	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	tag.SetAlbum(album)

	if w.backupEnabled {
		if err := w.backupFile(filePath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if err := tag.Save(); err != nil {
		w.restoreBackup(filePath)
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteYear writes year to TYER/TDRC frame
func (w *ID3Writer) WriteYear(filePath string, year string) error {
	if year == "" {
		return fmt.Errorf("year cannot be empty")
	}

	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	tag.SetYear(year)

	if w.backupEnabled {
		if err := w.backupFile(filePath); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if err := tag.Save(); err != nil {
		w.restoreBackup(filePath)
		return fmt.Errorf("failed to save tag: %w", err)
	}

	if w.backupEnabled {
		w.cleanupBackup(filePath)
	}

	return nil
}

// WriteMetadata writes both rating and label atomically
func (w *ID3Writer) WriteMetadata(filePath string, rating int, label string) error {
	// Check format support
	if !IsMetadataSupported(filePath) {
		return fmt.Errorf("metadata not supported for this file format (MP3 only)")
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open ID3 tag: %w", err)
	}
	defer tag.Close()

	// Write rating if provided
	if rating > 0 {
		if rating < 1 || rating > 5 {
			return fmt.Errorf("invalid rating: %d (must be 1-5)", rating)
		}

		ratingValue := ratingToPOPM(rating)
		popm := id3v2.PopularimeterFrame{
			Email:   AuxboxPOPMEmail,
			Rating:  ratingValue,
			Counter: big.NewInt(0),
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

	if w.backupEnabled {
		w.cleanupBackup(filePath)
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
