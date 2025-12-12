package main

import (
	"fmt"
	"os"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
)

// This tool plays just the first second of an AIFF and prints detailed sample data
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run aiff_debug_player.go <path-to-aiff>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	decoder := aiff.NewDecoder(file)
	if !decoder.IsValidFile() {
		fmt.Println("ERROR: Invalid AIFF file")
		os.Exit(1)
	}

	decoder.ReadInfo()
	format := decoder.Format()
	bitDepth := decoder.SampleBitDepth()

	fmt.Println("=== AIFF Debug Player ===")
	fmt.Printf("Sample Rate: %d Hz\n", format.SampleRate)
	fmt.Printf("Channels: %d\n", format.NumChannels)
	fmt.Printf("Bit Depth: %d bits\n", bitDepth)
	fmt.Println()

	// Read first 100 frames (200 samples for stereo)
	framesToRead := 100
	buf := &audio.IntBuffer{
		Data:   make([]int, framesToRead*format.NumChannels),
		Format: format,
	}

	n, err := decoder.PCMBuffer(buf)
	if err != nil {
		fmt.Printf("ERROR reading PCM: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Read %d sample values\n", n)
	fmt.Println("\n=== First 50 samples (raw int values) ===")

	// Look for suspiciously large values that might indicate sign extension issues
	var maxVal, minVal int
	suspiciousCount := 0

	for i := 0; i < 50 && i < len(buf.Data); i++ {
		val := buf.Data[i]

		if val > maxVal {
			maxVal = val
		}
		if val < minVal {
			minVal = val
		}

		// For 24-bit, valid range is roughly -8388608 to 8388607
		// If we see values way outside this, sign extension is wrong
		if bitDepth == 24 && (val > 8388607 || val < -8388608) {
			fmt.Printf("  Sample %d: %d ⚠️  OUT OF RANGE!\n", i, val)
			suspiciousCount++
		} else {
			fmt.Printf("  Sample %d: %d\n", i, val)
		}
	}

	fmt.Printf("\nMin value: %d\n", minVal)
	fmt.Printf("Max value: %d\n", maxVal)
	fmt.Printf("Suspicious samples: %d\n", suspiciousCount)

	// Check if values are reasonable for 24-bit
	expectedMax := int(1 << 23)
	if maxVal > expectedMax*2 || minVal < -expectedMax*2 {
		fmt.Println("\n⚠️  WARNING: Values exceed expected 24-bit range!")
		fmt.Println("This indicates the library may not be properly handling 24-bit samples.")
	}
}
