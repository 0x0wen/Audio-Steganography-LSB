package vigenere

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext []byte
		key       string
	}{
		{
			name:      "Simple Case",
			plaintext: []byte("Hello World!"),
			key:       "KEY123",
		},
		{
			name:      "Longer Plaintext",
			plaintext: []byte("This is a longer test message to see how the key repetition works."),
			key:       "secret",
		},
		{
			name:      "Key Longer Than Plaintext",
			plaintext: []byte("short"),
			key:       "averylongkey",
		},
		{
			name:      "Plaintext with various symbols",
			plaintext: []byte("`1234567890-=~!@#$%^&*()_+[]\\{}|;':\",./<>?"),
			key:       "symbols!@#",
		},
		{
			name:      "Empty Plaintext",
			plaintext: []byte(""),
			key:       "anykey",
		},
		{
			name:      "Single Character Key",
			plaintext: []byte("Testing with a single character key."),
			key:       "a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			
			encrypted := Encrypt(tc.plaintext, tc.key)

			decrypted := Decrypt(encrypted, tc.key)

			if !bytes.Equal(tc.plaintext, decrypted) {
				t.Errorf("Decrypted text does not match original plaintext.\nOriginal:  %s\nDecrypted: %s", string(tc.plaintext), string(decrypted))
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {

	testCases := []struct {
		name     string
		dataLen  int
		key      string
		expected []byte
	}{
		{
			name:     "Key shorter than data",
			dataLen:  10,
			key:      "KEY",
			expected: []byte("KEYKEYKEYK"),
		},
		{
			name:     "Key longer than data",
			dataLen:  4,
			key:      "SECRETKEY",
			expected: []byte("SECR"),
		},
		{
			name:     "Key same length as data",
			dataLen:  6,
			key:      "SAMELE",
			expected: []byte("SAMELE"),
		},
		{
			name:     "Zero length data",
			dataLen:  0,
			key:      "ANYKEY",
			expected: []byte(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			generated := generateKey(tc.dataLen, tc.key)

			if !bytes.Equal(generated, tc.expected) {
				t.Errorf("Generated key is incorrect.\nExpected: %s\nGot:      %s", string(tc.expected), string(generated))
			}
		})
	}
}
