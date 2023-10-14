package gcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	oauth2l "github.com/google/oauth2l/util"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

type ProjectConfig struct {
	Name        string                    `json:"name"`
	Client      *projectConfigClient      `json:"client"`
	SignIn      *ProjectConfigSignIn      `json:"signIn"`
	MultiTenant *projectConfigMultiTenant `json:"multiTenant"`
	Subtype     string                    `json:"subtype"`
}

type projectConfigClient struct {
	ApiKey            string `json:"apiKey"`
	FirebaseSubdomain string `json:"firebaseSubdomain"`
}

type projectConfigMultiTenant struct {
	AllowTenants bool `json:"allowTenants"`
}

type ProjectConfigSignIn struct {
	HashConfig *HashConfig `json:"hashConfig"`
}

type HashConfig struct {
	HashAlgorithm     string           `json:"algorithm"`
	HashSignerKey     Base64EncodedStr `json:"signerKey"`
	HashSaltSeparator Base64EncodedStr `json:"saltSeparator"`
	HashRounds        int              `json:"rounds"`
	HashMemoryCost    int              `json:"memoryCost"`
}

type Base64EncodedStr struct {
	Value        string
	DecodedBytes []byte
}

func (v *Base64EncodedStr) UnmarshalJSON(bytes []byte) error {
	// Strip the quotes
	if len(bytes) > 2 && bytes[0] == '"' && bytes[len(bytes)-1] == '"' {
		bytes = bytes[1 : len(bytes)-1]
	}

	strValue := string(bytes)
	v.Value = strValue

	decodedBytes, err := base64.StdEncoding.DecodeString(strValue)
	if err != nil {
		return errors.Wrap(err, "failed to decode string")
	}
	v.DecodedBytes = decodedBytes

	return nil
}

func (c *HashConfig) Key() []byte {
	if c == nil {
		return []byte{}
	}
	return c.HashSignerKey.DecodedBytes
}

func (c *HashConfig) SaltSeparator() []byte {
	if c == nil {
		return []byte{}
	}
	return c.HashSaltSeparator.DecodedBytes
}

func (c *HashConfig) Rounds() int {
	if c == nil {
		return 0
	}
	return c.HashRounds
}

func (c *HashConfig) MemoryCost() int {
	if c == nil {
		return 0
	}
	return c.HashMemoryCost
}

//Temporarily disable for now
/*func b64decode(s string) []byte {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Fatalln("Failed to decode string", err)
	}
	return b
}*/

func (app *App) GetProjectConfig(ctx context.Context) (*ProjectConfig, error) {
	credential, err := transport.Creds(ctx, option.WithCredentialsFile(app.credentialFile))
	if err != nil {
		return nil, err
	}

	oauth2lSetting := &oauth2l.Settings{
		CredentialsJSON: string(credential.JSON),
		AuthType:        oauth2l.AuthTypeOAuth,
		State:           "state",
		Scope:           "https://www.googleapis.com/auth/identitytoolkit",
	}

	token, err := oauth2l.FetchToken(ctx, oauth2lSetting)
	if err != nil {
		return nil, errors.Wrap(err, "FetchToken")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://identitytoolkit.googleapis.com/admin/v2/projects/%s/config", app.ProjectID), nil)
	if err != nil {
		return nil, errors.Wrap(err, "NewRequestWithContext")
	}
	token.SetAuthHeader(request)

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "httpClient.Do")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	default:
		return nil, errors.New(string(body))
	}

	config := &ProjectConfig{}
	if err := json.Unmarshal(body, config); err != nil {
		return nil, errors.Wrap(err, "Unmarshal")
	}

	return config, nil
}

type TenantConfigProvider interface {
	GetTenantConfig(ctx context.Context, tenantID string) (*TenantConfig, error)
}

type TenantConfig struct {
	Name                string      `json:"name"`
	DisplayName         string      `json:"displayName"`
	AllowPasswordSignup bool        `json:"allowPasswordSignup"`
	HashConfig          *HashConfig `json:"hashConfig"`
}

func (app *App) GetTenantConfig(ctx context.Context, tenantID string) (*TenantConfig, error) {
	credential, err := transport.Creds(ctx, option.WithCredentialsFile(app.credentialFile))
	if err != nil {
		return nil, err
	}

	oauth2lSetting := &oauth2l.Settings{
		CredentialsJSON: string(credential.JSON),
		AuthType:        oauth2l.AuthTypeOAuth,
		State:           "state",
		Scope:           "https://www.googleapis.com/auth/identitytoolkit",
	}

	token, err := oauth2l.FetchToken(ctx, oauth2lSetting)
	if err != nil {
		return nil, errors.Wrap(err, "FetchToken")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://identitytoolkit.googleapis.com/v2/projects/%s/tenants/%s", app.ProjectID, tenantID), nil)
	if err != nil {
		return nil, errors.Wrap(err, "NewRequestWithContext")
	}
	token.SetAuthHeader(request)

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "httpClient.Do")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	switch resp.StatusCode {
	case http.StatusOK:
		break
	default:
		return nil, errors.New(string(body))
	}

	config := &TenantConfig{}
	if err := json.Unmarshal(body, config); err != nil {
		return nil, errors.Wrap(err, "Unmarshal")
	}

	return config, nil
}
