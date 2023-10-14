package multitenant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsScryptHashValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		inputScryptHash ScryptHash
		expectedErr     error
	}{
		{
			name:            "scrypt hash is valid",
			inputScryptHash: aValidScryptHash(),
			expectedErr:     nil,
		},
		{
			name:            "scrypt hash is invalid because of rounds",
			inputScryptHash: anInvalidScryptHash("invalid rounds"),
			expectedErr:     ErrInvalidScryptRounds,
		},
		{
			name:            "scrypt hash is invalid because of key",
			inputScryptHash: anInvalidScryptHash("invalid key"),
			expectedErr:     ErrInvalidScryptKey,
		},
		{
			name:            "scrypt hash is invalid because of key",
			inputScryptHash: anInvalidScryptHash("invalid memory cost"),
			expectedErr:     ErrInvalidMemoryCost,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualErr := IsScryptHashValid(testCase.inputScryptHash)
			assert.Equal(t, testCase.expectedErr, actualErr)
		})
	}
}

func TestHashedPassword(t *testing.T) {
	t.Parallel()

	inputPassword := []byte("example-password")
	inputSalt := []byte("example-salt")

	testCases := []struct {
		name                   string
		inputScryptHash        ScryptHash
		inputPassword          []byte
		inputSalt              []byte
		expectedHashedPassword []byte
		expectedErr            error
	}{
		{
			name:                   "hash password successfully",
			inputScryptHash:        aValidScryptHash(),
			inputPassword:          inputPassword,
			inputSalt:              inputSalt,
			expectedHashedPassword: []byte{0x9a, 0xac, 0x62},
			expectedErr:            nil,
		},
		{
			name:          "scrypt hash is nil",
			inputPassword: inputPassword,
			inputSalt:     inputSalt,
			expectedErr:   ErrNilScryptHash,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualHashedPassword, actualErr := HashedPassword(testCase.inputScryptHash, testCase.inputPassword, testCase.inputSalt)
			assert.Equal(t, testCase.expectedErr, actualErr)
			assert.Equal(t, testCase.expectedHashedPassword, actualHashedPassword)
		})
	}
}
