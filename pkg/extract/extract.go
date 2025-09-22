package extract

import (
	"fmt"

	"audio-steganography-lsb/pkg/metadata"
	"audio-steganography-lsb/pkg/utils"
)

type ExtractConfig struct {
	StegoAudio string
	StegoKey   string
	OutputPath string
}

func Extract(config *ExtractConfig) error {
	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
		return fmt.Errorf("invalid stego key: %w", err)
	}

	messageData, err := metadata.RetrieveSecretMessage(config.StegoAudio)
	if err != nil {
		return fmt.Errorf("failed to retrieve secret message: %w", err)
	}

	if err := utils.WriteFile(config.OutputPath, messageData); err != nil {
		return fmt.Errorf("failed to write extracted message: %w", err)
	}

	fmt.Printf("Successfully extracted %d bytes to %s\n", len(messageData), config.OutputPath)
	return nil
}
