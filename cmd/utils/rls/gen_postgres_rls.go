package rls

import (
	"fmt"
	"strconv"
	"strings"
)

type Postgres struct {
	IOUtils interface {
		GetFileNamesOnDir(filename string) ([]string, error)
		WriteStringFile(filename string, content string) error
	}
}

const defaultLocationColName = "location_id"
const migrationFolder = "migrations"
const rlsFilter = `%s (
%s in (
	select			
		usp."%s"
	from
					granted_permissions p
	join %s usp on
					usp.%s = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = '%s'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)`
const templateRls = `CREATE POLICY %s ON "%s" AS PERMISSIVE FOR %s TO PUBLIC
%s
%s;`
const templateRlsLocationInsideTable = `CREATE POLICY %s ON "%s" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	%s in (
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
					p2.permission_name = '%s'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
)
with check (
	%s in (
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
					p2.permission_name = '%s'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);`
const templateDropPolicy = `DROP POLICY IF EXISTS %s on "%s";`
const templateRlsV4 = `CREATE POLICY %s ON "%s" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	current_setting('app.user_id') %s
)
with check (
	current_setting('app.user_id') %s
);`
const templateInsertAnyRLS = `CREATE POLICY %s ON "%s" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);`
const writePermission = ".write"
const readPermission = ".read"
const rlsPrefix = "rls_%s"
const templateACTemplate3 = `CREATE POLICY %s ON "%s" AS PERMISSIVE FOR ALL TO PUBLIC
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
					p2.permission_name = '%s'
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
					p2.permission_name = '%s'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
);`

func checkPGArgs() string {
	errMsg := ""
	if table == "" {
		errMsg += "table arg is missing. "
	}
	if pkey == "" && templateVersion != "3" {
		errMsg += "pkey arg is missing. "
	}
	if databaseName == "" {
		databaseName += "service arg is missing. "
	}
	if permissionPrefix == "" {
		permissionPrefix += "service arg is missing. "
	}

	return errMsg
}

func filterSQLFiles(files []string) []string {
	rs := files[:0]
	for _, f := range files {
		if strings.HasSuffix(f, ".sql") {
			rs = append(rs, f)
		}
	}
	return rs
}

func getNewMigrateNumber(filename string) (string, error) {
	s := strings.Split(filename, "_")

	curMigraNum, err := strconv.Atoi(s[0])

	if err != nil {
		return "", err
	}

	return strconv.Itoa(curMigraNum + 1), nil
}

func (p *Postgres) getNewMigrateFile(svc string) (string, error) {
	svcFolder := fmt.Sprintf("%s/%s", migrationFolder, svc)
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

type PolicyExpression struct {
	policyFor        string
	expression       []string
	permissionSuffix string
}

func getPolicyExpressions() []PolicyExpression {
	return []PolicyExpression{{policyFor: "select", expression: []string{"using"}, permissionSuffix: readPermission}, {policyFor: "update", expression: []string{"using", "with check"}, permissionSuffix: writePermission}, {policyFor: "delete", expression: []string{"using"}, permissionSuffix: writePermission}}
}

func buildTemplateV4() (string, string) {
	queryStr := fmt.Sprintf("= %s", pkey)
	policyName := fmt.Sprintf("rls_%s_permission_v4", table)
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	dropPermissivePolicyMutiTenant := fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"
	policy := fmt.Sprintf(templateRlsV4, policyName, table, queryStr, queryStr)
	return policyName, dropPermissivePolicyMutiTenant + policy
}

func buildTemplateV3() (string, string) {
	policyName := fmt.Sprintf("rls_%s_permission_v3", table)
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	dropPermissivePolicyMutiTenant := fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"
	policy := fmt.Sprintf(templateACTemplate3, policyName, table, permissionPrefix+readPermission, permissionPrefix+writePermission)
	return policyName, dropPermissivePolicyMutiTenant + policy
}

func buildTemplateV11() ([]PostgresPolicyStage, string) {
	policyContent := ""
	policyExpressions := getPolicyExpressions()
	policies := []PostgresPolicyStage{}

	for _, policyExpression := range policyExpressions {
		policyName := fmt.Sprintf("rls_%s_%s_location", table, policyExpression.policyFor)

		command := ""
		for _, express := range policyExpression.expression {
			command += fmt.Sprintf(rlsFilter, express, pkey, accessPathTableKey, accessPathTable, accessPathLocationCol, permissionPrefix+policyExpression.permissionSuffix)
		}

		content := fmt.Sprintf(templateRls, policyName, table, policyExpression.policyFor, command, "")

		policies = append(policies, PostgresPolicyStage{Name: policyName, Content: content})

		policyContent += content + "\n"
	}
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	dropPermissivePolicyMutiTenant := fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"

	insertPolicyName := fmt.Sprintf("rls_%s_insert_location", table)
	insertPolicy := fmt.Sprintf(templateInsertAnyRLS, insertPolicyName, table) + "\n"

	policies = append(policies, PostgresPolicyStage{Name: insertPolicyName, Content: insertPolicy})

	return policies, dropPermissivePolicyMutiTenant + insertPolicy + policyContent
}

func buildTemplateV1() (string, string) {
	policyContent := ""
	policyName := fmt.Sprintf("rls_%s_location", table)
	readRule := fmt.Sprintf(rlsFilter, "using", pkey, accessPathTableKey, accessPathTable, accessPathLocationCol, permissionPrefix+readPermission)
	writeRule := fmt.Sprintf(rlsFilter, "with check", pkey, accessPathTableKey, accessPathTable, accessPathLocationCol, permissionPrefix+writePermission)
	policyContent += fmt.Sprintf(templateRls, policyName, table, "ALL", readRule, writeRule)
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	dropPermissivePolicyMutiTenant := fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"
	return policyName, dropPermissivePolicyMutiTenant + policyContent
}

func correctInputForTemplate1() {
	if accessPathTableKey == "" {
		accessPathTableKey = pkey
	}
	if accessPathLocationCol == "" {
		accessPathLocationCol = defaultLocationColName
	}
}

func buildTemplateV1NoAccessPathTable() (string, string) {
	policyName := fmt.Sprintf("rls_%s_location", table)
	permissionRead := permissionPrefix + readPermission
	permissionWrite := permissionPrefix + writePermission
	dropPolicyName := fmt.Sprintf(rlsPrefix, table)
	dropPermissivePolicyMutiTenant := fmt.Sprintf(templateDropPolicy, dropPolicyName, table) + "\n"
	policyContent := dropPermissivePolicyMutiTenant + fmt.Sprintf(templateRlsLocationInsideTable, policyName, table, pkey, permissionRead, pkey, permissionWrite)
	return policyName, policyContent
}

func buildDropPolicy(policies []PostgresPolicyStage, tableName string) string {
	dropPolicies := ""
	for _, v := range policies {
		dropPolicies += fmt.Sprintf(templateDropPolicy, v.Name, tableName) + "\n"
	}
	return dropPolicies
}

func correctInputForTemplate1WithoutAccessPathTable() {
	if accessPathLocationCol != "" {
		pkey = accessPathLocationCol
	}
}

func (p *Postgres) genPostgresRLS() (*PostgresTemplateStage, error) {
	fmt.Println("Running genRLSPostgres")

	errMsg := checkPGArgs()
	if errMsg != "" {
		return nil, fmt.Errorf(errMsg)
	}

	policyContent := ""
	var policies []PostgresPolicyStage

	switch {
	case templateVersion == "3":
		policyName, content := buildTemplateV3()
		policyContent = content
		policies = []PostgresPolicyStage{{Name: policyName, Content: content}}
	case templateVersion == "4":
		policyName, content := buildTemplateV4()
		policyContent = content
		policies = []PostgresPolicyStage{{Name: policyName, Content: content}}
	case templateVersion == "1.1":
		correctInputForTemplate1()
		policyName, content := buildTemplateV11()
		policyContent = content
		policies = policyName
	case accessPathTable == "":
		correctInputForTemplate1WithoutAccessPathTable()
		policyName, content := buildTemplateV1NoAccessPathTable()
		policyContent = content
		policies = []PostgresPolicyStage{{Name: policyName, Content: content}}
	default:
		correctInputForTemplate1()
		policyName, content := buildTemplateV1()
		policyContent = content
		policies = []PostgresPolicyStage{{Name: policyName, Content: content}}
	}

	fmt.Printf("Policy content: %s \n", policyContent)

	newMigrateFile, err := p.getNewMigrateFile(databaseName)

	if err != nil {
		return nil, fmt.Errorf("getNewMigrateFile error %w", err)
	}

	err = p.IOUtils.WriteStringFile(newMigrateFile, policyContent)

	if err != nil {
		return nil, fmt.Errorf("file can't be write file %w", err)
	}

	fmt.Println("file generated to: ", newMigrateFile)
	return &PostgresTemplateStage{FileDir: newMigrateFile, Policies: policies}, nil
}
