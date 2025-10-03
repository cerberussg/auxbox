package daemon

import (
	"fmt"
	"io"

	"github.com/gopxl/beep/v2"
)

// DecoderFunc represents a function that can decode audio files
type DecoderFunc func(io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error)

// FormatRegistry manages audio format decoders
type FormatRegistry struct {
	decoders map[string]DecoderFunc
}

// NewFormatRegistry creates a new format registry with default decoders
func NewFormatRegistry() *FormatRegistry {
	registry := &FormatRegistry{
		decoders: make(map[string]DecoderFunc),
	}

	// Register default decoders
	registry.RegisterDecoder(".mp3", DecodeMP3)
	registry.RegisterDecoder(".wav", DecodeWAV)
	registry.RegisterDecoder(".aiff", DecodeAIFF)
	registry.RegisterDecoder(".aif", DecodeAIFF)

	return registry
}

// RegisterDecoder registers a decoder for a file extension
func (r *FormatRegistry) RegisterDecoder(ext string, decoder DecoderFunc) {
	r.decoders[ext] = decoder
}

// Decode decodes an audio file based on its extension
func (r *FormatRegistry) Decode(filePath string, file io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) {
	ext := getFileExtension(filePath)

	decoder, exists := r.decoders[ext]
	if !exists {
		return nil, beep.Format{}, fmt.Errorf("unsupported audio format: %s", ext)
	}

	return decoder(file)
}

// GetSupportedExtensions returns a list of supported file extensions
func (r *FormatRegistry) GetSupportedExtensions() []string {
	extensions := make([]string, 0, len(r.decoders))
	for ext := range r.decoders {
		extensions = append(extensions, ext)
	}
	return extensions
}