package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

func ValidateStegoKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("stego key cannot be empty")
	}
	if len(key) > 25 {
		return fmt.Errorf("stego key cannot exceed 25 characters")
	}
	return nil
}

func ValidateNLsb(nLsb int) error {
	if nLsb < 1 || nLsb > 4 {
		return fmt.Errorf("n_lsb must be between 1 and 4, got %d", nLsb)
	}
	return nil
}

func GeneratePositions(key string, useRandomSeed bool, totalSamples, nLsb int) ([]int, error) {
	if useRandomSeed {
		return generateRandomPositions(key, totalSamples, nLsb)
	}
	return generateSequentialPositions(totalSamples, nLsb), nil
}

func generateRandomPositions(key string, totalSamples, nLsb int) ([]int, error) {
	hash := sha256.Sum256([]byte(key))
	
	neededPositions := (totalSamples * nLsb) / 8
	positions := make([]int, 0, neededPositions)
	
	hashIndex := 0
	attempts := 0
	maxAttempts := neededPositions * 2
	
	for len(positions) < neededPositions && attempts < maxAttempts {
		pos := int(hash[hashIndex%len(hash)]) + int(hash[(hashIndex+1)%len(hash)])*256
		pos = pos % totalSamples
		
		if !contains(positions, pos) {
			positions = append(positions, pos)
		}
		
		hashIndex++
		attempts++
	}
	
	for len(positions) < neededPositions {
		pos := len(positions) % totalSamples
		if !contains(positions, pos) {
			positions = append(positions, pos)
		} else {
			for i := 0; i < totalSamples; i++ {
				if !contains(positions, i) {
					positions = append(positions, i)
					break
				}
			}
		}
	}
	
	return positions, nil
}

func generateSequentialPositions(totalSamples, nLsb int) []int {
	neededPositions := (totalSamples * nLsb) / 8
	positions := make([]int, neededPositions)
	
	for i := 0; i < neededPositions; i++ {
		positions[i] = i % totalSamples
	}
	
	return positions
}

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func WriteFile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func CalculateCapacity(totalSamples, nLsb int) int {
	return (totalSamples * nLsb) / 8
}
