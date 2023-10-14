package services

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/shamir/configurations"

	"github.com/stretchr/testify/assert"
)

func TestSalesforceService_TestGetAccessToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"access_token":"valid token"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	testCases := []struct {
		name        string
		userID      string
		clientID    string
		setup       func() SalesforceService
		expected    string
		expectedErr error
	}{
		{
			name: "happy case",
			setup: func() SalesforceService {
				pemBytes, err := os.ReadFile("salesforce_key_test.pem")
				if err != nil {
					log.Fatalf("Error reading PEM file: %v", err)
				}

				config := make(map[string]configurations.SalesforceOrgConfig)
				config["manabie"] = configurations.SalesforceOrgConfig{
					Key:      string(pemBytes),
					ClientID: "salesforce_client_id",
				}
				return SalesforceService{
					Config: configurations.SalesforceConfigs{
						Aud:                 "https://login.salesforce.com",
						AccessTokenEndpoint: server.URL,
						Configurations:      config,
					},
				}
			},
			expected: "valid token",
		},
		{
			name: "invalid private key",
			setup: func() SalesforceService {
				pemBytes, err := os.ReadFile("salesforce_key_test_invalid.pem")
				if err != nil {
					log.Fatalf("Error reading PEM file: %v", err)
				}

				config := make(map[string]configurations.SalesforceOrgConfig)
				config["manabie"] = configurations.SalesforceOrgConfig{
					Key:      string(pemBytes),
					ClientID: "salesforce_client_id",
				}
				return SalesforceService{
					Config: configurations.SalesforceConfigs{
						Aud:                 "https://login.salesforce.com",
						AccessTokenEndpoint: server.URL,
						Configurations:      config,
					},
				}
			},
			expectedErr: fmt.Errorf(fmt.Errorf("invalid private key").Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := testCase.setup()

			resp, err := service.GetAccessToken(fmt.Sprint(constants.ManabieSchool), testCase.userID)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expected, resp)
		})
	}
}
