package services

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func TestEncrypt(t *testing.T) {

	var (
		secret = generateSecretByte()
		c      = &CryptV2{}
	)

	testCases := []struct {
		name      string
		content   string
		secretKey []byte
		hasError  bool
	}{
		{
			name:      "happy path for encrypting string",
			secretKey: decodeSecret(secret),
			content:   "test-string",
			hasError:  false,
		},
		{
			name:      "using empty byte as secret key",
			secretKey: []byte{},
			content:   "test-string",
			hasError:  true,
		},
	}

	for _, tc := range testCases {

		_, err := c.Encrypt(tc.content, tc.secretKey)

		if tc.hasError && err == nil {
			t.Errorf("Expecting an error got nil")
		}

		if !tc.hasError && err != nil {
			t.Errorf("Unwanted error occured %v", err)
		}
	}
}

func TestDecrypt(t *testing.T) {

	secret := generateSecretByte()
	c := &CryptV2{}
	encryptedContent, _ := c.Encrypt("test-string", decodeSecret(secret))

	testCases := []struct {
		name      string
		content   string
		secretKey []byte
		hasError  bool
		expected  string
	}{
		{
			name:      "happy path for decrypting string",
			secretKey: decodeSecret(secret),
			content:   encryptedContent,
			hasError:  false,
			expected:  "test-string",
		},
		{
			name:      "using invalid encrypted content",
			secretKey: []byte{},
			content:   "invalid-content",
			hasError:  true,
		},
		{
			name:      "using invalid secret key",
			secretKey: decodeSecret("invalid-key"),
			content:   encryptedContent,
			hasError:  true,
		},
		{
			name:      "using empty byte as secret key",
			secretKey: []byte{},
			content:   encryptedContent,
			hasError:  true,
		},
	}

	for _, tc := range testCases {
		got, err := c.Decrypt(tc.content, tc.secretKey)

		if tc.hasError && err == nil {
			t.Errorf("Expecting an error got nil")
		}

		if !tc.hasError && err != nil {
			t.Errorf("Unwanted error occured %v", err)
		}

		if tc.expected != got {
			t.Errorf("Expecting decrypted value %s, got %s", tc.expected, got)
		}
	}
}

func decodeSecret(key string) []byte {
	k, _ := hex.DecodeString(key)
	return k
}

func generateSecretByte() string {
	key := make([]byte, 16)
	rand.Read(key)

	return hex.EncodeToString(key)
}
