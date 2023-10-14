package configs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	tmpSubDir            string
	credFilePath         string
	commonConfigFilePath string
	configFilePath       string
	secretFilePath       string
	sopsSecretFilePath   string
	once                 sync.Once
)

func init() {
	tmpDir := os.TempDir() // TODO: use testing.T.TempDir() instead
	tmpSubDir = filepath.Join(tmpDir, "manabie")
	credFilePath = filepath.Join(tmpSubDir, "service_credentials.json")
	commonConfigFilePath = filepath.Join(tmpSubDir, "bob.common.config.yaml")
	configFilePath = filepath.Join(tmpSubDir, "bob.config.yaml")
	sopsSecretFilePath = filepath.Join(tmpSubDir, "bob.secrets.encrypted.yaml")
}

type testConfig struct {
	Common    string `yaml:"common"`
	Config    string `yaml:"config"`
	Secret    string `yaml:"secret"`
	SubConfig struct {
		Common string `yaml:"common"`
		Config string `yaml:"config"`
		Secret string `yaml:"secret"`
	} `yaml:"sub_config"`
}

func setServiceCredential() error {
	// Note: This is secret info. Even though anyone with access to backend repo can get this info,
	// it should never be leaked outside.
	const bobServiceCredentialJSON = `{
"type": "service_account",
"project_id": "dev-manabie-online",
"private_key_id": "b9429198d79997fda6d093bcebe81a7d7ab69938",
"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC6EhNnLsCquIaL\nvTlBOU+O4q4r8ZHzLzXD7IqtPxq5uRm7I5H1eOxMOCQzzgWdTVIhvychuyvLC+Vq\nvOv3+RR4ZUejmLLJuFrDD2pWPr2Nvm1X1r57TTldIr1c3Qcf5ipSzQ2n6gX9Nt8D\nHEkg3svRIR4fC+NorFQYIRFU95+R54Ui5s5GhHYYZTT19oiJrm+QER57IfPZXYQr\n9QdBjBA6B0jr9wlyTLAjeFL4STEboxilfwMbSj80sQkar/KbR2hByinlf/LTgCwz\nskkBk+llb2c7nZKih2x5zpehAgekGzeoAB6QlLqM2mP6URSO+KdQP7pCCh0RQxol\nGMY4x1iTAgMBAAECggEAB4VdhWktXnkw7wsJ+mnvnk3pTltoU9UPrkisXk5TrTgf\nIyJP7wUhP/9w7ysfrPkIHdcVJNbk8UMc1dCnFRHbUvZ9C87LQz4RZRsFaFEG5mjR\nEKDceC1p6SrTTqKcfByYj1o8eBIMheym3QBSsGJxCJX3GrgnS/7TM1p60d1kdMg+\nDzw/j4b+mMuTzr7pkg7hbpdTXkroK0zf9t5SHZJi/WKP9oDwfHx6K62tfCNlu1Qi\nasqGOG6K2lse/RiTA4zqE5BgfaH8oyxaj1yD9SM7bTaWkUMfdLKiMGTfiPAWAzMa\n3pTfzrT9iaCWYh5yb5p4xqFsXC4aboSRGw+utS5E+QKBgQD+HWSzsMpcN0jwPRN2\nUIG0jGzYv6hHZlAD+pAacintjLJDZ5eMf/WFYgF1jRBWMTDN9JwlQcuOrLyKJ5vt\nh1P7c1CJko5VHZnYw94vDnosZWKMmry6Lc4Jn8uEi0SIPscs/K8QQv+Qbl5ynaGa\nVKTcggbGLoXwzKMg8qyJRg8wpwKBgQC7c3R/TkGUJkm6TVp86sGCVmOfvcSfMI9c\nEwWK7IUpXgikYyNg+fqFrdxrA7tU/7hhaJreVkfFOM8o0nS/AG+9jTL6GhK80C/M\nXH5BBpqBVIMpsdH9fIsbPQ/vP+l5R6X8ZZEv+W22MA+C7j332vPKsvNflBmFBTqi\nWnPMYB5KNQKBgBI5UWuBlkGexWBVQPwPMf4cxAGXXR4hvENMyODcpx0eJfqnhzrQ\nQm9aY/hmMXG8/V8H19rkKREGWk8eIBScy+0QjAoRtJtuEAZ3pYuCYkikzLiAsGA5\nwLj3+MR8qGGM/wO+618jLujQwX0+yMQkpd4ahRnZZEmso1ZNkQoXOCepAoGARt7t\n6rvhm2umcGOSlKwFIYwb+mc7EZzAduVSMSYfanZ8+fnphF6+0w/ayDMO/qH4SgvM\nkcc5N121JQ/8x8IYfSgHX/u/nddwWumVamxeugsD1B3A8P/HcDLz9VbKpOnr3bNg\n4yyAyGL/WldM4orLpZVm4noR8/L4Ki3cniaxDQkCgYA8RK6NjzOFlQ45Oc/R9w0L\njXfm/VdXxNgij87WTXKylS4Dmk/bhsViCCk2ZkE7Ruo7iq9uorBnvCtUaPKyhSSi\n6GIfmYo4GS2/E+usaKfVTru/5R6lQ3wwkHVpGNymVeu7XEqy5uI2hSIKj19V79hI\n63a8cwnIHu5JlRONCVYprg==\n-----END PRIVATE KEY-----\n",
"client_email": "bootstrap@dev-manabie-online.iam.gserviceaccount.com",
"client_id": "104103173896408079531",
"auth_uri": "https://accounts.google.com/o/oauth2/auth",
"token_uri": "https://oauth2.googleapis.com/token",
"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/bootstrap%40dev-manabie-online.iam.gserviceaccount.com"
}`
	err := os.WriteFile(credFilePath, []byte(bobServiceCredentialJSON), 0o666)
	if err != nil {
		return fmt.Errorf("failed to write to credential file: %s", err)
	}
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFilePath) // TODO: use testing.T.Setenv()
	if err != nil {
		return fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS env: %s", err)
	}
	return nil
}

// Here, we simulate the configmap.yaml from Helm chart
func setConfig() error {
	const commonConfigData = `
common: common.common
config: common.config
secret: common.secret
sub_config:
  common: common.common
  config: common.config
  secret: common.secret
`
	err := os.WriteFile(commonConfigFilePath, []byte(commonConfigData), 0o666)
	if err != nil {
		return fmt.Errorf("failed to write to common config file: %s", err)
	}

	const configData = `
config: config.config
secret: config.secret
sub_config:
  config: config.config
  secret: config.secret
`
	err = os.WriteFile(configFilePath, []byte(configData), 0o666)
	if err != nil {
		return fmt.Errorf("failed to write config file: %s", err)
	}
	return nil
}

func setSOPSSecret() error {
	const secretDataBase64 = `c2VjcmV0OiBFTkNbQUVTMjU2X0dDTSxkYXRhOlZ0VFVNZnlUTnc5THhKUk5iZz09LGl2OlFLb3BT
Z0kvQTRyd2xvamVWRkFDS2o0UEI3d3k5WmJZY1FhR0toNnRqWm89LHRhZzpVcFlTWmxialJmbjVy
MXUwMXlKNlB3PT0sdHlwZTpzdHJdCnN1Yl9jb25maWc6CiAgICBzZWNyZXQ6IEVOQ1tBRVMyNTZf
R0NNLGRhdGE6WXk5OEExVTBSWTRuOUM1S2R3PT0saXY6TUVyRzNpR0RCREZiSVRWZ1lQMVVaeVMw
bkFxWDB3SWwvTXlRN2hiSy9QTT0sdGFnOmJ2TDlvRkxwSklRTGpOT1p5TkgvM3c9PSx0eXBlOnN0
cl0Kc29wczoKICAgIGttczogW10KICAgIGdjcF9rbXM6CiAgICAgICAgLSByZXNvdXJjZV9pZDog
cHJvamVjdHMvZGV2LW1hbmFiaWUtb25saW5lL2xvY2F0aW9ucy9nbG9iYWwva2V5UmluZ3MvZGVw
bG95bWVudHMvY3J5cHRvS2V5cy9naXRodWItYWN0aW9ucwogICAgICAgICAgY3JlYXRlZF9hdDog
IjIwMjItMDEtMTFUMDQ6MjY6MDRaIgogICAgICAgICAgZW5jOiBDaVFBLzJsSGpTNThBaDNGcTBZ
ZmgzSDlrTFY5WmMrQVJ3MkR3OTFkMFFKNVRDczVWS01TU1FDYWJDb3RLUVV6QkJLSEJyWnFKNUY5
aFRtdllmK05zcEU0OGtnbXN2L3VCMElRRUJtTHFhWmc4SkRORnQ1U056cUg5eS91QXZpeFlCQlBz
VndJc3VWUEVNeDhBU1p5dVhnPQogICAgYXp1cmVfa3Y6IFtdCiAgICBoY192YXVsdDogW10KICAg
IGFnZTogW10KICAgIGxhc3Rtb2RpZmllZDogIjIwMjItMDEtMTFUMDQ6MjY6MDRaIgogICAgbWFj
OiBFTkNbQUVTMjU2X0dDTSxkYXRhOjc5RkdtMVk3M2pKeWhjZFJQa05nWWJZdkV6NWE0M0tFQ1Jr
YUtBMzA1aWJXSlZ4MkVvWFJsZXA1WXo4QUJpZDVCUVYzNndtMnBTQ0RyZXZQWWlMcWsyMktCakdl
ZTNrMUpDcVlpTHdxY1FBTmxsREtYcWpKemNQK21kZ2NMZW5xV0laMUNmeU5YMUFzOWRLTDZVREJz
NXRiL2xYZlZZalBWbE0rY0RUWitTMD0saXY6UXd3OUZxUmZnbVByUk1CNlFjNndPU1I1R0lkbnlY
eWMvNTYwTDJqQklLbz0sdGFnOkRnazlBbU9Ha0Q4dWsyMnhSWlR1aGc9PSx0eXBlOnN0cl0KICAg
IHBncDogW10KICAgIHVuZW5jcnlwdGVkX3N1ZmZpeDogX3VuZW5jcnlwdGVkCiAgICB2ZXJzaW9u
OiAzLjcuMQo=
`
	secretData, err := base64.StdEncoding.DecodeString(secretDataBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %s", err)
	}
	err = os.WriteFile(sopsSecretFilePath, []byte(secretData), 0o666)
	if err != nil {
		return fmt.Errorf("failed to write to secret file: %s", err)
	}
	return nil
}

func setConfigsAndSOPSSecrets() (err error, cleanup func()) {
	err = func() error {
		err := os.MkdirAll(tmpSubDir, 0o777)
		if err != nil {
			return fmt.Errorf("os.MkdirAll: %s", err)
		}
		if err := setServiceCredential(); err != nil {
			return fmt.Errorf("setServiceCredential: %s", err)
		}
		if err := setConfig(); err != nil {
			return fmt.Errorf("setConfig: %s", err)
		}
		if err := setSOPSSecret(); err != nil {
			return fmt.Errorf("setSecret: %s", err)
		}
		return nil
	}()

	cleanup = func() {
		err := os.RemoveAll(tmpSubDir)
		if err != nil {
			log.Printf("could not clean up temp directory: %s", err)
		}
	}

	if err != nil {
		cleanup()
		return err, nil
	}
	return nil, cleanup
}

// This test should never be run in parallel.
func TestMustLoadConfigWithSOPS(t *testing.T) {
	err, cleanup := setConfigsAndSOPSSecrets()
	defer cleanup() // must clean up here to minimize secret leaks
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c := &testConfig{}
	MustLoadConfig(
		ctx,
		commonConfigFilePath,
		configFilePath,
		sopsSecretFilePath,
		c,
	)
	assert.Equal(t, "common.common", c.Common)
	assert.Equal(t, "common.common", c.SubConfig.Common)
	assert.Equal(t, "config.config", c.Config)
	assert.Equal(t, "config.config", c.SubConfig.Config)
	assert.Equal(t, "secret.secret", c.Secret)
	assert.Equal(t, "secret.secret", c.SubConfig.Secret)
}
