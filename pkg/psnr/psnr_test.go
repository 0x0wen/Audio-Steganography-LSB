package psnr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculatePSNR(t *testing.T) {
	tests := []struct {
		name     string
		original []int16
		stego    []int16
		wantErr  bool
		minPSNR  float64
	}{
		{
			name:     "identical signals",
			original: []int16{1000, 2000, 3000, 4000},
			stego:    []int16{1000, 2000, 3000, 4000},
			wantErr:  false,
			minPSNR:  100.0, 
		},
		{
			name:     "small differences",
			original: []int16{1000, 2000, 3000, 4000},
			stego:    []int16{1001, 2001, 3001, 4001},
			wantErr:  false,
			minPSNR:  30.0,
		},
		{
			name:     "large differences",
			original: []int16{1000, 2000, 3000, 4000},
			stego:    []int16{2000, 4000, 6000, 8000},
			wantErr:  false,
			minPSNR:  0.0, 
		},
		{
			name:     "different lengths",
			original: []int16{1000, 2000, 3000},
			stego:    []int16{1000, 2000, 3000, 4000},
			wantErr:  true,
		},
		{
			name:     "empty signals",
			original: []int16{},
			stego:    []int16{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psnr, err := CalculatePSNR(tt.original, tt.stego)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, psnr, tt.minPSNR)
			}
		})
	}
}

func TestIsQualityAcceptable(t *testing.T) {
	tests := []struct {
		name     string
		psnr     float64
		expected bool
	}{
		{
			name:     "excellent quality",
			psnr:     50.0,
			expected: true,
		},
		{
			name:     "good quality",
			psnr:     40.0,
			expected: true,
		},
		{
			name:     "acceptable quality",
			psnr:     30.0,
			expected: true,
		},
		{
			name:     "poor quality",
			psnr:     25.0,
			expected: false,
		},
		{
			name:     "very poor quality",
			psnr:     10.0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsQualityAcceptable(tt.psnr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetQualityDescription(t *testing.T) {
	tests := []struct {
		name        string
		psnr        float64
		expected    string
	}{
		{
			name:     "excellent quality",
			psnr:     60.0,
			expected: "Excellent quality",
		},
		{
			name:     "good quality",
			psnr:     45.0,
			expected: "Good quality",
		},
		{
			name:     "acceptable quality",
			psnr:     35.0,
			expected: "Acceptable quality",
		},
		{
			name:     "poor quality",
			psnr:     25.0,
			expected: "Poor quality",
		},
		{
			name:     "very poor quality",
			psnr:     15.0,
			expected: "Very poor quality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetQualityDescription(tt.psnr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePSNRWithKnownMSE(t *testing.T) {
	tests := []struct {
		name     string
		original []int16
		stego    []int16
		minPSNR  float64
		maxPSNR  float64
	}{
		{
			name:     "identical signals (MSE = 0)",
			original: []int16{1000, 2000, 3000},
			stego:    []int16{1000, 2000, 3000},
			minPSNR:  99.0,
			maxPSNR:  101.0,
		},
		{
			name:     "constant difference of 1 (MSE = 1)",
			original: []int16{1000, 2000, 3000},
			stego:    []int16{1001, 2001, 3001},
			minPSNR:  90.0,
			maxPSNR:  100.0,
		},
		{
			name:     "larger differences",
			original: []int16{1000, 2000, 3000},
			stego:    []int16{1002, 2000, 2998},
			minPSNR:  80.0,
			maxPSNR:  95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			psnr, err := CalculatePSNR(tt.original, tt.stego)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, psnr, tt.minPSNR)
			assert.LessOrEqual(t, psnr, tt.maxPSNR)
		})
	}
}
