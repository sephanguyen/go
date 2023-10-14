package database

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func VerifyAllTableWithRLS() error {
	svc, err := getDirectories(snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to get sub directory in %s: %w", snapshotDir, err)
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "../../../migrations")
	ignoreFile := filepath.Join(dir, "public_tables.json")
	ignoreTableMap, err := LoadIgnoreTableJSON(ignoreFile)
	if err != nil {
		return fmt.Errorf("failed to load ignore table json file: %w", err)
	}
	ignoreACTableMap, err := loadIgnoreACTables()
	if err != nil {
		return fmt.Errorf("failed to load ignore table json file: %w", err)
	}
	for _, s := range svc {
		if ignoreRLSService[s] {
			continue
		}

		files, err := getFilesInDirectory(filepath.Join(snapshotDir, s), ".json")
		if err != nil {
			return fmt.Errorf("failed to read json file in directory %s: %w", snapshotDir, err)
		}

		bypassRLSAccounts := BypassRLSAccountList{}

		if len(files) > 0 {
			err = loadJSON(filepath.Join(snapshotDir, s, bypassRLSAccountFileName), &bypassRLSAccounts)
			if err != nil {
				return fmt.Errorf("failed to load bypass rls account list: %w", err)
			}
		}

		for _, f := range files {
			tblSchema := &tableSchema{}
			err = loadJSON(f, tblSchema)
			if err != nil {
				return fmt.Errorf("failed to read json file %s: %s", f, err)
			}
			// skip check for schema_versioning file
			if len(tblSchema.Schema) == 0 && len(tblSchema.Policies) == 0 && tblSchema.TableName == "" {
				continue
			}
			err = VerifyRLS(s, tblSchema, ignoreTableMap, bypassRLSAccounts.Account, ignoreACTableMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type PublicTableList struct {
	Database string   `json:"name"`
	Tables   []string `json:"tables"`
}

func LoadIgnoreTableJSON(ignoreFilePath string) (map[string]map[string]bool, error) {
	ignoreTableJSON := make(map[string][]PublicTableList)
	err := loadJSON(ignoreFilePath, &ignoreTableJSON)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("failed to load ignore table json file %s: %s", ignoreFilePath, err))
	}
	ignoreTableMap := make(map[string]map[string]bool)
	for _, databases := range ignoreTableJSON {
		for _, database := range databases {
			t := make(map[string]bool)
			for _, table := range database.Tables {
				t[table] = true
			}
			ignoreTableMap[database.Database] = t
		}
	}
	return ignoreTableMap, nil
}

func tableCheck(sv string, tblSchema *tableSchema) error {
	permissiveFlag := false
	restrictiveFlag := false

	for _, p := range tblSchema.Policies {
		if !p.Relforcerowsecurity.Bool {
			return fmt.Errorf("please force row level security for table %s in service %s", p.Name.String, sv)
		}
		if !p.RelrowSecurity.Bool {
			return fmt.Errorf("row security is not enable for table %s in service %s", p.Name.String, sv)
		}
		if p.Qual.String != fmt.Sprintf("permission_check(resource_path, '%s'::text)", p.Name.String) {
			return fmt.Errorf("function permission_check is not in policy for table %s in service %s. Please change to permission_check(resource_path, '%s'::text)", p.Name.String, sv, p.Name.String)
		}
		if p.WithCheck.String != fmt.Sprintf("permission_check(resource_path, '%s'::text)", p.Name.String) {
			return fmt.Errorf("with_check in policy does not use function permission_check for table %s in service %s. Please use with_check in policy with permission_check(resource_path, '%s'::text)", p.Name.String, sv, p.Name.String)
		}

		if !isGrantedToPublic(p) {
			return fmt.Errorf("policy for table %s in service %s is not granted to public", p.Name.String, sv)
		}

		// Check both policy restrictive and permissive must exist
		switch p.PolicyName.String {
		case "rls_" + p.Name.String:
			if p.Permissive.String != "PERMISSIVE" {
				return fmt.Errorf("permission is not set to permissive in rls_"+p.Name.String+" for table %s in service %s", p.Name.String, sv)
			}
			permissiveFlag = true
		case "rls_" + p.Name.String + "_restrictive":
			if p.Permissive.String != "RESTRICTIVE" {
				return fmt.Errorf("permission is not set to restrictive in rls_"+p.Name.String+"_restrictive for table %s in service %s", p.Name.String, sv)
			}
			restrictiveFlag = true
		default:
			return fmt.Errorf("policy name is not in format rls_"+p.Name.String+" or rls_"+p.Name.String+"_restrictive for table %s in service %s", p.Name.String, sv)
		}
	}
	// schema versioning special
	if tblSchema.TableName != "" {
		if !permissiveFlag {
			return fmt.Errorf("table %s in service %s missing permissive rls policy", tblSchema.TableName, sv)
		}
		if !restrictiveFlag {
			return fmt.Errorf("table %s in service %s missing restrictive rls policy", tblSchema.TableName, sv)
		}
	}
	return nil
}

func isGrantedToPublic(p *tablePolicy) bool {
	for _, role := range p.Roles.Elements {
		if role.String == publicSchema {
			return true
		}
	}
	return false
}

func VerifyRLS(sv string, tblSchema *tableSchema, ignoreTableMap map[string]map[string]bool, bypassRLSAccounts []string, ignoreACTables map[string]map[string]bool) error {
	if ignoreTableMap[sv][tblSchema.TableName] || ignoreACTables[sv][tblSchema.TableName] {
		return nil
	}
	if tblSchema.Type == "BASE TABLE" {
		return tableCheck(sv, tblSchema)
	}
	if tblSchema.Type == "VIEW" {
		return fmt.Errorf("view %s in service %s is not supported. Please remove your view", tblSchema.TableName, sv)
	}
	return fmt.Errorf("table %s in %s has unexpected type. Expecting BASE TABLE or VIEW got %s", tblSchema.TableName, sv, tblSchema.Type)
}
