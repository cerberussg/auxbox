package decoders

import (
	"fmt"
	"io"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
)

// DecodeMP3 decodes MP3 files using beep's MP3 decoder
func DecodeMP3(r io.ReadCloser) (beep.StreamSeekCloser, beep.Format, error) {
	streamer, format, err := mp3.Decode(r)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("failed to decode MP3: %w", err)
	}

	return streamer, format, nil
}