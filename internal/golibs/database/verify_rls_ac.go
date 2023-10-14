package database

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/cmd/utils/rls"
	fileio "github.com/manabie-com/backend/internal/golibs/io"

	"go.uber.org/multierr"
	"gopkg.in/yaml.v2"
)

const fileACStageDir = "../../../accesscontrol/stage.json"

func buildKeyStages(svc, tableName string) string {
	return svc + "-" + tableName
}

func loadACStagesBy(fileDir string) (map[string][]string, map[string]rls.FileStage, error) {
	var fileStages []rls.FileStage

	err := loadJSON(fileDir, &fileStages)
	if err != nil {
		return nil, nil, err
	}

	stageObj := map[string]rls.FileStage{}
	svc := map[string][]string{}
	for _, stage := range fileStages {
		stageObj[buildKeyStages(stage.Service, stage.TableName)] = stage
		if svc[stage.Service] != nil {
			svc[stage.Service] = append(svc[stage.Service], stage.TableName)
		} else {
			svc[stage.Service] = []string{stage.TableName}
		}
	}

	return svc, stageObj, nil
}

func groupACByServiceAndTable(fileACStageDir string) (map[string]map[string]bool, error) {
	var fileStages []rls.FileStage

	err := loadJSON(fileACStageDir, &fileStages)
	if err != nil {
		return nil, err
	}

	acGrouped := make(map[string]map[string]bool)
	for _, stage := range fileStages {
		if _, ok := acGrouped[stage.Service]; !ok {
			acGrouped[stage.Service] = make(map[string]bool)
		}
		acGrouped[stage.Service][stage.TableName] = true
	}

	return acGrouped, nil
}

func loadIgnoreACTablesBy(fileDir string) (map[string]map[string]bool, error) {
	var fileStages []rls.FileStage

	err := loadJSON(fileDir, &fileStages)
	if err != nil {
		return nil, err
	}
	ignoreTableMap := make(map[string]map[string]bool)

	for _, stage := range fileStages {
		_, ok := ignoreTableMap[stage.Service]
		if !ok {
			ignoreTableMap[stage.Service] = make(map[string]bool)
		}
		ignoreTableMap[stage.Service][stage.TableName] = true
	}

	return ignoreTableMap, nil
}

func loadIgnoreACTables() (map[string]map[string]bool, error) {
	return loadIgnoreACTablesBy(fileACStageDir)
}

func returnFailIsEmpty(rs *bool) bool {
	if rs == nil {
		return false
	}
	return *rs
}

var (
	RESTRICTIVE = "RESTRICTIVE"
	PERMISSIVE  = "PERMISSIVE"
)

type PostgresPolicyGetter struct {
	Temp        rls.TemplateStage
	PolicyNames map[string]bool
	Policies    map[string]rls.PostgresPolicyStage
}

func newPostgresPolicyGetter(temp rls.TemplateStage, policyNames map[string]bool, policies map[string]rls.PostgresPolicyStage) PostgresPolicyGetter {
	return PostgresPolicyGetter{Temp: temp, PolicyNames: policyNames, Policies: policies}
}

func (p PostgresPolicyGetter) isValidPolicy() bool {
	return p.Temp.Postgres != nil
}

func (p *PostgresPolicyGetter) setPostgresPolicy() {
	if !p.isValidPolicy() {
		return
	}
	for _, policy := range p.Temp.Postgres.Policies {
		p.PolicyNames[policy.Name] = false
		p.Policies[policy.Name] = policy
	}
}

func (p PostgresPolicyGetter) isValidCustomPolicy() bool {
	return returnFailIsEmpty(p.Temp.UseCustomPolicy) && p.Temp.PostgresPolicy != nil
}

func (p *PostgresPolicyGetter) setPostgresCustomPolicy() {
	if !p.isValidCustomPolicy() {
		return
	}

	for _, policy := range *p.Temp.PostgresPolicy {
		p.PolicyNames[policy.Name] = false
		p.Policies[policy.Name] = rls.PostgresPolicyStage{Name: policy.Name, Content: policy.Using + policy.WithCheck}
	}
}

func getAllPoliciesBy(templateStages []rls.TemplateStage) (map[string]bool, map[string]rls.PostgresPolicyStage) {
	objPoliciesNames := map[string]bool{}
	objPolicies := map[string]rls.PostgresPolicyStage{}
	for _, temp := range templateStages {
		postgresPolicyGetter := newPostgresPolicyGetter(temp, objPoliciesNames, objPolicies)
		postgresPolicyGetter.setPostgresPolicy()
		postgresPolicyGetter.setPostgresCustomPolicy()
		objPoliciesNames = postgresPolicyGetter.PolicyNames
		objPolicies = postgresPolicyGetter.Policies
	}
	return objPoliciesNames, objPolicies
}

type PostgresAccessControlVerifier struct {
	TablePolicy      *tablePolicy
	StagePolicies    map[string]rls.PostgresPolicyStage
	StagePolicyNames map[string]bool
	SV               string
	RestrictiveFlag  bool
	PermissiveFlag   bool
	RestrictiveNo    int
	PermissiveNo     int
}

func newAccessControlVerifier(table *tablePolicy, stagePolicyNames map[string]bool, stagePolices map[string]rls.PostgresPolicyStage, sv string) PostgresAccessControlVerifier {
	return PostgresAccessControlVerifier{TablePolicy: table, StagePolicyNames: stagePolicyNames, StagePolicies: stagePolices, SV: sv, RestrictiveFlag: false, RestrictiveNo: 0, PermissiveNo: 0}
}

func (p PostgresAccessControlVerifier) policyName() string {
	return p.TablePolicy.PolicyName.String
}

func (p *PostgresAccessControlVerifier) setPolicyNameExisted() {
	if !p.StagePolicyNames[p.policyName()] {
		p.StagePolicyNames[p.policyName()] = true
	}
}

func (p PostgresAccessControlVerifier) verifyEnabledRLS() error {
	if !p.TablePolicy.Relforcerowsecurity.Bool {
		return fmt.Errorf("please force row level security for table %s in service %s", p.TablePolicy.Name.String, p.SV)
	}
	if !p.TablePolicy.RelrowSecurity.Bool {
		return fmt.Errorf("row security is not enable for table %s in service %s", p.TablePolicy.Name.String, p.SV)
	}
	if !isGrantedToPublic(p.TablePolicy) {
		return fmt.Errorf("policy for table %s in service %s is not granted to public", p.TablePolicy.Name.String, p.SV)
	}
	return nil
}

func (p PostgresAccessControlVerifier) restrictivePolicyName() string {
	return "rls_" + p.TablePolicy.Name.String + "_restrictive"
}

func (p *PostgresAccessControlVerifier) verifyRestrictiveMultiTenant() error {
	restrictiveMultiTenantPolicy := p.restrictivePolicyName()

	if p.policyName() != restrictiveMultiTenantPolicy {
		return nil
	}
	if p.TablePolicy.Permissive.String != RESTRICTIVE {
		return fmt.Errorf("policy %s on table %s in service %s must be restrictive policy", restrictiveMultiTenantPolicy, p.TablePolicy.Name.String, p.SV)
	}
	if p.TablePolicy.Qual.String != fmt.Sprintf("permission_check(resource_path, '%s'::text)", p.TablePolicy.Name.String) {
		return fmt.Errorf("function permission_check is not in policy for table %s in service %s. Please change to permission_check(resource_path, '%s'::text)", p.TablePolicy.Name.String, p.SV, p.TablePolicy.Name.String)
	}
	if p.TablePolicy.WithCheck.String != fmt.Sprintf("permission_check(resource_path, '%s'::text)", p.TablePolicy.Name.String) {
		return fmt.Errorf("with_check in policy does not use function permission_check for table %s in service %s. Please use with_check in policy with permission_check(resource_path, '%s'::text)", p.TablePolicy.Name.String, p.SV, p.TablePolicy.Name.String)
	}

	return nil
}

func (p *PostgresAccessControlVerifier) incrRestrictiveCount() {
	if p.TablePolicy.Permissive.String != RESTRICTIVE {
		return
	}
	p.RestrictiveFlag = true
	p.RestrictiveNo++
}

func (p *PostgresAccessControlVerifier) incrPermissiveCount() {
	if p.TablePolicy.Permissive.String != PERMISSIVE {
		return
	}
	p.PermissiveFlag = true
	p.PermissiveNo++
}

func correctACPolicyStr(str string) string {
	str = strings.ReplaceAll(str, "::text", "")
	str = strings.ToLower(str)
	str = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(str, "")
	str = strings.ReplaceAll(str, "trueasbool", "true")
	return str
}

func (p *PostgresAccessControlVerifier) verifyPostgresContentPolicy() error {
	if p.policyName() == p.restrictivePolicyName() {
		return nil
	}
	if _, ok := p.StagePolicyNames[p.policyName()]; !ok {
		return fmt.Errorf("policy %s on table %s in service %s is unexpected policy", p.policyName(), p.TablePolicy.Name.String, p.SV)
	}

	policy := p.StagePolicies[p.policyName()]

	if !strings.Contains(correctACPolicyStr(policy.Content), correctACPolicyStr(p.TablePolicy.Qual.String)) {
		return fmt.Errorf("policy %s on table %s in service %s have content using is not correct", p.policyName(), p.TablePolicy.Name.String, p.SV)
	}

	if !strings.Contains(correctACPolicyStr(policy.Content), correctACPolicyStr(p.TablePolicy.WithCheck.String)) {
		return fmt.Errorf("policy %s on table %s in service %s have content with check is not correct", p.policyName(), p.TablePolicy.Name.String, p.SV)
	}

	return nil
}

func isIgnoreChecKPostgresRLS(tableStage rls.FileStage) bool {
	isIgnore := true
	for _, stage := range tableStage.TemplateStage {
		if stage.Permissions != nil && stage.Permissions.Postgres != nil {
			isIgnore = false
		}
		if stage.PostgresPolicy != nil {
			isIgnore = false
		}
	}
	return isIgnore
}

func (p *PostgresAccessControlVerifier) run() error {
	if err := p.verifyEnabledRLS(); err != nil {
		return err
	}

	if err := p.verifyRestrictiveMultiTenant(); err != nil {
		return err
	}

	if err := p.verifyPostgresContentPolicy(); err != nil {
		return err
	}

	p.setPolicyNameExisted()
	p.incrRestrictiveCount()
	p.incrPermissiveCount()
	return nil
}

func VerifyPostgresRls(sv string, tblSchema *tableSchema, tableStage rls.FileStage) error {
	permissiveFlag := false
	restrictiveFlag := false
	restrictiveNo := 0
	permissiveNo := 0

	if isIgnoreChecKPostgresRLS(tableStage) {
		return nil
	}

	stagePolicyNames, stagePolicies := getAllPoliciesBy(tableStage.TemplateStage)
	for _, p := range tblSchema.Policies {
		verifier := newAccessControlVerifier(p, stagePolicyNames, stagePolicies, sv)
		err := verifier.run()
		if err != nil {
			return err
		}

		if !restrictiveFlag {
			restrictiveFlag = verifier.RestrictiveFlag
		}
		if !permissiveFlag {
			permissiveFlag = verifier.PermissiveFlag
		}

		restrictiveNo += verifier.RestrictiveNo
		permissiveNo += verifier.PermissiveNo
		stagePolicies = verifier.StagePolicies
		stagePolicyNames = verifier.StagePolicyNames
	}

	if !permissiveFlag {
		return fmt.Errorf("table %s in service %s missing permissive rls policy", tblSchema.TableName, sv)
	}
	if !restrictiveFlag {
		return fmt.Errorf("table %s in service %s missing restrictive rls policy", tblSchema.TableName, sv)
	}
	if restrictiveNo != 1 {
		return fmt.Errorf("table %s in service %s must have only one restrictive policy instead of %d policies", tblSchema.TableName, sv, restrictiveNo)
	}
	if permissiveNo == 0 {
		return fmt.Errorf("table %s in service %s must have at least one permissive policy", tblSchema.TableName, sv)
	}
	for k := range stagePolicies {
		if !stagePolicyNames[k] {
			return fmt.Errorf("table %s in service %s missing %s policy", tblSchema.TableName, sv, k)
		}
	}

	return nil
}

func findTableSchema(tableName string, files []string) string {
	for _, f := range files {
		if strings.Contains(f, "/"+tableName+".json") {
			return f
		}
	}
	return ""
}

func loadStageFile() (map[string][]string, map[string]rls.FileStage, error) {
	existedSvc, stages, err := loadACStagesBy(fileACStageDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load AC stages %w", err)
	}
	return existedSvc, stages, err
}

func VerifyPostgresACInSv(s string, existedSvc map[string][]string, files []string, stages map[string]rls.FileStage) error {
	tableNames := existedSvc[s]
	for _, tableName := range tableNames {
		f := findTableSchema(tableName, files)
		if f == "" {
			return fmt.Errorf("failed to read json file %s", tableName)
		}

		tblSchema := &tableSchema{}

		err := loadJSON(f, tblSchema)
		if err != nil {
			return fmt.Errorf("failed to read json file %s: %s", f, err)
		}

		err = VerifyPostgresRls(s, tblSchema, stages[buildKeyStages(s, tableName)])
		if err != nil {
			return err
		}
	}
	return nil
}

func filterFileByService(files []string, svc string) []string {
	filteredFiles := []string{}
	for _, f := range files {
		if strings.Contains(f, fmt.Sprintf("/%s/", svc)) {
			filteredFiles = append(filteredFiles, f)
		}
	}
	return filteredFiles
}

func VerifyPostgresAC() error {
	svc, err := getDirectories(snapshotDir)
	if err != nil {
		return fmt.Errorf("failed to get sub directory in %s: %w", snapshotDir, err)
	}

	existedSvc, stages, err := loadStageFile()
	if err != nil {
		return err
	}

	for _, s := range svc {
		if existedSvc[s] != nil && len(existedSvc[s]) == 0 {
			continue
		}

		files, err := getFilesInDirectory(filepath.Join(snapshotDir, s), ".json")
		if err != nil {
			return fmt.Errorf("failed to read json file in directory %s: %w", snapshotDir, err)
		}
		svcFiles := filterFileByService(files, s)

		err = VerifyPostgresACInSv(s, existedSvc, svcFiles, stages)
		if err != nil {
			return err
		}
	}
	return nil
}

type HasuraAccessControlVerifier struct {
	Metadata rls.HasuraTable
	Stage    rls.FileStage

	SelectPermissionMap map[string][]HasuraPolicyPermission
	InsertPermissionMap map[string][]HasuraPolicyPermission
	UpdatePermissionMap map[string][]HasuraPolicyPermission
	DeletePermissionMap map[string][]HasuraPolicyPermission
}

func newHasuraAccessControlVerifier(metadata rls.HasuraTable, stage rls.FileStage) HasuraAccessControlVerifier {
	selectPermissionMap := map[string][]HasuraPolicyPermission{}
	insertPermissionMap := map[string][]HasuraPolicyPermission{}
	updatePermissionMap := map[string][]HasuraPolicyPermission{}
	deletePermissionMap := map[string][]HasuraPolicyPermission{}

	return HasuraAccessControlVerifier{Metadata: metadata, Stage: stage, SelectPermissionMap: selectPermissionMap, InsertPermissionMap: insertPermissionMap, UpdatePermissionMap: updatePermissionMap, DeletePermissionMap: deletePermissionMap}
}

func (h *HasuraAccessControlVerifier) tableName() string {
	return h.Stage.TableName
}

type HasuraPolicyPermission struct {
	Filter *interface{}
	Check  *interface{}
}

func buildPermissionMap(permissions *[]rls.TemplateHasuraRole, permissionMap map[string][]HasuraPolicyPermission) {
	if permissions != nil {
		for _, permission := range *permissions {
			permissionMap[permission.Name] = append(permissionMap[permission.Name], HasuraPolicyPermission{Filter: permission.Filter, Check: permission.Check})
		}
	}
}

func (h *HasuraAccessControlVerifier) mapPermissions() {
	for _, stage := range h.Stage.TemplateStage {
		if stage.Hasura == nil || stage.Hasura.HasuraPolicy == nil {
			continue
		}
		buildPermissionMap(stage.Hasura.HasuraPolicy.SelectPermission, h.SelectPermissionMap)
		buildPermissionMap(stage.Hasura.HasuraPolicy.InsertPermission, h.InsertPermissionMap)
		buildPermissionMap(stage.Hasura.HasuraPolicy.UpdatePermission, h.UpdatePermissionMap)
		buildPermissionMap(stage.Hasura.HasuraPolicy.DeletePermission, h.DeletePermissionMap)
	}
}

func getStrFromYMLFilter(filter interface{}) (string, error) {
	if filter == nil {
		return "", nil
	}
	b, err := yaml.Marshal(filter)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getStrFromJSONFilter(filter interface{}) (string, error) {
	if filter == nil {
		return "", nil
	}
	b, err := json.Marshal(filter)
	if err != nil {
		return "", err
	}
	if string(b) == "null" {
		return "", nil
	}
	return string(b), nil
}

func (h *HasuraAccessControlVerifier) permissionMapBy(command string) map[string][]HasuraPolicyPermission {
	permissions := map[string]map[string][]HasuraPolicyPermission{
		"select": h.SelectPermissionMap,
		"insert": h.InsertPermissionMap,
		"update": h.UpdatePermissionMap,
		"delete": h.DeletePermissionMap,
	}
	return permissions[command]
}

func (h *HasuraAccessControlVerifier) comparePermission(role string, filter interface{}, withCheck interface{}, command string) error {
	stages, ok := h.permissionMapBy(command)[role]
	if !ok {
		fmt.Printf("role (%s) %s is not tracked in table %s \n", command, role, h.tableName())
		return nil
	}

	for _, stage := range stages {
		filterStr, _ := getStrFromYMLFilter(filter)
		filterStageStr, _ := getStrFromJSONFilter(stage.Filter)

		if !strings.Contains(correctACPolicyStr(filterStr), correctACPolicyStr(filterStageStr)) {
			return fmt.Errorf("role (%s) %s on table %s have content of filter is not correct", command, role, h.tableName())
		}

		checkStr, _ := getStrFromYMLFilter(withCheck)
		checkStageStr, _ := getStrFromJSONFilter(stage.Check)
		if !strings.Contains(correctACPolicyStr(checkStr), correctACPolicyStr(checkStageStr)) {
			return fmt.Errorf("role (%s) %s on table %s have content of check is not correct", command, role, h.tableName())
		}
	}

	return nil
}

var (
	selectCommand = "select"
	insertCommand = "insert"
	updateCommand = "update"
	deleteCommand = "delete"
)

func validState(stage map[string][]HasuraPolicyPermission) bool {
	return len(stage) > 0
}

func (h *HasuraAccessControlVerifier) verifySelectPermissions() error {
	stage := h.permissionMapBy(selectCommand)
	selectPermission := h.Metadata.SelectPermissions

	if selectPermission == nil && validState(stage) {
		return fmt.Errorf("role (%s) on table %s missing select permission", selectCommand, h.tableName())
	}

	if selectPermission == nil {
		return nil
	}

	rolesMap := h.getRolesFromStage(selectCommand)

	for _, s := range *selectPermission {
		if err := h.comparePermission(s.Role, s.Permission.Filter, nil, selectCommand); err != nil {
			return err
		}
		rolesMap[s.Role] = true
	}
	return h.checkNotCompareRole(rolesMap)
}

func (h *HasuraAccessControlVerifier) verifyInsertPermissions() error {
	stage := h.permissionMapBy(insertCommand)
	insertPermission := h.Metadata.InsertPermissions
	if insertPermission == nil && validState(stage) {
		return fmt.Errorf("role (%s) on table %s missing insert permission", insertCommand, h.tableName())
	}

	if insertPermission == nil {
		return nil
	}
	rolesMap := h.getRolesFromStage(insertCommand)

	for _, s := range *insertPermission {
		if err := h.comparePermission(s.Role, nil, s.Permission.Check, insertCommand); err != nil {
			return err
		}
		rolesMap[s.Role] = true
	}
	return h.checkNotCompareRole(rolesMap)
}

func (h *HasuraAccessControlVerifier) verifyUpdatePermissions() error {
	stage := h.permissionMapBy(updateCommand)
	updatePermission := h.Metadata.UpdatePermissions

	if updatePermission == nil && validState(stage) {
		return fmt.Errorf("role (%s) on table %s missing update permission", updateCommand, h.tableName())
	}

	if updatePermission == nil {
		return nil
	}

	rolesMap := h.getRolesFromStage(updateCommand)

	for _, s := range *updatePermission {
		if err := h.comparePermission(s.Role, s.Permission.Filter, s.Permission.Check, updateCommand); err != nil {
			return err
		}
		rolesMap[s.Role] = true
	}
	return h.checkNotCompareRole(rolesMap)
}

func (h *HasuraAccessControlVerifier) getRolesFromStage(command string) map[string]bool {
	stage := h.permissionMapBy(command)
	roleMap := map[string]bool{}
	for k := range stage {
		roleMap[k] = false
	}
	return roleMap
}

func (h *HasuraAccessControlVerifier) checkNotCompareRole(rolesMap map[string]bool) error {
	for k := range rolesMap {
		if !rolesMap[k] {
			return fmt.Errorf("role (%s) on table %s missing permission: %s", deleteCommand, h.tableName(), k)
		}
	}
	return nil
}

func (h *HasuraAccessControlVerifier) verifyDeletePermissions() error {
	stage := h.permissionMapBy(deleteCommand)
	deletePermission := h.Metadata.DeletePermissions

	if deletePermission == nil && validState(stage) {
		return fmt.Errorf("role (%s) on table %s missing permission", deleteCommand, h.tableName())
	}

	if deletePermission == nil {
		return nil
	}

	rolesMap := h.getRolesFromStage(deleteCommand)

	for _, s := range *deletePermission {
		if err := h.comparePermission(s.Role, s.Permission.Filter, s.Permission.Check, deleteCommand); err != nil {
			return err
		}
		rolesMap[s.Role] = true
	}

	return h.checkNotCompareRole(rolesMap)
}

func (h *HasuraAccessControlVerifier) run() error {
	h.mapPermissions()
	return multierr.Combine(h.verifySelectPermissions(),
		h.verifyInsertPermissions(),
		h.verifyUpdatePermissions(),
		h.verifyDeletePermissions())
}

func getHasuraMetadataTables(svc string) (map[string]rls.HasuraTable, error) {
	dir := fmt.Sprintf("../../../deployments/helm/manabie-all-in-one/charts/%s/files/hasura/metadata/tables.yaml", svc)
	fileUtil := fileio.FileUtils{}
	content, err := fileUtil.GetFileContent(dir)
	if err != nil {
		return nil, err
	}

	tableMetadata := []rls.HasuraTable{}

	err = yaml.Unmarshal(content, &tableMetadata)

	if err != nil {
		return nil, err
	}

	tablesMap := map[string]rls.HasuraTable{}

	for _, table := range tableMetadata {
		tablesMap[table.Table.Name] = table
	}

	return tablesMap, nil
}

func verifyTableHasuraAC(stage rls.FileStage, metadata rls.HasuraTable) error {
	verifier := newHasuraAccessControlVerifier(metadata, stage)
	return verifier.run()
}

func checkFileNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such file or directory")
}

func VerifyHasuraAC() error {
	existedSvc, stages, err := loadStageFile()
	if err != nil {
		return err
	}
	for svc := range existedSvc {
		metadataTables, err := getHasuraMetadataTables(svc)
		if checkFileNotFound(err) {
			fmt.Println("Bypass if not found hasura metadata")
			continue
		} else if err != nil {
			return fmt.Errorf("find metadata of service %s not found %w", svc, err)
		}
		for _, table := range existedSvc[svc] {
			stage, okStage := stages[buildKeyStages(svc, table)]
			metadataTable, okMetadata := metadataTables[table]

			if !okMetadata || !okStage {
				continue
			}

			err := verifyTableHasuraAC(stage, metadataTable)

			if err != nil {
				return fmt.Errorf("verify hasura ac fail %v", err)
			}
		}
	}
	return nil
}
