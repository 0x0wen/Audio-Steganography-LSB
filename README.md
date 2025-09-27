# Audio Steganography LSB

A Go implementation of audio steganography using multiple LSB (Least Significant Bit) techniques for embedding secret messages in MP3 audio files with MP3-aware robustness features.

## Overview

This project implements advanced audio steganography on MP3 files using multiple LSB embedding techniques specifically designed to survive MP3 compression. The implementation includes several MP3-robust steganography methods, from traditional LSB manipulation to sophisticated codec-aware and quantization noise techniques.

## Features

- **Multiple LSB Steganography**: True LSB embedding on audio samples (1-4 bits per sample)
- **MP3-Robust Techniques**: Multiple embedding methods designed to survive MP3 compression
- **MP3 Bitstream Embedding**: Direct manipulation of MP3 bitstream data
- **Codec-Aware Steganography**: Advanced techniques that account for MP3 quantization
- **Quantization Noise Manipulation**: Dithering-based embedding that survives compression
- **Random Position Generation**: SHA256-based position selection using stego key as seed
- **File Type Support**: Accept any file type as secret message
- **Metadata Preservation**: Store original filename, extension, and embedding parameters
- **PSNR Calculation**: Audio quality measurement for steganography assessment
- **CLI Interface**: Command-line tool with comprehensive parameter support

## Project Structure

```
audio-steganography-lsb/
├── cmd/                    # Main application entry point
│   └── main.go
├── internal/
│   └── cli/               # CLI interface using Cobra
│       └── cli.go
├── pkg/
│   ├── embed/             # Multiple LSB embedding techniques
│   │   ├── embed.go
│   │   └── embed_test.go
│   ├── extract/           # Multi-method extraction with fallbacks
│   │   ├── extract.go
│   │   └── extract_test.go
│   ├── lame/              # MP3 encoding wrapper
│   │   └── lame.go
│   ├── metadata/          # Metadata handling
│   │   ├── metadata.go
│   │   └── metadata_test.go
│   ├── psnr/              # Audio quality measurement
│   │   ├── psnr.go
│   │   └── psnr_test.go
│   └── utils/             # Common utilities and validation
│       ├── utils.go
│       └── utils_test.go
├── test/                  # Test files and demo scripts
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build and test commands
└── README.md             # This file
```

## Installation

### Prerequisites

For optimal MP3 encoding quality, install LAME MP3 encoder:
```bash
# macOS
brew install lame

# Ubuntu/Debian
sudo apt-get install lame

# CentOS/RHEL
sudo yum install lame
```

### Build Instructions

1. Clone the repository:
```bash
git clone <repository-url>
cd audio-steganography-lsb
```

2. Install Go dependencies:
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
- `--message, -m`: Secret message file (any file type)
- `--key, -k`: Steganography key (max 25 characters, used for encryption and position generation)
- `--lsb, -l`: Number of LSB bits to use (1-4, affects capacity and robustness)
- `--random, -r`: Use random seed for embedding positions (improves security)
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
- `--key, -k`: Steganography key (must match embedding key)
- `--output, -o`: Output extracted file

## Technical Implementation

### Core Steganography Methods

The implementation provides multiple embedding techniques optimized for different scenarios:

#### 1. MP3 Bitstream Embedding (Primary Method)
- **Function**: `embedMP3Bitstream()`
- **Approach**: Direct manipulation of MP3 bitstream data
- **Advantages**: Avoids lossy re-encoding, preserves data integrity
- **Process**:
  - Identifies embeddable positions in MP3 frames
  - Avoids sync patterns and headers
  - Uses parameter header for extraction configuration

#### 2. Traditional LSB Steganography
- **Function**: `embedLSB()`
- **Approach**: Classic LSB modification on decoded audio samples
- **Process**: MP3 decode → LSB modification → MP3 re-encode
- **Use Case**: When maximum compatibility is needed

#### 3. MP3-Robust LSB
- **Function**: `embedLSBRobust()`
- **Features**:
  - Error correction coding (3x redundancy)
  - Higher-order bit usage (LSB+2, LSB+3)
  - Larger magnitude changes for compression survival
- **Robustness**: Designed to survive MP3 re-compression

#### 4. Magnitude-Based Encoding
- **Function**: `embedMP3Compatible()`
- **Technique**: Odd/even magnitude encoding
- **Advantage**: More resistant to quantization than direct LSB
- **Method**: Bit 1 = odd magnitude, Bit 0 = even magnitude

#### 5. Quantization Noise Manipulation
- **Function**: `embedQuantizationNoise()`
- **Approach**: Controlled dithering that survives MP3 quantization
- **Innovation**: Uses triangular dithering patterns preserved by MP3
- **Calculation**: Adaptive quantization step estimation

#### 6. Codec-Aware Steganography
- **Function**: `embedCodecAwareLSB()`
- **Integration**: Works with LAME encoder parameters
- **Optimization**: Embedding aligned with MP3 psychoacoustic model

### Position Generation

#### Random Positions
- **Method**: SHA256 hash of stego key generates deterministic pseudo-random positions
- **Security**: Prevents sequential pattern detection
- **Implementation**: Hash-based position selection with collision avoidance

#### Sequential Positions
- **Fallback**: Used when random generation insufficient
- **Predictability**: Lower security but guaranteed capacity

### Metadata Management

**Embedded Information:**
- Original filename and extension
- File size and data length
- Embedding configuration (nLsb, random seed, encryption flags)
- Serialized as binary data with length prefixes

**Storage Format:**
```
[4 bytes: metadata length] + [metadata] + [4 bytes: message length] + [message data]
```

### Extraction Strategy

**Multi-Method Fallback System:**
1. **MP3 Bitstream**: Attempts parameter header extraction first
2. **Codec-Aware**: Falls back to codec-aware extraction
3. **Traditional LSB**: Final fallback to basic LSB extraction
4. **Parameter Guessing**: Tries all nLsb/random combinations if needed

### Audio Quality Assessment

**PSNR Calculation:**
```
PSNR = 10 × log₁₀(MAX² / MSE)
```
Where:
- MAX = 32767 (16-bit audio maximum)
- MSE = Mean Squared Error between original and stego audio

**Quality Thresholds:**
- PSNR ≥ 30 dB: Acceptable quality (minimal distortion)
- PSNR ≥ 40 dB: Good quality (barely perceptible)
- PSNR ≥ 50 dB: Excellent quality (imperceptible)

## Testing

### Run Tests
```bash
# All tests
make test

# With coverage report
make test-coverage

# Specific package
go test ./pkg/embed/
```