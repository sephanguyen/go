package usermgmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchoolIdAndTenant(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name                 string
		inputEnv             string
		inputFirebaseProject string
		expectedOutput       map[int]string
	}{
		{name: "local", inputEnv: "local", expectedOutput: LocalSchoolAndTenantIDMap},
		{name: "stag", inputEnv: "stag", expectedOutput: stagingSchoolAndTenantIDMap},
		{name: "uat", inputEnv: "uat", expectedOutput: uatSchoolAndTenantIDMap},
		{
			name:                 "prod manabie",
			inputEnv:             "prod",
			inputFirebaseProject: prodManabieFirebaseProject,
			expectedOutput:       prodManabieSchoolAndTenantIDMap,
		},
		{
			name:                 "prod synersia",
			inputEnv:             "prod",
			inputFirebaseProject: prodSynerisaFirebaseProject,
			expectedOutput:       prodSynersiaSchoolAndTenantIDMap,
		},
		{
			name:                 "prod renseikai",
			inputEnv:             "prod",
			inputFirebaseProject: prodRenseikaiFirebaseProject,
			expectedOutput:       prodRenseikaiSchoolAndTenantIDMap,
		},
		{
			name:                 "prod GA",
			inputEnv:             "prod",
			inputFirebaseProject: prodGAFirebaseProject,
			expectedOutput:       prodGAdSchoolAndTenantIDMap,
		},
		{
			name:                 "prod KEC",
			inputEnv:             "prod",
			inputFirebaseProject: prodKECFirebaseProject,
			expectedOutput:       prodKECSchoolAndTenantIDMap,
		},
		{
			name:                 "prod AIC",
			inputEnv:             "prod",
			inputFirebaseProject: prodAICFirebaseProject,
			expectedOutput:       prodAICSchoolAndTenantIDMap,
		},
		{
			name:                 "prod NSG",
			inputEnv:             "prod",
			inputFirebaseProject: prodNSGFirebaseProject,
			expectedOutput:       prodNSGSchoolAndTenantIDMap,
		},
		{
			name:                 "prod E2E Tokyo",
			inputEnv:             "prod",
			inputFirebaseProject: prodE2ETokyoFirebaseProject,
			expectedOutput:       prodE2ETokyoSchoolAndTenantIDMap,
		},
		{name: "undefined", inputEnv: "undefined", expectedOutput: nil},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, schoolIDAndTenant(testcase.inputEnv, testcase.inputFirebaseProject), testcase.expectedOutput)
		})
	}
}
