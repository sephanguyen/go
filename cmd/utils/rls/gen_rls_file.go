package rls

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	firstLvQuery = ""
)

type FileTemplate struct {
	IOUtils interface {
		GetFileNamesOnDir(filename string) ([]string, error)
		WriteStringFile(filename string, content string) error
		GetFoldersOnDir(folder string) ([]string, error)
		GetFileContent(filepath string) ([]byte, error)

		WriteFile(filename string, content []byte) error
		AppendStrToFile(filePath string, content string) error

		Copy(src, dst string) (int64, error)
	}
}

type ACFileInfo struct {
	filepath string
	dirname  string
}

const accessCtrFolder = "accesscontrol"

func (s *FileTemplate) getFileACBy(dir string) ([]ACFileInfo, error) {
	folders, err := s.IOUtils.GetFoldersOnDir(dir)
	if err != nil {
		return nil, err
	}

	files := []ACFileInfo{}
	for _, folder := range folders {
		dirFolder := fmt.Sprintf("%s/%s", dir, folder)

		fileNamesOnFolder, err := s.IOUtils.GetFileNamesOnDir(dirFolder)
		if err != nil {
			return nil, err
		}
		for _, fileName := range fileNamesOnFolder {
			if strings.HasPrefix(fileName, "_") {
				continue
			}
			fileDir := fmt.Sprintf("%s/%s", dirFolder, fileName)
			files = append(files, ACFileInfo{filepath: fileDir, dirname: folder})
		}
	}
	return files, nil
}

func (s *FileTemplate) getFilesContent(files []ACFileInfo) ([]TemplateFile, error) {
	templates := []TemplateFile{}
	for _, file := range files {
		content, err := s.IOUtils.GetFileContent(file.filepath)
		if err != nil {
			return nil, err
		}

		data := []Template{}
		err = yaml.Unmarshal(content, &data)
		if err != nil {
			return nil, err
		}
		template := TemplateFile{
			FileDir:      file.filepath,
			DatabaseName: file.dirname,
			Templates:    &data,
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func getMappingColumns(template Template) (string, string, string) {
	accessPathTable := template.AccessPathTable

	if template.Template == "4" {
		return "", *template.OwnerCol, ""
	}

	if accessPathTable == nil || accessPathTable.ColumnMapping == nil {
		return "", "location_id", ""
	}

	colMapping := *accessPathTable.ColumnMapping
	relationshipTable := accessPathTable.Name
	key := ""
	relationshipKey := ""

	for k, v := range colMapping {
		key = k
		relationshipKey = v
	}
	return relationshipTable, key, relationshipKey
}

func returnEmptyIfNull(data *string) string {
	if data == nil {
		return ""
	}
	return *data
}

func resetCommandVariable() {
	table = ""
	accessPathTable, pkey, accessPathTableKey = "", "", ""
	databaseName = ""
	permissionPrefix = ""
	templateVersion = ""
	ownerCol = ""
	accessPathLocationCol = ""
}

func setCommandVariable(template Template, service string) {
	table = template.TableName
	accessPathTable, pkey, accessPathTableKey = getMappingColumns(template)
	databaseName = service
	permissionPrefix = returnEmptyIfNull(template.PermissionPrefix)
	templateVersion = template.Template
	ownerCol = returnEmptyIfNull(template.OwnerCol)
	accessPathLocationCol = returnEmptyIfNull(template.LocationCol)
	addRLSToAllPermissionHasura = true
	mapHasuraDirectly = true
}

func dropUnusedPolicy(dropPolicy string, pg *Postgres, databaseName string) error {
	if dropPolicy != "" {
		newMigrateFile, err := pg.getNewMigrateFile(databaseName)
		if err != nil {
			return fmt.Errorf("getNewMigrateFile error %w", err)
		}

		fmt.Println(Warn("drop old policy: ", newMigrateFile, ":", strings.Trim(dropPolicy, "\n")))

		return pg.IOUtils.WriteStringFile(newMigrateFile, dropPolicy)
	}
	return nil
}

type HasuraRemoveStep struct {
	HasuraState HasuraTemplateStage
	TableName   string
	TemplateAccessPathTable
}

func dropRelationship(dropTemplateStages []TemplateStage, hasura *Hasura, svc string, tableName string) error {
	numberOfTemplate := len(dropTemplateStages)
	if numberOfTemplate > 0 {
		for _, dropTemplateStage := range dropTemplateStages {
			err := hasura.dropRelationship(svc, dropTemplateStage, tableName, numberOfTemplate)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getTableNameOnStages(templateFile TemplateFile) string {
	tableName := ""
	for _, template := range *templateFile.Templates {
		tableName = template.TableName
		break
	}
	return tableName
}

func detectAndRemoveUnusedPolicy(fileStages map[string]FileStage, pg *Postgres, ha *Hasura, templateFile TemplateFile) (FileStage, int, map[string]int, error) {
	if fileStages == nil {
		return FileStage{}, 0, nil, nil
	}

	tableName := getTableNameOnStages(templateFile)
	keyStage := templateFile.DatabaseName + ":" + tableName
	currentStage := fileStages[keyStage]

	revision := currentStage.Revision
	dropPolicyPostgres, dropPoliciesHasura, changed := compareStageWithSameTemplate(templateFile, currentStage)
	err := dropUnusedPolicy(dropPolicyPostgres, pg, templateFile.DatabaseName)
	if err != nil {
		return currentStage, 0, nil, err
	}
	err = dropHasuraSecurityFilter(fileStages[keyStage], dropPoliciesHasura, ha)
	if err != nil {
		return currentStage, 0, nil, err
	}
	return fileStages[keyStage], revision, changed, nil
}

func setBuiltTemplateToStage(templateStage *TemplateStage, template Template) {
	if template.AccessPathTable != nil {
		templateStage.AccessPathTable = &TemplateStageAccessPathTable{
			Name:          template.AccessPathTable.Name,
			ColumnMapping: template.AccessPathTable.ColumnMapping,
		}
	}
	templateStage.PermissionPrefix = template.PermissionPrefix
	templateStage.LocationCol = template.LocationCol
	templateStage.OwnerCol = template.OwnerCol
	templateStage.UseCustomPolicy = template.UseCustomPolicy
	templateStage.UseCustomHasuraPolicy = template.UseCustomHasuraPolicy
	templateStage.HasuraPolicy = template.HasuraPolicy
	templateStage.PostgresPolicy = template.PostgresPolicy
	templateStage.PostgresPolicyVersion = template.PostgresPolicyVersion
	if template.Permissions != nil {
		templateStage.Permissions = &TemplateStagePermission{
			Hasura:   template.Permissions.Hasura,
			Postgres: template.Permissions.Postgres,
		}
	}
}

func isGenPg(templateChanged int) bool {
	return templateChanged == 5 || templateChanged == 4
}

func isGenHasura(templateChanged int) bool {
	return templateChanged == 5 || templateChanged == 1
}

func isCustom(template Template) bool {
	return template.UseCustomPolicy != nil && *template.UseCustomPolicy
}

func isCustomHasura(template Template) bool {
	return template.UseCustomHasuraPolicy != nil && *template.UseCustomHasuraPolicy
}

func isGenCustomHasura(templateChanged int, template Template) bool {
	return isGenHasura(templateChanged) && isCustomHasura(template)
}

func isGenCustomPostgres(templateChanged int, template Template) bool {
	return isGenPg(templateChanged) && isCustom(template)
}

func buildPolicies(templateChanged int, template Template, templateStage *TemplateStage, pg *Postgres, hasura *Hasura) error {
	if isGenCustomPostgres(templateChanged, template) {
		err := pg.genCustomPolicy(template.PostgresPolicy, template.TableName)
		if err != nil {
			return fmt.Errorf("genCustomPolicy - %v", err)
		}
	}
	if isGenCustomHasura(templateChanged, template) {
		hasuraState, err := hasura.getCustomPolicy(template.HasuraPolicy)
		if err != nil {
			return fmt.Errorf("genRLSMetadata - %v", err)
		}
		templateStage.Hasura = hasuraState
		firstLvQuery = hasuraState.FirstLvQuery
	}
	if isGenPg(templateChanged) && !isCustom(template) && template.Permissions != nil && template.Permissions.Postgres != nil {
		postgresStage, err := pg.genPostgresRLS()
		if err != nil {
			return fmt.Errorf("genPostgresRLS - %v", err)
		}
		templateStage.Postgres = postgresStage
	}
	if isGenHasura(templateChanged) && !isCustomHasura(template) && template.Permissions != nil && template.Permissions.Hasura != nil {
		hasuraState, err := hasura.genRLSMetadata()
		if err != nil {
			return fmt.Errorf("genRLSMetadata - %v", err)
		}
		templateStage.Hasura = hasuraState
		firstLvQuery = hasuraState.FirstLvQuery
	}
	return nil
}

func getCurrentTemplateStage(currentStage FileStage, template string) TemplateStage {
	stage := TemplateStage{}
	for _, tempStage := range currentStage.TemplateStage {
		if tempStage.Template == template {
			stage = tempStage
		}
	}
	return stage
}

func (s *FileTemplate) runCommand(templateFile TemplateFile, fileStages map[string]FileStage) ([]FileStage, error) {
	fmt.Println(Info("----"))
	fmt.Println(Info("Running on service: ", templateFile.DatabaseName))
	fmt.Println(Info("Running on file: ", templateFile.FileDir))
	firstLvQuery = ""
	stages := []FileStage{}
	pg := &Postgres{
		IOUtils: s.IOUtils,
	}
	hasura := &Hasura{
		IOUtils: s.IOUtils,
	}

	currentStage, revision, fileChanged, err := detectAndRemoveUnusedPolicy(fileStages, pg, hasura, templateFile)
	if err != nil {
		return nil, err
	}

	newStage := s.newStage(templateFile.FileDir, templateFile.DatabaseName, revision)
	newStage.TemplateStage = []TemplateStage{}

	for _, template := range *templateFile.Templates {
		fmt.Println(Info("Generating template version: ", template.Template))
		templateChanged := 5
		if fileStages != nil {
			templateChanged = fileChanged[template.Template]
		}
		fmt.Println(Info("next template changed: ", templateChanged))

		newStage.TableName = template.TableName
		stage := getCurrentTemplateStage(currentStage, template.Template)
		if templateChanged == 0 {
			if stage.Hasura != nil {
				firstLvQuery = stage.Hasura.FirstLvQuery
			}
			newStage.CreatedAt = currentStage.CreatedAt
			newStage.UpdatedAt = currentStage.UpdatedAt
			newStage.Revision = currentStage.Revision
			newStage.TemplateStage = append(newStage.TemplateStage, stage)
			continue
		}

		resetCommandVariable()
		setCommandVariable(template, templateFile.DatabaseName)

		templateStage := TemplateStage{}
		templateStage.Template = template.Template
		templateStage.Postgres = nil
		templateStage.Hasura = nil
		if template.Permissions != nil && template.Permissions.Postgres != nil {
			templateStage.Postgres = stage.Postgres
		}
		if template.Permissions != nil && template.Permissions.Hasura != nil {
			templateStage.Hasura = stage.Hasura
		}

		newStage.CreatedAt = time.Now()
		newStage.UpdatedAt = time.Now()

		err = buildPolicies(templateChanged, template, &templateStage, pg, hasura)
		if err != nil {
			return nil, err
		}
		setBuiltTemplateToStage(&templateStage, template)
		newStage.TemplateStage = append(newStage.TemplateStage, templateStage)
	}
	stages = append(stages, newStage)

	return stages, nil
}

func (s *FileTemplate) createStgMetadataTables(dir string) error {
	if !stgHasura {
		return nil
	}

	folders, err := s.IOUtils.GetFoldersOnDir(dir)
	if err != nil {
		return err
	}
	for _, folder := range folders {
		if ignoreFile(folder) {
			continue
		}
		destDir := fmt.Sprintf(hasuraTemplateStgTableMetadataPath, folder)
		srcDir := fmt.Sprintf(hasuraTemplateTableMetadataPath, folder)

		_, err := s.IOUtils.Copy(srcDir, destDir)
		if err != nil {
			fmt.Println(Fata("error when copy file %w", err))
			return err
		}
	}
	return nil
}

func ignoreFile(svc string) bool {
	if acFolder == "" || acFolder == svc {
		return false
	}
	return true
}

func mergeLastStageAndCurrentStage(lastStages []FileStage, currentStage []FileStage) []FileStage {
	if acFolder == "" {
		return currentStage
	}
	rsStage := []FileStage{}

	for _, stage := range lastStages {
		if ignoreFile(stage.Service) {
			rsStage = append(rsStage, stage)
		}
	}
	rsStage = append(rsStage, currentStage...)
	return rsStage
}

func (s *FileTemplate) genFromFile() error {
	err := s.createStgMetadataTables(accessCtrFolder)
	if err != nil {
		return err
	}

	fileDirs, err := s.getFileACBy(accessCtrFolder)
	if err != nil {
		return err
	}

	filesContent, err := s.getFilesContent(fileDirs)
	if err != nil {
		return err
	}

	tablesStages := []FileStage{}
	latestStages, rawLatestStages, err := s.getLatestStage()
	if err != nil {
		return err
	}
	for _, fileContent := range filesContent {
		if ignoreFile(fileContent.DatabaseName) {
			continue
		}
		tableStages, err := s.runCommand(fileContent, latestStages)
		if err != nil {
			fmt.Println(Fata("got error genFromFile : ", err))
			return err
		}

		tablesStages = append(tablesStages, tableStages...)
	}

	err = s.writeFileStage(mergeLastStageAndCurrentStage(rawLatestStages, tablesStages))

	return err
}
