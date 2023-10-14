package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for entryexitmgmt
type Config struct {
	Common                   configs.CommonConfig
	Issuers                  []configs.TokenIssuerConfig
	Storage                  configs.StorageConfig
	PostgresV2               configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS                   configs.NatsJetStreamConfig
	QrCodeEncryption         QrCodeEncryptionConfig      `yaml:"qrcode_encryption"`
	QrCodeEncryptionSynersia QrCodeEncryptionConfig      `yaml:"qrcode_encryption_synersia"`
	QrCodeEncryptionTokyo    QrCodeEncryptionConfig      `yaml:"qrcode_encryption_tokyo"`
	UnleashClientConfig      configs.UnleashClientConfig `yaml:"unleash_client"`
}

type QrCodeEncryptionConfig struct {
	SecretKey string `yaml:"secret_key"`
}
