package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStegoKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "mykey123",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "key too long",
			key:     "thiskeyistoolongandexceedsthelimit",
			wantErr: true,
		},
		{
			name:    "key at limit",
			key:     "1234567890123456789012345",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStegoKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNLsb(t *testing.T) {
	tests := []struct {
		name    string
		nLsb    int
		wantErr bool
	}{
		{
			name:    "valid n_lsb 1",
			nLsb:    1,
			wantErr: false,
		},
		{
			name:    "valid n_lsb 4",
			nLsb:    4,
			wantErr: false,
		},
		{
			name:    "invalid n_lsb 0",
			nLsb:    0,
			wantErr: true,
		},
		{
			name:    "invalid n_lsb 5",
			nLsb:    5,
			wantErr: true,
		},
		{
			name:    "invalid n_lsb negative",
			nLsb:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNLsb(tt.nLsb)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateCapacity(t *testing.T) {
	tests := []struct {
		name         string
		totalSamples int
		nLsb         int
		expected     int
	}{
		{
			name:         "1000 samples, 1 LSB",
			totalSamples: 1000,
			nLsb:         1,
			expected:     125,
		},
		{
			name:         "1000 samples, 4 LSB",
			totalSamples: 1000,
			nLsb:         4,
			expected:     500, 
		},
		{
			name:         "8000 samples, 2 LSB",
			totalSamples: 8000,
			nLsb:         2,
			expected:     2000, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCapacity(tt.totalSamples, tt.nLsb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeneratePositions(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		useRandomSeed bool
		totalSamples  int
		nLsb          int
	}{
		{
			name:          "sequential positions",
			key:           "testkey",
			useRandomSeed: false,
			totalSamples:  1000,
			nLsb:          2,
		},
		{
			name:          "random positions",
			key:           "testkey",
			useRandomSeed: true,
			totalSamples:  1000,
			nLsb:          2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			positions, err := GeneratePositions(tt.key, tt.useRandomSeed, tt.totalSamples, tt.nLsb)
			require.NoError(t, err)
			
			expectedCount := (tt.totalSamples * tt.nLsb) / 8
			assert.Len(t, positions, expectedCount)
			
			for _, pos := range positions {
				assert.GreaterOrEqual(t, pos, 0)
				assert.Less(t, pos, tt.totalSamples)
			}
		})
	}
}

func TestGeneratePositionsDeterministic(t *testing.T) {
	key := "testkey"
	totalSamples := 1000
	nLsb := 2

	positions1, err1 := GeneratePositions(key, true, totalSamples, nLsb)
	require.NoError(t, err1)

	positions2, err2 := GeneratePositions(key, true, totalSamples, nLsb)
	require.NoError(t, err2)

	assert.Equal(t, positions1, positions2)
}

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	tests := []struct {
		name     string
		value    int
		expected bool
	}{
		{
			name:     "contains value",
			value:    3,
			expected: true,
		},
		{
			name:     "does not contain value",
			value:    6,
			expected: false,
		},
		{
			name:     "contains first value",
			value:    1,
			expected: true,
		},
		{
			name:     "contains last value",
			value:    5,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(slice, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	
	content := "This is test content for file reading"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
		expected    string
	}{
		{
			name:        "valid file",
			filePath:    testFile,
			expectError: false,
			expected:    content,
		},
		{
			name:        "nonexistent file",
			filePath:    "nonexistent.txt",
			expectError: true,
		},
		{
			name:        "empty file",
			filePath:    filepath.Join(tempDir, "empty.txt"),
			expectError: false,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "empty file" {
				err := os.WriteFile(tt.filePath, []byte(""), 0644)
				require.NoError(t, err)
			}

			data, err := ReadFile(tt.filePath)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, string(data))
			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		filePath    string
		data        []byte
		expectError bool
	}{
		{
			name:        "valid write",
			filePath:    filepath.Join(tempDir, "test.txt"),
			data:        []byte("This is test data"),
			expectError: false,
		},
		{
			name:        "empty data",
			filePath:    filepath.Join(tempDir, "empty.txt"),
			data:        []byte(""),
			expectError: false,
		},
		{
			name:        "binary data",
			filePath:    filepath.Join(tempDir, "binary.bin"),
			data:        []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
			expectError: false,
		},
		{
			name:        "invalid path",
			filePath:    "/nonexistent/directory/file.txt",
			data:        []byte("test"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteFile(tt.filePath, tt.data)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				
				content, err := os.ReadFile(tt.filePath)
				require.NoError(t, err)
				assert.Equal(t, tt.data, content)
			}
		})
	}
}

func TestReadFileWithLargeFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.txt")
	
	largeContent := make([]byte, 10000)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	
	err := os.WriteFile(testFile, largeContent, 0644)
	require.NoError(t, err)

	data, err := ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, largeContent, data)
}

func TestWriteFileOverwrite(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "overwrite.txt")
	
	initialContent := []byte("Initial content")
	err := WriteFile(testFile, initialContent)
	require.NoError(t, err)
	
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, initialContent, content)
	
	newContent := []byte("New content")
	err = WriteFile(testFile, newContent)
	require.NoError(t, err)
	
	content, err = os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, newContent, content)
}
