package crypt

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testSignerKey     = "Impassphrasegood" //not base64
	testInitialVector = "1234567890123456" //not base64

	testData            = "hello world"
	base64TestData      = "aGVsbG8gd29ybGQ="
	keyString           = "d77fb06a46e432cb657d9d996c89d1ef"
	base64EncryptedData = "9f4yohBU0rUoq6ajOcC3hA=="
)

func TestAESEncrypt(t *testing.T) {
	testCases := []struct {
		name                 string
		signerKey            string
		initialVector        string
		dataToEncrypt        string
		expectedBase64Output string
		expectedErr          error
	}{
		{
			name:                 "happy case",
			signerKey:            testSignerKey,
			initialVector:        testInitialVector,
			dataToEncrypt:        testData,
			expectedBase64Output: "9f4yohBU0rUoq6ajOcC3hA==",
			expectedErr:          nil,
		},
		{
			name:                 "signer key size is empty",
			signerKey:            "",
			initialVector:        testInitialVector,
			dataToEncrypt:        testData,
			expectedBase64Output: "",
			expectedErr:          aes.KeySizeError(0),
		},
		{
			name:                 "signer key is invalid",
			signerKey:            "123456789",
			initialVector:        testInitialVector,
			dataToEncrypt:        testData,
			expectedBase64Output: "",
			expectedErr:          aes.KeySizeError(9),
		},
		{
			name:                 "data to encrypt is empty",
			signerKey:            testSignerKey,
			initialVector:        testInitialVector,
			dataToEncrypt:        "",
			expectedBase64Output: "",
			expectedErr:          ErrDataToEncryptIsEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			encryptedData, err := AESEncrypt(testCase.dataToEncrypt, []byte(testCase.signerKey), []byte(testCase.initialVector))
			assert.ErrorIs(t, err, testCase.expectedErr)
			assert.Equal(t, testCase.expectedBase64Output, EncodeBase64(encryptedData))
		})
	}
}

func TestAESDecrypt(t *testing.T) {
	encryptedData, err := DecodeBase64(base64EncryptedData)
	assert.NoError(t, err)

	testCases := []struct {
		name                   string
		dataToDecrypt          []byte
		signerKeyToDecrypt     string
		initialVectorToDecrypt string
		expectedDecryptedData  string
		expectedErr            error
	}{
		{
			name:                   "happy case",
			dataToDecrypt:          encryptedData,
			signerKeyToDecrypt:     testSignerKey,
			initialVectorToDecrypt: testInitialVector,
			expectedDecryptedData:  testData,
			expectedErr:            nil,
		},
		{
			name:                   "signer key size is empty",
			dataToDecrypt:          encryptedData,
			signerKeyToDecrypt:     "",
			initialVectorToDecrypt: testInitialVector,
			expectedDecryptedData:  "",
			expectedErr:            aes.KeySizeError(0),
		},
		{
			name:                   "signer key is invalid",
			dataToDecrypt:          encryptedData,
			signerKeyToDecrypt:     "123456789",
			initialVectorToDecrypt: testInitialVector,
			expectedDecryptedData:  "",
			expectedErr:            aes.KeySizeError(9),
		},
		{
			name:                   "data to decrypt is empty",
			dataToDecrypt:          []byte(""),
			signerKeyToDecrypt:     testSignerKey,
			initialVectorToDecrypt: testInitialVector,
			expectedDecryptedData:  "",
			expectedErr:            ErrDataToDecryptIsEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			decryptedData, err := AESDecrypt(testCase.dataToDecrypt, []byte(testCase.signerKeyToDecrypt), []byte(testCase.initialVectorToDecrypt))
			assert.ErrorIs(t, err, testCase.expectedErr)
			assert.Equal(t, testCase.expectedDecryptedData, string(decryptedData))
		})
	}
}

func TestPKCS7Padding(t *testing.T) {
	testCases := []struct {
		name string

		inputData                []byte
		inputBlockLength         int
		expectedPaddedDataOutput []byte
		expectedErr              error
	}{
		{
			name:                     "invalid block length",
			inputData:                nil,
			inputBlockLength:         0,
			expectedPaddedDataOutput: nil,
			expectedErr:              ErrInvalidBlockLength(0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			paddedData, err := PKCS7Padding(testCase.inputData, testCase.inputBlockLength)
			assert.ErrorIs(t, err, testCase.expectedErr)
			assert.Equal(t, testCase.expectedPaddedDataOutput, paddedData)
		})
	}
}

func TestEncrypt(t *testing.T) {
	testCases := []struct {
		name            string
		stringToEncrypt string
		keyString       string
		hasError        bool
	}{
		{
			name:            "happy case",
			stringToEncrypt: testData,
			keyString:       keyString,
			hasError:        false,
		},
		{
			name:            "key invalid",
			stringToEncrypt: testData,
			keyString:       "12345667",
			hasError:        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := Encrypt(testCase.stringToEncrypt, testCase.keyString)
			if testCase.hasError && err == nil {
				t.Errorf("Expecting an error got nil")
			}
			if !testCase.hasError && err != nil {
				t.Errorf("Unwanted error occured %v", err)
			}

		})
	}
}

func TestDecrypt(t *testing.T) {
	testCases := []struct {
		name            string
		stringToEncrypt string
		keyString       string
		hasError        bool
	}{
		{
			name:            "happy case",
			stringToEncrypt: testData,
			keyString:       keyString,
			hasError:        false,
		},
		{
			name:            "key invalid",
			stringToEncrypt: testData,
			keyString:       "12345667",
			hasError:        true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			encryptedContent, err := Encrypt(testCase.stringToEncrypt, testCase.keyString)
			decryptedContent, err := Decrypt(encryptedContent, testCase.keyString)

			if testCase.hasError && err == nil {
				t.Errorf("Expecting an error got nil")
			}
			if !testCase.hasError && err != nil {
				t.Errorf("Unwanted error occured %v", err)
				assert.Equal(t, decryptedContent, testCase.stringToEncrypt)

			}

		})
	}
}
