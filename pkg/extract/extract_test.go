package extract

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractConfig(t *testing.T) {
	config := &ExtractConfig{
		StegoAudio: "stego.mp3",
		StegoKey:   "testkey",
		OutputPath: "extracted.txt",
	}

	assert.Equal(t, "stego.mp3", config.StegoAudio)
	assert.Equal(t, "testkey", config.StegoKey)
	assert.Equal(t, "extracted.txt", config.OutputPath)
}

func TestExtractValidation(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	tests := []struct {
		name        string
		config      *ExtractConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid stego key - empty",
			config: &ExtractConfig{
				StegoAudio: "stego.mp3",
				StegoKey:   "",
				OutputPath: outputFile,
			},
			expectError: true,
			errorMsg:    "invalid stego key",
		},
		{
			name: "invalid stego key - too long",
			config: &ExtractConfig{
				StegoAudio: "stego.mp3",
				StegoKey:   "thiskeyistoolongandexceedsthelimit",
				OutputPath: outputFile,
			},
			expectError: true,
			errorMsg:    "invalid stego key",
		},
		{
			name: "valid config with nonexistent file",
			config: &ExtractConfig{
				StegoAudio: "stego.mp3",
				StegoKey:   "testkey",
				OutputPath: outputFile,
			},
			expectError: true,
			errorMsg:    "failed to read audio samples",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Extract(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractWithNonexistentFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	config := &ExtractConfig{
		StegoAudio: "nonexistent.mp3",
		StegoKey:   "testkey",
		OutputPath: outputFile,
	}

	err := Extract(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read audio samples")
}

func TestExtractWithInvalidOutputPath(t *testing.T) {
	config := &ExtractConfig{
		StegoAudio: "stego.mp3",
		StegoKey:   "testkey",
		OutputPath: "/nonexistent/directory/output.txt",
	}

	err := Extract(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read audio samples")
}

func TestExtractConfigFields(t *testing.T) {
	config := &ExtractConfig{
		StegoAudio: "test_stego.mp3",
		StegoKey:   "my_secret_key",
		OutputPath: "/path/to/output.txt",
	}

	assert.Equal(t, "test_stego.mp3", config.StegoAudio)
	assert.Equal(t, "my_secret_key", config.StegoKey)
	assert.Equal(t, "/path/to/output.txt", config.OutputPath)

	config.StegoAudio = "another_stego.mp3"
	config.StegoKey = "another_key"
	config.OutputPath = "another_output.txt"

	assert.Equal(t, "another_stego.mp3", config.StegoAudio)
	assert.Equal(t, "another_key", config.StegoKey)
	assert.Equal(t, "another_output.txt", config.OutputPath)
}

func TestExtractKeyValidation(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	testCases := []struct {
		name      string
		key       string
		shouldErr bool
	}{
		{"empty key", "", true},
		{"single character", "a", false},
		{"normal key", "testkey123", false},
		{"key at limit", "1234567890123456789012345", false},
		{"key over limit", "12345678901234567890123456", true},
		{"key with special chars", "test@key#123", false},
		{"key with spaces", "test key 123", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &ExtractConfig{
				StegoAudio: "stego.mp3",
				StegoKey:   tc.key,
				OutputPath: outputFile,
			}

			err := Extract(config)
			if tc.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid stego key")
			} else {
				assert.Error(t, err)
				assert.NotContains(t, err.Error(), "invalid stego key")
			}
		})
	}
}
