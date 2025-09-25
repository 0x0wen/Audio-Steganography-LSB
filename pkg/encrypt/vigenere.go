package vigenere

import (
	"fmt"
)

// generateKey creates a key that is the same length as the input text by repeating the original key.
// Example: text="HELLO", key="KEY" -> newKey="KEYKE"
func generateKey(text string, key string) []byte {
	keyBytes := []byte(key)
	textLen := len(text)
	keyLen := len(keyBytes)
	newKey := make([]byte, textLen)

	for i := 0; i < textLen; i++ {
		newKey[i] = keyBytes[i%keyLen]
	}
	return newKey
}

// Extended Vigenère encryption.
// Formula : C = (P + K) mod 256
func Encrypt(plaintext string, key string) string {
	plaintextBytes := []byte(plaintext)
	keyBytes := generateKey(plaintext, key)
	ciphertext := make([]byte, len(plaintextBytes))

	for i := 0; i < len(plaintextBytes); i++ {
		// Formula : (plaintext_char + key_char) % 256
		// Cast to int to perform calculation and avoid byte overflow.
		ciphertext[i] = byte((int(plaintextBytes[i]) + int(keyBytes[i])) % 256)
	}
	return string(ciphertext)
}

// Extended Vigenère decryption.
// Formula : P = (C - K + 256) mod 256
func Decrypt(ciphertext string, key string) string {
	ciphertextBytes := []byte(ciphertext)
	keyBytes := generateKey(ciphertext, key)
	plaintext := make([]byte, len(ciphertextBytes))

	for i := 0; i < len(ciphertextBytes); i++ {
		// Formula : (ciphertext_char - key_char + 256) % 256
		plaintext[i] = byte((int(ciphertextBytes[i]) - int(keyBytes[i]) + 256) % 256)
	}
	return string(plaintext)
}

func main() {
	originalText := "This is a secret message! It includes numbers 123 and symbols @#$%."
	secretKey := "KEY123"

	fmt.Println("--- Extended Vigenère Cipher ---")
	fmt.Printf("Original Text:\n%s\n\n", originalText)
	fmt.Printf("Secret Key:\n%s\n\n", secretKey)

	// Encrypt
	encryptedText := Encrypt(originalText, secretKey)
	fmt.Printf("Encrypted Text (Ciphertext):\n%s\n\n", encryptedText)

	// Decrypt
	decryptedText := Decrypt(encryptedText, secretKey)
	fmt.Printf("Decrypted Text (Plaintext):\n%s\n\n", decryptedText)

	// Verify
	if originalText == decryptedText {
		fmt.Println("Success: The decrypted text matches the original text.")
	} else {
		fmt.Println("Error: Decryption failed.")
	}
}

