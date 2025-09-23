package cli

import (
	"audio-steganography-lsb/pkg/embed"
	"audio-steganography-lsb/pkg/extract"

	"github.com/spf13/cobra"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "steganography",
		Short: "Audio steganography using LSB method",
		Long:  "A tool for embedding and extracting secret messages in MP3 audio files using the Least Significant Bit (LSB) method.",
	}

	rootCmd.AddCommand(embedCmd())
	rootCmd.AddCommand(extractCmd())

	return rootCmd.Execute()
}

func embedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "embed",
		Short: "Embed a secret message into an MP3 file",
		Long:  "Embed a secret message into an MP3 audio file using the LSB method.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cover, _ := cmd.Flags().GetString("cover")
			message, _ := cmd.Flags().GetString("message")
			key, _ := cmd.Flags().GetString("key")
			lsb, _ := cmd.Flags().GetInt("lsb")
			random, _ := cmd.Flags().GetBool("random")
			output, _ := cmd.Flags().GetString("output")

			config := &embed.EmbedConfig{
				CoverAudio:    cover,
				SecretMessage: message,
				StegoKey:      key,
				NLsb:          lsb,
				UseRandomSeed: random,
				UseEncryption: false,
				OutputPath:    output,
			}

			return embed.Embed(config)
		},
	}

	cmd.Flags().StringP("cover", "c", "", "Cover audio file (MP3)")
	cmd.Flags().StringP("message", "m", "", "Secret file to embed (any file type)")
	cmd.Flags().StringP("key", "k", "", "Steganography key (max 25 characters)")
	cmd.Flags().IntP("lsb", "l", 1, "Number of LSB bits to use (1-4)")
	cmd.Flags().BoolP("random", "r", false, "Use random seed for embedding positions")
	cmd.Flags().StringP("output", "o", "", "Output stego audio file")

	cmd.MarkFlagRequired("cover")
	cmd.MarkFlagRequired("message")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("output")

	return cmd
}

func extractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract a secret message from an MP3 file",
		Long:  "Extract a secret message from an MP3 audio file that contains embedded data.",
		RunE: func(cmd *cobra.Command, args []string) error {
			stego, _ := cmd.Flags().GetString("stego")
			key, _ := cmd.Flags().GetString("key")
			output, _ := cmd.Flags().GetString("output")

			config := &extract.ExtractConfig{
				StegoAudio: stego,
				StegoKey:   key,
				OutputPath: output,
			}

			return extract.Extract(config)
		},
	}

	cmd.Flags().StringP("stego", "s", "", "Stego audio file (MP3)")
	cmd.Flags().StringP("key", "k", "", "Steganography key (max 25 characters)")
	cmd.Flags().StringP("output", "o", "", "Output extracted file")

	cmd.MarkFlagRequired("stego")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("output")

	return cmd
}
