package embed

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"audio-steganography-lsb/pkg/utils"

	"github.com/hajimehoshi/go-mp3"

	"audio-steganography-lsb/pkg/encrypt"
)

type EmbedConfig struct {
	CoverAudio     string
	SecretMessage  string
	StegoKey       string
	NLsb           int
	UseRandomSeed  bool
	UseEncryption  bool
	OutputPath     string
}

type FileMetadata struct {
	OriginalFilename string
	FileExtension    string
	FileSize         int64
	UseEncryption    bool
	UseRandomSeed    bool
	NLsb             int
	DataSize         int64
}

func Embed(config *EmbedConfig) error {
	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
		return fmt.Errorf("invalid stego key: %w", err)
	}

	if err := utils.ValidateNLsb(config.NLsb); err != nil {
		return fmt.Errorf("invalid n_lsb: %w", err)
	}

	messageData, err := utils.ReadFile(config.SecretMessage)
	if err != nil {
		return fmt.Errorf("failed to read secret message: %w", err)
	}

	if config.UseEncryption {
		messageData = vigenere.Encrypt(messageData, config.StegoKey)
	}

	stegoMetadata, err := metadata.CreateMetadataFromFile(
		config.SecretMessage,
		config.UseEncryption,
		config.UseRandomSeed,
		config.NLsb,
	)
	fileInfo, err := os.Stat(config.SecretMessage)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	metadata := &FileMetadata{
		OriginalFilename: filepath.Base(config.SecretMessage),
		FileExtension:    filepath.Ext(config.SecretMessage),
		FileSize:         fileInfo.Size(),
		UseEncryption:    config.UseEncryption,
		UseRandomSeed:    config.UseRandomSeed,
		NLsb:             config.NLsb,
		DataSize:         int64(len(messageData)),
	}

	_, err = readAudioSamples(config.CoverAudio)
	if err != nil {
		return fmt.Errorf("failed to read audio samples: %w", err)
	}

	metadataSize := calculateMetadataSize(metadata)
	totalDataSize := len(messageData) + metadataSize
	
	if totalDataSize > 1000000 {
		return fmt.Errorf("message too large: need %d bytes, max capacity is 1000000 bytes", totalDataSize)
	}

	if err := simpleEmbed(config.CoverAudio, config.OutputPath, messageData, metadata); err != nil {
		return fmt.Errorf("failed to embed data: %w", err)
	}

	fmt.Printf("Successfully embedded %d bytes using LSB steganography in audio samples\n", len(messageData))
	return nil
}

func readAudioSamples(filePath string) ([]int16, error) {
	coverFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cover audio: %w", err)
	}
	defer coverFile.Close()

	decoder, err := mp3.NewDecoder(coverFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	var audioSamples []int16
	buffer := make([]byte, 1024)
	for {
		n, err := decoder.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("failed to read audio data: %w", err)
		}
		if n == 0 {
			break
		}

		for i := 0; i < n; i += 2 {
			if i+1 < n {
				sample := int16(buffer[i]) | (int16(buffer[i+1]) << 8)
				audioSamples = append(audioSamples, sample)
			}
		}
	}

	return audioSamples, nil
}

func serializeMetadata(metadata *FileMetadata) []byte {
	data := make([]byte, 0, 1024)

	data = append(data, byte(len(metadata.OriginalFilename)))

	data = append(data, []byte(metadata.OriginalFilename)...)

	data = append(data, byte(len(metadata.FileExtension)))

	data = append(data, []byte(metadata.FileExtension)...)

	sizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(sizeBytes, uint64(metadata.FileSize))
	data = append(data, sizeBytes...)

	flags := byte(0)
	if metadata.UseEncryption {
		flags |= 1
	}
	if metadata.UseRandomSeed {
		flags |= 2
	}
	data = append(data, flags)

	data = append(data, byte(metadata.NLsb))

	dataSizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(dataSizeBytes, uint64(metadata.DataSize))
	data = append(data, dataSizeBytes...)

	return data
}

func calculateMetadataSize(metadata *FileMetadata) int {
	return 1 + len(metadata.OriginalFilename) + 1 + len(metadata.FileExtension) + 8 + 1 + 1 + 8
}

func simpleEmbed(originalPath, outputPath string, messageData []byte, metadata *FileMetadata) error {
	originalData, err := os.ReadFile(originalPath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	outputData := make([]byte, len(originalData))
	copy(outputData, originalData)

	startOffset := 1000
	
	metadataBytes := serializeMetadata(metadata)
	
	if startOffset+4 > len(outputData) {
		return fmt.Errorf("not enough space for metadata length")
	}
	binary.LittleEndian.PutUint32(outputData[startOffset:], uint32(len(metadataBytes)))
	startOffset += 4
	
	if startOffset+len(metadataBytes) > len(outputData) {
		return fmt.Errorf("not enough space for metadata")
	}
	copy(outputData[startOffset:], metadataBytes)
	startOffset += len(metadataBytes)
	
	if startOffset+4 > len(outputData) {
		return fmt.Errorf("not enough space for message length")
	}
	binary.LittleEndian.PutUint32(outputData[startOffset:], uint32(len(messageData)))
	startOffset += 4
	
	if startOffset+len(messageData) > len(outputData) {
		return fmt.Errorf("not enough space for message data")
	}
	copy(outputData[startOffset:], messageData)

	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
