package auth

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/gcp"

	"github.com/stretchr/testify/assert"
)

func aValidScryptHash() *gcp.HashConfig {
	return &gcp.HashConfig{
		HashAlgorithm:  "SCRYPT",
		HashRounds:     8,
		HashMemoryCost: 8,
		HashSaltSeparator: gcp.Base64EncodedStr{
			Value:        "salt",
			DecodedBytes: []byte("salt"),
		},
		HashSignerKey: gcp.Base64EncodedStr{
			Value:        "key",
			DecodedBytes: []byte("key"),
		},
	}
}

func anInvalidScryptHash(invalid string) *gcp.HashConfig {
	scryptHash := aValidScryptHash()

	switch invalid {
	case "invalid rounds":
		scryptHash.HashRounds = 0
	case "invalid key":
		scryptHash.HashSignerKey = gcp.Base64EncodedStr{
			DecodedBytes: []byte{},
		}
	case "invalid memory cost":
		scryptHash.HashMemoryCost = 0
	}

	return scryptHash
}

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
