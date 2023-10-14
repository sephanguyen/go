package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func AESEncrypt(dataToEncrypt string, key []byte, initialVector []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if dataToEncrypt == "" {
		return nil, ErrDataToEncryptIsEmpty
	}

	ecb := cipher.NewCBCEncrypter(block, initialVector)

	content, err := PKCS7Padding([]byte(dataToEncrypt), block.BlockSize())
	if err != nil {
		return nil, errors.Wrap(err, "PKCS7Padding()")
	}

	encryptedContent := make([]byte, len(content))
	ecb.CryptBlocks(encryptedContent, content)

	return encryptedContent, nil
}

func Encrypt(stringToEncrypt string, keyString string) (string, error) {
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", cipherText), nil
}

func Decrypt(encryptedString string, keyString string) (string, error) {
	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	nonce, cipherText := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func AESDecrypt(encryptedData []byte, key []byte, initialVector []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(encryptedData) == 0 {
		return nil, ErrDataToDecryptIsEmpty
	}

	ecb := cipher.NewCBCDecrypter(block, initialVector)
	decrypted := make([]byte, len(encryptedData))
	ecb.CryptBlocks(decrypted, encryptedData)

	content, err := PKCS7Unpadding(decrypted, block.BlockSize())
	if err != nil {
		return nil, err
	}

	return content, nil
}

func PKCS7Padding(data []byte, blockLen int) ([]byte, error) {
	if blockLen <= 0 {
		return nil, ErrInvalidBlockLength(blockLen)
	}
	paddingLen := 1
	for ((len(data) + paddingLen) % blockLen) != 0 {
		paddingLen++
	}

	pad := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(data, pad...), nil
}

func PKCS7Unpadding(data []byte, blockLen int) ([]byte, error) {
	if blockLen <= 0 {
		return nil, ErrInvalidBlockLength(blockLen)
	}
	if len(data)%blockLen != 0 || len(data) == 0 {
		return nil, ErrInvalidDataLength(len(data))
	}

	paddingLen := int(data[len(data)-1])
	if paddingLen > blockLen || paddingLen == 0 {
		return nil, ErrInvalidPadding
	}
	// check padding
	pad := data[len(data)-paddingLen:]
	for i := 0; i < paddingLen; i++ {
		if pad[i] != byte(paddingLen) {
			return nil, ErrInvalidPadding
		}
	}

	return data[:len(data)-paddingLen], nil
}
