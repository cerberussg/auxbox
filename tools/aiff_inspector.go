package main

import (
	"fmt"
	"os"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run aiff_inspector.go <path-to-aiff>")
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

	fmt.Println("=== AIFF File Inspector ===")
	fmt.Printf("File:        %s\n", filePath)
	fmt.Printf("Sample Rate: %d Hz\n", format.SampleRate)
	fmt.Printf("Channels:    %d\n", format.NumChannels)
	fmt.Printf("Bit Depth:   %d bits\n", decoder.SampleBitDepth())
	fmt.Printf("Total Frames: %d\n", decoder.NumSampleFrames)

	// Calculate duration
	duration := float64(decoder.NumSampleFrames) / float64(format.SampleRate)
	fmt.Printf("Duration:    %.2f seconds\n", duration)

	// Check encoding
	bitDepth := decoder.SampleBitDepth()
	fmt.Printf("\nDecoding Details:\n")
	fmt.Printf("Bit Depth:   %d\n", bitDepth)

	var maxValue float64
	switch bitDepth {
	case 8:
		maxValue = 128
	case 16:
		maxValue = 32768
	case 24:
		maxValue = 8388608
	case 32:
		maxValue = 2147483648
	case 64:
		maxValue = float64(uint64(1) << 63)
	default:
		maxValue = 32768
		fmt.Printf("WARNING: Unusual bit depth %d, defaulting to 16-bit normalization\n", bitDepth)
	}
	fmt.Printf("Max Value:   %.0f (for normalization)\n", maxValue)

	// Try reading a small sample
	fmt.Println("\n=== Testing Sample Read ===")

	// Create a buffer for reading samples
	bufSize := 1024
	buf := &audio.IntBuffer{
		Data:   make([]int, bufSize*format.NumChannels),
		Format: decoder.Format(),
	}

	n, err := decoder.PCMBuffer(buf)
	if err != nil {
		fmt.Printf("ERROR reading PCM data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Read %d samples successfully\n", n)

	// Show first few samples
	if n > 0 && len(buf.Data) > 10 {
		fmt.Println("\nFirst 10 raw samples:")
		for i := 0; i < 10 && i < len(buf.Data); i++ {
			normalized := float64(buf.Data[i]) / maxValue
			fmt.Printf("  Sample %d: %d (normalized: %.6f)\n", i, buf.Data[i], normalized)

			// Check for clipping
			if normalized > 1.0 || normalized < -1.0 {
				fmt.Printf("    ⚠️  WARNING: Sample out of range!\n")
			}
		}
	}

	fmt.Println("\n✓ File appears readable")
}
