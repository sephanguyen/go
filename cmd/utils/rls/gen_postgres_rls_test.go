package rls_test

import (
	"bytes"
	"testing"

	"github.com/manabie-com/backend/cmd/utils/rls"
	mock_fileio "github.com/manabie-com/backend/mock/golibs/io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRootCmd(t *testing.T) {
	mockFileIO := &mock_fileio.MockFileUtils{}
	rlsPostgres := &rls.Postgres{
		IOUtils: mockFileIO,
	}
	migrationFolder := "migrations/"
	messageError := "actual is not expected"
	rlsTypeFlag := "--rlsType"
	tableFlag := "--table"
	pkeyFlag := "--pkey"
	accessPathTableFlag := "--accessPathTable"
	databaseFlag := "--databaseName"
	permissionPrefixFlag := "--permissionPrefix"
	accessPathTableKeyFlag := "--accessPathTableKey"
	expectedFileMigration := "1001_migrate.up.sql"
	databaseName := "bob"
	expectedMigrateFolder := migrationFolder + databaseName
	expectedMigrateFile := migrationFolder + databaseName + "/" + expectedFileMigration
	lastFileMigrateFolder := "1005_migrate.up.sql"
	expectedNextFileMigration := "1006_migrate.up.sql"
	studentPrefixPermission := "user.student"
	templateVersionFlag := "--templateVersion"

	t.Run("Should generate migration file when migration folder is empty", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		expectedPolicy := `DROP POLICY IF EXISTS rls_students on "students";
CREATE POLICY rls_students_location ON "students" AS PERMISSIVE FOR ALL TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
with check (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
);`
		table := "students"

		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "student_id", accessPathTableFlag, "user_access_paths", databaseFlag, databaseName, permissionPrefixFlag, studentPrefixPermission, accessPathTableKeyFlag, "user_id"}

		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetErr(error)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should generate migration file when migration folder is empty and accessPathTableKey is missing", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		expectedPolicy := `DROP POLICY IF EXISTS rls_courses on "courses";
CREATE POLICY rls_courses_location ON "courses" AS PERMISSIVE FOR ALL TO PUBLIC
using (
course_id in (
	select			
		usp."course_id"
	from
					granted_permissions p
	join course_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.course.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
with check (
course_id in (
	select			
		usp."course_id"
	from
					granted_permissions p
	join course_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.course.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
);`
		table := "courses"
		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "course_id", accessPathTableFlag, "course_access_paths", databaseFlag, databaseName, permissionPrefixFlag, "user.course"}

		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetArgs(args)
		cmd.SetErr(error)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should generate migration file when migration folder is empty and accessPathTable is missing", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		expectedPolicy := `DROP POLICY IF EXISTS rls_lessons on "lessons";
CREATE POLICY rls_lessons_location ON "lessons" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	center_id in (
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
					p2.permission_name = 'user.lesson.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
)
with check (
	center_id in (
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
					p2.permission_name = 'user.lesson.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);`
		table := "lessons"
		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "center_id", databaseFlag, databaseName, permissionPrefixFlag, "user.lesson"}

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)

		cmd.SetArgs(args)
		cmd.SetErr(error)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should generate migration file when last file on migration folder is 1005_migrate.up.sql", func(t *testing.T) {
		// given
		table := "students"
		databaseName := "bob"
		expectedMigrateFolder := migrationFolder + databaseName
		expectedMigrateFile := migrationFolder + databaseName + "/" + expectedNextFileMigration
		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "student_id", accessPathTableFlag, "user_access_paths", databaseFlag, databaseName, permissionPrefixFlag, studentPrefixPermission}

		error := new(bytes.Buffer)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, mock.AnythingOfType("string")).Once().Return(nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetErr(error)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should throw error when missing one of args", func(t *testing.T) {
		// given
		args := []string{rlsTypeFlag, "pg", pkeyFlag, "user_id", accessPathTableFlag, "user_access_paths", permissionPrefixFlag, "user.user.read"}

		actual := new(bytes.Buffer)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetOut(actual)
		cmd.SetErr(actual)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := "Error: table arg is missing."
		assert.Contains(t, actual.String(), expected, messageError)
	})

	t.Run("Should generate RLS to permission template 4", func(t *testing.T) {
		// given

		expectedContent := `DROP POLICY IF EXISTS rls_students on "students";
CREATE POLICY rls_students_permission_v4 ON "students" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	current_setting('app.user_id') = owners
)
with check (
	current_setting('app.user_id') = owners
);`
		expectedMigrateFile := migrationFolder + databaseName + "/" + expectedNextFileMigration

		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)

		actual := new(bytes.Buffer)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetOut(actual)
		cmd.SetErr(actual)

		args := []string{rlsTypeFlag, "pg", pkeyFlag, "owners", tableFlag, "students", templateVersionFlag, "4", databaseFlag, databaseName}
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Contains(t, actual.String(), expected, messageError)
	})

	t.Run("Should generate RLS to permission template 1.1", func(t *testing.T) {
		// given
		table := "students"
		expectedContent := `DROP POLICY IF EXISTS rls_students on "students";
CREATE POLICY rls_students_insert_location ON "students" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_students_select_location ON "students" AS PERMISSIVE FOR select TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
CREATE POLICY rls_students_update_location ON "students" AS PERMISSIVE FOR update TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)with check (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
CREATE POLICY rls_students_delete_location ON "students" AS PERMISSIVE FOR delete TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
`
		expectedMigrateFile := migrationFolder + databaseName + "/" + expectedNextFileMigration

		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)

		actual := new(bytes.Buffer)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetOut(actual)
		cmd.SetErr(actual)

		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "student_id", accessPathTableFlag, "user_access_paths", databaseFlag, databaseName, permissionPrefixFlag, studentPrefixPermission, accessPathTableKeyFlag, "user_id", templateVersionFlag, "1.1"}
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Contains(t, actual.String(), expected, messageError)
	})

	t.Run("Should generate RLS to permission template 1 with access path table have column location name is center_id", func(t *testing.T) {
		// given
		table := "lesson_members"
		expectedContent := `DROP POLICY IF EXISTS rls_lesson_members on "lesson_members";
CREATE POLICY rls_lesson_members_insert_location ON "lesson_members" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_lesson_members_select_location ON "lesson_members" AS PERMISSIVE FOR select TO PUBLIC
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
;
CREATE POLICY rls_lesson_members_update_location ON "lesson_members" AS PERMISSIVE FOR update TO PUBLIC
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
				p2.permission_name = 'user.lesson.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)with check (
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
)
;
CREATE POLICY rls_lesson_members_delete_location ON "lesson_members" AS PERMISSIVE FOR delete TO PUBLIC
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
				p2.permission_name = 'user.lesson.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
`
		expectedMigrateFile := migrationFolder + databaseName + "/" + expectedNextFileMigration

		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)

		actual := new(bytes.Buffer)

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		cmd.SetOut(actual)
		cmd.SetErr(actual)

		args := []string{rlsTypeFlag, "pg", tableFlag, table, pkeyFlag, "lesson_id", accessPathTableFlag, "lessons", databaseFlag, databaseName, permissionPrefixFlag, "user.lesson", accessPathTableKeyFlag, "lesson_id", "--accessPathLocationCol", "center_id", templateVersionFlag, "1.1"}
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Contains(t, actual.String(), expected, messageError)
	})

	t.Run("Should generate AC template 3 success with lesson table", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		expectedPolicy := `DROP POLICY IF EXISTS rls_lessons on "lessons";
CREATE POLICY rls_lessons_permission_v3 ON "lessons" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	true <= (
		select			
			true
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
					p2.permission_name = 'user.lesson.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
)
with check (
	true <= (
		select			
			true
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
					p2.permission_name = 'user.lesson.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
);`
		args := []string{rlsTypeFlag, "pg", tableFlag, "lessons", templateVersionFlag, "3", databaseFlag, databaseName, permissionPrefixFlag, "user.lesson"}

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
		cmd.SetArgs(args)
		cmd.SetErr(error)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})

	t.Run("Should generate AC template 3 success with student table", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		expectedPolicy := `DROP POLICY IF EXISTS rls_students on "students";
CREATE POLICY rls_students_permission_v3 ON "students" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	true <= (
		select			
			true
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
					p2.permission_name = 'user.students.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
)
with check (
	true <= (
		select			
			true
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
					p2.permission_name = 'user.students.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
);`
		args := []string{rlsTypeFlag, "pg", tableFlag, "students", templateVersionFlag, "3", databaseFlag, databaseName, permissionPrefixFlag, "user.students"}

		cmd := rls.GetCmd(rlsPostgres, nil, nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedPolicy).Once().Return(nil, nil)

		cmd.SetArgs(args)
		cmd.SetErr(error)
		// when
		cmd.Execute()

		// then
		assert.Equal(t, error.String(), "", messageError)
	})
}
