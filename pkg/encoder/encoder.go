package encoder

import (
	"fmt"
	"os"

	"github.com/braheezy/shine-mp3/pkg/mp3"
)

func EncodePCMToMP3(samples []int16, sampleRate, channels int, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create MP3 file: %w", err)
	}
	defer file.Close()

	encoder := mp3.NewEncoder(sampleRate, channels)
	
	encoder.Write(file, samples)
	
	return nil
}
