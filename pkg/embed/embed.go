package embed

import (
	"encoding/json"
	"fmt"
	"os"

	"audio-steganography-lsb/pkg/metadata"
	"audio-steganography-lsb/pkg/utils"

	"github.com/hajimehoshi/go-mp3"
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

	stegoMetadata, err := metadata.CreateMetadataFromFile(
		config.SecretMessage,
		config.UseEncryption,
		config.UseRandomSeed,
		config.NLsb,
	)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	coverFile, err := os.Open(config.CoverAudio)
	if err != nil {
		return fmt.Errorf("failed to open cover audio: %w", err)
	}
	defer coverFile.Close()

	decoder, err := mp3.NewDecoder(coverFile)
	if err != nil {
		return fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	var audioSamples []int16
	buffer := make([]byte, 1024)
	for {
		n, err := decoder.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("failed to read audio data: %w", err)
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

	capacity := utils.CalculateCapacity(len(audioSamples), config.NLsb)
	
	metadataSize := calculateMetadataSize(stegoMetadata)
	totalDataSize := len(messageData) + metadataSize

	if totalDataSize > capacity {
		return fmt.Errorf("message too large: need %d bytes, capacity is %d bytes", totalDataSize, capacity)
	}

	
	if err := copyFile(config.CoverAudio, config.OutputPath); err != nil {
		return fmt.Errorf("failed to copy cover audio: %w", err)
	}

	if err := metadata.StoreMetadata(config.OutputPath, stegoMetadata); err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	if err := metadata.StoreSecretMessage(config.OutputPath, messageData); err != nil {
		return fmt.Errorf("failed to store secret message: %w", err)
	}

	fmt.Printf("Successfully embedded %d bytes using ID3 tag steganography\n", len(messageData))

	return nil
}

func calculateMetadataSize(metadata *metadata.StegoMetadata) int {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return 200 
	}
	return len(metadataJSON)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
