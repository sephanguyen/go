package rls

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const fileStage = "stage.json"

type HasuraTemplateStage struct {
	FileDir      string                `json:"stage_dir"`
	Permissions  []string              `json:"permissions"`
	Relationship string                `json:"relationship"`
	FirstLvQuery string                `json:"first_level_query"`
	HasuraPolicy *TemplateHasuraPolicy `json:"hasura_policies" yaml:"hasuraPolicy"`
}

type PostgresPolicyStage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
type PostgresTemplateStage struct {
	FileDir  string                `json:"stage_dir"`
	Policies []PostgresPolicyStage `json:"policies"`
}

type TemplateStageAccessPathTable struct {
	Name          string             `json:"name"`
	ColumnMapping *map[string]string `json:"columnMapping"`
}

type TemplateStagePermission struct {
	Postgres *[]string `json:"postgres"`
	Hasura   *[]string `json:"hasura"`
}

type TemplateStage struct {
	Template string                 `json:"template"`
	Hasura   *HasuraTemplateStage   `json:"hasura"`
	Postgres *PostgresTemplateStage `json:"postgres"`

	AccessPathTable  *TemplateStageAccessPathTable `json:"accessPathTable"`
	LocationCol      *string                       `json:"locationCol"`
	PermissionPrefix *string                       `json:"permissionPrefix"`
	Permissions      *TemplateStagePermission      `json:"permissions"`
	OwnerCol         *string                       `json:"ownerCol"`

	TemplatesPolicy
}

type FileStage struct {
	FileDir       string          `json:"filename"`
	Service       string          `json:"service"`
	Revision      int             `json:"revision"`
	TableName     string          `json:"table_name"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	TemplateStage []TemplateStage `json:"stages"`
}

func (s *FileTemplate) newStage(filename string, databaseName string, latestRevision int) FileStage {
	state := FileStage{
		FileDir:  filename,
		Service:  databaseName,
		Revision: latestRevision + 1,
	}

	return state
}

func (s *FileTemplate) getLatestStage() (map[string]FileStage, []FileStage, error) {
	if stgHasura {
		return nil, nil, nil
	}

	fileDir := fmt.Sprintf("%s/%s", accessCtrFolder, fileStage)
	fileContentStr, err := s.IOUtils.GetFileContent(fileDir)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	stages := []FileStage{}
	err = json.Unmarshal(fileContentStr, &stages)
	if err != nil {
		return nil, nil, err
	}

	stageObj := map[string]FileStage{}

	for _, stage := range stages {
		keyStage := stage.Service + ":" + stage.TableName
		stageObj[keyStage] = stage
	}

	return stageObj, stages, nil
}

func removeSlice[T any](slice []T, s int) []T {
	if s < 0 {
		return slice
	}
	return append(slice[:s], slice[s+1:]...)
}

func findCurrentStageByTemplateVersion(templateVersion string, templates []TemplateStage) (int, *TemplateStage) {
	for i, v := range templates {
		if v.Template == templateVersion {
			return i, &v
		}
	}
	return -1, nil
}

func flatArrayStrToString(arrayStr *[]string) string {
	if arrayStr == nil {
		return ""
	}
	return strings.Join(*arrayStr, ",") + ","
}

func incr(t int, n int) int {
	if t < 5 {
		return t + n
	}
	return t
}

func convertBoolToStr(val *bool) string {
	if val == nil {
		return ""
	}
	return strconv.FormatBool(*val)
}

func convertPtIntToInt(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

// 1 change hasura
// 4 change postgres
// 5 change all
type ChangeDetecter struct {
	currentTemplate TemplateStage
	nextTemplate    Template
	changed         int
}

func newChangeDetecter(currentTemplate TemplateStage, nextTemplate Template, changed int) *ChangeDetecter {
	return &ChangeDetecter{currentTemplate, nextTemplate, changed}
}

func (ch *ChangeDetecter) checkSameLocationAndPermissionPrefix() {
	if returnEmptyIfNull(ch.currentTemplate.LocationCol) != returnEmptyIfNull(ch.nextTemplate.LocationCol) || returnEmptyIfNull(ch.currentTemplate.PermissionPrefix) != returnEmptyIfNull(ch.nextTemplate.PermissionPrefix) {
		ch.changed = incr(ch.changed, 5)
	}
}
func (ch *ChangeDetecter) checkSameAccessPath() {
	if ch.currentTemplate.AccessPathTable != nil && ch.nextTemplate.Permissions != nil {
		if ch.currentTemplate.AccessPathTable.Name != ch.nextTemplate.AccessPathTable.Name ||
			!reflect.DeepEqual(ch.currentTemplate.AccessPathTable.ColumnMapping, ch.nextTemplate.AccessPathTable.ColumnMapping) {
			ch.changed = incr(ch.changed, 5)
		}
	}
}

func (ch *ChangeDetecter) checkSamePolicyType() {
	if ch.currentTemplate.Permissions != nil && ch.nextTemplate.Permissions != nil {
		if flatArrayStrToString(ch.currentTemplate.Permissions.Hasura) != flatArrayStrToString(ch.nextTemplate.Permissions.Hasura) {
			ch.changed = incr(ch.changed, 1)
		}

		if flatArrayStrToString(ch.currentTemplate.Permissions.Postgres) != flatArrayStrToString(ch.nextTemplate.Permissions.Postgres) {
			ch.changed = incr(ch.changed, 4)
		}
	} else if (ch.currentTemplate.Permissions == nil && ch.nextTemplate.Permissions != nil) || (ch.currentTemplate.Permissions != nil && ch.nextTemplate.Permissions == nil) {
		ch.changed = incr(ch.changed, 5)
	}
}

func (ch *ChangeDetecter) checkChangeCustomHasuraPolicy() {
	if convertBoolToStr(ch.currentTemplate.UseCustomHasuraPolicy) != convertBoolToStr(ch.nextTemplate.UseCustomHasuraPolicy) {
		ch.changed = incr(ch.changed, 1)
	}
}

func (ch *ChangeDetecter) checkChangePostgresPolicyVersion() {
	if convertPtIntToInt(ch.currentTemplate.PostgresPolicyVersion) < convertPtIntToInt(ch.nextTemplate.PostgresPolicyVersion) {
		ch.changed = incr(ch.changed, 4)
	}
}

func detectChange(currentTemplate TemplateStage, nextTemplate Template) int {
	changeDetecter := newChangeDetecter(currentTemplate, nextTemplate, 0)
	changeDetecter.checkSameLocationAndPermissionPrefix()
	changeDetecter.checkSameAccessPath()
	changeDetecter.checkSamePolicyType()
	changeDetecter.checkChangeCustomHasuraPolicy()
	changeDetecter.checkChangePostgresPolicyVersion()
	return changeDetecter.changed
}

func getDropCommandFromDeletedTemplate(currentTemplates []TemplateStage, tableName string) (string, []TemplateStage) {
	dropPolicyPostgres := ""
	dropPoliciesHasura := []TemplateStage{}
	for _, currentTemplate := range currentTemplates {
		if currentTemplate.Postgres != nil {
			dropPolicyPostgres += buildDropPolicy(currentTemplate.Postgres.Policies, tableName) + "\n"
		}
		if currentTemplate.Hasura != nil {
			dropPoliciesHasura = append(dropPoliciesHasura, currentTemplate)
		}
	}
	return dropPolicyPostgres, dropPoliciesHasura
}

func getPolicies(temp TemplateStage) []PostgresPolicyStage {
	policies := []PostgresPolicyStage{}
	if temp.PostgresPolicy != nil && temp.UseCustomPolicy != nil && *temp.UseCustomPolicy {
		for _, policy := range *temp.PostgresPolicy {
			policies = append(policies, PostgresPolicyStage{Name: policy.Name, Content: ""})
		}
	} else if temp.Postgres != nil {
		policies = temp.Postgres.Policies
	}
	return policies
}

func compareStageWithSameTemplate(templateFile TemplateFile, currentStage FileStage) (string, []TemplateStage, map[string]int) {
	currentTemplates := make([]TemplateStage, len(currentStage.TemplateStage))
	copy(currentTemplates, currentStage.TemplateStage)
	nextTemplates := templateFile.Templates
	dropPolicyPostgres := ""
	dropPoliciesHasura := []TemplateStage{}
	templatesChanged := make(map[string]int)
	for _, nextTemplate := range *nextTemplates {
		indexCurrentTemplate, currentTemplate := findCurrentStageByTemplateVersion(nextTemplate.Template, currentTemplates)

		currentTemplates = removeSlice(currentTemplates, indexCurrentTemplate)

		if currentTemplate == nil {
			templatesChanged[nextTemplate.Template] = 5
			continue
		}

		templatesChanged[nextTemplate.Template] = detectChange(*currentTemplate, nextTemplate)

		if templatesChanged[nextTemplate.Template] == 5 || templatesChanged[nextTemplate.Template] == 4 {
			dropPolicyPostgres += buildDropPolicy(getPolicies(*currentTemplate), currentStage.TableName) + "\n"
		}
		if (templatesChanged[nextTemplate.Template] == 5 || templatesChanged[nextTemplate.Template] == 1) && currentTemplate.Hasura != nil {
			dropPoliciesHasura = append(dropPoliciesHasura, *currentTemplate)
		}
	}

	dropStr, dropRelation := getDropCommandFromDeletedTemplate(currentTemplates, currentStage.TableName)
	dropPolicyPostgres += dropStr
	dropPoliciesHasura = append(dropPoliciesHasura, dropRelation...)

	return dropPolicyPostgres, dropPoliciesHasura, templatesChanged
}

func (s *FileTemplate) writeFileStage(tablesStages []FileStage) error {
	if stgHasura {
		return nil
	}

	fileDir := fmt.Sprintf("%s/%s", accessCtrFolder, fileStage)

	data, err := json.MarshalIndent(tablesStages, "", "    ")

	if err != nil {
		return err
	}

	return s.IOUtils.WriteFile(fileDir, data)
}
