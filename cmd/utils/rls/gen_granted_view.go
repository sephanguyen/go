package rls

import (
	"fmt"

	"github.com/spf13/cobra"
)

type GrantedView struct {
	IOUtils interface {
		GetFileNamesOnDir(filename string) ([]string, error)
		WriteStringFile(filename string, content string) error
	}
}

const grantedViewWithCacheTableFunc = `CREATE OR REPLACE FUNCTION get_granted_permissions_with_cache_table () 
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
LANGUAGE 'plpgsql';`

const template = `create or replace
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

const cachingViewTemplate = `create or replace
view granted_permissions as
select
	user_id,
	permission_name,
	location_id,
	resource_path
from get_granted_permissions_with_cache_table();`

func checkGrantedViewArgs() string {
	errMsg := ""
	if databaseName == "" {
		errMsg += "databaseName arg is missing."
	}

	return errMsg
}

func (p *GrantedView) getNewMigrateFile() (string, error) {
	svcFolder := fmt.Sprintf("%s/%s", migrationFolder, databaseName)
	files, err := p.IOUtils.GetFileNamesOnDir(svcFolder)

	if err != nil {
		return "", err
	}

	sqlFiles := filterSQLFiles(files)
	lastFile := ""
	newMigraNum := "1001"

	if len(sqlFiles) > 0 {
		lastFile = sqlFiles[len(sqlFiles)-1]
		newMigraNum, err = getNewMigrateNumber(lastFile)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s/%s_migrate.up.sql", svcFolder, newMigraNum), nil
}

func getTemplate() string {
	twoNewLine := "\n\n"
	if rlsType == ViewType {
		return template
	}
	return grantedViewWithCacheTableFunc + twoNewLine + cachingViewTemplate
}

func (p *GrantedView) genGrantedView(cmd *cobra.Command, args []string) error {
	fmt.Println("Running genGrantedView")

	errMsg := checkGrantedViewArgs()
	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}

	newMigrateFile, err := p.getNewMigrateFile()

	if err != nil {
		return fmt.Errorf("getNewMigrateFile error %w", err)
	}

	err = p.IOUtils.WriteStringFile(newMigrateFile, getTemplate())

	if err != nil {
		return fmt.Errorf("file can't be write file %w", err)
	}

	fmt.Println("file generated to: ", newMigrateFile)
	return nil
}
