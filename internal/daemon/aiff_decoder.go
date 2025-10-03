package daemon

import (
	"fmt"
	"io"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/gopxl/beep/v2"
)

// aiffStreamer implements beep.StreamSeekCloser for AIFF files
type aiffStreamer struct {
	decoder *aiff.Decoder
	format  beep.Format
	buffer  *audio.IntBuffer
	pos     int
	samples [][2]float64
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

	// Debug: Log AIFF format info
	fmt.Printf("AIFF Format - Sample Rate: %d, Channels: %d\n",
		format.SampleRate, format.NumChannels)

	// Convert to beep format
	beepFormat := beep.Format{
		SampleRate:  beep.SampleRate(format.SampleRate),
		NumChannels: format.NumChannels,
		Precision:   4, // 32-bit samples
	}

	// Read the entire PCM buffer for now (could be optimized for streaming later)
	pcmBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("failed to read AIFF PCM data: %w", err)
	}

	streamer := &aiffStreamer{
		decoder: decoder,
		format:  beepFormat,
		buffer:  pcmBuffer,
		pos:     0,
	}

	fmt.Printf("PCM Buffer length: %d, Sample count will be: %d\n",
		len(pcmBuffer.Data), len(pcmBuffer.Data)/format.NumChannels)

	// Convert audio data to beep's float64 format
	streamer.convertSamples()

	return streamer, beepFormat, nil
}

func (s *aiffStreamer) convertSamples() {
	if s.buffer == nil || s.buffer.Data == nil {
		return
	}

	numSamples := len(s.buffer.Data) / s.format.NumChannels
	s.samples = make([][2]float64, numSamples)

	// Determine the bit depth from the buffer SourceBitDepth (bytes)
	sourceBitDepth := s.buffer.SourceBitDepth
	bitDepth := int(sourceBitDepth * 8) // Convert bytes to bits
	var maxValue float64

	// Calculate the maximum value for normalization based on bit depth
	switch bitDepth {
	case 8:
		maxValue = float64(1 << 7) // 128
	case 16:
		maxValue = float64(1 << 15) // 32768
	case 24:
		maxValue = float64(1 << 23) // 8388608
	case 32:
		maxValue = float64(1 << 31) // 2147483648
	case 64:
		maxValue = float64(1 << 63) // For int64
	default:
		// Default to 16-bit if unknown
		maxValue = float64(1 << 15)
		fmt.Printf("Unknown bit depth %d (from %d bytes), defaulting to 16-bit normalization\n", bitDepth, sourceBitDepth)
	}

	fmt.Printf("Converting %d samples, source bytes: %d, bit depth: %d, max value: %f\n",
		numSamples, sourceBitDepth, bitDepth, maxValue)

	for i := 0; i < numSamples; i++ {
		if s.format.NumChannels == 1 {
			// Mono: duplicate to both channels
			sample := float64(s.buffer.Data[i]) / maxValue
			s.samples[i] = [2]float64{sample, sample}
		} else if s.format.NumChannels >= 2 {
			// Stereo or more: take first two channels
			left := float64(s.buffer.Data[i*s.format.NumChannels]) / maxValue
			right := float64(s.buffer.Data[i*s.format.NumChannels+1]) / maxValue
			s.samples[i] = [2]float64{left, right}
		}
	}

	if len(s.samples) > 0 {
		sampleCount := 5
		if len(s.samples) < sampleCount {
			sampleCount = len(s.samples)
		}
		fmt.Printf("Sample conversion complete. First few samples: %v\n", s.samples[:sampleCount])
	}
}

func (s *aiffStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.pos >= len(s.samples) {
		return 0, false
	}

	n = len(samples)
	if s.pos+n > len(s.samples) {
		n = len(s.samples) - s.pos
	}

	copy(samples[:n], s.samples[s.pos:s.pos+n])
	s.pos += n

	return n, true
}

func (s *aiffStreamer) Err() error {
	return nil
}

func (s *aiffStreamer) Len() int {
	return len(s.samples)
}

func (s *aiffStreamer) Position() int {
	return s.pos
}

func (s *aiffStreamer) Seek(p int) error {
	if p < 0 || p >= len(s.samples) {
		return fmt.Errorf("seek position out of range")
	}
	s.pos = p
	return nil
}

func (s *aiffStreamer) Close() error {
	// The decoder doesn't need explicit closing in go-audio/aiff
	return nil
}