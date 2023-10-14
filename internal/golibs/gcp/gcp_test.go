package gcp

import (
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/stretchr/testify/assert"
)

const (
	validGCPProjectID        = "example-project-id"
	validGCPServiceAccountID = "example-service-account-id"
)

type mockAppConfig struct {
	gcpProjectID        string
	gcpServiceAccountID string
}

func (config mockAppConfig) GetGCPProjectID() string {
	return config.gcpProjectID
}

func (config mockAppConfig) GetGCPServiceAccountID() string {
	return config.gcpServiceAccountID
}

func TestAppConfigToFirebaseConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                   string
		appConfig              AppConfig
		expectedFirebaseConfig firebase.Config
	}{
		{
			name: "has both project id and service account id",
			appConfig: mockAppConfig{
				gcpProjectID:        validGCPProjectID,
				gcpServiceAccountID: validGCPServiceAccountID,
			},
			expectedFirebaseConfig: firebase.Config{
				ProjectID:        validGCPProjectID,
				ServiceAccountID: validGCPServiceAccountID,
			},
		},
		{
			name: "has project id but service account id is empty",
			appConfig: mockAppConfig{
				gcpProjectID: validGCPProjectID,
			},
			expectedFirebaseConfig: firebase.Config{
				ProjectID: validGCPProjectID,
			},
		},
		{
			name:                   "both project id and service account id are empty",
			appConfig:              mockAppConfig{},
			expectedFirebaseConfig: firebase.Config{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			firebaseConfig := appConfigToFirebaseConfig(testCase.appConfig)
			assert.Equal(t, testCase.expectedFirebaseConfig, *firebaseConfig)
		})
	}
}
