package configurations

import (
	"github.com/manabie-com/backend/internal/golibs/configs"
)

// Config for Enigma
type Config struct {
	Common                    configs.CommonConfig
	Upload                    configs.UploadConfig
	PostgresV2                configs.PostgresConfigV2 `yaml:"postgres_v2"`
	ASIAPAYHashSecret         string                   `yaml:"asiapay_hash_secret"`
	JPREPSignatureSecret      string                   `yaml:"jprep_signature_secret"`
	JPREPPayloadExpiredSec    int                      `yaml:"jprep_payload_expired_sec"`
	CloudConvertSigningSecret string                   `yaml:"cloud_convert_signing_secret"`
	NatsJS                    configs.NatsJetStreamConfig
	RouteCheckerHosts         []string `yaml:"route_checker_hosts"`
	RouteCheckerServices      []string `yaml:"route_checker_services"`
	Jira                      configs.JiraConfig
}
