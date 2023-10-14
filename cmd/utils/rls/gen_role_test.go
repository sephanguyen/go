package rls_test

import (
	"bytes"
	"testing"

	"github.com/manabie-com/backend/cmd/utils/rls"
	mock_fileio "github.com/manabie-com/backend/mock/golibs/io"

	"github.com/stretchr/testify/assert"
)

func setupComonMock(mockFileIO *mock_fileio.MockFileUtils, fileInput string, fileExpected string, expectedHasuraFile string) error {
	inputContent, expectedContent, err := getMockData(fileInput, fileExpected)

	mockFileIO.On("WriteFile", expectedHasuraFile, expectedContent).Once().Return(nil, nil, nil)
	mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(inputContent, nil, nil)

	return err
}

func TestGenRoleRootCmd(t *testing.T) {
	mockFileIO := &mock_fileio.MockFileUtils{}
	rlsHasura := &rls.Hasura{
		IOUtils: mockFileIO,
	}
	messageError := "actual is not expected"
	rlsTypeFlag := "--rlsType"
	databaseFlag := "--databaseName"
	expectedHasuraFile := "deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/tables.yaml"
	databaseName := "bob"

	input := "../../../mock/testing/testdata/rls/gen_role.yaml"
	expected := "../../../mock/testing/testdata/rls/gen_role_expected.yaml"
	expectedWithRemoveOptionExpected := "../../../mock/testing/testdata/rls/gen_role_with_remove_option_expected.yaml"

	genRoleTestCases := []RLSTestCase{
		{
			name:           "Should add new MANABIE role and keep all roles exists into metadata of hasura for each table in file",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRole, databaseFlag, databaseName},
			setup: func() error {
				return setupComonMock(mockFileIO, input, expected, expectedHasuraFile)
			},
		},
		{
			name:           "Should add new MANABIE role and remove all roles exists into metadata of hasura for each table in file",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRole, databaseFlag, databaseName, "--removeOldRole", "true"},
			setup: func() error {
				return setupComonMock(mockFileIO, input, expectedWithRemoveOptionExpected, expectedHasuraFile)
			},
		},
		{
			name:           "Should throw error when databaseName arg is missing",
			expectedErrStr: "Error: databaseName arg is missing. \n",
			args:           []string{rlsTypeFlag, rls.GenRole},
			setup: func() error {
				return setupComonMock(mockFileIO, input, expected, expectedHasuraFile)
			},
		},
	}

	for _, genRoleTestCase := range genRoleTestCases {
		t.Run(genRoleTestCase.name, func(t *testing.T) {
			// case
			error := new(bytes.Buffer)
			err := genRoleTestCase.setup()

			cmd := rls.GetCmd(nil, rlsHasura, nil, nil)

			cmd.SetArgs(genRoleTestCase.args)
			cmd.SetErr(error)

			// when
			cmd.Execute()

			// then
			assert.Equal(t, genRoleTestCase.expectedErrStr, error.String(), messageError)
			assert.Equal(t, err, nil, messageError)
		})
	}
}
