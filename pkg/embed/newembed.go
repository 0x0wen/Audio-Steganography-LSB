package embed

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"

	"audio-steganography-lsb/pkg/utils"
	"audio-steganography-lsb/pkg/vigenere"
	"audio-steganography-lsb/pkg/psnr"
	"audio-steganography-lsb/pkg/lame"

	"github.com/hajimehoshi/go-mp3"
)

// EmbedConfig mendefinisikan parameter untuk proses penyisipan.
type EmbedConfig struct {
	CoverAudio    string
	SecretMessage string
	StegoKey      string
	NLsb          int
	UseRandomSeed bool
	UseEncryption bool
	OutputPath    string
}

// signature represents a start or end marker for the hidden message.
type signature []bool

var signatures = map[int]struct {
	start signature
	end   signature
}{
	1: { // 1-LSB
		start: signature{true, false, true, false, true, false, true, false, true, false, true, false, true, false},
		end:   signature{true, false, true, false, true, false, true, false, true, false, true, false, true, false},
	},
	2: { // 2-LSB
		start: signature{false, true, false, true, false, true, false, true, false, true, false, true, false, true},
		end:   signature{false, true, false, true, false, true, false, true, false, true, false, true, false, true},
	},
	3: { // 3-LSB
		start: signature{true, false, true, false, true, false, true, false, false, true, false, true, false, true, false, true},
		end:   signature{false, true, false, true, false, true, false, true, true, false, true, false, true, false, true, false},
	},
	4: { // 4-LSB
		start: signature{false, true, false, true, false, true, false, true, true, false, true, false, true, false, true, false},
		end:   signature{true, false, true, false, true, false, true, false, false, true, false, true, false, true, false, true},
	},
}

// Embed menyisipkan pesan rahasia ke dalam file audio.
func Embed(config *EmbedConfig) error {
	fmt.Printf("flag0\n")

	if err := utils.ValidateStegoKey(config.StegoKey); err != nil {
		return fmt.Errorf("kunci stego tidak valid: %w", err)
	}
	if err := utils.ValidateNLsb(config.NLsb); err != nil {
		return fmt.Errorf("n_lsb tidak valid: %w", err)
	}
	fmt.Printf("flag1\n")
	// 1. Baca dan siapkan pesan rahasia
	messageData, err := os.ReadFile(config.SecretMessage)
	if err != nil {
		return fmt.Errorf("gagal membaca pesan rahasia: %w", err)
	}
	if config.UseEncryption {
		messageData = vigenere.Encrypt(messageData, config.StegoKey)
	}
	fmt.Printf("flag2\n")

	// 2. Decode file audio cover menjadi sampel PCM
	originalSamples, sampleRate, err := decodeMP3ToSamples(config.CoverAudio)
	if err != nil {
		return fmt.Errorf("gagal mendekode file MP3: %w", err)
	}

	fmt.Printf("flag3\n")

	// 3. Siapkan payload data untuk disisipkan (metadata + pesan)
	payload := createPayload(config, messageData)
	payloadBits := bytesToBits(payload)

	// Dapatkan signature untuk n-LSB yang dipilih
	sigPair, ok := signatures[config.NLsb]
	if !ok {
		return fmt.Errorf("tidak ada signature yang didefinisikan untuk n-LSB=%d", config.NLsb)
	}
	dataToEmbedBits := append(sigPair.start, payloadBits...)
	dataToEmbedBits = append(dataToEmbedBits, sigPair.end...)

	fmt.Printf("flag4\n")

	// 4. Hitung kapasitas dan periksa apakah pesan muat
	capacityBits := len(originalSamples) * config.NLsb
	if len(dataToEmbedBits) > capacityBits {
		return fmt.Errorf("pesan terlalu besar untuk disisipkan. Kapasitas: %d bit, Dibutuhkan: %d bit", capacityBits, len(dataToEmbedBits))
	}

	fmt.Printf("flag5\n")

	// 5. Terapkan steganografi LSB pada sampel audio
	stegoSamples, err := embedBitsIntoSamples(originalSamples, dataToEmbedBits, config)
	if err != nil {
		return fmt.Errorf("gagal menyisipkan data: %w", err)
	}

	fmt.Printf("flag6\n")

	// 6. Encode sampel yang dimodifikasi kembali ke file MP3
	if err := encodeSamplesToMP3(stegoSamples, config.OutputPath, sampleRate, 2); err != nil {
		return fmt.Errorf("gagal mengenkode ke MP3: %w", err)
	}

	fmt.Printf("flag7\n")
	// 7. Hitung dan kembalikan PSNR
	psnrValue, err := psnr.CalculatePSNR(originalSamples, stegoSamples)
	if err != nil {
		return fmt.Errorf("gagal menghitung PSNR: %w", err)
	}else{
		fmt.Printf("PSNR: %.2f dB\n", psnrValue)
	}

	fmt.Printf("Berhasil menyisipkan %d bytes ke %s\n", len(messageData), config.OutputPath)
	return nil
}

// createPayload membuat array byte yang berisi metadata dan data pesan.
func createPayload(config *EmbedConfig, messageData []byte) []byte {
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.LittleEndian, uint32(len(messageData)))
	payload.Write(messageData)
	return payload.Bytes()
}

// embedBitsIntoSamples menyisipkan bit data ke dalam sampel audio.
func embedBitsIntoSamples(originalSamples []int16, bits []bool, config *EmbedConfig) ([]int16, error) {
	stegoSamples := make([]int16, len(originalSamples))
	copy(stegoSamples, originalSamples)

	startSample := 0
	if config.UseRandomSeed {
		hash := utils.GetSHA256Hash(config.StegoKey)
		seed := int64(binary.BigEndian.Uint64(hash[:8]))
		r := rand.New(rand.NewSource(seed))

		maxStartPos := len(stegoSamples) - (len(bits) / config.NLsb) - 1
		if maxStartPos <= 0 {
			maxStartPos = 1
		}
		startSample = r.Intn(maxStartPos)
	}

	bitIndex := 0
	for i := startSample; i < len(stegoSamples) && bitIndex < len(bits); i++ {
		sample := stegoSamples[i]
		mask := ^((1 << config.NLsb) - 1)
		clearedSample := int16(int(sample) & mask)

		var bitsToEmbed int16
		for j := 0; j < config.NLsb && bitIndex < len(bits); j++ {
			if bits[bitIndex] {
				bitsToEmbed |= (1 << j)
			}
			bitIndex++
		}
		stegoSamples[i] = clearedSample | bitsToEmbed
	}

	return stegoSamples, nil
}


// decodeMP3ToSamples membaca file MP3 dan mengembalikannya sebagai slice dari int16.
func decodeMP3ToSamples(filePath string) ([]int16, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, 0, err
	}
	
	sampleRate := decoder.SampleRate()

	var samples []int16
	buf := make([]byte, 4096)
	for {
		n, err := decoder.Read(buf)
		if n > 0 {
			for i := 0; i < n; i += 2 {
				if i+1 < n {
					// Asumsi 16-bit little-endian PCM
					sample := int16(binary.LittleEndian.Uint16(buf[i : i+2]))
					samples = append(samples, sample)
				}
			}
		}
		if err != nil {
			break // Keluar dari loop jika ada error (termasuk EOF)
		}
	}
	return samples, sampleRate, nil
}

// encodeSamplesToMP3 mengenkode sampel PCM ke file MP3 menggunakan LAME.
func encodeSamplesToMP3(samples []int16, outputPath string, sampleRate int, channels int) error {
	// Inisialisasi encoder LAME dengan parameter dari audio asli
	// Bitrate 320 adalah kualitas tinggi, bisa disesuaikan jika perlu
	encoder := lame.NewCodecAwareEncoder(sampleRate, channels, 320)

	// Fungsi EncodeWithSteganography di paket lame Anda sebenarnya tidak menggunakan
	// argumen secretBits. Ia hanya membuat file WAV sementara dari sampel
	// dan kemudian mengenkodenya ke MP3 menggunakan LAME. Jadi kita bisa
	// memberikan slice kosong untuk argumen tersebut.
	return encoder.EncodeWithSteganography(samples, []bool{}, outputPath)
}

// bytesToBits mengonversi slice byte menjadi slice boolean.
func bytesToBits(data []byte) []bool {
	bits := make([]bool, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b>>j)&1 == 1
		}
	}
	return bits
}