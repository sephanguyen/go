package gcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	oauth2l "github.com/google/oauth2l/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type MockHttpClient struct {
	mockDo func(req *http.Request) (*http.Response, error)
}

func (client *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return client.mockDo(req)
}

func aValidGCPHashConfig() *HashConfig {
	return &HashConfig{
		HashAlgorithm:  "SCRYPT",
		HashRounds:     8,
		HashMemoryCost: 8,
		HashSaltSeparator: Base64EncodedStr{
			Value:        "salt",
			DecodedBytes: []byte("salt"),
		},
		HashSignerKey: Base64EncodedStr{
			Value:        "key",
			DecodedBytes: []byte("key"),
		},
	}
}

func TestHashConfig_GetMethods(t *testing.T) {
	t.Parallel()

	hashConfig := aValidGCPHashConfig()

	assert.Equal(t, hashConfig.HashAlgorithm, "SCRYPT")
	assert.Equal(t, hashConfig.HashSignerKey.DecodedBytes, hashConfig.Key())
	assert.Equal(t, hashConfig.HashSaltSeparator.DecodedBytes, hashConfig.SaltSeparator())
	assert.Equal(t, hashConfig.HashMemoryCost, hashConfig.MemoryCost())
	assert.Equal(t, hashConfig.Rounds(), hashConfig.Rounds())
}

func TestBase64EncodedStr_UnmarshalJSON(t *testing.T) {
	response :=
		`
		{
		  "name": "projects/888663408192/config",
		  "signIn": {
			"email": {
			  "enabled": true,
			  "passwordRequired": true
			},
			"hashConfig": {
			  "algorithm": "SCRYPT",
			  "signerKey": "ZXhhbXBsZS1zaWduZXIta2V5",
			  "saltSeparator": "ZXhhbXBsZS1zYWx0LXNlcGFyYXRvcg==",
			  "rounds": 8,
			  "memoryCost": 14
			}
		  }
		}
		`

	projectConfig := &ProjectConfig{}
	err := json.Unmarshal([]byte(response), projectConfig)

	assert.Nil(t, err)
	assertHashConfig(t, projectConfig.SignIn.HashConfig)
}

func assertHashConfig(t *testing.T, hashConfig *HashConfig) {
	assert.NotNil(t, hashConfig)
	assert.Equal(t, hashConfig.HashSignerKey.Value, "ZXhhbXBsZS1zaWduZXIta2V5")
	assert.NotNil(t, hashConfig.HashSignerKey.DecodedBytes)
	decodedHashSignerKey, err := base64.StdEncoding.DecodeString(hashConfig.HashSignerKey.Value)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, decodedHashSignerKey, hashConfig.HashSignerKey.DecodedBytes)

	assert.Equal(t, hashConfig.HashSaltSeparator.Value, "ZXhhbXBsZS1zYWx0LXNlcGFyYXRvcg==")
	assert.NotNil(t, hashConfig.HashSaltSeparator.DecodedBytes)
	decodedHashSaltSeparator, err := base64.StdEncoding.DecodeString(hashConfig.HashSaltSeparator.Value)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, decodedHashSaltSeparator, hashConfig.HashSaltSeparator.DecodedBytes)

	assert.Equal(t, hashConfig.HashAlgorithm, "SCRYPT")
	assert.Equal(t, hashConfig.HashRounds, 8)
	assert.Equal(t, hashConfig.HashMemoryCost, 14)
}

const validGetProjectConfigResponse = `
	{
	  "name": "projects/1234566789/config",
	  "signIn": {
		"email": {
		  "enabled": true,
		  "passwordRequired": true
		},
		"hashConfig": {
		  "algorithm": "SCRYPT",
		  "signerKey": "ZXhhbXBsZS1zaWduZXIta2V5",
		  "saltSeparator": "ZXhhbXBsZS1zYWx0LXNlcGFyYXRvcg==",
		  "rounds": 8,
		  "memoryCost": 14
		}
	  },
	  "notification": {
		"sendEmail": {
		  "method": "DEFAULT",
		  "resetPasswordTemplate": {
			"senderLocalPart": "noreply",
			"subject": "Reset your password for %APP_NAME%",
			"body": "\u003cp\u003eHello,\u003c/p\u003e\n\u003cp\u003eFollow this link to reset your %APP_NAME% password for your %EMAIL% account.\u003c/p\u003e\n\u003cp\u003e\u003ca href='%LINK%'\u003e%LINK%\u003c/a\u003e\u003c/p\u003e\n\u003cp\u003eIf you didn’t ask to reset your password, you can ignore this email.\u003c/p\u003e\n\u003cp\u003eThanks,\u003c/p\u003e\n\u003cp\u003eYour %APP_NAME% team\u003c/p\u003e",
			"bodyFormat": "HTML",
			"replyTo": "noreply"
		  },
		  "verifyEmailTemplate": {
			"senderLocalPart": "noreply",
			"subject": "Verify your email for %APP_NAME%",
			"body": "\u003cp\u003eHello %DISPLAY_NAME%,\u003c/p\u003e\n\u003cp\u003eFollow this link to verify your email address.\u003c/p\u003e\n\u003cp\u003e\u003ca href='%LINK%'\u003e%LINK%\u003c/a\u003e\u003c/p\u003e\n\u003cp\u003eIf you didn’t ask to verify this address, you can ignore this email.\u003c/p\u003e\n\u003cp\u003eThanks,\u003c/p\u003e\n\u003cp\u003eYour %APP_NAME% team\u003c/p\u003e",
			"bodyFormat": "HTML",
			"replyTo": "noreply"
		  },
		  "changeEmailTemplate": {
			"senderLocalPart": "noreply",
			"subject": "Your sign-in email was changed for %APP_NAME%",
			"body": "\u003cp\u003eHello %DISPLAY_NAME%,\u003c/p\u003e\n\u003cp\u003eYour sign-in email for %APP_NAME% was changed to %NEW_EMAIL%.\u003c/p\u003e\n\u003cp\u003eIf you didn’t ask to change your email, follow this link to reset your sign-in email.\u003c/p\u003e\n\u003cp\u003e\u003ca href='%LINK%'\u003e%LINK%\u003c/a\u003e\u003c/p\u003e\n\u003cp\u003eThanks,\u003c/p\u003e\n\u003cp\u003eYour %APP_NAME% team\u003c/p\u003e",
			"bodyFormat": "HTML",
			"replyTo": "noreply"
		  },
		  "callbackUri": "https://dev-manabie-online.firebaseapp.com/__/auth/action",
		  "dnsInfo": {
			"customDomainState": "NOT_STARTED",
			"domainVerificationRequestTime": "1970-01-01T00:00:00Z"
		  },
		  "revertSecondFactorAdditionTemplate": {
			"senderLocalPart": "noreply",
			"subject": "You've added 2 step verification to your %APP_NAME% account.",
			"body": "\u003cp\u003eHello %DISPLAY_NAME%,\u003c/p\u003e\n\u003cp\u003eYour account in %APP_NAME% has been updated with %SECOND_FACTOR% for 2-step verification.\u003c/p\u003e\n\u003cp\u003eIf you didn't add this 2-step verification, click the link below to remove it.\u003c/p\u003e\n\u003cp\u003e\u003ca href='%LINK%'\u003e%LINK%\u003c/a\u003e\u003c/p\u003e\n\u003cp\u003eThanks,\u003c/p\u003e\n\u003cp\u003eYour %APP_NAME% team\u003c/p\u003e",
			"bodyFormat": "HTML",
			"replyTo": "noreply"
		  }
		},
		"sendSms": {
		  "smsTemplate": {
			"content": "%LOGIN_CODE% is your verification code for %APP_NAME%."
		  }
		},
		"defaultLocale": "en"
	  },
	  "quota": {},
	  "monitoring": {
		"requestLogging": {}
	  },
	  "multiTenant": {
		"allowTenants": true
	  },
	  "authorizedDomains": [
		"localhost",
		"dev-manabie-online.firebaseapp.com",
		"dev-manabie-online.web.app"
	  ],
	  "subtype": "IDENTITY_PLATFORM",
	  "client": {
		"apiKey": "example-api-key",
		"permissions": {},
		"firebaseSubdomain": "dev-manabie-online"
	  },
	  "mfa": {
		"state": "DISABLED"
	  },
	  "blockingFunctions": {
		"forwardInboundCredentials": {}
	  },
	  "smsRegionConfig": {}
	}
`

func TestApp_GetProjectConfig(t *testing.T) {
	ctx := context.Background()

	assert.NoError(t, setServiceCredential())

	mockHTTPClient := &MockHttpClient{
		func(req *http.Request) (*http.Response, error) {
			response := &http.Response{
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(validGetProjectConfigResponse))),
				StatusCode: http.StatusOK,
			}
			return response, nil
		},
	}

	tokenFetcher := func(ctx context.Context, settings *oauth2l.Settings) (*oauth2.Token, error) {
		return &oauth2.Token{}, nil
	}

	app, err := NewApp(ctx, mockHTTPClient, tokenFetcher, "", "example-project-id")
	assert.NoError(t, err)
	assert.NotNil(t, app)
}

const validGetTenantConfigResponse = `
	{
	  "name": "projects/123456789/tenants/end-to-end-school-5xn27",
	  "displayName": "end-to-end-school",
	  "allowPasswordSignup": true,
	  "hashConfig": {
		"algorithm": "SCRYPT",
		"signerKey": "ZXhhbXBsZS1zaWduZXIta2V5",
		"saltSeparator": "ZXhhbXBsZS1zYWx0LXNlcGFyYXRvcg==",
		"rounds": 8,
		"memoryCost": 14
	  },
	  "inheritance": {}
	}
`

func TestApp_GetTenantConfig(t *testing.T) {
	ctx := context.Background()

	assert.NoError(t, setServiceCredential())

	mockHTTPClient := &MockHttpClient{
		func(req *http.Request) (*http.Response, error) {
			response := &http.Response{
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(validGetProjectConfigResponse))),
				StatusCode: http.StatusOK,
			}
			return response, nil
		},
	}

	tokenFetcher := func(ctx context.Context, settings *oauth2l.Settings) (*oauth2.Token, error) {
		return &oauth2.Token{}, nil
	}

	app, err := NewApp(ctx, mockHTTPClient, tokenFetcher, "", "example-project-id")
	assert.NoError(t, err)
	assert.NotNil(t, app)

	app.HttpClient = &MockHttpClient{
		func(req *http.Request) (*http.Response, error) {
			response := &http.Response{
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(validGetTenantConfigResponse))),
				StatusCode: http.StatusOK,
			}
			return response, nil
		},
	}

	tenantConfig, err := app.GetTenantConfig(ctx, "example-tenant-id")
	assert.NoError(t, err)
	assert.NotNil(t, tenantConfig)
	assertHashConfig(t, tenantConfig.HashConfig)
}

var (
	tmpDir       string
	credFilePath string
)

func init() {
	tmpDir = os.TempDir()
	credFilePath = filepath.Join(tmpDir, "service_credentials.json")
}

func setServiceCredential() error {
	//The private-key value is randomly generated for testing
	const defaultServiceCredentialJSON = `{
"type": "service_account",
"project_id": "dev-manabie-online",
"private_key_id": "example-private-key-id",
"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDGd0vGdoTx1QFv\nMDF0kGDEXpeisb6CA46FmY/yaKP9CemgJtdRixirEGoikeBXpTQbgVbc8TnZ+OG7\n0IuaDPSkCJw1ykJ9gwToXkN0Edszx3N13+VYTmnGzrpdEam3uXdPYXLtMwAAkJ9I\n+DoM+8CFbDIFhqUaoa9jTGoSS9NXtiLBYwDFxvbRJFUAh7e6IADtkl3CWM0L0RRG\nRfJyAbw2X5G+MUS0tDEIIvQh6HJR3TLKGIT9fso/lfvnECSzifXwgCmg72MP5FPW\nfJd+batZEJ6Wtxhs3UmoWpEbMetkmsVKWqzKjfQV57pcUgJzOcSyt2BhGR7nqWFN\nmbmf0z8RAgMBAAECggEAGQ5ioLHB4w4zWihJdh/sN56BomayWJO+YJuckswnOAES\nX8fHk2HuQVqXK7ojCq2uwHI51zcVSLGlPiL8HPzZvgPgROI+Nr5d1kBgX70JYaYi\nq5USzW1I6XKcELf0J1/g8kKpUc0IiQm5Mms0WQNHsRCR2CTBn3UeQXkaQykTi5UY\nJZ1TlqNgOyMJR2qCrhoi+Sft3JF/dUj+iNk1qRc79PvECd/SGb1N6NWWM2plS4ge\n2Z013noUJOrZnPDPbFz4l1qzuW8YkhwL/CcM372t6gGyqWaJZuCrCMf2nsAczOum\nlqY2Nac1SQiM18/6EFUPz0Y0gQlWYftQcgQQGZ3GqQKBgQDqCdWQuhhpp2+t26CR\nWmn6QpHa888Lc41AqGLFUJvJLXGJsylYfY1vC93wL/Q0ZsO9nnmnz5MbCTSggufl\nQVLcpQU3XIHZmMo/HHfq8MXLT9iXZzstK0wtJL4M/e7YfPaizkkh7d2B9c8Ry6iK\nlUUofY4KmySuwxR7K1nxO9AoswKBgQDZFu1A5Mp+muhvINNIg1Q4sHtAJWBbd7Om\ngrrQfGWbh2wcyMLdYLESXTCvwaszZqGVS2zhZZ1i4gGcfwATeYFDnHEixwhM2MrP\nGtRQQ9CqRYCnu0wPeXFejsfpszzmUaDunpK496AUjfCAdIr3BaXkyH+qU1tdCvjy\nWFEq6hxzKwKBgQC9kyW5S+TGgGhILiVMWC6MFyxKbT+DCSCcBUmshvUJ6pOTdNrC\n4UCVeMlX66Amai+YAyyML+n69mP4uNDatSVHsSweggJ0nf0FTiwc1NeDLrRFP8uB\ndRcJYj/IClFUbzTg/7PhlendgZ0vzwZA61TPzZQnJzB5l2+Zra33Z/nfXQKBgCjW\nDzstzomSSjbdTeFOEwG28PhYD5AlLD4eSVX+kH55MvUXLtDF54k0znvBSpsYqzyS\nO6EKpFh9eyAdI76GFLLLMtz/46fRABWFTnrqxs3A1Tq4GM6wYYsQALsNZF9O6573\nZVI2An7bVGpVge6FuXcX4CwCEiWmcr3jryELeN6RAoGAcJPTTYjygbg8YEHM34zv\nFns0s/qHjRq9luwTMCm5OGn4RY7EM0H+Z4S3fyDInf1fjMAZPS0I/5bLhG55GVvX\ne6ptzR8mPLUxeO2fKSZ7FjwXgSbEcRfwUidUUkAggvSpeYk9jD1If6I43Ih19Svr\nqMLjb8NkISDKA50+r8XP90A=\n-----END PRIVATE KEY-----",
"client_email": "example-email@dev-manabie-online.iam.gserviceaccount.com",
"client_id": "example-client-id",
"auth_uri": "https://accounts.google.com/o/oauth2/auth",
"token_uri": "https://oauth2.googleapis.com/token",
"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/bootstrap%40dev-manabie-online.iam.gserviceaccount.com"
}`
	err := os.WriteFile(credFilePath, []byte(defaultServiceCredentialJSON), 0666)
	if err != nil {
		return fmt.Errorf("failed to write to credential file: %s", err)
	}
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFilePath)
	if err != nil {
		return fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS env: %s", err)
	}
	return nil
}

func TestNewGCPApp(t *testing.T) {
	ctx := context.Background()

	assert.NoError(t, setServiceCredential())

	mockHTTPClient := &MockHttpClient{
		func(req *http.Request) (*http.Response, error) {
			response := &http.Response{
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(validGetProjectConfigResponse))),
				StatusCode: http.StatusOK,
			}
			return response, nil
		},
	}

	tokenFetcher := func(ctx context.Context, settings *oauth2l.Settings) (*oauth2.Token, error) {
		return &oauth2.Token{}, nil
	}

	app, err := NewGCPApp(ctx, mockHTTPClient, tokenFetcher, "", &configurations.MultiTenantConfig{})
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, app)
}
