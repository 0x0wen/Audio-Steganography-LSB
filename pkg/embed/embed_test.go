package embed

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbedConfig(t *testing.T) {
	config := &EmbedConfig{
		CoverAudio:    "test.mp3",
		SecretMessage: "secret.txt",
		StegoKey:      "testkey",
		NLsb:          2,
		UseRandomSeed: true,
		UseEncryption: false,
		OutputPath:    "output.mp3",
	}

	assert.Equal(t, "test.mp3", config.CoverAudio)
	assert.Equal(t, "secret.txt", config.SecretMessage)
	assert.Equal(t, "testkey", config.StegoKey)
	assert.Equal(t, 2, config.NLsb)
	assert.True(t, config.UseRandomSeed)
	assert.False(t, config.UseEncryption)
	assert.Equal(t, "output.mp3", config.OutputPath)
}

func TestEmbedValidation(t *testing.T) {
	tempDir := t.TempDir()
	
	secretFile := filepath.Join(tempDir, "secret.txt")
	coverFile := filepath.Join(tempDir, "cover.mp3")
	outputFile := filepath.Join(tempDir, "output.mp3")
	
	secretContent := "This is a secret message"
	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(coverFile, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		config      *EmbedConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid stego key - empty",
			config: &EmbedConfig{
				CoverAudio:    coverFile,
				SecretMessage: secretFile,
				StegoKey:      "",
				NLsb:          2,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "invalid stego key",
		},
		{
			name: "invalid stego key - too long",
			config: &EmbedConfig{
				CoverAudio:    coverFile,
				SecretMessage: secretFile,
				StegoKey:      "thiskeyistoolongandexceedsthelimit",
				NLsb:          2,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "invalid stego key",
		},
		{
			name: "invalid n_lsb - too low",
			config: &EmbedConfig{
				CoverAudio:    coverFile,
				SecretMessage: secretFile,
				StegoKey:      "testkey",
				NLsb:          0,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "invalid n_lsb",
		},
		{
			name: "invalid n_lsb - too high",
			config: &EmbedConfig{
				CoverAudio:    coverFile,
				SecretMessage: secretFile,
				StegoKey:      "testkey",
				NLsb:          5,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "invalid n_lsb",
		},
		{
			name: "nonexistent secret message file",
			config: &EmbedConfig{
				CoverAudio:    coverFile,
				SecretMessage: "nonexistent.txt",
				StegoKey:      "testkey",
				NLsb:          2,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "failed to read secret message",
		},
		{
			name: "nonexistent cover audio file",
			config: &EmbedConfig{
				CoverAudio:    "nonexistent.mp3",
				SecretMessage: secretFile,
				StegoKey:      "testkey",
				NLsb:          2,
				UseRandomSeed: true,
				UseEncryption: false,
				OutputPath:    outputFile,
			},
			expectError: true,
			errorMsg:    "failed to open cover audio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Embed(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateMetadataSize(t *testing.T) {
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.txt")
	coverFile := filepath.Join(tempDir, "cover.mp3")
	outputFile := filepath.Join(tempDir, "output.mp3")
	
	secretContent := "Test message"
	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(coverFile, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	config := &EmbedConfig{
		CoverAudio:    coverFile,
		SecretMessage: secretFile,
		StegoKey:      "testkey",
		NLsb:          2,
		UseRandomSeed: true,
		UseEncryption: false,
		OutputPath:    outputFile,
	}

	err = Embed(config)
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "metadata")
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")
	
	content := "This is test content"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)
	
	err = copyFile(srcFile, dstFile)
	require.NoError(t, err)
	
	copiedContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))
}

func TestCopyFileErrors(t *testing.T) {
	tempDir := t.TempDir()
	dstFile := filepath.Join(tempDir, "destination.txt")
	
	tests := []struct {
		name     string
		srcFile  string
		dstFile  string
		wantErr  bool
	}{
		{
			name:    "source file does not exist",
			srcFile: "nonexistent.txt",
			dstFile: dstFile,
			wantErr: true,
		},
		{
			name:    "destination directory does not exist",
			srcFile: "source.txt",
			dstFile: "/nonexistent/directory/file.txt",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.srcFile == "source.txt" {
				srcFile := filepath.Join(tempDir, "source.txt")
				err := os.WriteFile(srcFile, []byte("test"), 0644)
				require.NoError(t, err)
				tt.srcFile = srcFile
			}
			
			err := copyFile(tt.srcFile, tt.dstFile)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
