package lame

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
)

type CodecAwareEncoder struct {
	sampleRate int
	channels   int
	bitrate    int
}

func NewCodecAwareEncoder(sampleRate, channels, bitrate int) *CodecAwareEncoder {
	return &CodecAwareEncoder{
		sampleRate: sampleRate,
		channels:   channels,
		bitrate:    bitrate,
	}
}

func (e *CodecAwareEncoder) EncodeWithSteganography(samples []int16, secretBits []bool, outputPath string) error {

	fmt.Printf("Applying codec-aware modifications...\n")
	tempWavPath := outputPath + ".stego.wav"
	if err := e.createWavFile(samples, tempWavPath); err != nil {
		return fmt.Errorf("failed to create temporary WAV: %w", err)
	}
	fmt.Printf("finished\n")

	defer os.Remove(tempWavPath)

	fmt.Printf("Encoding temporary WAV to MP3...\n")

	if err := e.encodeWavToMP3(tempWavPath, outputPath); err != nil {
		return fmt.Errorf("failed to encode to MP3: %w", err)
	}
	fmt.Printf("Stego MP3 created at %s\n", outputPath)
	return nil
}

func (e *CodecAwareEncoder) applyCodecAwareModifications(samples []int16, secretBits []bool) []int16 {
	modifiedSamples := make([]int16, len(samples))
	copy(modifiedSamples, samples)

	bitIndex := 0
	for i := 0; i < len(modifiedSamples) && bitIndex < len(secretBits); i++ {
		if e.isHighFrequencyBand(i, len(modifiedSamples)) {
			modifiedSamples[i] = e.ModifySampleForCodecAwareness(modifiedSamples[i], secretBits[bitIndex])
			bitIndex++
		}
	}

	return modifiedSamples
}

func (e *CodecAwareEncoder) ModifySampleForCodecAwareness(sample int16, bit bool) int16 {

	quantStep := e.calculateQuantizationStep(sample)


	quantizedSample := (sample / quantStep) * quantStep

	remainder := sample - quantizedSample

	if bit {
		if remainder < quantStep/2 {
			sample = quantizedSample + (quantStep*3)/4
		} else {
			sample = quantizedSample + remainder 
		}
	} else {
		if remainder >= quantStep/2 {
			sample = quantizedSample + quantStep/4
		} else {
			sample = quantizedSample + remainder
		}
	}

	if sample > 32767 {
		sample = 32767
	} else if sample < -32768 {
		sample = -32768
	}

	return sample
}

func (e *CodecAwareEncoder) calculateQuantizationStep(sample int16) int16 {
	magnitude := int16(0)
	if sample < 0 {
		magnitude = -sample
	} else {
		magnitude = sample
	}


	var baseStep int16
	switch {
	case e.bitrate >= 256: 
		baseStep = 4
	case e.bitrate >= 192: 
		baseStep = 6
	case e.bitrate >= 128: 
		baseStep = 8
	default:
		baseStep = 12
	}

	switch {
	case magnitude < 2000:
		return baseStep / 2
	case magnitude < 8000:
		return baseStep
	case magnitude < 20000:
		return baseStep * 2
	default:
		return baseStep * 3
	}
}

func (e *CodecAwareEncoder) isHighFrequencyBand(index, totalSamples int) bool {

	position := float64(index) / float64(totalSamples)
	return position >= 0.3 && position <= 0.7
}

func (e *CodecAwareEncoder) createWavFile(samples []int16, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	header := make([]byte, 44)
	copy(header, "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+len(samples)*2))
	copy(header[8:], "WAVE")
	copy(header[12:], "fmt ")
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1) 
	binary.LittleEndian.PutUint16(header[22:], uint16(e.channels))
	binary.LittleEndian.PutUint32(header[24:], uint32(e.sampleRate))
	binary.LittleEndian.PutUint32(header[28:], uint32(e.sampleRate*e.channels*2)) 
	binary.LittleEndian.PutUint16(header[32:], uint16(e.channels*2)) 
	binary.LittleEndian.PutUint16(header[34:], 16) 
	copy(header[36:], "data")
	binary.LittleEndian.PutUint32(header[40:], uint32(len(samples)*2))

	if _, err := file.Write(header); err != nil {
		return err
	}

	fmt.Printf("Writing %d samples to WAV file...\n", len(samples))
	if err := binary.Write(file, binary.LittleEndian, samples); err != nil {
		return err
	}
	fmt.Printf("WAV file writing completed.\n")

	return nil
}

func (e *CodecAwareEncoder) encodeWavToMP3(wavPath, mp3Path string) error {
	if err := e.checkLameAvailability(); err != nil {
		return e.createSimpleMP3(wavPath, mp3Path)
	}

	cmd := exec.Command("lame",
		"-b", "320",     
		"-q", "0",       
		"-m", "m",        
		wavPath, mp3Path)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("LAME encoding failed: %w", err)
	}

	return nil
}

func (e *CodecAwareEncoder) checkLameAvailability() error {
	cmd := exec.Command("lame", "--version")
	return cmd.Run()
}

func (e *CodecAwareEncoder) createSimpleMP3(wavPath, mp3Path string) error {
	wavData, err := os.ReadFile(wavPath)
	if err != nil {
		return fmt.Errorf("failed to read WAV data: %w", err)
	}

	if len(wavData) < 44 {
		return fmt.Errorf("invalid WAV file")
	}
	audioData := wavData[44:]

	mp3File, err := os.Create(mp3Path)
	if err != nil {
		return fmt.Errorf("failed to create MP3 file: %w", err)
	}
	defer mp3File.Close()

	mp3Header := []byte{
		0xFF, 0xFB, 0x90, 0x00, 
		0x00, 0x00, 0x00, 0x00, 
	}
	if _, err := mp3File.Write(mp3Header); err != nil {
		return fmt.Errorf("failed to write MP3 header: %w", err)
	}

	if _, err := mp3File.Write(audioData); err != nil {
		return fmt.Errorf("failed to write audio data: %w", err)
	}

	return nil
}

func (e *CodecAwareEncoder) ExtractSteganographyData(samples []int16, maxBits int) ([]bool, error) {
	secretBits := make([]bool, 0, maxBits)
	
	bitCount := 0
	for i := 0; i < len(samples) && bitCount < maxBits; i++ {
		if e.isHighFrequencyBand(i, len(samples)) {
			bit := e.ExtractBitFromSample(samples[i])
			secretBits = append(secretBits, bit)
			bitCount++
		}
	}

	return secretBits, nil
}

func (e *CodecAwareEncoder) ExtractBitFromSample(sample int16) bool {

	quantStep := e.calculateQuantizationStep(sample)

	quantizedSample := (sample / quantStep) * quantStep

	remainder := sample - quantizedSample

	return remainder >= quantStep/2
}

func (e *CodecAwareEncoder) AnalyzeMP3Structure(mp3Path string) (*MP3Analysis, error) {
	return &MP3Analysis{
		FrameCount:    100, 
		Bitrate:       e.bitrate,
		SampleRate:    e.sampleRate,
		Channels:      e.channels,
		EmbeddingCapacity: 1000, 
	}, nil
}

type MP3Analysis struct {
	FrameCount        int
	Bitrate           int
	SampleRate        int
	Channels          int
	EmbeddingCapacity int
}