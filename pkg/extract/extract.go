package extract

import (
	"encoding/binary"
	"fmt"
	"os"

	"audio-steganography-lsb/pkg/lame"
	"audio-steganography-lsb/pkg/utils"
	"audio-steganography-lsb/pkg/vigenere"

	"github.com/hajimehoshi/go-mp3"
)

type ExtractConfig struct {
	StegoAudio string
	StegoKey   string
	OutputPath string
	UseDecryption bool
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

	fmt.Printf("DEBUG: Trying MP3 bitstream extraction\n")
	messageData, err := extractMP3Bitstream(config.StegoAudio, config.StegoKey)
	if err != nil {
		fmt.Printf("DEBUG: MP3 bitstream extraction failed: %v\n", err)

		fmt.Printf("DEBUG: Falling back to sample-based extraction\n")
		audioSamples, err := readAudioSamples(config.StegoAudio)
		if err != nil {
			return fmt.Errorf("failed to read audio samples: %w", err)
		}

		fmt.Printf("DEBUG: Trying codec-aware extraction with %d samples\n", len(audioSamples))
		messageData, err = extractCodecAwareLSB(audioSamples, config.StegoKey)
		if err != nil {
			fmt.Printf("DEBUG: Codec-aware extraction failed: %v\n", err)
			fmt.Printf("DEBUG: Trying original LSB extraction\n")
			messageData, err = extractLSB(audioSamples, config.StegoKey)
			if err != nil {
				fmt.Printf("DEBUG: Original LSB extraction failed: %v\n", err)
				return fmt.Errorf("failed to extract data using all methods: %w", err)
			}
		}
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

func extractMP3Bitstream(stegoPath, stegoKey string) ([]byte, error) {
	mp3Data, err := os.ReadFile(stegoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MP3 file: %w", err)
	}

	embeddablePositions := findEmbeddablePositions(mp3Data)
	if len(embeddablePositions) < 64 {
		return nil, fmt.Errorf("not enough embeddable positions")
	}

	paramHeader, err := extractParameterHeader(mp3Data, embeddablePositions[:64])
	if err != nil {
		fmt.Printf("DEBUG: Failed to extract parameter header: %v\n", err)
		return extractMP3BitstreamLegacy(mp3Data, embeddablePositions, stegoKey)
	}

	params, err := parseParameterHeader(paramHeader, stegoKey)
	if err != nil {
		fmt.Printf("DEBUG: Invalid parameter header: %v\n", err)
		return extractMP3BitstreamLegacy(mp3Data, embeddablePositions, stegoKey)
	}

	fmt.Printf("DEBUG: Found valid parameter header - nLsb=%d, useRandom=%t\n", params.nLsb, params.useRandomSeed)

	dataPositions := embeddablePositions[64:]

	maxPositionsNeeded := 100000 
	segmentSize := 50000

	if len(dataPositions) > maxPositionsNeeded {
		fmt.Printf("DEBUG: Large file detected with %d positions, trying segmented extraction\n", len(dataPositions))

		dataPositions = dataPositions[:segmentSize]
	}

	positions, err := utils.GeneratePositions(stegoKey, params.useRandomSeed, len(dataPositions), params.nLsb)
	if err != nil {
		return nil, fmt.Errorf("failed to generate positions: %w", err)
	}

	headerData, err := extractLimitedMP3Data(mp3Data, dataPositions, positions, params.nLsb, 1024) 
	if err != nil {
		return nil, fmt.Errorf("failed to extract header data: %w", err)
	}

	if len(headerData) < 8 {
		return nil, fmt.Errorf("not enough header data extracted")
	}

	metadataLen := binary.LittleEndian.Uint32(headerData[0:4])
	fmt.Printf("DEBUG: MP3 bitstream metadata length: %d\n", metadataLen)

	if int(metadataLen) > len(headerData)-4 || metadataLen > 10000 { 
		return nil, fmt.Errorf("invalid metadata length: %d", metadataLen)
	}

	if len(headerData) < int(4+metadataLen+4) {
		headerData, err = extractLimitedMP3Data(mp3Data, dataPositions, positions, params.nLsb, int(metadataLen)+100)
		if err != nil {
			return nil, fmt.Errorf("failed to extract extended header: %w", err)
		}
	}

	if len(headerData) < int(4+metadataLen+4) {
		return nil, fmt.Errorf("insufficient data for message length")
	}

	messageLen := binary.LittleEndian.Uint32(headerData[4+metadataLen:4+metadataLen+4])
	fmt.Printf("DEBUG: MP3 bitstream message length: %d\n", messageLen)

	if messageLen > 100*1024*1024 {
		return nil, fmt.Errorf("message length too large: %d", messageLen)
	}

	totalDataSize := int(4 + metadataLen + 4 + messageLen)

	data, err := extractLimitedMP3Data(mp3Data, dataPositions, positions, params.nLsb, totalDataSize)
	if err != nil {
		return nil, fmt.Errorf("failed to extract full data: %w", err)
	}

	if len(data) < totalDataSize {
		return nil, fmt.Errorf("insufficient data extracted: got %d, need %d", len(data), totalDataSize)
	}

	messageStart := 4 + int(metadataLen) + 4
	messageEnd := messageStart + int(messageLen)
	if messageEnd <= len(data) && messageLen > 0 {
		fmt.Printf("DEBUG: MP3 bitstream validation passed, returning message\n")
		return data[messageStart:messageEnd], nil
	}

	return nil, fmt.Errorf("failed to extract message data")
}

func extractLimitedMP3Data(mp3Data []byte, dataPositions []int, positions []int, nLsb int, maxBytes int) ([]byte, error) {
	maxBits := maxBytes * 8

	bitsPerPosition := nLsb
	positionsNeeded := (maxBits + bitsPerPosition - 1) / bitsPerPosition 

	if positionsNeeded > len(positions) {
		positionsNeeded = len(positions)
	}

	limitedPositions := positions[:positionsNeeded]

	fmt.Printf("DEBUG: Extracting limited data - need %d bytes (%d bits), using %d positions\n", maxBytes, maxBits, len(limitedPositions))

	var bits []bool
	for _, posIndex := range limitedPositions {
		if posIndex >= len(dataPositions) {
			break
		}

		actualPos := dataPositions[posIndex]
		if actualPos >= len(mp3Data) {
			continue
		}

		for i := 0; i < nLsb; i++ {
			bit := (mp3Data[actualPos] >> i) & 1
			bits = append(bits, bit == 1)
		}
	}

	if len(bits) < 8 {
		return nil, fmt.Errorf("not enough bits extracted")
	}

	bits = bits[:len(bits)-(len(bits)%8)]

	bytes := make([]byte, len(bits)/8)
	for i := 0; i < len(bytes); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << j
			}
		}
		bytes[i] = b
	}

	return bytes, nil
}

type ParameterHeader struct {
	nLsb          int
	useRandomSeed bool
}

func extractParameterHeader(mp3Data []byte, positions []int) ([]byte, error) {
	if len(positions) < 64 {
		return nil, fmt.Errorf("not enough positions for header")
	}

	var headerBits []bool
	for i := 0; i < 64; i++ {
		pos := positions[i]
		if pos >= len(mp3Data) {
			return nil, fmt.Errorf("position out of bounds")
		}

		bit := (mp3Data[pos] & 0x01) == 1
		headerBits = append(headerBits, bit)
	}

	headerBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if headerBits[i*8+j] {
				b |= 1 << j
			}
		}
		headerBytes[i] = b
	}

	return headerBytes, nil
}

func parseParameterHeader(header []byte, stegoKey string) (*ParameterHeader, error) {
	if len(header) != 8 {
		return nil, fmt.Errorf("invalid header length")
	}

	if header[0] != 0xAB || header[1] != 0xCD {
		return nil, fmt.Errorf("invalid magic bytes")
	}

	nLsb := int(header[2])
	if nLsb < 1 || nLsb > 4 {
		return nil, fmt.Errorf("invalid nLsb value: %d", nLsb)
	}

	useRandomSeed := header[3] == 1

	expectedKeySum := uint32(0)
	for _, b := range []byte(stegoKey) {
		expectedKeySum += uint32(b)
	}
	actualKeySum := binary.LittleEndian.Uint32(header[4:8])

	if actualKeySum != expectedKeySum {
		return nil, fmt.Errorf("key checksum mismatch")
	}

	return &ParameterHeader{
		nLsb:          nLsb,
		useRandomSeed: useRandomSeed,
	}, nil
}

func extractMP3BitstreamLegacy(mp3Data []byte, embeddablePositions []int, stegoKey string) ([]byte, error) {
	fmt.Printf("DEBUG: Using legacy extraction method (guessing parameters)\n")

	for nLsb := 1; nLsb <= 4; nLsb++ {
		for useRandom := 0; useRandom < 2; useRandom++ {
			useRandomSeed := useRandom == 1

			positions, err := utils.GeneratePositions(stegoKey, useRandomSeed, len(embeddablePositions), nLsb)
			if err != nil {
				continue
			}

			data, err := extractFromMP3Frames(mp3Data, embeddablePositions, positions, nLsb)
			if err != nil {
				continue
			}

			if len(data) > 0 {
				fmt.Printf("DEBUG: MP3 bitstream got %d bytes of data\n", len(data))
				if len(data) >= 8 {
					metadataLen := binary.LittleEndian.Uint32(data[0:4])
					fmt.Printf("DEBUG: MP3 bitstream metadata length: %d\n", metadataLen)
					if int(metadataLen) < len(data)-4 {
						messageLen := binary.LittleEndian.Uint32(data[4+metadataLen:4+metadataLen+4])
						fmt.Printf("DEBUG: MP3 bitstream message length: %d\n", messageLen)
						if int(4+metadataLen+4+messageLen) <= len(data) {
							messageStart := 4 + int(metadataLen) + 4
							messageEnd := messageStart + int(messageLen)
							if messageEnd <= len(data) && messageLen > 0 {
								fmt.Printf("DEBUG: MP3 bitstream validation passed, returning message\n")
								return data[messageStart:messageEnd], nil
							}
						}
					}
				}
				fmt.Printf("DEBUG: MP3 bitstream data validation failed for nLsb=%d, useRandom=%d\n", nLsb, useRandom)
				continue 
			}
		}
	}

	return nil, fmt.Errorf("failed to extract data from MP3 bitstream")
}

func extractFromMP3Frames(mp3Data []byte, embeddablePositions []int, positions []int, nLsb int) ([]byte, error) {
	var bits []bool

	fmt.Printf("DEBUG: MP3 extraction from %d positions with nLsb=%d\n", len(positions), nLsb)

	for _, posIndex := range positions {
		if posIndex >= len(embeddablePositions) {
			break
		}

		actualPos := embeddablePositions[posIndex]
		if actualPos >= len(mp3Data) {
			continue
		}

		for i := 0; i < nLsb; i++ {
			bit := (mp3Data[actualPos] >> i) & 1
			bits = append(bits, bit == 1)
		}
	}

	fmt.Printf("DEBUG: MP3 extracted %d bits\n", len(bits))

	if len(bits) < 8 {
		return nil, fmt.Errorf("not enough bits extracted")
	}

	bits = bits[:len(bits)-(len(bits)%8)]

	bytes := make([]byte, len(bits)/8)
	for i := 0; i < len(bytes); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << j
			}
		}
		bytes[i] = b
	}

	return bytes, nil
}

func findEmbeddablePositions(mp3Data []byte) []int {
	var positions []int


	skipStart := 512 

	frameHeaders := make(map[int]bool)
	for i := 0; i < len(mp3Data)-1; i++ {
		if mp3Data[i] == 0xFF && (mp3Data[i+1]&0xE0) == 0xE0 {
			for j := 0; j < 4 && i+j < len(mp3Data); j++ {
				frameHeaders[i+j] = true
			}
		}
	}

	for i := skipStart; i < len(mp3Data); i++ {
		if frameHeaders[i] {
			continue
		}

		if mp3Data[i] == 0xFF && i < len(mp3Data)-1 && (mp3Data[i+1]&0xE0) == 0xE0 {
			continue
		}

		positions = append(positions, i)
	}

	if len(positions) < 10000 && len(mp3Data) > 2000 {
		positions = []int{}
		start := 2000
		step := 1 

		for i := start; i < len(mp3Data); i += step {
			if mp3Data[i] == 0xFF {
				continue
			}

			if i > 0 && mp3Data[i-1] == 0xFF && (mp3Data[i]&0xE0) == 0xE0 {
				continue
			}

			if i < len(mp3Data)-1 && mp3Data[i+1] == 0xFF {
				continue
			}

			positions = append(positions, i)
		}
	}

	return positions
}

func readAudioSamples(filePath string) ([]int16, error) {
	coverFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stego audio: %w", err)
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

func extractCodecAwareLSB(audioSamples []int16, stegoKey string) ([]byte, error) {
	encoder := lame.NewCodecAwareEncoder(44100, 1, 320) 
	// fmt.Printf("DEBUG: Codec-aware extraction with %d samples\n", len(audioSamples))
	
	for nLsb := 1; nLsb <= 4; nLsb++ {
		for useRandom := 0; useRandom < 2; useRandom++ {
			useRandomSeed := useRandom == 1
			
			positions, err := utils.GeneratePositions(stegoKey, useRandomSeed, len(audioSamples), nLsb)
			if err != nil {
				continue
			}
		
			data, err := extractCodecAwareData(audioSamples, positions, nLsb, encoder)
			if err != nil {
				continue
			}
			
			if len(data) > 0 {
				// fmt.Printf("DEBUG: Codec-aware got %d bytes of data\n", len(data))
				if len(data) >= 8 { 
					metadataLen := binary.LittleEndian.Uint32(data[0:4])
					// fmt.Printf("DEBUG: Metadata length: %d\n", metadataLen)
					if int(metadataLen) < len(data)-4 {
						messageLen := binary.LittleEndian.Uint32(data[4+metadataLen:4+metadataLen+4])
						// fmt.Printf("DEBUG: Message length: %d\n", messageLen)
						if int(4+metadataLen+4+messageLen) <= len(data) {
							messageStart := 4 + int(metadataLen) + 4
							messageEnd := messageStart + int(messageLen)
							if messageEnd <= len(data) && messageLen > 0 {
								// fmt.Printf("DEBUG: Codec-aware validation passed, returning message\n")
								return data[messageStart:messageEnd], nil
							}
						}
					}
				}
				// fmt.Printf("DEBUG: Codec-aware data validation failed for nLsb=%d, useRandom=%d\n", nLsb, useRandom)
				continue 
			}
		}
	}
	
	// fmt.Printf("DEBUG: No valid codec-aware embedding found\n")
	return nil, fmt.Errorf("failed to extract data - no valid codec-aware embedding found")
}

func extractCodecAwareData(audioSamples []int16, positions []int, nLsb int, encoder *lame.CodecAwareEncoder) ([]byte, error) {
	var secretBits []bool

	bitCount := 0
	for _, pos := range positions {
		if pos >= len(audioSamples) {
			break
		}

		sample := audioSamples[pos]

		for i := 0; i < nLsb && bitCount < len(positions)*nLsb; i++ {
			bit := encoder.ExtractBitFromSample(sample)
			secretBits = append(secretBits, bit)
			bitCount++
		}
	}

	// fmt.Printf("DEBUG: Codec-aware extracted %d bits from %d positions\n", len(secretBits), len(positions))

	if len(secretBits) < 8 {
		return nil, fmt.Errorf("not enough bits extracted: got %d bits", len(secretBits))
	}

	secretBits = secretBits[:len(secretBits)-(len(secretBits)%8)]

	bytes := make([]byte, len(secretBits)/8)
	for i := 0; i < len(bytes); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if secretBits[i*8+j] {
				b |= 1 << j
			}
		}
		bytes[i] = b
	}

	return bytes, nil
}

func extractLSB(audioSamples []int16, stegoKey string) ([]byte, error) {
	
	for nLsb := 1; nLsb <= 4; nLsb++ {
		for useRandom := 0; useRandom < 2; useRandom++ {
			useRandomSeed := useRandom == 1
			
			positions, err := utils.GeneratePositions(stegoKey, useRandomSeed, len(audioSamples), nLsb)
			if err != nil {
				continue
			}
		
		// fmt.Printf("DEBUG: Original LSB trying nLsb=%d, useRandom=%d\n", nLsb, useRandom)
		data, err := extractDataFromSamples(audioSamples, positions, nLsb)
		if err != nil {
			// fmt.Printf("DEBUG: Original LSB extraction failed: %v\n", err)
			continue
		}
		// fmt.Printf("DEBUG: Original LSB extracted %d bytes\n", len(data))
			
			if len(data) > 0 {
				if len(data) >= 8 { 
					metadataLen := binary.LittleEndian.Uint32(data[0:4])
					// fmt.Printf("DEBUG: Original LSB metadata length: %d\n", metadataLen)
					if int(metadataLen) < len(data)-4 {
						messageLen := binary.LittleEndian.Uint32(data[4+metadataLen:4+metadataLen+4])
						// fmt.Printf("DEBUG: Original LSB message length: %d\n", messageLen)
						if int(4+metadataLen+4+messageLen) <= len(data) {
							messageStart := 4 + int(metadataLen) + 4
							messageEnd := messageStart + int(messageLen)
							if messageEnd <= len(data) && messageLen > 0 {
								// fmt.Printf("DEBUG: Original LSB validation passed, returning message\n")
								return data[messageStart:messageEnd], nil
							}
						}
					}
				}
				// fmt.Printf("DEBUG: Original LSB data validation failed for nLsb=%d, useRandom=%d\n", nLsb, useRandom)
			}
		}
	}
	
	return nil, fmt.Errorf("failed to extract data - no valid LSB embedding found")
}

func extractDataFromSamples(audioSamples []int16, positions []int, nLsb int) ([]byte, error) {
	var bits []bool
	
	// fmt.Printf("DEBUG: Extracting from %d positions with nLsb=%d\n", len(positions), nLsb)
	
	for _, pos := range positions {
		if pos >= len(audioSamples) {
			break
		}
		
		sample := audioSamples[pos]
		
		for i := 0; i < nLsb; i++ {
			bit := (sample >> i) & 1
			bits = append(bits, bit == 1)
		}
	}
	
	// fmt.Printf("DEBUG: Extracted %d bits\n", len(bits))
	
	if len(bits) < 8 {
		return nil, fmt.Errorf("not enough bits extracted")
	}
	
	bits = bits[:len(bits)-(len(bits)%8)]
	
	bytes := make([]byte, len(bits)/8)
	for i := 0; i < len(bytes); i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << j
			}
		}
		bytes[i] = b
	}
	
	return bytes, nil
}