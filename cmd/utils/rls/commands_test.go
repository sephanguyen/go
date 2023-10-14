package rls_test

import (
	"bytes"
	"testing"

	"github.com/manabie-com/backend/cmd/utils/rls"
	mock_fileio "github.com/manabie-com/backend/mock/golibs/io"

	"github.com/stretchr/testify/assert"
)

func TestRLSRootCmd(t *testing.T) {
	mockFileIO := &mock_fileio.MockFileUtils{}
	rlsHasura := &rls.Hasura{
		IOUtils: mockFileIO,
	}
	pg := &rls.Postgres{
		IOUtils: mockFileIO,
	}
	messageError := "actual is not expected"
	rlsTypeFlag := "--rlsType"
	tableFlag := "--table"
	pkeyFlag := "--pkey"
	databaseFlag := "--databaseName"
	permissionPrefixFlag := "--permissionPrefix"
	accessPathTableFlag := "--accessPathTable"
	accessPathTableKeyFlag := "--accessPathTableKey"

	expectedHasuraFile := "deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata/tables.yaml"
	databaseName := "bob"

	classCaseInput := "../../../mock/testing/testdata/rls/class_case.yaml"
	classCaseExpected := "../../../mock/testing/testdata/rls/class_case_expected.yaml"

	expectedFileMigration := "1001_migrate.up.sql"
	userClassPrefixPermission := "user.class"
	migrationFolder := "migrations/"
	expectedMigrateFolder := migrationFolder + databaseName
	expectedMigrateFile := migrationFolder + databaseName + "/" + expectedFileMigration

	testCases := []RLSTestCase{
		{
			name:           "Should call success both hasura rls and pg rls",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, "rls", tableFlag, "class", pkeyFlag, "location_id", databaseFlag, databaseName, permissionPrefixFlag, userClassPrefixPermission},
			setup: func() error {
				classContentInput, classContentExpected, err := getMockData(classCaseInput, classCaseExpected)
				expectedPolicy := `DROP POLICY IF EXISTS rls_class on "class";
CREATE POLICY rls_class_location ON "class" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	location_id in (
		select			
			p.location_id
		from
						granted_permissions p
		where
			p.user_id = current_setting('app.user_id')
			and p.permission_id = (
				select
					p2.permission_id
				from
					"permission" p2
				where
					p2.permission_name = 'user.class.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
)
with check (
	location_id in (
		select			
			p.location_id
		from
						granted_permissions p
		where
			p.user_id = current_setting('app.user_id')
			and p.permission_id = (
				select
					p2.permission_id
				from
					"permission" p2
				where
					p2.permission_name = 'user.class.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);`

				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(classContentInput, nil)
				mockFileIO.On("WriteFile", expectedHasuraFile, classContentExpected).Once().Return(nil, nil)

				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
				mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)

				return err
			},
		},
		{
			name:           "Should call success both hasura rls and pg rls with access paths table",
			expectedErrStr: "",
			args:           []string{rlsTypeFlag, "rls", tableFlag, "lesson_members", pkeyFlag, "lesson_id", databaseFlag, databaseName, permissionPrefixFlag, "user.lesson", accessPathTableFlag, "lessons", accessPathTableKeyFlag, "lesson_id", "--accessPathLocationCol", "center_id"},
			setup: func() error {
				expectedPolicy := `DROP POLICY IF EXISTS rls_lesson_members on "lesson_members";
CREATE POLICY rls_lesson_members_location ON "lesson_members" AS PERMISSIVE FOR ALL TO PUBLIC
using (
lesson_id in (
	select			
		usp."lesson_id"
	from
					granted_permissions p
	join lessons usp on
					usp.center_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.lesson.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
with check (
lesson_id in (
	select			
		usp."lesson_id"
	from
					granted_permissions p
	join lessons usp on
					usp.center_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.lesson.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
);`
				mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
				mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)

				inputFile := "../../../mock/testing/testdata/rls/lesson_members_case.yaml"
				expectedFile := "../../../mock/testing/testdata/rls/lesson_members_expected.yaml"
				lessonMemberContentInput, lessonMemberContentExpected, err := getMockData(inputFile, expectedFile)
				mockFileIO.On("WriteFile", expectedHasuraFile, lessonMemberContentExpected).Once().Return(nil, nil)
				mockFileIO.On("GetFileContent", expectedHasuraFile).Once().Return(lessonMemberContentInput, nil)
				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// case
			err := testCase.setup()

			error := new(bytes.Buffer)

			cmd := rls.GetCmd(pg, rlsHasura, nil, nil)
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
