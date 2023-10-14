package configurations

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
)

const (
	validGoogleCloudProject      = "example-google-cloud-project"
	validIdentityPlatformProject = "example-identity-platform-project"
)

func TestConfig_GetMultiTenantProjectID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                         string
		config                       Config
		expectedMultiTenantProjectID string
	}{
		{
			name: "has both GoogleCloudProjectID and IdentityPlatformProjectID",
			config: Config{
				Common: configs.CommonConfig{
					GoogleCloudProject:      validGoogleCloudProject,
					IdentityPlatformProject: validIdentityPlatformProject,
				},
			},
			expectedMultiTenantProjectID: validIdentityPlatformProject,
		},
		{
			name: "has GoogleCloudProjectID but doesn't has IdentityPlatformProjectID",
			config: Config{
				Common: configs.CommonConfig{
					GoogleCloudProject: validGoogleCloudProject,
				},
			},
			expectedMultiTenantProjectID: validGoogleCloudProject,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedMultiTenantProjectID, testCase.config.GetMultiTenantProjectID())
		})
	}
}
