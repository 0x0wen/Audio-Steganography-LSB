package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bogem/id3v2"
)

type StegoMetadata struct {
	OriginalFilename string `json:"original_filename"`
	FileExtension    string `json:"file_extension"`
	FileSize         int64  `json:"file_size"`
	UseEncryption    bool   `json:"use_encryption"`
	UseRandomSeed    bool   `json:"use_random_seed"`
	NLsb             int    `json:"n_lsb"`
}

func StoreMetadata(filePath string, metadata *StegoMetadata) error {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	udtf := id3v2.UserDefinedTextFrame{
		Encoding:    id3v2.EncodingUTF8,
		Description: "STEGO_METADATA",
		Value:       string(metadataJSON),
	}
	tag.AddFrame("TXXX", udtf)

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

func StoreSecretMessage(filePath string, messageData []byte) error {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	udtf := id3v2.UserDefinedTextFrame{
		Encoding:    id3v2.EncodingUTF8,
		Description: "SECRET_MESSAGE",
		Value:       string(messageData),
	}
	tag.AddFrame("TXXX", udtf)

	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save secret message: %w", err)
	}

	return nil
}

func RetrieveSecretMessage(filePath string) ([]byte, error) {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	frames := tag.GetFrames("TXXX")
	for _, frame := range frames {
		if udtf, ok := frame.(id3v2.UserDefinedTextFrame); ok {
			if udtf.Description == "SECRET_MESSAGE" {
				return []byte(udtf.Value), nil
			}
		}
	}

	return nil, fmt.Errorf("no secret message found in file")
}

func RetrieveMetadata(filePath string) (*StegoMetadata, error) {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open MP3 file: %w", err)
	}
	defer tag.Close()

	frames := tag.GetFrames("TXXX")
	for _, frame := range frames {
		if udtf, ok := frame.(id3v2.UserDefinedTextFrame); ok {
			if udtf.Description == "STEGO_METADATA" {
				var metadata StegoMetadata
				if err := json.Unmarshal([]byte(udtf.Value), &metadata); err != nil {
					return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
				}
				return &metadata, nil
			}
		}
	}

	return nil, fmt.Errorf("no steganography metadata found in file")
}

func CreateMetadataFromFile(filePath string, useEncryption, useRandomSeed bool, nLsb int) (*StegoMetadata, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &StegoMetadata{
		OriginalFilename: filepath.Base(filePath),
		FileExtension:    filepath.Ext(filePath),
		FileSize:         fileInfo.Size(),
		UseEncryption:    useEncryption,
		UseRandomSeed:    useRandomSeed,
		NLsb:             nLsb,
	}, nil
}
