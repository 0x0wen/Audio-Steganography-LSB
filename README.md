# Audio Steganography LSB

A Go implementation of audio steganography using ID3 tag storage for embedding secret messages in MP3 audio files.

## Overview

This project implements steganography on MP3 audio files using ID3 tag storage. It allows users to embed secret messages into MP3 metadata tags and extract them later, providing a simple and reliable method for hiding data in audio files.

## Features

- **Message Embedding**: Embed secret messages into MP3 files using ID3 tag storage
- **Message Extraction**: Extract embedded messages from stego-audio files
- **Metadata Storage**: Store embedding configuration and message data in ID3 tags
- **PSNR Calculation**: Measure audio quality using Peak Signal-to-Noise Ratio (available but not used in main workflow)
- **CLI Interface**: Command-line tool for easy usage
- **Input Validation**: Validate stego key length and LSB parameters
- **JSON Metadata**: Structured metadata storage in ID3 TXXX frames

## Project Structure

```
audio-steganography-lsb/
├── cmd/                    # Main application entry point
│   └── main.go
├── internal/
│   └── cli/               # CLI interface using Cobra
│       └── cli.go
├── pkg/
│   ├── embed/             # Message embedding functionality
│   │   ├── embed.go
│   │   └── embed_test.go
│   ├── extract/           # Message extraction functionality
│   │   ├── extract.go
│   │   └── extract_test.go
│   ├── encoder/           # Audio encoding utilities
│   │   ├── encoder.go
│   │   └── encoder_test.go
│   ├── metadata/          # ID3 metadata handling
│   │   ├── metadata.go
│   │   └── metadata_test.go
│   ├── psnr/              # Audio quality measurement
│   │   ├── psnr.go
│   │   └── psnr_test.go
│   └── utils/             # Common utilities
│       ├── utils.go
│       └── utils_test.go
├── test/                  # Test files and demo
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build and test commands
├── test_demo.sh          # Demo script
└── README.md             # This file
```

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd audio-steganography-lsb
```

2. Install dependencies:
```bash
make deps
```

3. Build the application:
```bash
make build
```

## Usage

### Embedding a Message

```bash
./bin/steganography embed \
  --cover cover.mp3 \
  --message secret.txt \
  --key mykey123 \
  --lsb 2 \
  --random \
  --output stego.mp3
```

**Parameters:**
- `--cover, -c`: Cover audio file (MP3)
- `--message, -m`: Secret message file
- `--key, -k`: Steganography key (max 25 characters)
- `--lsb, -l`: Number of LSB bits to use (1-4) - *Note: Accepted for compatibility but not used in ID3 tag storage*
- `--random, -r`: Use random seed for embedding positions - *Note: Accepted for compatibility but not used in ID3 tag storage*
- `--output, -o`: Output stego audio file

### Extracting a Message

```bash
./bin/steganography extract \
  --stego stego.mp3 \
  --key mykey123 \
  --output extracted.txt
```

**Parameters:**
- `--stego, -s`: Stego audio file (MP3)
- `--key, -k`: Steganography key (max 25 characters)
- `--output, -o`: Output message file

## Technical Implementation

### ID3 Tag Steganography

The implementation uses ID3 tag storage where:
- Secret messages are stored in ID3 TXXX (User Defined Text) frames
- No audio sample modification is performed
- Data is stored in MP3 metadata tags, not in the audio stream itself
- Two separate TXXX frames are used for different purposes

### Metadata Storage

Embedding configuration and message data are stored in ID3 tags:

#### STEGO_METADATA Frame
- **Description**: "STEGO_METADATA"
- **Content**: JSON containing:
  - Original filename and extension
  - File size
  - Configuration flags (encryption, random seed, n_lsb)
  - UTF-8 encoding

#### SECRET_MESSAGE Frame
- **Description**: "SECRET_MESSAGE"
- **Content**: Raw message data as string
- **Encoding**: UTF-8

### Data Flow

1. **Embedding Process**:
   - Read secret message file
   - Create metadata JSON with file information
   - Copy original MP3 file to output location
   - Store metadata in ID3 TXXX frame
   - Store message data in separate ID3 TXXX frame

2. **Extraction Process**:
   - Open stego MP3 file
   - Search for "SECRET_MESSAGE" TXXX frame
   - Extract and write message data to output file

### Audio Quality Measurement

PSNR (Peak Signal-to-Noise Ratio) calculation is available but not used in the main workflow:
```
PSNR = 10 × log₁₀(MAX² / MSE)
```
Where:
- MAX = 32767 (maximum value for 16-bit audio)
- MSE = Mean Squared Error between original and stego audio

Quality thresholds:
- PSNR ≥ 30 dB: Acceptable quality
- PSNR ≥ 40 dB: Good quality
- PSNR ≥ 50 dB: Excellent quality

**Note**: Since no audio samples are modified, PSNR calculation is not performed during the embedding process.

## Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

### Test Coverage

The test suite covers:
- Input validation (stego key length, n_lsb values)
- ID3 metadata creation and validation
- Message embedding and extraction
- PSNR calculations (utility functions)
- Error handling scenarios
- File I/O operations
- JSON metadata serialization/deserialization

## Dependencies

- `github.com/hajimehoshi/go-mp3`: MP3 decoding
- `github.com/bogem/id3v2`: ID3 tag handling
- `github.com/spf13/cobra`: CLI framework
- `github.com/stretchr/testify`: Testing framework

## Limitations

1. **Implementation Method**: Uses ID3 tag steganography, not LSB audio sample modification
2. **Detection**: ID3 tags can be easily detected and modified by standard MP3 tools
3. **Capacity**: Limited by ID3 tag size limits (typically ~256MB per tag)
4. **Security**: No encryption of stored data - messages are stored in plain text
5. **Compatibility**: Requires MP3 files with ID3 tag support
6. **Steganography**: Not true steganography as data is stored in metadata, not hidden in audio
7. **MP3 Encoding**: Uses MP3 decoding but requires external tools for re-encoding

## Future Enhancements

1. **True LSB Steganography**: Implement actual LSB audio sample modification
2. **MP3 Encoding**: Add complete MP3 encoding capability
3. **Encryption Support**: Encrypt message data before storing in ID3 tags
4. **Additional Audio Formats**: Support for WAV, FLAC, and other formats
5. **GUI Interface**: Graphical user interface for easier usage
6. **Batch Processing**: Process multiple files simultaneously
7. **Steganalysis Resistance**: Implement more sophisticated hiding techniques

## License

This project is part of a cryptography course assignment.
