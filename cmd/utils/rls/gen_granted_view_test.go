package rls_test

import (
	"bytes"
	"testing"

	"github.com/manabie-com/backend/cmd/utils/rls"
	mock_fileio "github.com/manabie-com/backend/mock/golibs/io"

	"github.com/stretchr/testify/assert"
)

func TestGenGrantedViewCmd(t *testing.T) {
	mockFileIO := &mock_fileio.MockFileUtils{}
	grantedView := &rls.GrantedView{
		IOUtils: mockFileIO,
	}
	migrationFolder := "migrations/"
	messageError := "actual is not expected"
	rlsTypeFlag := "--rlsType"
	databaseFlag := "--databaseName"
	expectedFileMigration := "1001_migrate.up.sql"
	databaseName := "bob"
	expectedMigrateFolder := migrationFolder + databaseName
	expectedMigrateFile := migrationFolder + databaseName + "/" + expectedFileMigration
	expectedContent := `create or replace
view public.granted_permissions
as
select
	ugm.user_id,
	p.permission_name,
	l1.location_id,
	ugm.resource_path,
	p.permission_id
from
	user_group_member ugm
join user_group ug on
	ugm.user_group_id = ug.user_group_id
join granted_role gr on
	ug.user_group_id = gr.user_group_id
join role r on
	gr.role_id = r.role_id
join permission_role pr on
	r.role_id = pr.role_id
join permission p on
	p.permission_id = pr.permission_id
join granted_role_access_path grap on
	gr.granted_role_id = grap.granted_role_id
join locations l on
	l.location_id = grap.location_id
join locations l1 on
	l1.access_path ~~ (l.access_path || '%'::text)
where
	ugm.deleted_at is null
	and ug.deleted_at is null
	and gr.deleted_at is null
	and r.deleted_at is null
	and pr.deleted_at is null
	and p.deleted_at is null
	and grap.deleted_at is null
	and l.deleted_at is null
	and l1.deleted_at is null;`

	t.Run("Should generate migration file when migration folder is empty", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)

		args := []string{rlsTypeFlag, "view", databaseFlag, databaseName}

		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)

		cmd := rls.GetCmd(nil, nil, grantedView, nil)
		cmd.SetErr(error)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should generate migration file when last file on migration folder is 1005_migrate.up.sql", func(t *testing.T) {
		// given
		lastFileMigrateFolder := "1005_migrate.up.sql"
		expectedFileMigration := "1006_migrate.up.sql"
		databaseName := "bob"
		expectedMigrateFolder := migrationFolder + databaseName
		expectedMigrateFile := migrationFolder + databaseName + "/" + expectedFileMigration
		args := []string{rlsTypeFlag, "view", databaseFlag, databaseName}

		error := new(bytes.Buffer)
		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return([]string{lastFileMigrateFolder}, nil)

		cmd := rls.GetCmd(nil, nil, grantedView, nil)
		cmd.SetErr(error)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
	t.Run("Should throw error when missing databaseName arg", func(t *testing.T) {
		// given
		args := []string{rlsTypeFlag, "view"}

		actual := new(bytes.Buffer)

		cmd := rls.GetCmd(nil, nil, grantedView, nil)
		cmd.SetOut(actual)
		cmd.SetErr(actual)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := "Error: databaseName arg is missing."
		assert.Contains(t, actual.String(), expected, messageError)
	})
	t.Run("Should caching view for granted permissions correctly", func(t *testing.T) {
		// given
		error := new(bytes.Buffer)
		args := []string{rlsTypeFlag, "cache-view", databaseFlag, databaseName}
		expectedContent = `CREATE OR REPLACE FUNCTION get_granted_permissions_with_cache_table () 
RETURNS TABLE (
	user_id TEXT,
	permission_name TEXT,
	location_id TEXT,
	resource_path TEXT
) 
SECURITY INVOKER
AS $$
BEGIN
RETURN QUERY select 
	ugm.user_id as user_id,
	gp.permission_name as permission_name,
	l1.location_id as location_id,
	ugm.resource_path as resource_path
from
	user_group_member ugm
	join granted_permission gp on ugm.user_group_id = gp.user_group_id
		and ugm.resource_path = gp.resource_path
	join locations l on l.location_id = gp.location_id
		and l.resource_path = gp.resource_path
	join locations l1 on l1.access_path ~ l.access_path
		and l1.resource_path = l.resource_path
where 
	ugm.deleted_at is null
	and l.deleted_at is null
	and l1.deleted_at is null;
END; $$ 
LANGUAGE 'plpgsql';

create or replace
view granted_permissions as
select
	user_id,
	permission_name,
	location_id,
	resource_path
from get_granted_permissions_with_cache_table();`

		mockFileIO.On("WriteStringFile", expectedMigrateFile, expectedContent).Once().Return(nil, nil)
		mockFileIO.On("GetFileNamesOnDir", expectedMigrateFolder).Once().Return(nil, nil)

		cmd := rls.GetCmd(nil, nil, grantedView, nil)
		cmd.SetErr(error)
		cmd.SetArgs(args)

		// when
		cmd.Execute()

		// then
		expected := ""
		assert.Equal(t, error.String(), expected, messageError)
	})
}
