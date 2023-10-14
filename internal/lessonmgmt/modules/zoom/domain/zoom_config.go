package domain

import (
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/golibs/crypt"
)

type ZoomConfigEncrypted struct {
	AccountID    string `json:"account_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (c *ZoomConfigEncrypted) ToJSONString() (string, error) {
	bData, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(bData), nil
}

func InitZoomConfig(strData string) (*ZoomConfigEncrypted, error) {
	data := &ZoomConfigEncrypted{}
	err := json.Unmarshal([]byte(strData), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type ZoomConfig struct {
	AccountID    string
	ClientID     string
	ClientSecret string
}

type ZoomConfigCache struct {
	*ZoomConfig
	ExpireIn *time.Time
}

func (c *ZoomConfigEncrypted) ToDecrypt(encryptKey string) (*ZoomConfig, error) {
	accountID, err := crypt.Decrypt(c.AccountID, encryptKey)
	if err != nil {
		return nil, err
	}
	clientID, err := crypt.Decrypt(c.ClientID, encryptKey)
	if err != nil {
		return nil, err
	}
	clientSecret, err := crypt.Decrypt(c.ClientSecret, encryptKey)
	if err != nil {
		return nil, err
	}

	return &ZoomConfig{
		AccountID:    accountID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

func (s *ZoomConfig) ToEncrypted(encryptKey string) (*ZoomConfigEncrypted, error) {
	accountIDEncrypted, err := crypt.Encrypt(s.AccountID, encryptKey)
	if err != nil {
		return nil, err
	}
	clientIDEncrypted, err := crypt.Encrypt(s.ClientID, encryptKey)
	if err != nil {
		return nil, err
	}
	clientSecretEncrypted, err := crypt.Encrypt(s.ClientSecret, encryptKey)
	if err != nil {
		return nil, err
	}

	return &ZoomConfigEncrypted{
		AccountID:    accountIDEncrypted,
		ClientID:     clientIDEncrypted,
		ClientSecret: clientSecretEncrypted,
	}, nil
}
