package decoders

import (
	"fmt"
	"io"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/gopxl/beep/v2"
)

const (
	// Buffer size for streaming AIFF data (in samples per channel)
	// This determines how much data we keep in memory at once
	aiffBufferSize = 4096 // ~93ms at 44.1kHz, uses ~32KB for stereo float64
)

// aiffStreamer implements beep.StreamSeekCloser for AIFF files
type aiffStreamer struct {
	decoder       *aiff.Decoder
	format        beep.Format
	reader        io.ReadSeeker

	// Streaming buffers
	rawBuffer     *audio.IntBuffer // Raw PCM data buffer for chunks
	sampleBuffer  [][2]float64     // Converted samples ready for playback
	bufferPos     int              // Current position in sampleBuffer
	totalSamples  int              // Total samples in the file
	currentSample int              // Current sample position in file

	// Format info for conversion
	sourceBitDepth int
	maxValue       float64
}

// DecodeAIFF creates a beep-compatible streamer from an AIFF file
func DecodeAIFF(r io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) {
	// AIFF decoder needs ReadSeeker, so we need to convert
	readSeeker, ok := r.(io.ReadSeeker)
	if !ok {
		return nil, beep.Format{}, fmt.Errorf("AIFF decoder requires ReadSeeker interface")
	}

	decoder := aiff.NewDecoder(readSeeker)

	if !decoder.IsValidFile() {
		return nil, beep.Format{}, fmt.Errorf("invalid AIFF file")
	}

	decoder.ReadInfo() // ReadInfo doesn't return an error

	format := decoder.Format()
	if format == nil {
		return nil, beep.Format{}, fmt.Errorf("could not get AIFF format")
	}

	// Convert to beep format
	beepFormat := beep.Format{
		SampleRate:  beep.SampleRate(format.SampleRate),
		NumChannels: format.NumChannels,
		Precision:   4, // 32-bit samples
	}

	// Calculate total samples in the file
	totalSamples := int(decoder.NumSampleFrames)

	// Setup bit depth and normalization values
	sourceBitDepth := int(decoder.SampleBitDepth())
	var maxValue float64
	switch sourceBitDepth {
	case 8:
		maxValue = float64(1 << 7)  // 128
	case 16:
		maxValue = float64(1 << 15) // 32768
	case 24:
		maxValue = float64(1 << 23) // 8388608
	case 32:
		maxValue = float64(1 << 31) // 2147483648
	case 64:
		maxValue = float64(1 << 63) // For int64
	default:
		maxValue = float64(1 << 15) // Default to 16-bit
	}

	streamer := &aiffStreamer{
		decoder:        decoder,
		format:         beepFormat,
		reader:         readSeeker,
		sampleBuffer:   make([][2]float64, aiffBufferSize),
		bufferPos:      0,
		totalSamples:   totalSamples,
		currentSample:  0,
		sourceBitDepth: sourceBitDepth,
		maxValue:       maxValue,
	}

	// Load initial buffer
	if err := streamer.fillBuffer(); err != nil {
		return nil, beep.Format{}, fmt.Errorf("failed to load initial AIFF buffer: %w", err)
	}

	return streamer, beepFormat, nil
}

// fillBuffer reads a chunk of audio data from the file and converts it to float64 samples
func (s *aiffStreamer) fillBuffer() error {
	if s.currentSample >= s.totalSamples {
		return fmt.Errorf("end of file reached")
	}

	// Calculate how many samples to read (don't exceed file bounds)
	samplesToRead := aiffBufferSize
	if s.currentSample+samplesToRead > s.totalSamples {
		samplesToRead = s.totalSamples - s.currentSample
	}

	// Create a buffer for this chunk
	s.rawBuffer = &audio.IntBuffer{
		Data:   make([]int, samplesToRead*s.format.NumChannels),
		Format: s.decoder.Format(),
	}

	// Read PCM data for this chunk
	n, err := s.decoder.PCMBuffer(s.rawBuffer)
	if err != nil {
		return fmt.Errorf("failed to read PCM chunk: %w", err)
	}

	if n == 0 {
		return fmt.Errorf("no data read from decoder")
	}

	// Convert the raw samples to float64 format
	actualSamples := n / s.format.NumChannels
	if actualSamples > len(s.sampleBuffer) {
		actualSamples = len(s.sampleBuffer)
	}

	for i := 0; i < actualSamples; i++ {
		if s.format.NumChannels == 1 {
			// Mono: duplicate to both channels
			sample := float64(s.rawBuffer.Data[i]) / s.maxValue
			s.sampleBuffer[i] = [2]float64{sample, sample}
		} else if s.format.NumChannels >= 2 {
			// Stereo or more: take first two channels
			left := float64(s.rawBuffer.Data[i*s.format.NumChannels]) / s.maxValue
			right := float64(s.rawBuffer.Data[i*s.format.NumChannels+1]) / s.maxValue
			s.sampleBuffer[i] = [2]float64{left, right}
		}
	}

	// Reset buffer position and update file position
	s.bufferPos = 0
	s.currentSample += actualSamples

	return nil
}

func (s *aiffStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.currentSample >= s.totalSamples {
		return 0, false // End of file
	}

	totalCopied := 0
	requestedSamples := len(samples)

	for totalCopied < requestedSamples && s.currentSample < s.totalSamples {
		// Check if we need to refill the buffer
		if s.bufferPos >= aiffBufferSize || (s.bufferPos > 0 && s.bufferPos >= s.totalSamples-s.currentSample+s.bufferPos) {
			if err := s.fillBuffer(); err != nil {
				// If we can't fill buffer but have copied some samples, return what we have
				if totalCopied > 0 {
					return totalCopied, true
				}
				return 0, false
			}
		}

		// Calculate how many samples to copy from current buffer
		remainingInRequest := requestedSamples - totalCopied
		remainingInBuffer := aiffBufferSize - s.bufferPos
		samplesRemaining := s.totalSamples - (s.currentSample - (aiffBufferSize - s.bufferPos))

		toCopy := remainingInRequest
		if toCopy > remainingInBuffer {
			toCopy = remainingInBuffer
		}
		if toCopy > samplesRemaining {
			toCopy = samplesRemaining
		}

		// Copy samples from buffer
		copy(samples[totalCopied:totalCopied+toCopy], s.sampleBuffer[s.bufferPos:s.bufferPos+toCopy])

		s.bufferPos += toCopy
		totalCopied += toCopy
	}

	return totalCopied, totalCopied > 0
}

func (s *aiffStreamer) Err() error {
	return nil
}

func (s *aiffStreamer) Len() int {
	return s.totalSamples
}

func (s *aiffStreamer) Position() int {
	// Return current position in the file (samples played so far)
	return s.currentSample - (aiffBufferSize - s.bufferPos)
}

func (s *aiffStreamer) Seek(p int) error {
	if p < 0 || p >= s.totalSamples {
		return fmt.Errorf("seek position out of range: %d (total: %d)", p, s.totalSamples)
	}

	// Seek in the underlying reader
	// Calculate byte position for seeking (this is approximate)
	bytesPerSample := s.sourceBitDepth / 8
	bytePos := int64(p * s.format.NumChannels * bytesPerSample)

	if _, err := s.reader.Seek(bytePos, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek in AIFF file: %w", err)
	}

	// Update our position tracking
	s.currentSample = p
	s.bufferPos = aiffBufferSize // Force buffer refill on next Stream call

	return nil
}

func (s *aiffStreamer) Close() error {
	// Clean up buffers to free memory
	s.rawBuffer = nil
	s.sampleBuffer = nil
	// The decoder doesn't need explicit closing in go-audio/aiff
	return nil
}