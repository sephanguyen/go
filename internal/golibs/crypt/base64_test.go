package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAESDecryptBase64(t *testing.T) {
	encryptedData, err := AESEncrypt(testData, []byte(testSignerKey), []byte(testInitialVector))
	assert.NoError(t, err)

	base64EncryptedData := EncodeBase64(encryptedData)

	decryptedData, err := AESDecryptBase64(base64EncryptedData, []byte(testSignerKey), []byte(testInitialVector))
	assert.NoError(t, err)

	assert.Equal(t, testData, string(decryptedData))
}

func TestEncodeBase64(t *testing.T) {
	assert.Equal(t, base64TestData, EncodeBase64([]byte(testData)))
}

func TestDecodeBase64(t *testing.T) {
	decodedData, err := DecodeBase64(base64TestData)
	assert.NoError(t, err)
	assert.Equal(t, testData, string(decodedData))
}
