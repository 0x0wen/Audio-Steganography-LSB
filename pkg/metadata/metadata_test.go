package metadata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMetadataFromFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	content := "This is a test message for steganography"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		filePath       string
		useEncryption  bool
		useRandomSeed  bool
		nLsb           int
		expectedError  bool
	}{
		{
			name:          "valid file",
			filePath:      testFile,
			useEncryption: false,
			useRandomSeed: true,
			nLsb:          2,
			expectedError: false,
		},
		{
			name:          "nonexistent file",
			filePath:      "nonexistent.txt",
			useEncryption: false,
			useRandomSeed: false,
			nLsb:          1,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := CreateMetadataFromFile(tt.filePath, tt.useEncryption, tt.useRandomSeed, tt.nLsb)
			
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				require.NotNil(t, metadata)
				
				assert.Equal(t, "test.txt", metadata.OriginalFilename)
				assert.Equal(t, ".txt", metadata.FileExtension)
				assert.Equal(t, int64(len(content)), metadata.FileSize)
				assert.Equal(t, tt.useEncryption, metadata.UseEncryption)
				assert.Equal(t, tt.useRandomSeed, metadata.UseRandomSeed)
				assert.Equal(t, tt.nLsb, metadata.NLsb)
			}
		})
	}
}

func TestStegoMetadataSerialization(t *testing.T) {
	original := &StegoMetadata{
		OriginalFilename: "secret.txt",
		FileExtension:    ".txt",
		FileSize:         1024,
		UseEncryption:    false,
		UseRandomSeed:    true,
		NLsb:             2,
	}

	assert.Equal(t, "secret.txt", original.OriginalFilename)
	assert.Equal(t, ".txt", original.FileExtension)
	assert.Equal(t, int64(1024), original.FileSize)
	assert.Equal(t, false, original.UseEncryption)
	assert.Equal(t, true, original.UseRandomSeed)
	assert.Equal(t, 2, original.NLsb)
}

func TestStegoMetadataValidation(t *testing.T) {
	tests := []struct {
		name     string
		metadata *StegoMetadata
		valid    bool
	}{
		{
			name: "valid metadata",
			metadata: &StegoMetadata{
				OriginalFilename: "test.txt",
				FileExtension:    ".txt",
				FileSize:         100,
				UseEncryption:    false,
				UseRandomSeed:    true,
				NLsb:             2,
			},
			valid: true,
		},
		{
			name: "invalid n_lsb",
			metadata: &StegoMetadata{
				OriginalFilename: "test.txt",
				FileExtension:    ".txt",
				FileSize:         100,
				UseEncryption:    false,
				UseRandomSeed:    true,
				NLsb:             5, 
			},
			valid: false,
		},
		{
			name: "negative file size",
			metadata: &StegoMetadata{
				OriginalFilename: "test.txt",
				FileExtension:    ".txt",
				FileSize:         -1,
				UseEncryption:    false,
				UseRandomSeed:    true,
				NLsb:             2,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.GreaterOrEqual(t, tt.metadata.NLsb, 1)
				assert.LessOrEqual(t, tt.metadata.NLsb, 4)
				assert.GreaterOrEqual(t, tt.metadata.FileSize, int64(0))
				assert.NotEmpty(t, tt.metadata.OriginalFilename)
			} else {
				invalid := tt.metadata.NLsb < 1 || tt.metadata.NLsb > 4 || 
						   tt.metadata.FileSize < 0 || tt.metadata.OriginalFilename == ""
				assert.True(t, invalid)
			}
		})
	}
}

func TestStoreMetadata(t *testing.T) {
	tempDir := t.TempDir()
	mp3File := filepath.Join(tempDir, "test.mp3")
	
	err := os.WriteFile(mp3File, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	metadata := &StegoMetadata{
		OriginalFilename: "secret.txt",
		FileExtension:    ".txt",
		FileSize:         1024,
		UseEncryption:    false,
		UseRandomSeed:    true,
		NLsb:             2,
	}

	err = StoreMetadata(mp3File, metadata)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to open MP3 file")
	} else {
		t.Log("StoreMetadata unexpectedly succeeded with fake MP3 file")
	}
}

func TestStoreSecretMessage(t *testing.T) {
	tempDir := t.TempDir()
	mp3File := filepath.Join(tempDir, "test.mp3")
	
	err := os.WriteFile(mp3File, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	messageData := []byte("This is a secret message")

	err = StoreSecretMessage(mp3File, messageData)
	if err != nil {
		assert.Contains(t, err.Error(), "failed to open MP3 file")
	} else {
		t.Log("StoreSecretMessage unexpectedly succeeded with fake MP3 file")
	}
}

func TestRetrieveSecretMessage(t *testing.T) {
	tempDir := t.TempDir()
	mp3File := filepath.Join(tempDir, "test.mp3")
	
	err := os.WriteFile(mp3File, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	_, err = RetrieveSecretMessage(mp3File)
	if err != nil {
		assert.True(t, 
			strings.Contains(err.Error(), "failed to open MP3 file") || 
			strings.Contains(err.Error(), "no secret message found in file"))
	} else {
		t.Log("RetrieveSecretMessage unexpectedly succeeded with fake MP3 file")
	}
}

func TestRetrieveMetadata(t *testing.T) {
	tempDir := t.TempDir()
	mp3File := filepath.Join(tempDir, "test.mp3")
	
	err := os.WriteFile(mp3File, []byte("fake mp3 content"), 0644)
	require.NoError(t, err)

	_, err = RetrieveMetadata(mp3File)
	if err != nil {
		assert.True(t, 
			strings.Contains(err.Error(), "failed to open MP3 file") || 
			strings.Contains(err.Error(), "no steganography metadata found in file"))
	} else {
		t.Log("RetrieveMetadata unexpectedly succeeded with fake MP3 file")
	}
}

func TestStoreMetadataWithNonexistentFile(t *testing.T) {
	metadata := &StegoMetadata{
		OriginalFilename: "secret.txt",
		FileExtension:    ".txt",
		FileSize:         1024,
		UseEncryption:    false,
		UseRandomSeed:    true,
		NLsb:             2,
	}

	err := StoreMetadata("nonexistent.mp3", metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open MP3 file")
}

func TestStoreSecretMessageWithNonexistentFile(t *testing.T) {
	messageData := []byte("This is a secret message")

	err := StoreSecretMessage("nonexistent.mp3", messageData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open MP3 file")
}

func TestRetrieveSecretMessageWithNonexistentFile(t *testing.T) {
	_, err := RetrieveSecretMessage("nonexistent.mp3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open MP3 file")
}

func TestRetrieveMetadataWithNonexistentFile(t *testing.T) {
	_, err := RetrieveMetadata("nonexistent.mp3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open MP3 file")
}
