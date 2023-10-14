package crypt

import "encoding/base64"

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeBase64(content string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(content)
}

func AESDecryptBase64(encryptedBase64 string, key []byte, initialVector []byte) ([]byte, error) {
	data, err := DecodeBase64(encryptedBase64)
	if err != nil {
		return nil, err
	}

	return AESDecrypt(data, key, initialVector)
}
