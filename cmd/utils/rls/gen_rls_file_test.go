package rls_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/manabie-com/backend/cmd/utils/rls"
	mock_fileio "github.com/manabie-com/backend/mock/golibs/io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getDirByFileName(filename string) string {
	rootMockFolder := "../../../mock/testing/testdata/rls"
	return fmt.Sprintf("%s/%s", rootMockFolder, filename)
}

func getFileMockAc(filename string) ([]byte, error) {
	fileDirMock := getDirByFileName(filename)
	content, err := ioutil.ReadFile(fileDirMock)
	if err != nil {
		return nil, err
	}
	return content, nil
}

const fileTemplate1Stage = `[
    {
        "filename": "accesscontrol/bob/archtechture.lessons.yaml",
        "service": "bob",
        "revision": 1,
        "table_name": "lessons",
        "created_at": "2022-10-05T11:33:32.051291717+07:00",
        "updated_at": "2022-10-05T11:33:32.051291779+07:00",
        "stages": [
            {
                "template": "1",
                "hasura": {
                    "stage_dir": "deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/tables.yaml",
                    "permissions": [
                        "SELECT"
                    ],
                    "relationship": "lessons_location_permission",
					"first_level_query": "",
					"hasura_policies": {}
                },
                "postgres": {
                    "stage_dir": "migrations/bob/1022_migrate.up.sql",
					"policies": [
                        {
                            "name": "rls_lessons_location",
                            "content": "content"
                        }
                    ]
                },
                "accessPathTable": null,
                "locationCol": null,
                "permissionPrefix": "lesson.lessons",
                "permissions": {
                    "postgres": [],
                    "hasura": []
                },
                "ownerCol": null
            }
        ]
    }
]`

func TestGenFileTemplateRootCmd(t *testing.T) {
	mockFileIO := &mock_fileio.MockFileUtils{}
	fileGen := &rls.FileTemplate{
		IOUtils: mockFileIO,
	}
	rlsTypeFlag := "--rlsType"
	messageError := "actual is not expected"
	folderName := []string{"bob"}
	folderDir := "accesscontrol/bob"
	expectedMigrateFolder := "migrations/bob"
	lastFileMigrateFolder := "1001_migrate.up.sql"
	expectedMigrateFile := expectedMigrateFolder + "/" + "1002_migrate.up.sql"
	stagesFile := "accesscontrol/stage.json"
	notFoundFileErrMsg := "no such file or directory"
	expectedHasuraFile := "deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/tables.yaml"
	unitByteType := "[]uint8"
	lessonCaseFile := "lesson_case.yaml"
	lessonCaseExpectedFile := "lesson_case_expected.yaml"

	testCases := []RLSTestCase{
		{
			name:           "Should call successfully file generate for file with template 1.1 and 4",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				expectedRelationshipAccessPath := `- name: students_location_permission
    using:
      manual_configuration:
        remote_table:
          schema: public
          name: granted_permissions
        column_mapping:
          location_id: location_id
  - name: students
    using:
      manual_configuration:
        remote_table:
          schema: public
          name: students
        column_mapping:
          user_id: student_id`
				expectedMainTableRelationship := `  - name: user_access_paths
    using:
      manual_configuration:
        remote_table:
          schema: public
          name: user_access_paths
        column_mapping:
          student_id: user_id`
				acStudentsFilename := "ac_students.yaml"
				acStudentsFilenames := []string{acStudentsFilename}
				acStudentsFileDir := folderDir + "/" + acStudentsFilename

				content, err := getFileMockAc(acStudentsFilename)
				mockFileIO.On("GetFoldersOnDir", "accesscontrol").Once().Return(folderName, nil)
				mockFileIO.On("GetFileContent", stagesFile).Once().Return(nil, fmt.Errorf(notFoundFileErrMsg))
				mockFileIO.On("GetFileContent", acStudentsFileDir).Once().Return(content, nil)
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return(acStudentsFilenames, nil).Once()

				// template 1.1
				// verify gen postgres
				mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.AnythingOfType("string")).Once().Return(nil, nil)
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil).Once()

				mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.MatchedBy(func(content string) bool {
					return strings.Contains(content, "CREATE POLICY rls_students_insert_location ON \"students\" AS PERMISSIVE FOR INSERT TO PUBLIC")
				})).Once().Return(nil, nil)

				// verify input of output of hasura gen
				studentCaseInput := getDirByFileName("student_case.yaml")
				studentCaseExpected := getDirByFileName("student_case_expected.yaml")

				studentContentInput, _, err := getMockData(studentCaseInput, studentCaseExpected)

				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(studentContentInput, nil)
				mockFileIO.On("WriteFile", stagesFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, mock.MatchedBy(func(content []byte) bool {
					contentStr := string(content)
					return strings.Contains(contentStr, expectedRelationshipAccessPath) && strings.Contains(contentStr, expectedMainTableRelationship)
				})).Once().Return(nil, nil)

				// template 4
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil).Once()
				mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.AnythingOfType("string")).Once().Return(nil, nil)

				// verify input of output of hasura gen
				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(studentContentInput, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil)

				return err
			},
		},
		{
			name:           "Should call successfully file generate for file with template 1",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				acLessonFilename := "ac_lessons.yaml"
				acLessonFilenames := []string{acLessonFilename}
				acLessonFileDir := folderDir + "/" + acLessonFilename

				content, err := getFileMockAc(acLessonFilename)

				mockFileIO.On("GetFileContent", acLessonFileDir).Once().Return(content, nil)
				mockFileIO.On("GetFileContent", stagesFile).Once().Return(nil, fmt.Errorf(notFoundFileErrMsg))
				mockFileIO.On("GetFoldersOnDir", "accesscontrol").Once().Return(folderName, nil)
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return(acLessonFilenames, nil).Once()

				// template 1
				// verify gen postgres
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil).Once()
				mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.MatchedBy(func(content string) bool {
					return strings.Contains(content, "CREATE POLICY rls_lessons_location ON \"lessons\" AS PERMISSIVE FOR ALL TO PUBLIC")
				})).Once().Return(nil, nil)

				// verify input of output of hasura gen
				lessonCaseInput := getDirByFileName(lessonCaseFile)
				lessonCaseExpected := getDirByFileName(lessonCaseExpectedFile)
				lessonContentInput, _, err := getMockData(lessonCaseInput, lessonCaseExpected)

				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(lessonContentInput, nil)
				mockFileIO.On("WriteFile", stagesFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, mock.MatchedBy(func(content []byte) bool {
					contentStr := string(content)
					return strings.Contains(contentStr, "- permission_name:") && strings.Contains(contentStr, "_eq: lesson.lessons.read")
				})).Once().Return(nil, nil)

				return err
			},
		},
		{
			name:           "Should throw error when missing table name",
			expectedErrStr: "Error: genPostgresRLS - table arg is missing. \n",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				filename := "ac_missing_template.yaml"
				filenames := []string{filename}
				fileDir := folderDir + "/" + filename
				content, err := getFileMockAc(filename)

				mockFileIO.On("GetFileContent", stagesFile).Once().Return(nil, fmt.Errorf(notFoundFileErrMsg))
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return(filenames, nil).Once()
				mockFileIO.On("GetFileContent", fileDir).Once().Return(content, nil)
				mockFileIO.On("GetFoldersOnDir", "accesscontrol").Once().Return(folderName, nil)

				return err
			},
		},
		{
			name:           "Should stages capture the change when users change from template 1 to template 4",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				acLessonFilename := "ac_lessons_template4.yaml"
				acLessonFileDir := folderDir + "/" + acLessonFilename
				acMainFolder := "accesscontrol"

				content, err := getFileMockAc(acLessonFilename)

				mockFileIO.On("GetFileContent", acLessonFileDir).Once().Return(content, nil)
				mockFileIO.On("GetFileContent", stagesFile).Once().Return([]byte(fileTemplate1Stage), nil)
				mockFileIO.On("GetFoldersOnDir", acMainFolder).Once().Return(folderName, nil)
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return([]string{acLessonFilename}, nil)

				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{"1000_migrate.up.sql"}, nil).Once()
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil).Once()

				// template 1
				// verify gen postgres
				dropPolicyFile := expectedMigrateFolder + "/" + lastFileMigrateFolder
				mockFileIO.On("WriteStringFile", dropPolicyFile, mock.MatchedBy(func(content string) bool {
					return strings.Contains(content, "DROP POLICY IF EXISTS rls_lessons_location on \"lessons\";")
				})).Once().Return(nil, nil)

				createPolicyFile := expectedMigrateFolder + "/" + "1002_migrate.up.sql"
				mockFileIO.On("WriteStringFile", createPolicyFile, mock.MatchedBy(func(content string) bool {
					return strings.Contains(content, "CREATE POLICY rls_lessons_permission_v4 ON \"lessons\" AS PERMISSIVE FOR ALL TO PUBLIC")
				})).Once().Return(nil, nil)

				// verify input of output of hasura gen
				lessonCaseInput := getDirByFileName(lessonCaseFile)
				lessonCaseExpected := getDirByFileName(lessonCaseExpectedFile)
				lessonContentInput, _, err := getMockData(lessonCaseInput, lessonCaseExpected)

				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(lessonContentInput, nil)
				mockFileIO.On("WriteFile", stagesFile, mock.MatchedBy(func(content []byte) bool {
					contentStr := string(content)
					return strings.Contains(contentStr, "\"revision\": 2") && strings.Contains(contentStr, "\"template\": \"4\"") && strings.Contains(contentStr, "rls_lessons_permission_v4")
				})).Once().Return(nil, nil).Once()

				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(lessonContentInput, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil).Once()

				mockFileIO.On("WriteFile", expectedHasuraFile, mock.MatchedBy(func(content []byte) bool {
					contentStr := string(content)
					return strings.Contains(contentStr, "- center_id:") && strings.Contains(contentStr, "_eq: X-Hasura-User-Id")
				})).Once().Return(nil, nil).Once()

				return err
			},
		},
		{
			name:           "Should call successfully file generate for file with template 3",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				acLessonTemplate3Filename := "ac_lessons_template_3.yaml"
				acLessonTemplate3Filenames := []string{acLessonTemplate3Filename}
				acLessonTemplate3FileDir := folderDir + "/" + acLessonTemplate3Filename

				content, err := getFileMockAc(acLessonTemplate3Filename)

				mockFileIO.On("GetFileContent", acLessonTemplate3FileDir).Once().Return(content, nil)
				mockFileIO.On("GetFileContent", stagesFile).Once().Return(nil, fmt.Errorf(notFoundFileErrMsg))
				mockFileIO.On("GetFoldersOnDir", "accesscontrol").Once().Return(folderName, nil)
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return(acLessonTemplate3Filenames, nil).Once()

				// template 3
				// verify gen postgres
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil).Once()
				mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.MatchedBy(func(content string) bool {
					return strings.Contains(content, "CREATE POLICY rls_lessons_permission_v3 ON \"lessons\" AS PERMISSIVE FOR ALL TO PUBLIC")
				})).Once().Return(nil, nil)

				// template 3
				// verify input of output of hasura gen
				lessonContentInput, _, err := getMockData(getDirByFileName(lessonCaseFile), getDirByFileName(lessonCaseExpectedFile))
				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(lessonContentInput, nil)
				mockFileIO.On("WriteFile", stagesFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, mock.MatchedBy(func(content []byte) bool {
					contentStr := string(content)
					queryStr := `_and:
        - _exists:
            _table:
              name: granted_permissions
              schema: public
            _where:
              _and:
              - user_id:
                  _eq: X-Hasura-User-Id
              - permission_name:
                  _eq: lesson.lessons.read`
					return strings.Contains(contentStr, queryStr)
				})).Once().Return(nil, nil)

				return err
			},
		},
		{
			name:           "Should call successfully when gen hasura custom policy",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, rls.GenRLSFile},
			setup: func() error {
				acHasuraCustomPolicyFilename := "ac_hasura_custom_policy.yaml"
				acHasuraCustomPolicyFilenames := []string{acHasuraCustomPolicyFilename}
				acHasuraCustomPolicyFileDir := folderDir + "/" + acHasuraCustomPolicyFilename

				content, err := getFileMockAc(acHasuraCustomPolicyFilename)

				mockFileIO.On("GetFileContent", acHasuraCustomPolicyFileDir).Once().Return(content, nil)
				mockFileIO.On("GetFileContent", stagesFile).Once().Return(nil, fmt.Errorf(notFoundFileErrMsg))
				mockFileIO.On("GetFoldersOnDir", "accesscontrol").Once().Return(folderName, nil)
				mockFileIO.On("GetFileNamesOnDir", folderDir).Once().Return(acHasuraCustomPolicyFilenames, nil).Once()

				// template custom
				// verify input of output of hasura gen
				input, expected, err := getMockData(getDirByFileName("ac_hasura_custom_policy_input.yaml"), getDirByFileName("ac_hasura_custom_policy_expected.yaml"))
				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(input, nil)
				mockFileIO.On("WriteFile", stagesFile, mock.AnythingOfType(unitByteType)).Once().Return(nil, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, expected).Once().Return(nil, nil)

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// case
			err := testCase.setup()

			error := new(bytes.Buffer)

			cmd := rls.GetCmd(nil, nil, nil, fileGen)
			// cmd.SetOut(error)
			cmd.SetErr(error)
			cmd.SetArgs(testCase.args)

			// when
			cmd.Execute()

			// then
			assert.Equal(t, err, nil, messageError)
			assert.Equal(t, testCase.expectedErrStr, error.String(), messageError)
		})
	}
}
