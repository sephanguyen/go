package entity

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/argon2"
)

type mockUserToTestEncryptedUserIDByPasswordFromUser struct {
	EmptyUser
	randomID field.String
}

func (user *mockUserToTestEncryptedUserIDByPasswordFromUser) UserID() field.String {
	return field.NewString(fmt.Sprintf("uid-%s", user.randomID))
}

func (user *mockUserToTestEncryptedUserIDByPasswordFromUser) Password() field.String {
	return field.NewString(fmt.Sprintf("pwd-%s", user.randomID))
}

func TestEncryptedUserIDByPasswordFromUser(t *testing.T) {
	testUser := &mockUserToTestEncryptedUserIDByPasswordFromUser{randomID: field.NewString(idutil.ULIDNow())}

	encryptedDataBase64, err := EncryptedUserIDByPasswordFromUser(testUser)
	assert.NoError(t, err)
	assert.NotEmpty(t, encryptedDataBase64)

	key := argon2.IDKey([]byte(testUser.Password().String()), []byte(testUser.UserID().String()), 2, 19*1024, 1, 32)
	block, err := aes.NewCipher(key)
	assert.NoError(t, err)

	encryptedData, err := base64.RawStdEncoding.DecodeString(encryptedDataBase64.String())
	assert.NoError(t, err)

	plainText := make([]byte, len(encryptedData))
	stream := cipher.NewCTR(block, make([]byte, 16))

	stream.XORKeyStream(plainText, encryptedData)

	assert.Equal(t, testUser.UserID().String(), string(plainText))
}
