package embed

// import (
// 	"encoding/binary"
// 	"fmt"
// 	"os"
// 	"path/filepath"

// 	"audio-steganography-lsb/pkg/utils"

// 	// "github.com/hajimehoshi/go-mp3"

// 	"audio-steganography-lsb/pkg/vigenere"
// )

// type EmbedConfig struct {
// 	CoverAudio     string
// 	SecretMessage  string
// 	StegoKey       string
// 	NLsb           int
// 	UseRandomSeed  bool
// 	UseEncryption  bool
// 	OutputPath     string
// }

// type FileMetadata struct {
// 	OriginalFilename string
// 	FileExtension    string
// 	FileSize         int64
// 	UseEncryption    bool
// 	UseRandomSeed    bool
// 	NLsb             int
// 	DataSize         int64
// }

// func Embed(config *EmbedConfig) error {
// 	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
// 		return fmt.Errorf("invalid stego key: %w", err)
// 	}

// 	if err := utils.ValidateNLsb(config.NLsb); err != nil {
// 		return fmt.Errorf("invalid n_lsb: %w", err)
// 	}

// 	messageData, err := utils.ReadFile(config.SecretMessage)
// 	if err != nil {
// 		return fmt.Errorf("failed to read secret message: %w", err)
// 	}else if config.UseEncryption{
// 		messageData = vigenere.Encrypt(messageData, config.StegoKey)
// 	}

// 	// stegoMetadata, err := metadata.CreateMetadataFromFile(
// 	// 	config.SecretMessage,
// 	// 	config.UseEncryption,
// 	// 	config.UseRandomSeed,
// 	// 	config.NLsb,
// 	// )
// 	fileInfo, err := os.Stat(config.SecretMessage)
// 	if err != nil {
// 		return fmt.Errorf("failed to get file info: %w", err)
// 	}

// 	metadata := &FileMetadata{
// 		OriginalFilename: filepath.Base(config.SecretMessage),
// 		FileExtension:    filepath.Ext(config.SecretMessage),
// 		FileSize:         fileInfo.Size(),
// 		UseEncryption:    config.UseEncryption,
// 		UseRandomSeed:    config.UseRandomSeed,
// 		NLsb:             config.NLsb,
// 		DataSize:         int64(len(messageData)),
// 	}

// 	if err := embedMP3Bitstream(config.CoverAudio, messageData, metadata, config.StegoKey, config.UseRandomSeed, config.NLsb, config.OutputPath); err != nil {
// 		return fmt.Errorf("failed to embed data in MP3 bitstream: %w", err)
// 	}

// 	fmt.Printf("Successfully embedded %d bytes using MP3 bitstream steganography\n", len(messageData))
// 	return nil
// }

// func embedMP3Bitstream(coverPath string, messageData []byte, metadata *FileMetadata, stegoKey string, useRandomSeed bool, nLsb int, outputPath string) error {
// 	mp3Data, err := os.ReadFile(coverPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to read MP3 file: %w", err)
// 	}

// 	paramHeader := createParameterHeader(nLsb, useRandomSeed, stegoKey)

// 	embeddablePositions := findEmbeddablePositions(mp3Data)
// 	if len(embeddablePositions) < 64 {
// 		return fmt.Errorf("not enough embeddable positions for parameter header")
// 	}

// 	modifiedMP3Data := make([]byte, len(mp3Data))
// 	copy(modifiedMP3Data, mp3Data)

// 	if err := embedParameterHeader(modifiedMP3Data, embeddablePositions[:64], paramHeader); err != nil {
// 		return fmt.Errorf("failed to embed parameter header: %w", err)
// 	}

// 	metadataBytes := serializeMetadata(metadata)

// 	dataToEmbed := make([]byte, 0, 4+len(metadataBytes)+4+len(messageData))

// 	metadataLenBytes := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(metadataLenBytes, uint32(len(metadataBytes)))
// 	dataToEmbed = append(dataToEmbed, metadataLenBytes...)

// 	dataToEmbed = append(dataToEmbed, metadataBytes...)

// 	messageLenBytes := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(messageLenBytes, uint32(len(messageData)))
// 	dataToEmbed = append(dataToEmbed, messageLenBytes...)

// 	dataToEmbed = append(dataToEmbed, messageData...)

// 	fmt.Printf("Total data to embed (metadata + message): %d bytes\n", len(dataToEmbed))

// 	if err := embedDataInMP3Frames(modifiedMP3Data, embeddablePositions[64:], dataToEmbed, stegoKey, useRandomSeed, nLsb); err != nil {
// 		return fmt.Errorf("failed to embed data in MP3 frames: %w", err)
// 	}

// 	if err := os.WriteFile(outputPath, modifiedMP3Data, 0644); err != nil {
// 		return fmt.Errorf("failed to write output MP3: %w", err)
// 	}

// 	return nil
// }

// func createParameterHeader(nLsb int, useRandomSeed bool, stegoKey string) []byte {
// 	header := make([]byte, 8)

// 	header[0] = 0xAB 
// 	header[1] = 0xCD 

// 	header[2] = byte(nLsb)

// 	if useRandomSeed {
// 		header[3] = 1
// 	} else {
// 		header[3] = 0
// 	}

// 	keySum := uint32(0)
// 	for _, b := range []byte(stegoKey) {
// 		keySum += uint32(b)
// 	}
// 	binary.LittleEndian.PutUint32(header[4:8], keySum)

// 	return header
// }

// func embedParameterHeader(mp3Data []byte, positions []int, header []byte) error {
// 	if len(positions) < len(header)*8 {
// 		return fmt.Errorf("not enough positions for header")
// 	}

// 	headerBits := bytesToBits(header)

// 	for i, bit := range headerBits {
// 		if i >= len(positions) {
// 			break
// 		}

// 		pos := positions[i]
// 		if pos >= len(mp3Data) {
// 			continue
// 		}

// 		mp3Data[pos] = mp3Data[pos] & 0xFE 
// 		if bit {
// 			mp3Data[pos] = mp3Data[pos] | 0x01 
// 		}
// 	}

// 	return nil
// }

// func embedDataInMP3Frames(mp3Data []byte, positions []int, data []byte, stegoKey string, useRandomSeed bool, nLsb int) error {
// 	bits := bytesToBits(data)

// 	// fmt.Printf("Generating Positions")
// 	dataPositions, err := utils.GeneratePositions(stegoKey, useRandomSeed, len(positions), nLsb)
// 	if err != nil {
// 		return fmt.Errorf("failed to generate positions: %w", err)
// 	}
// 	// fmt.Printf("calculating capacity")
// 	capacity := len(dataPositions) * nLsb
// 	if len(bits) > capacity {
// 		return fmt.Errorf("data too large: need %d bits, capacity is %d bits", len(bits), capacity)
// 	}

// 	fmt.Printf("Embedding %d bits into %d positions using %d LSBs (capacity %d bits)\n", len(bits), len(dataPositions), nLsb, capacity)
// 	bitIndex := 0
// 	for _, posIndex := range dataPositions {
// 		if bitIndex >= len(bits) || posIndex >= len(positions) {
// 			break
// 		}

// 		actualPos := positions[posIndex]
// 		if actualPos >= len(mp3Data) {
// 			continue
// 		}

// 		for i := 0; i < nLsb && bitIndex < len(bits); i++ {
// 			bit := bits[bitIndex]

// 			mask := ^(1 << i)
// 			mp3Data[actualPos] = mp3Data[actualPos] & byte(mask)

// 			if bit {
// 				mp3Data[actualPos] = mp3Data[actualPos] | (1 << i)
// 			}
// 			bitIndex++
// 		}
// 	}

// 	return nil
// }

// func findEmbeddablePositions(mp3Data []byte) []int {
// 	var positions []int


// 	skipStart := 512 

// 	frameHeaders := make(map[int]bool)
// 	for i := 0; i < len(mp3Data)-1; i++ {
// 		if mp3Data[i] == 0xFF && (mp3Data[i+1]&0xE0) == 0xE0 {
// 			for j := 0; j < 4 && i+j < len(mp3Data); j++ {
// 				frameHeaders[i+j] = true
// 			}
// 		}
// 	}

// 	for i := skipStart; i < len(mp3Data); i++ {
// 		if frameHeaders[i] {
// 			continue
// 		}

// 		if mp3Data[i] == 0xFF && i < len(mp3Data)-1 && (mp3Data[i+1]&0xE0) == 0xE0 {
// 			continue
// 		}

// 		positions = append(positions, i)
// 	}

// 	if len(positions) < 10000 && len(mp3Data) > 2000 {
// 		positions = []int{}
// 		start := 2000 
// 		step := 1 

// 		for i := start; i < len(mp3Data); i += step {
// 			if mp3Data[i] == 0xFF {
// 				continue 
// 			}

// 			if i > 0 && mp3Data[i-1] == 0xFF && (mp3Data[i]&0xE0) == 0xE0 {
// 				continue
// 			}

// 			if i < len(mp3Data)-1 && mp3Data[i+1] == 0xFF {
// 				continue
// 			}

// 			positions = append(positions, i)
// 		}
// 	}

// 	return positions
// }

// func serializeMetadata(metadata *FileMetadata) []byte {
// 	data := make([]byte, 0, 1024)

// 	data = append(data, byte(len(metadata.OriginalFilename)))

// 	data = append(data, []byte(metadata.OriginalFilename)...)

// 	data = append(data, byte(len(metadata.FileExtension)))

// 	data = append(data, []byte(metadata.FileExtension)...)

// 	sizeBytes := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(sizeBytes, uint64(metadata.FileSize))
// 	data = append(data, sizeBytes...)

// 	flags := byte(0)
// 	if metadata.UseEncryption {
// 		flags |= 1
// 	}
// 	if metadata.UseRandomSeed {
// 		flags |= 2
// 	}
// 	data = append(data, flags)

// 	data = append(data, byte(metadata.NLsb))

// 	dataSizeBytes := make([]byte, 8)
// 	binary.LittleEndian.PutUint64(dataSizeBytes, uint64(metadata.DataSize))
// 	data = append(data, dataSizeBytes...)

// 	return data
// }

// func bytesToBits(data []byte) []bool {
// 	bits := make([]bool, len(data)*8)
// 	for i, b := range data {
// 		for j := 0; j < 8; j++ {
// 			bits[i*8+j] = (b>>j)&1 == 1
// 		}
// 	}
// 	return bits
// }