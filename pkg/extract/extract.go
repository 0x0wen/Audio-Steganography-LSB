package extract

import (
	"encoding/binary"
	"fmt"
	"os"

	"audio-steganography-lsb/pkg/utils"
)

type ExtractConfig struct {
	StegoAudio string
	StegoKey   string
	OutputPath string
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

func Extract(config *ExtractConfig) error {
	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
		return fmt.Errorf("invalid stego key: %w", err)
	}

	messageData, err := simpleExtract(config.StegoAudio)
	if err != nil {
		return fmt.Errorf("failed to extract data: %w", err)
	}

	if err := utils.WriteFile(config.OutputPath, messageData); err != nil {
		return fmt.Errorf("failed to write extracted message: %w", err)
	}

	fmt.Printf("Successfully extracted %d bytes to %s\n", len(messageData), config.OutputPath)
	return nil
}

func simpleExtract(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stego file: %w", err)
	}

	startOffset := 1000 
	
	if startOffset+4 > len(data) {
		return nil, fmt.Errorf("invalid file format")
	}
	metadataLen := binary.LittleEndian.Uint32(data[startOffset:])
	startOffset += 4
	
	startOffset += int(metadataLen)
	
	if startOffset+4 > len(data) {
		return nil, fmt.Errorf("invalid file format")
	}
	messageLen := binary.LittleEndian.Uint32(data[startOffset:])
	startOffset += 4
	
	if startOffset+int(messageLen) > len(data) {
		return nil, fmt.Errorf("invalid file format")
	}
	messageData := data[startOffset : startOffset+int(messageLen)]
	
	return messageData, nil
}
