package domain

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/crypt"

	"gotest.tools/assert"
)

func TestZoomConfigToZoomConfigEncrypted(t *testing.T) {
	t.Run("Encrypted succeeded", func(t *testing.T) {
		accountID := "AccountID"
		clientID := "ClientID"
		clientSecret := "ClientSecret"
		encryptKey := "48404D635166546A576E5A7234753778"
		zoomConfig := ZoomConfig{AccountID: accountID, ClientID: clientID, ClientSecret: clientSecret}
		zoomConfigEncrypted, _ := zoomConfig.ToEncrypted(encryptKey)
		accountIDDecrypted, _ := crypt.Decrypt(zoomConfigEncrypted.AccountID, encryptKey)
		clientIDDecrypted, _ := crypt.Decrypt(zoomConfigEncrypted.ClientID, encryptKey)
		clientSecretEncrypted, _ := crypt.Decrypt(zoomConfigEncrypted.ClientSecret, encryptKey)

		assert.Equal(t, accountIDDecrypted, accountID)
		assert.Equal(t, clientIDDecrypted, clientID)
		assert.Equal(t, clientSecretEncrypted, clientSecret)

	})
	t.Run("Should error when key encrypt wrong format", func(t *testing.T) {

		accountID := "AccountID"
		clientID := "ClientID"
		clientSecret := "ClientSecret"
		encryptKey := "1"
		zoomConfig := ZoomConfig{AccountID: accountID, ClientID: clientID, ClientSecret: clientSecret}
		_, err := zoomConfig.ToEncrypted(encryptKey)
		assert.Equal(t, "crypto/aes: invalid key size 0", err.Error())
	})
}

func TestZoomConfigDecrypt(t *testing.T) {
	t.Run("Decrypted succeeded", func(t *testing.T) {
		accountID := "AccountID"
		clientID := "ClientID"
		clientSecret := "ClientSecret"
		encryptKey := "48404D635166546A576E5A7234753778"
		zoomConfigExpected := ZoomConfig{AccountID: accountID, ClientID: clientID, ClientSecret: clientSecret}
		zoomConfigEncrypted, _ := zoomConfigExpected.ToEncrypted(encryptKey)
		zoomConfigFromEncryptedData, _ := zoomConfigEncrypted.ToDecrypt(encryptKey)

		assert.Equal(t, zoomConfigExpected.AccountID, zoomConfigFromEncryptedData.AccountID)
		assert.Equal(t, zoomConfigExpected.ClientID, zoomConfigFromEncryptedData.ClientID)
		assert.Equal(t, zoomConfigExpected.ClientSecret, zoomConfigFromEncryptedData.ClientSecret)

	})

	t.Run("Should error when key encrypt wrong format", func(t *testing.T) {

		accountID := "AccountID"
		clientID := "ClientID"
		clientSecret := "ClientSecret"
		encryptKey := "48404D635166546A576E5A7234753778"
		zoomConfigExpected := ZoomConfig{AccountID: accountID, ClientID: clientID, ClientSecret: clientSecret}
		zoomConfigEncrypted, _ := zoomConfigExpected.ToEncrypted(encryptKey)
		_, err := zoomConfigEncrypted.ToDecrypt("1")
		assert.Equal(t, "crypto/aes: invalid key size 0", err.Error())
	})
}

func TestZoomConfigEncryptedToJSON(t *testing.T) {
	t.Run("Encrypted succeeded", func(t *testing.T) {
		accountID := "AccountID"
		clientID := "ClientID"
		clientSecret := "ClientSecret"
		encryptKey := "48404D635166546A576E5A7234753778"
		zoomConfigExpected := ZoomConfig{AccountID: accountID, ClientID: clientID, ClientSecret: clientSecret}
		zoomConfigEncrypted, _ := zoomConfigExpected.ToEncrypted(encryptKey)
		strJson, _ := zoomConfigEncrypted.ToJSONString()

		zoomConfigActual, _ := InitZoomConfig(strJson)

		assert.Equal(t, zoomConfigEncrypted.AccountID, zoomConfigActual.AccountID)
		assert.Equal(t, zoomConfigEncrypted.ClientID, zoomConfigActual.ClientID)
		assert.Equal(t, zoomConfigEncrypted.ClientSecret, zoomConfigActual.ClientSecret)

	})

	t.Run("Should error when string zoom config wrong format", func(t *testing.T) {

		_, err := InitZoomConfig("")
		assert.Equal(t, "unexpected end of JSON input", err.Error())
	})
}
