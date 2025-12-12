package commands

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/cerberussg/auxbox/internal/metadata"
	"github.com/cerberussg/auxbox/internal/playlist"
	"github.com/cerberussg/auxbox/internal/shared"
)

// MetadataHandler handles metadata operations (rating, labeling)
type MetadataHandler struct {
	playlist *playlist.Playlist
	writer   metadata.Writer
	reader   metadata.Reader
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(playlist *playlist.Playlist) *MetadataHandler {
	return &MetadataHandler{
		playlist: playlist,
		writer:   metadata.NewWriter(),
		reader:   metadata.NewReader(),
	}
}

// HandleRate rates the current track (1-5 stars)
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

	// Build star display
	stars := ""
	for i := 0; i < rating; i++ {
		stars += "⭐"
	}

	return shared.NewSuccessResponse(
		fmt.Sprintf("Rated %d/5 %s: %s", rating, stars, currentTrack.Filename),
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

// HandleGenre tags the current track with a genre
func (h *MetadataHandler) HandleGenre(cmd shared.Command) shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewErrorResponse("No track is currently playing")
	}

	genre := cmd.Genre
	if genre == "" {
		return shared.NewErrorResponse("Genre cannot be empty")
	}

	// Check format support
	if !metadata.IsMetadataSupported(currentTrack.Path) {
		ext := filepath.Ext(currentTrack.Path)
		return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
	}

	// Write genre
	if err := h.writer.WriteGenre(currentTrack.Path, genre); err != nil {
		log.Printf("Failed to write genre: %v", err)
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set genre: %v", err))
	}

	log.Printf("Tagged '%s' with genre: %s", currentTrack.Filename, genre)

	return shared.NewSuccessResponse(
		fmt.Sprintf("🎵 Genre set to '%s': %s", genre, currentTrack.Filename),
		nil,
	)
}

// HandleTitle sets the track title
func (h *MetadataHandler) HandleTitle(cmd shared.Command) shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewErrorResponse("No track is currently playing")
	}

	title := cmd.Title
	if title == "" {
		return shared.NewErrorResponse("Title cannot be empty")
	}

	if !metadata.IsMetadataSupported(currentTrack.Path) {
		ext := filepath.Ext(currentTrack.Path)
		return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
	}

	if err := h.writer.WriteTitle(currentTrack.Path, title); err != nil {
		log.Printf("Failed to write title: %v", err)
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set title: %v", err))
	}

	log.Printf("Set title '%s' for track: %s", title, currentTrack.Filename)

	return shared.NewSuccessResponse(
		fmt.Sprintf("📝 Title set to '%s': %s", title, currentTrack.Filename),
		nil,
	)
}

// HandleArtist sets the track artist
func (h *MetadataHandler) HandleArtist(cmd shared.Command) shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewErrorResponse("No track is currently playing")
	}

	artist := cmd.Artist
	if artist == "" {
		return shared.NewErrorResponse("Artist cannot be empty")
	}

	if !metadata.IsMetadataSupported(currentTrack.Path) {
		ext := filepath.Ext(currentTrack.Path)
		return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
	}

	if err := h.writer.WriteArtist(currentTrack.Path, artist); err != nil {
		log.Printf("Failed to write artist: %v", err)
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set artist: %v", err))
	}

	log.Printf("Set artist '%s' for track: %s", artist, currentTrack.Filename)

	return shared.NewSuccessResponse(
		fmt.Sprintf("🎤 Artist set to '%s': %s", artist, currentTrack.Filename),
		nil,
	)
}

// HandleAlbum sets the album name
func (h *MetadataHandler) HandleAlbum(cmd shared.Command) shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewErrorResponse("No track is currently playing")
	}

	album := cmd.Album
	if album == "" {
		return shared.NewErrorResponse("Album cannot be empty")
	}

	if !metadata.IsMetadataSupported(currentTrack.Path) {
		ext := filepath.Ext(currentTrack.Path)
		return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
	}

	if err := h.writer.WriteAlbum(currentTrack.Path, album); err != nil {
		log.Printf("Failed to write album: %v", err)
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set album: %v", err))
	}

	log.Printf("Set album '%s' for track: %s", album, currentTrack.Filename)

	return shared.NewSuccessResponse(
		fmt.Sprintf("💿 Album set to '%s': %s", album, currentTrack.Filename),
		nil,
	)
}

// HandleYear sets the release year
func (h *MetadataHandler) HandleYear(cmd shared.Command) shared.Response {
	currentTrack := h.playlist.GetCurrentTrack()
	if currentTrack == nil {
		return shared.NewErrorResponse("No track is currently playing")
	}

	year := cmd.Year
	if year == "" {
		return shared.NewErrorResponse("Year cannot be empty")
	}

	if !metadata.IsMetadataSupported(currentTrack.Path) {
		ext := filepath.Ext(currentTrack.Path)
		return shared.NewErrorResponse(fmt.Sprintf("Metadata not supported for %s files (MP3 only in Phase 4)", ext))
	}

	if err := h.writer.WriteYear(currentTrack.Path, year); err != nil {
		log.Printf("Failed to write year: %v", err)
		return shared.NewErrorResponse(fmt.Sprintf("Failed to set year: %v", err))
	}

	log.Printf("Set year '%s' for track: %s", year, currentTrack.Filename)

	return shared.NewSuccessResponse(
		fmt.Sprintf("📅 Year set to '%s': %s", year, currentTrack.Filename),
		nil,
	)
}
