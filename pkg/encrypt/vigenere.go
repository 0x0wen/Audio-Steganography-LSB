package vigenere

import(
	"fmt"
)

// generateKey creates a key that is the same length as the input data by repeating the original key.
// Example: text="HELLO", key="KEY" -> newKey="KEYKE"
func generateKey(dataLen int, key string) []byte {
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	newKey := make([]byte, dataLen)

	for i := 0; i < dataLen; i++ {
		newKey[i] = keyBytes[i%keyLen]
	}
	return newKey
}

// Extended Vigenère encryption.
// Formula : C = (P + K) mod 256
func Encrypt(plaintext []byte, key string) []byte {
	keyBytes := generateKey(len(plaintext), key)
	ciphertext := make([]byte, len(plaintext))

	for i := 0; i < len(plaintext); i++ {
		ciphertext[i] = byte((int(plaintext[i]) + int(keyBytes[i])) % 256)
	}
	return ciphertext
}

// Extended Vigenère decryption.
// Formula : P = (C - K + 256) mod 256
func Decrypt(ciphertext []byte, key string) []byte {
	keyBytes := generateKey(len(ciphertext), key)
	plaintext := make([]byte, len(ciphertext))

	for i := 0; i < len(ciphertext); i++ {
		plaintext[i] = byte((int(ciphertext[i]) - int(keyBytes[i]) + 256) % 256)
	}
	return plaintext
}


func main(){
	messageData := []byte("Hello World!")
	key := "KEY123"

	// Enkripsi
	encrypted := Encrypt(messageData, key)
	fmt.Println("Encrypted (raw bytes):", encrypted)

	// Dekripsi
	decrypted := Decrypt(encrypted, key)
	fmt.Println("Decrypted (string):", string(decrypted))

}