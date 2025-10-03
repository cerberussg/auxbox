package decoders

import (
	"fmt"
	"io"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/wav"
)

// DecodeWAV decodes WAV files using beep's WAV decoder
func DecodeWAV(r io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) {
	streamer, format, err := wav.Decode(r)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("failed to decode WAV: %w", err)
	}

	return streamer, format, nil
}