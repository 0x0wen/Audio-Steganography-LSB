package cli

import (
	"audio-steganography-lsb/pkg/embed"
	"audio-steganography-lsb/pkg/extract"
//	"audio-steganography-lsb/pkg/encrypt" // added import for encryption

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
			encrypt, _ := cmd.Flags().GetBool("encrypt") // args untuk enkripsi
			output, _ := cmd.Flags().GetString("output")

			config := &embed.EmbedConfig{
				CoverAudio:    cover,
				SecretMessage: message,
				StegoKey:      key,
				NLsb:          lsb,
				UseRandomSeed: random,
				UseEncryption: encrypt, // set config sesuai var encrypt
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
	cmd.Flags().BoolP("encrypt", "e", false, "Encrypt the message before embedding") // flag untuk enkripsi
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
			decrypt, _ := cmd.Flags().GetBool("decrypt") // args untuk enkripsi


			config := &extract.ExtractConfig{
				StegoAudio: stego,
				StegoKey:   key,
				OutputPath: output,
				UseDecryption: decrypt, // set config sesuai var decrypt

			}

			return extract.Extract(config)
		},
	}

	cmd.Flags().StringP("stego", "s", "", "Stego audio file (MP3)")
	cmd.Flags().StringP("key", "k", "", "Steganography key (max 25 characters)")
	cmd.Flags().StringP("output", "o", "", "Output extracted file")
	cmd.Flags().BoolP("decrypt", "d", false, "Decrypt the message after extracting") // flag untuk enkripsi


	cmd.MarkFlagRequired("stego")
	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("output")

	return cmd
}
