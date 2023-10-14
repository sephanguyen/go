package multitenant

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/gcp"

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

func TestNewUserFromGCPUserRecord(t *testing.T) {
	t.Parallel()

	gcpAuthUserRecord := aValidUserRecord()

	user := NewUserFromGCPUserRecord(gcpAuthUserRecord)

	assert.Equal(t, gcpAuthUserRecord.UserInfo.UID, user.UserID().String())
	assert.Equal(t, gcpAuthUserRecord.UserInfo.Email, user.Email().String())
	assert.Equal(t, gcpAuthUserRecord.UserInfo.PhoneNumber, user.PhoneNumber().String())
	assert.Equal(t, gcpAuthUserRecord.UserInfo.PhotoURL, user.Avatar().String())
	assert.Equal(t, gcpAuthUserRecord.UserInfo.DisplayName, user.FullName().String())
	assert.Equal(t, "", user.Password().String())
}

func TestToGCPUsersToCreate(t *testing.T) {
	t.Parallel()

	validUser := anUserWithValidInfo()

	userWithoutEmail := anUserWithEmptyFields(UserFieldEmail)

	testCases := []struct {
		name        string
		inputUser   User
		expectedErr error
	}{
		{
			name:        "parse valid user",
			inputUser:   validUser,
			expectedErr: nil,
		},
		{
			name:        "parse user without email",
			inputUser:   userWithoutEmail,
			expectedErr: ErrUserEmailEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ToGCPUsersToCreate(testCase.inputUser)

			assert.Equal(t, testCase.expectedErr, err)

			//values in auth.UserToCreate are unexported without provide interface to test
		})
	}
}

func TestToGCPUsersToImport(t *testing.T) {
	t.Parallel()

	validUser := anUserWithValidInfo()
	userWithoutEmail := anUserWithEmptyFields(UserFieldEmail)

	testCases := []struct {
		name            string
		inputUser       Users
		inputHashConfig ScryptHash
		expectedErr     error
	}{
		{
			name:            "parse users without hash config",
			inputUser:       Users{validUser},
			inputHashConfig: nil,
			expectedErr:     nil,
		},
		{
			name:            "parse users with hash config",
			inputUser:       Users{validUser},
			inputHashConfig: aValidScryptHash(),
			expectedErr:     nil,
		},
		{
			name:            "parse users without email",
			inputUser:       Users{userWithoutEmail},
			inputHashConfig: aValidScryptHash(),
			expectedErr:     ErrUserEmailEmpty,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ToGCPUsersToImport(testCase.inputUser, testCase.inputHashConfig)

			assert.Equal(t, testCase.expectedErr, err)

			//values in auth.UserToCreate are unexported without provide interface to test
		})
	}
}
