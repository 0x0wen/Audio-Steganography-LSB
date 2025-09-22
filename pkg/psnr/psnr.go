package psnr

import (
	"fmt"
	"math"
)

func CalculatePSNR(original, stego []int16) (float64, error) {
	if len(original) != len(stego) {
		return 0, fmt.Errorf("audio signals must have the same length")
	}

	if len(original) == 0 {
		return 0, fmt.Errorf("audio signals cannot be empty")
	}

	mse := calculateMSE(original, stego)
	
	maxValue := float64(32767)
	
	if mse == 0 {
		return 100.0, nil
	}
	
	psnr := 10 * math.Log10((maxValue*maxValue)/mse)
	return psnr, nil
}

func calculateMSE(original, stego []int16) float64 {
	var sum float64
	n := len(original)
	
	for i := 0; i < n; i++ {
		diff := float64(original[i] - stego[i])
		sum += diff * diff
	}
	
	return sum / float64(n)
}

func IsQualityAcceptable(psnr float64) bool {
	return psnr >= 30.0
}

func GetQualityDescription(psnr float64) string {
	switch {
	case psnr >= 50:
		return "Excellent quality"
	case psnr >= 40:
		return "Good quality"
	case psnr >= 30:
		return "Acceptable quality"
	case psnr >= 20:
		return "Poor quality"
	default:
		return "Very poor quality"
	}
}
