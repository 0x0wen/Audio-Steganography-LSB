package extract

import (
	"fmt"

	"audio-steganography-lsb/pkg/metadata"
	"audio-steganography-lsb/pkg/utils"
	"audio-steganography-lsb/pkg/encrypt"
)

type ExtractConfig struct {
	StegoAudio string
	StegoKey   string
	OutputPath string
	UseDecryption bool
}

func Extract(config *ExtractConfig) error {
	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
		return fmt.Errorf("invalid stego key: %w", err)
	}

	messageData, err := metadata.RetrieveSecretMessage(config.StegoAudio)
	if err != nil {
		return fmt.Errorf("failed to retrieve secret message: %w", err)
	}

	if config.UseDecryption {
		messageData = vigenere.Decrypt(messageData, config.StegoKey)
	}


	if err := utils.WriteFile(config.OutputPath, messageData); err != nil {
		return fmt.Errorf("failed to write extracted message: %w", err)
	}

	fmt.Printf("Successfully extracted %d bytes to %s\n", len(messageData), config.OutputPath)
	return nil
}
