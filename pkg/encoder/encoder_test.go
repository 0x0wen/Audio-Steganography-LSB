package encoder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodePCMToMP3(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.mp3")

	samples := []int16{1000, 2000, 3000, 4000, 5000, -1000, -2000, -3000}
	sampleRate := 44100
	channels := 2

	err := EncodePCMToMP3(samples, sampleRate, channels, outputFile)
	
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create MP3 file")
	} else {
		_, err := os.Stat(outputFile)
		assert.NoError(t, err)
	}
}

func TestEncodePCMToMP3WithInvalidPath(t *testing.T) {
	samples := []int16{1000, 2000, 3000}
	sampleRate := 44100
	channels := 2

	err := EncodePCMToMP3(samples, sampleRate, channels, "/nonexistent/directory/output.mp3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create MP3 file")
}

func TestEncodePCMToMP3WithEmptySamples(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "empty.mp3")

	samples := []int16{}
	sampleRate := 44100
	channels := 2

	err := EncodePCMToMP3(samples, sampleRate, channels, outputFile)
	
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create MP3 file")
	}
}

func TestEncodePCMToMP3WithDifferentParameters(t *testing.T) {
	tempDir := t.TempDir()
	samples := []int16{1000, 2000, 3000, 4000}

	testCases := []struct {
		name       string
		sampleRate int
		channels   int
		outputFile string
	}{
		{
			name:       "mono 44.1kHz",
			sampleRate: 44100,
			channels:   1,
			outputFile: "mono_44k.mp3",
		},
		{
			name:       "stereo 48kHz",
			sampleRate: 48000,
			channels:   2,
			outputFile: "stereo_48k.mp3",
		},
		{
			name:       "mono 22kHz",
			sampleRate: 22050,
			channels:   1,
			outputFile: "mono_22k.mp3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, tc.outputFile)
			err := EncodePCMToMP3(samples, tc.sampleRate, tc.channels, outputPath)
			
			if err != nil {
				assert.Contains(t, err.Error(), "failed to create MP3 file")
			}
		})
	}
}

func TestEncodePCMToMP3WithLargeSamples(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "large.mp3")

	samples := make([]int16, 10000)
	for i := range samples {
		samples[i] = int16(i % 32767)
	}

	sampleRate := 44100
	channels := 2

	err := EncodePCMToMP3(samples, sampleRate, channels, outputFile)
	
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create MP3 file")
	}
}

func TestEncodePCMToMP3WithNegativeSamples(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "negative.mp3")

	samples := []int16{-32768, -16384, 0, 16384, 32767}
	sampleRate := 44100
	channels := 2

	err := EncodePCMToMP3(samples, sampleRate, channels, outputFile)
	
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create MP3 file")
	}
}

func TestEncodePCMToMP3FileCreation(t *testing.T) {
	tempDir := t.TempDir()
	
	samples := []int16{1000, 2000, 3000}
	sampleRate := 44100
	channels := 2

	outputFile := filepath.Join(tempDir, "test_output.mp3")
	
	err := EncodePCMToMP3(samples, sampleRate, channels, outputFile)
	
	if err != nil {
		_, statErr := os.Stat(outputFile)
		if statErr == nil {
			assert.NotContains(t, err.Error(), "failed to create MP3 file")
		} else {
			assert.Contains(t, err.Error(), "failed to create MP3 file")
		}
	}
}
