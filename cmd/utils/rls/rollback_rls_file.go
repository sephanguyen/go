package rls

import (
	"bytes"
	texttemplate "text/template"
)

func genTemp(tmp string, data map[string]string) (string, error) {
	t := texttemplate.New("")
	t, _ = t.Parse(tmp)
	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func genPostgresRollbackFor(tableName string, dropPolicy string) (string, error) {
	template := `{{ .DropPolicy}}CREATE POLICY rls_{{ .TableName }} ON "{{ .TableName }}" 
USING (permission_check(resource_path, '{{ .TableName }}')) WITH CHECK (permission_check(resource_path, '{{ .TableName }}'));
` + "\n"
	return genTemp(template, map[string]string{"TableName": tableName, "DropPolicy": dropPolicy})
}

func genDropPGPolicyForTemplate(temp TemplateStage, tableName string) string {
	return buildDropPolicy(getPolicies(temp), tableName)
}

func isCustomPolicyHasura(currentTemplates []TemplateStage) bool {
	for _, currentTemplate := range currentTemplates {
		if currentTemplate.UseCustomHasuraPolicy != nil && *currentTemplate.UseCustomHasuraPolicy {
			return true
		}
	}
	return false
}

func dropHasuraSecurityFilter(stage FileStage, currentTemplates []TemplateStage, ha *Hasura) error {
	if isCustomPolicyHasura(currentTemplates) {
		for _, currentTemplate := range currentTemplates {
			err := ha.dropCustomRelationship(stage.Service, *currentTemplate.HasuraPolicy, stage.TableName)
			if err != nil {
				return err
			}
		}
	}

	return dropRelationship(currentTemplates, ha, stage.Service, stage.TableName)
}

func ignoreTable(tableName string) bool {
	if table == "" || table == tableName {
		return false
	}

	return true
}

func (s *FileTemplate) rollbackRLSFiles() error {
	_, stages, err := s.getLatestStage()
	if err != nil {
		return err
	}

	pgRollbackTables := map[string]string{}
	pg := &Postgres{
		IOUtils: s.IOUtils,
	}
	ha := &Hasura{
		IOUtils: s.IOUtils,
	}

	for _, stage := range stages {
		if ignoreFile(stage.Service) || ignoreTable(stage.TableName) {
			continue
		}
		if _, ok := pgRollbackTables[stage.Service]; !ok {
			pgRollbackTables[stage.Service] = ""
		}
		dropPolicyForTable := ""
		for _, currentTemplate := range stage.TemplateStage {
			dropPolicyForTemplate := genDropPGPolicyForTemplate(currentTemplate, stage.TableName)
			dropPolicyForTable += dropPolicyForTemplate

			if currentTemplate.Hasura == nil {
				continue
			}
		}
		pgRollbackTable, err := genPostgresRollbackFor(stage.TableName, dropPolicyForTable)
		if err != nil {
			return err
		}
		pgRollbackTables[stage.Service] += pgRollbackTable

		err = dropHasuraSecurityFilter(stage, stage.TemplateStage, ha)
		if err != nil {
			return err
		}
	}
	for svc := range pgRollbackTables {
		if ignoreFile(svc) {
			continue
		}
		err := dropUnusedPolicy(pgRollbackTables[svc], pg, svc)
		if err != nil {
			return err
		}
	}

	return nil
}
