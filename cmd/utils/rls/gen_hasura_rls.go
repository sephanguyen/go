package rls

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

const hasuraTemplateTableMetadataPath = "deployments/helm/manabie-all-in-one/charts/%s/files/hasura/metadata/tables.yaml"
const hasuraTemplateStgTableMetadataPath = "deployments/helm/manabie-all-in-one/charts/%s/files/hasura/metadata/tables_stg.yaml"
const hasuraTemplateTableFolderV2 = "deployments/helm/manabie-all-in-one/charts/%s/files/hasurav2/metadata/databases/%s/tables"
const grantedTableName = "granted_permissions"
const defaultRoleName = "MANABIE"
const hasuraTemplateTableFileV2 = "deployments/helm/manabie-all-in-one/charts/%s/files/hasurav2/metadata/databases/%s/tables/public_%s.yaml"
const grantedTableIncluded = `- "!include public_granted_permissions.yaml"`
const hasuraUserIDInput = "X-Hasura-User-Id"
const hasuraResourcePathInput = "X-Hasura-Resource-Path"
const writeSuffix = ".write"

type Hasura struct {
	IOUtils interface {
		GetFileContent(filepath string) ([]byte, error)
		WriteFile(filename string, content []byte) error
		GetFileNamesOnDir(filename string) ([]string, error)
		AppendStrToFile(filePath string, content string) error
		Copy(src, dst string) (int64, error)
	}
}

func getAllSetsOnPermission(insertPermissions []HasuraInsertPermissions) map[string]string {
	setObj := make(map[string]string)
	for _, insertPermission := range insertPermissions {
		if sets := insertPermission.Permission.Set; sets != nil {
			for set, value := range *sets {
				setObj[set] = value
			}
		}
	}
	return setObj
}

func buildSelectPermission(columns []string, filters interface{}, _ []HasuraSelectPermissions) HasuraSelectPermissions {
	return HasuraSelectPermissions{
		Role: defaultRoleName,
		Permission: &HasuraPermission{
			Columns:           columns,
			AllowAggregations: true,
			Filter:            &filters,
		},
	}
}

func buildInsertPermission(columns []string, checks interface{}, insertPermissions []HasuraInsertPermissions) HasuraInsertPermissions {
	sets := getAllSetsOnPermission(insertPermissions)
	permission := &HasuraInsertPermission{
		Columns: columns,
		Check:   &checks,
	}
	if len(sets) > 0 {
		permission = &HasuraInsertPermission{
			Columns: columns,
			Check:   &checks,
			Set:     &sets,
		}
	}
	return HasuraInsertPermissions{
		Role:       defaultRoleName,
		Permission: permission,
	}
}

func buildUpdatePermission(columns []string, checks interface{}, _ []HasuraInsertPermissions) HasuraInsertPermissions {
	return HasuraInsertPermissions{
		Role: defaultRoleName,
		Permission: &HasuraInsertPermission{
			Columns: columns,
			Check:   &checks,
			Filter:  &checks,
		},
	}
}

func buildDeletePermission(_ []string, checks interface{}, _ []HasuraDeletePermissions) HasuraDeletePermissions {
	return HasuraDeletePermissions{
		Role: defaultRoleName,
		Permission: &HasuraDeletePermission{
			Check:  &checks,
			Filter: &checks,
		},
	}
}

func getFieldNamesOfInterface(nestedField interface{}) []string {
	searchFields := []string{}
	searchField, ok := nestedField.(map[interface{}]interface{})
	if ok {
		for k := range searchField {
			searchFields = append(searchFields, k.(string))
		}
	} else {
		strFields, ok := nestedField.(map[string]interface{})
		if ok {
			for strField := range strFields {
				searchFields = append(searchFields, strField)
			}
		}
	}
	return searchFields
}

func getAllKeyAndValueOfFilter(nestedField interface{}) string {
	searchFields := ""
	if searchField, ok := nestedField.(map[string]interface{}); ok {
		for strField := range searchField {
			searchFields += strField
			searchFields += getAllKeyAndValueOfFilter(searchField[strField])
		}
	} else if searchField, ok := nestedField.(map[interface{}]interface{}); ok {
		for k := range searchField {
			searchFields += getAllKeyAndValueOfFilter(searchField[k])
			searchFields += getAllKeyAndValueOfFilter(k)
		}
	} else if searchField, ok := nestedField.([]interface{}); ok {
		for _, data := range searchField {
			searchFields += getAllKeyAndValueOfFilter(data)
		}
	} else if searchField, ok := nestedField.(string); ok {
		searchFields += searchField
	} else if searchField, ok := nestedField.(bool); ok {
		searchFields += strconv.FormatBool(searchField)
	}
	return searchFields
}

func getInterfaceValue(data interface{}, fieldExisted string) interface{} {
	value, ok := data.(map[interface{}]interface{})
	if ok {
		return value[fieldExisted]
	}
	str, ok := data.(map[string]interface{})
	if ok {
		return str[fieldExisted]
	}
	return nil
}

func isGreater(nestedField interface{}, filter interface{}, fieldExisted string) bool {
	existedValue := getInterfaceValue(nestedField, fieldExisted)
	verifyValue := getInterfaceValue(filter, fieldExisted)
	return len(getAllKeyAndValueOfFilter(existedValue)) > len(getAllKeyAndValueOfFilter(verifyValue))
}

func checkQueryExisted(filters []interface{}, nestedField interface{}) bool {
	searchFields := getFieldNamesOfInterface(nestedField)

	if searchFields == nil {
		return false
	}
	field := searchFields[0]

	for i, filter := range filters {
		filterField := getFieldNamesOfInterface(filter)
		for _, fieldExisted := range filterField {
			if fieldExisted == field {
				if isGreater(nestedField, filter, fieldExisted) {
					filters[i] = nestedField
				}
				return true
			}
		}
	}
	return false
}

func getFilterFromOtherTemplate(filters []interface{}, field string) (int, interface{}) {
	for i, filter := range filters {
		filterField := getFieldNamesOfInterface(filter)
		for _, fieldExisted := range filterField {
			if fieldExisted == field {
				return i, filter
			}
		}
	}
	return -1, nil
}

func getLocationPermissionQuery(table string, permissionName string) interface{} {
	eqUserID := map[string]string{"_eq": hasuraUserIDInput}
	eqPermission := map[string]string{"_eq": permissionName}

	userFilter := make([]interface{}, 2)
	userFilter[0] = map[string]map[string]string{"user_id": eqUserID}
	userFilter[1] = map[string]map[string]string{"permission_name": eqPermission}

	andQuery := make(map[string]interface{})
	andQuery["_and"] = userFilter
	permissionLocationQuery := make(map[string]interface{})
	switch templateVersion {
	case "4":
		permissionLocationQuery[ownerCol] = map[string]string{"_eq": hasuraUserIDInput}
	case "3":
		table := map[string]string{
			"schema": "public",
			"name":   grantedTableName,
		}
		existsQuery := map[string]interface{}{
			"_exists": map[string]interface{}{
				"_table": table,
				"_where": andQuery,
			},
		}
		permissionLocationQuery = existsQuery
	default:
		permissionLocationQuery[table+"_location_permission"] = andQuery
	}

	return permissionLocationQuery
}

func getReferID(accessPathCol string) string {
	refID := "location_id"
	switch {
	case accessPathLocationCol != "":
		refID = accessPathLocationCol
	case accessPathCol != "":
		refID = accessPathCol
	case pkey != "" && accessPathCol == "":
		refID = pkey
	}
	return refID
}

func buildGrantedPermission(remoteTable string, accessPathCol string) HasuraUsingObjectRelationships {
	refID := getReferID(accessPathCol)

	return HasuraUsingObjectRelationships{
		ManualConfiguration: &HasuraManualConfiguration{
			RemoteTable: HasuraRemoteTable{
				Schema: "public",
				Name:   remoteTable,
			},
			ColumnMapping: map[string]string{
				"location_id": refID,
			},
		},
	}
}

func buildRelationShip(accessPathCol string) HasuraUsingObjectRelationships {
	refID := getReferID(accessPathCol)

	return HasuraUsingObjectRelationships{
		ManualConfiguration: &HasuraManualConfiguration{
			RemoteTable: HasuraRemoteTable{
				Schema: "public",
				Name:   "granted_permissions",
			},
			ColumnMapping: map[string]string{
				refID: "location_id",
			},
		},
	}
}

func buildManualObjectRelationship(table, tableID, refTableID string) HasuraUsingObjectRelationships {
	return HasuraUsingObjectRelationships{
		ManualConfiguration: &HasuraManualConfiguration{
			RemoteTable: HasuraRemoteTable{
				Schema: "public",
				Name:   table,
			},
			ColumnMapping: map[string]string{
				tableID: refTableID,
			},
		},
	}
}

func buildManualArrRelationship(table, tableID, refTableID string) HasuraUsing {
	return HasuraUsing{
		ManualConfiguration: &HasuraManualConfiguration{
			RemoteTable: HasuraRemoteTable{
				Schema: "public",
				Name:   table,
			},
			ColumnMapping: map[string]string{
				tableID: refTableID,
			},
		},
	}
}

func checkHasuraArgs() string {
	errMsg := ""
	if table == "" {
		errMsg += "table arg is missing. "
	}
	if pkey == "" && templateVersion != "3" {
		errMsg += "pkey arg is missing. "
	}
	if databaseName == "" {
		errMsg += "databaseName arg is missing. "
	}
	if permissionPrefix == "" && templateVersion != "4" {
		errMsg += "permissionPrefix arg is missing. "
	}

	if templateVersion == "4" {
		if ownerCol == "" {
			errMsg += "ownerCol is required in template Version 4"
		}
	}

	return errMsg
}

func getHasuraVersion1MetadataPath() string {
	if stgHasura {
		return hasuraTemplateStgTableMetadataPath
	}
	return hasuraTemplateTableMetadataPath
}

func (h *Hasura) getFileContent(svc string) ([]byte, error) {
	filepath := fmt.Sprintf(getHasuraVersion1MetadataPath(), svc)
	return h.IOUtils.GetFileContent(filepath)
}

func (h *Hasura) getHasuraV2Tables(svc string) ([]HasuraTable, error) {
	folderPath := fmt.Sprintf(hasuraTemplateTableFolderV2, svc, svc)
	fileNames, err := h.IOUtils.GetFileNamesOnDir(folderPath)
	if err != nil {
		return nil, err
	}
	tables := []HasuraTable{}

	for _, fileName := range fileNames {
		if fileName == "tables.yaml" {
			continue
		}

		path := fmt.Sprintf("%s/%s", folderPath, fileName)
		f, err := h.IOUtils.GetFileContent(path)

		if err != nil {
			return nil, err
		}

		table := HasuraTable{}
		err = yaml.Unmarshal(f, &table)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (h *Hasura) getHasuraV1Tables(svc string) ([]HasuraTable, error) {
	fileContent, err := h.getFileContent(svc)

	if err != nil {
		return nil, err
	}

	tableMetadata := []HasuraTable{}

	err = yaml.Unmarshal(fileContent, &tableMetadata)

	if err != nil {
		return nil, err
	}
	return tableMetadata, nil
}

func (h *Hasura) getTableMetadataContent(svc string) ([]HasuraTable, error) {
	if hasuraVersion == "2" {
		return h.getHasuraV2Tables(svc)
	}
	return h.getHasuraV1Tables(svc)
}

func findByTableName(tableName string, tables []HasuraTable, path string) *HasuraTable {
	for _, table := range tables {
		if table.Table.Name == tableName && !strings.Contains(path, table.Table.Name) {
			return &table
		}
	}

	return nil
}

func getObjRelationTableName(objRef HasuraObjectRelationships) (tableName string, relationName string) {
	if objRef.Using.ManualConfiguration != nil {
		return objRef.Using.ManualConfiguration.RemoteTable.Name, objRef.Name
	}
	return objRef.Name + "s", objRef.Name
}

func getArrRelationTableName(arrRef HasuraArrayRelationships) (tableName string, relationName string) {
	if arrRef.Using.ManualConfiguration != nil {
		return arrRef.Using.ManualConfiguration.RemoteTable.Name, arrRef.Name
	}
	return arrRef.Using.ForeignKeyConstraintOn.Table.Name, arrRef.Name
}

func checkMaxLv(path string) bool {
	level := strings.Split(path, "/")
	return len(level) > 6
}

func findAllRefAccessPath(hasuraTable HasuraTable, tables []HasuraTable, path string) []string {
	result := []string{}

	if checkMaxLv(path) {
		return result
	}

	if hasuraTable.ObjectRelationships != nil {
		for _, objRef := range *hasuraTable.ObjectRelationships {
			tableName, relationName := getObjRelationTableName(objRef)
			refTable := findByTableName(tableName, tables, path)
			if refTable == nil {
				continue
			}

			newPath := path + "/o:" + relationName

			if refTable.Table.Name == accessPathTable {
				result = append(result, newPath)
				continue
			}

			if refTable != nil {
				founds := findAllRefAccessPath(*refTable, tables, newPath)
				result = append(result, founds...)
			}
		}
	}
	if hasuraTable.ArrayRelationships != nil {
		for _, arrRef := range *hasuraTable.ArrayRelationships {
			tableName, relationName := getArrRelationTableName(arrRef)
			refTable := findByTableName(tableName, tables, path)
			if refTable == nil {
				continue
			}

			newPath := path + "/a:" + relationName

			if refTable.Table.Name == accessPathTable {
				result = append(result, newPath)
				continue
			}

			if refTable != nil {
				founds := findAllRefAccessPath(*refTable, tables, newPath)
				result = append(result, founds...)
			}
		}
	}
	return result
}

func getShortestRelationShip(relationships []string) string {
	result := ""
	shortestLv := 1000
	for _, relationship := range relationships {
		lv := strings.Split(relationship, "/")
		if len(lv) < shortestLv {
			shortestLv = len(lv)
			result = relationship
		}
	}
	return result
}

func replacePrefix(content string) string {
	content = strings.ReplaceAll(content, "o:", "")
	content = strings.ReplaceAll(content, "a:", "")
	return content
}

func buildPoliciesWithOtherTemplate(currentFilter interface{}, otherTemplateFilter interface{}) interface{} {
	if otherTemplateFilter == nil {
		return currentFilter
	}
	query := make([]interface{}, 2)
	query[0] = otherTemplateFilter
	query[1] = currentFilter
	locationAndOwnerPermission := make(map[string]interface{})
	locationAndOwnerPermission["_or"] = query
	return locationAndOwnerPermission
}

func buildFilterWithDeletedAt(filter interface{}) map[string]interface{} {
	deletedAtNull := map[string]interface{}{"deleted_at": map[string]bool{"_is_null": true}}
	arrQuery := make([]interface{}, 2)
	arrQuery[0] = filter
	arrQuery[1] = deletedAtNull
	andQuery := map[string]interface{}{"_and": arrQuery}
	return andQuery
}

func buildRls(shortestRelationship string, filter interface{}, restFilters []interface{}) (interface{}, interface{}) {
	fullQueries := make([]interface{}, 0)
	var permissionQuery interface{}
	fmt.Println(Info("templateVersion:", templateVersion))
	version := templateVersion

	indexTemplateFilter, otherTemplateFilter := getFilterFromOtherTemplate(restFilters, firstLvQuery)

	if version == "1" || version == "1.1" {
		tables := strings.Split(shortestRelationship, "/")
		nestedLocationFilter := make(map[string]interface{})
		lastIndex := len(tables) - 1
		for i := lastIndex; i > 0; i-- {
			table := tables[i]
			table = replacePrefix(table)
			temp := make(map[string]interface{})

			if i == lastIndex {
				temp[table] = buildFilterWithDeletedAt(filter)
				nestedLocationFilter = temp
			} else {
				temp[table] = nestedLocationFilter
				nestedLocationFilter = temp
			}
		}

		if lastIndex != 0 {
			filter = nestedLocationFilter
		}
		endFilters := buildPoliciesWithOtherTemplate(filter, otherTemplateFilter)
		permissionQuery = filter

		fullQueries = append(fullQueries, endFilters)
	} else {
		endFilters := buildPoliciesWithOtherTemplate(filter, otherTemplateFilter)
		fullQueries = append(fullQueries, endFilters)
		permissionQuery = filter
	}

	for i, v := range restFilters {
		if !checkQueryExisted(fullQueries, v) && i != indexTemplateFilter {
			fullQueries = append(fullQueries, v)
		}
	}

	andQuery := make(map[string]interface{})
	and := "_and"
	andQuery[and] = fullQueries
	return andQuery, permissionQuery
}

func (h *Hasura) writeMetadataFileHasuraV1(svc string, content []byte) error {
	filepath := fmt.Sprintf(getHasuraVersion1MetadataPath(), svc)
	fmt.Println("Wrote RLS for permission and location to metadata: ", filepath)
	return h.IOUtils.WriteFile(filepath, content)
}

const (
	SelectColumn = "SelectColumn"
	InsertColumn = "InsertColumn"
	UpdateColumn = "UpdateColumn"
)

func addPermissionCols(columns []string, cols map[string]bool) {
	for _, col := range columns {
		cols[col] = true
	}
}

type getCol[T any] func(input T) []string

func getColsFromPermission[T any](permissions *[]T, col getCol[T]) map[string]bool {
	cols := make(map[string]bool)
	if permissions == nil {
		return cols
	}
	for _, role := range *permissions {
		addPermissionCols(col(role), cols)
	}
	return cols
}

func getAllowColumns(table HasuraTable, columnType string) []string {
	cols := make(map[string]bool)
	switch columnType {
	case SelectColumn:
		cols = getColsFromPermission(table.SelectPermissions, func(role HasuraSelectPermissions) []string {
			if role.Permission.Columns != nil {
				return role.Permission.Columns
			}
			return []string{}
		})
	case UpdateColumn:
		cols = getColsFromPermission(table.UpdatePermissions, func(role HasuraInsertPermissions) []string {
			if role.Permission.Columns != nil {
				return role.Permission.Columns
			}
			return []string{}
		})
	case InsertColumn:
		cols = getColsFromPermission(table.InsertPermissions, func(role HasuraInsertPermissions) []string {
			if role.Permission.Columns != nil {
				return role.Permission.Columns
			}
			return []string{}
		})
	}

	keys := make([]string, 0, len(cols))
	for k := range cols {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if templateVersion == "4" {
		keys = append(keys, ownerCol)
	}

	return keys
}

func getFilterGrantedView() interface{} {
	eqUserID := map[string]string{"_eq": hasuraUserIDInput}
	eqResourcePath := map[string]string{"_eq": hasuraResourcePathInput}

	userFilter := make([]interface{}, 2)
	userFilter[0] = map[string]interface{}{"user_id": eqUserID}
	userFilter[1] = map[string]interface{}{"resource_path": eqResourcePath}

	andQuery := make(map[string]interface{})
	andQuery["_and"] = userFilter
	return andQuery
}

func getGrantedView() HasuraTable {
	filter := getFilterGrantedView()
	permission := HasuraSelectPermissions{
		Role: defaultRoleName,
		Permission: &HasuraPermission{
			Columns: []string{"user_id", "permission_name", "location_id", "resource_path"},
			Filter:  &filter,
		},
	}
	selectPermissions := []HasuraSelectPermissions{}
	selectPermissions = append(selectPermissions, permission)
	return HasuraTable{
		Table: HasuraTableSchema{
			Schema: "public",
			Name:   grantedTableName,
		},
		SelectPermissions: &selectPermissions,
	}
}

func (h *Hasura) writeTableFileHasuraV2(svc string, tableName string, table HasuraTable) error {
	filePath := fmt.Sprintf(hasuraTemplateTableFileV2, svc, svc, tableName)
	data, err := yaml.Marshal(&table)
	if err != nil {
		return err
	}

	return h.IOUtils.WriteFile(filePath, data)
}

func (h *Hasura) updateHasuraMetadata(hasuraVersion string, svc string, hasuraTableMetadata []HasuraTable) error {
	if hasuraVersion == "2" {
		return h.updateHasuraV2Files(svc, hasuraTableMetadata)
	}
	return h.updateHasuraV1File(svc, hasuraTableMetadata)
}

func (h *Hasura) updateHasuraV2Files(svc string, hasuraTableMetadata []HasuraTable) error {
	for _, hasuraTable := range hasuraTableMetadata {
		if hasuraTable.Table.Name == table || hasuraTable.Table.Name == accessPathTable && accessPathTable != table || hasuraTable.Table.Name == grantedTableName || rlsType == GenRole {
			err := h.writeTableFileHasuraV2(svc, hasuraTable.Table.Name, hasuraTable)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hasura) updateHasuraV1File(databaseName string, hasuraTableMetadata []HasuraTable) error {
	data, err := yaml.Marshal(&hasuraTableMetadata)

	if err != nil {
		return err
	}

	err = h.writeMetadataFileHasuraV1(databaseName, data)

	if err != nil {
		return err
	}

	return nil
}

func (h *Hasura) addIncludedTableToHasuraMetadataV2(svc string) error {
	filePath := fmt.Sprintf(hasuraTemplateTableFolderV2, svc, svc) + "/tables.yaml"

	err := h.IOUtils.AppendStrToFile(filePath, grantedTableIncluded)

	if err != nil {
		return err
	}

	return nil
}

func updateSelectPermission(selectPermissions []HasuraSelectPermissions, newSelectPermission HasuraSelectPermissions) *[]HasuraSelectPermissions {
	isExisted := false
	for i, permission := range selectPermissions {
		if permission.Role == newSelectPermission.Role {
			selectPermissions[i] = newSelectPermission
			isExisted = true
		}
	}
	if !isExisted {
		selectPermissions = append(selectPermissions, newSelectPermission)
	}

	return &selectPermissions
}

func updateInsertPermission(insertPermissions []HasuraInsertPermissions, newInsertPermission HasuraInsertPermissions) *[]HasuraInsertPermissions {
	isExisted := false
	for i, permission := range insertPermissions {
		if permission.Role == newInsertPermission.Role {
			insertPermissions[i] = newInsertPermission
			isExisted = true
		}
	}
	if !isExisted {
		insertPermissions = append(insertPermissions, newInsertPermission)
	}

	return &insertPermissions
}

func updateDeletePermission(deletePermissions []HasuraDeletePermissions, newDeletePermission HasuraDeletePermissions) *[]HasuraDeletePermissions {
	isExisted := false
	for i, permission := range deletePermissions {
		if permission.Role == newDeletePermission.Role {
			deletePermissions[i] = newDeletePermission
			isExisted = true
		}
	}
	if !isExisted {
		deletePermissions = append(deletePermissions, newDeletePermission)
	}

	return &deletePermissions
}

func updateObjRelationship(objectRelationships []HasuraObjectRelationships, newObjectRelationship HasuraObjectRelationships) *[]HasuraObjectRelationships {
	isExisted := false
	for i, objectRelationship := range objectRelationships {
		if objectRelationship.Name == newObjectRelationship.Name {
			objectRelationships[i] = newObjectRelationship
			isExisted = true
		}
	}
	if !isExisted {
		objectRelationships = append(objectRelationships, newObjectRelationship)
	}

	return &objectRelationships
}

func updateArrRelationship(arrRelationships []HasuraArrayRelationships, newArrRelationship HasuraArrayRelationships) *[]HasuraArrayRelationships {
	isExisted := false
	for _, objectRelationship := range arrRelationships {
		if objectRelationship.Name == newArrRelationship.Name {
			isExisted = true
		}
	}
	if !isExisted {
		arrRelationships = append(arrRelationships, newArrRelationship)
	}

	return &arrRelationships
}

type getAllPermissions[T any] func(input []T) []interface{}
type updateExitedPermission[T any] func(inputs []T, input T) *[]T
type buildPermission[T any] func(columns []string, checks interface{}, input []T) T
type BuildPermissionInput[T any] struct {
	input                  *[]T
	shortestRelationship   string
	columns                []string
	getAllPermissions      getAllPermissions[T]
	buildPermission        buildPermission[T]
	updateExitedPermission updateExitedPermission[T]
	permissionName         string
}

func buildPermissionsFor[T any](permissionInput BuildPermissionInput[T]) *[]T {
	permissions := []T{}
	if permissionInput.input != nil {
		permissions = *permissionInput.input
	}
	locationFilter := getLocationPermissionQuery(table, permissionInput.permissionName)
	filters := permissionInput.getAllPermissions(permissions)
	rlsQueries, _ := buildRls(permissionInput.shortestRelationship, locationFilter, filters)

	permission := permissionInput.buildPermission(permissionInput.columns, rlsQueries, permissions)

	return permissionInput.updateExitedPermission(permissions, permission)
}

type ManbieRoleInput struct {
}

func buildManabieRole(hasuraTableMetadata *HasuraTable, shortestRelationship string, columns, insertColumns, updateColumns []string) {
	permissionRead := permissionPrefix + ".read"
	permissionWrite := permissionPrefix + writeSuffix

	hasuraTableMetadata.SelectPermissions = buildPermissionsFor(
		BuildPermissionInput[HasuraSelectPermissions]{
			input:                  hasuraTableMetadata.SelectPermissions,
			shortestRelationship:   shortestRelationship,
			columns:                columns,
			getAllPermissions:      getAllSelectPermission,
			buildPermission:        buildSelectPermission,
			updateExitedPermission: updateSelectPermission,
			permissionName:         permissionRead,
		})

	if strings.Contains(writePermissionHasura, "INSERT") {
		hasuraTableMetadata.InsertPermissions = buildPermissionsFor(
			BuildPermissionInput[HasuraInsertPermissions]{
				input:                  hasuraTableMetadata.InsertPermissions,
				shortestRelationship:   shortestRelationship,
				columns:                insertColumns,
				getAllPermissions:      getAllInsertPermission,
				buildPermission:        buildInsertPermission,
				updateExitedPermission: updateInsertPermission,
				permissionName:         permissionWrite,
			})
	}
	if strings.Contains(writePermissionHasura, "UPDATE") {
		hasuraTableMetadata.UpdatePermissions = buildPermissionsFor(
			BuildPermissionInput[HasuraInsertPermissions]{
				input:                  hasuraTableMetadata.UpdatePermissions,
				shortestRelationship:   shortestRelationship,
				columns:                updateColumns,
				getAllPermissions:      getAllInsertPermission,
				buildPermission:        buildUpdatePermission,
				updateExitedPermission: updateInsertPermission,
				permissionName:         permissionWrite,
			})
	}
	if strings.Contains(writePermissionHasura, "DELETE") {
		hasuraTableMetadata.DeletePermissions = buildPermissionsFor(
			BuildPermissionInput[HasuraDeletePermissions]{
				input:                  hasuraTableMetadata.DeletePermissions,
				shortestRelationship:   shortestRelationship,
				columns:                updateColumns,
				getAllPermissions:      getAllDeletePermission,
				buildPermission:        buildDeletePermission,
				updateExitedPermission: updateDeletePermission,
				permissionName:         permissionWrite,
			})
	}

	fmt.Println("main table:", hasuraTableMetadata.Table.Name, hasuraTableMetadata.SelectPermissions)
}

func addRlsToSelectPermission(selectPermission HasuraSelectPermissions, shortestRelationship string) (interface{}, interface{}) {
	permissionRead := permissionPrefix + ".read"
	locationFilter := getLocationPermissionQuery(table, permissionRead)
	filters := getAllSelectPermission([]HasuraSelectPermissions{selectPermission})
	return buildRls(shortestRelationship, locationFilter, filters)
}

func addRLSSelectPermissionToAllRole(hasuraTableMetadata *HasuraTable, shortestRelationship string, permissionHasura *[]TemplateHasuraRole) {
	if hasuraTableMetadata.SelectPermissions == nil {
		return
	}
	for _, selectPermission := range *hasuraTableMetadata.SelectPermissions {
		filters, permissionCheck := addRlsToSelectPermission(selectPermission, shortestRelationship)
		selectPermission.Permission.Filter = &filters

		*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
			Name:   selectPermission.Role,
			Filter: &permissionCheck,
		})
	}
}

func findAndRemoveObjRelationship(hasuraTable *HasuraTable, relationship string) {
	if hasuraTable.ObjectRelationships == nil {
		return
	}
	foundIndex := -1
	for i, objectRelationship := range *hasuraTable.ObjectRelationships {
		if objectRelationship.Name == relationship {
			foundIndex = i
		}
	}
	if foundIndex > -1 {
		objectRelationships := removeSlice(*hasuraTable.ObjectRelationships, foundIndex)
		hasuraTable.ObjectRelationships = &objectRelationships
	}
}

func findAndRemoveArrRelationship(hasuraTable *HasuraTable, relationship string) {
	if hasuraTable.ArrayRelationships == nil {
		return
	}
	foundIndex := -1
	for i, arrayRelationship := range *hasuraTable.ArrayRelationships {
		if arrayRelationship.Name == relationship {
			foundIndex = i
		}
	}
	if foundIndex > -1 {
		arrayRelationships := removeSlice(*hasuraTable.ArrayRelationships, foundIndex)
		hasuraTable.ArrayRelationships = &arrayRelationships
	}
}

func findConditionHaveRelationship(itemObj []interface{}, relationship string) int {
	indexRemove := -1
	for i, v := range itemObj {
		n := v.(map[interface{}]interface{})
		for field := range n {
			str, ok := field.(string)
			if ok && str == relationship {
				indexRemove = i
				break
			}
		}
	}
	return indexRemove
}

func findAndRemoveFilter(hasuraTable *HasuraTable, relationship string) {
	if hasuraTable.SelectPermissions == nil || relationship == "" {
		return
	}
	selectPermissions := *hasuraTable.SelectPermissions
	isModified := false
	for i, selectPermission := range selectPermissions {
		filters := *selectPermission.Permission.Filter
		if filters != nil {
			filtersArr, _ := filters.(map[interface{}]interface{})
			for key, item := range filtersArr {
				itemObj, _ := item.([]interface{})
				keyStr, _ := key.(string)
				if indexRemove := findConditionHaveRelationship(itemObj, relationship); indexRemove > -1 && (keyStr == "_and" || keyStr == "_or") {
					itemObj = removeSlice(itemObj, indexRemove)
					isModified = true
				}
				filtersArr[key] = itemObj
			}
			filters = filtersArr
		}
		selectPermissions[i].Permission.Filter = &filters
	}
	if isModified {
		hasuraTable.SelectPermissions = &selectPermissions
	}
}

func (h *Hasura) dropRelationship(svc string, dropTemplateStage TemplateStage, tableName string, numberOfTemplate int) error {
	fmt.Println(Warn("dropRelationships: ", tableName))
	hasuraTableMetadata, err := h.getTableMetadataContent(svc)
	if err != nil {
		return fmt.Errorf("error when get content from metadata file %v", err)
	}

	for i, table := range hasuraTableMetadata {
		if table.Table.Name == grantedTableName || (dropTemplateStage.AccessPathTable != nil && table.Table.Name == dropTemplateStage.AccessPathTable.Name) || table.Table.Name == tableName {
			findAndRemoveObjRelationship(&hasuraTableMetadata[i], dropTemplateStage.Hasura.Relationship)
			findAndRemoveArrRelationship(&hasuraTableMetadata[i], dropTemplateStage.Hasura.Relationship)
			if table.Table.Name == tableName {
				firstLvQuery := dropTemplateStage.Hasura.FirstLvQuery
				if numberOfTemplate > 1 {
					firstLvQuery = "_or"
				}
				findAndRemoveFilter(&hasuraTableMetadata[i], firstLvQuery)
			}
		}
	}

	hasuraVersion := "1"
	err = h.updateHasuraMetadata(hasuraVersion, svc, hasuraTableMetadata)
	if err != nil {
		return err
	}
	return nil
}

func (h *Hasura) dropCustomRelationship(svc string, hasuraPolicy TemplateHasuraPolicy, tableName string) error {
	fmt.Println(Warn("dropCustomRelationship: ", tableName))
	hasuraTableMetadata, err := h.getTableMetadataContent(svc)
	if err != nil {
		return fmt.Errorf("error when get content from metadata file %v", err)
	}
	filterMap := map[string]string{}
	objectReMap := map[string][]string{}
	arrayReMap := map[string][]string{}
	for _, policy := range *hasuraPolicy.SelectPermission {
		policyFilter := *policy.Filter
		filters, _ := policyFilter.([]interface{})
		for _, filter := range filters {
			if filterObj, ok := filter.(map[string]interface{}); ok {
				for k := range filterObj {
					filterMap[policy.Name] = k
				}
			}
		}
	}
	if hasuraPolicy.ObjectCustomRelationship != nil {
		for _, objRe := range *hasuraPolicy.ObjectCustomRelationship {
			if _, ok := objectReMap[objRe.TableName]; !ok {
				objectReMap[objRe.TableName] = []string{}
			}
			objectReMap[objRe.TableName] = append(objectReMap[objRe.TableName], objRe.ManualConfig.Name)
		}
	}

	if hasuraPolicy.ArrayCustomRelationship != nil {
		for _, arrRe := range *hasuraPolicy.ArrayCustomRelationship {
			if _, ok := arrayReMap[arrRe.TableName]; !ok {
				arrayReMap[arrRe.TableName] = []string{}
			}
			arrayReMap[arrRe.TableName] = append(arrayReMap[arrRe.TableName], arrRe.ManualConfig.Name)
		}
	}

	for i, table := range hasuraTableMetadata {
		if objRels, ok := objectReMap[table.Table.Name]; ok {
			for _, objRel := range objRels {
				findAndRemoveObjRelationship(&hasuraTableMetadata[i], objRel)
			}
		}
		if arrRels, ok := arrayReMap[table.Table.Name]; ok {
			for _, arrRel := range arrRels {
				findAndRemoveArrRelationship(&hasuraTableMetadata[i], arrRel)
			}
		}
		if tableName == table.Table.Name {
			for _, v := range filterMap {
				findAndRemoveFilter(&hasuraTableMetadata[i], v)
			}
		}
	}

	hasuraVersion := "1"
	err = h.updateHasuraMetadata(hasuraVersion, svc, hasuraTableMetadata)
	if err != nil {
		return err
	}
	return nil
}

func getNilIfEmptyArr[T any](arr []T) *[]T {
	if len(arr) > 0 {
		return &arr
	}
	return nil
}

func addRlsToInsertPermission(insertPermission HasuraInsertPermissions, shortestRelationship string) (interface{}, interface{}) {
	permissionWrite := permissionPrefix + writeSuffix
	locationFilter := getLocationPermissionQuery(table, permissionWrite)
	filters := getAllInsertPermission([]HasuraInsertPermissions{insertPermission})
	return buildRls(shortestRelationship, locationFilter, filters)
}

func addRLSInsertPermissionToAllRole(hasuraTableMetadata *HasuraTable, shortestRelationship string, permissionHasura *[]TemplateHasuraRole) {
	if templateVersion == "1.1" {
		return
	}
	insertPermissions := []HasuraInsertPermissions{}
	if hasuraTableMetadata.InsertPermissions != nil {
		insertPermissions = *hasuraTableMetadata.InsertPermissions
	}

	for _, insertPermission := range insertPermissions {
		checks, permissionCheck := addRlsToInsertPermission(insertPermission, shortestRelationship)
		insertPermission.Permission.Check = &checks

		*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
			Name:  insertPermission.Role,
			Check: &permissionCheck,
		})
	}
	hasuraTableMetadata.InsertPermissions = getNilIfEmptyArr(insertPermissions)
}

func addRlsToUpdatePermission(updatePermission HasuraInsertPermissions, shortestRelationship string) (interface{}, interface{}) {
	permissionWrite := permissionPrefix + writeSuffix
	locationFilter := getLocationPermissionQuery(table, permissionWrite)
	filters := getAllInsertPermission([]HasuraInsertPermissions{updatePermission})
	return buildRls(shortestRelationship, locationFilter, filters)
}

func addRLSUpdatePermissionToAllRole(hasuraTableMetadata *HasuraTable, shortestRelationship string, permissionHasura *[]TemplateHasuraRole) {
	updatePermissions := []HasuraInsertPermissions{}
	if hasuraTableMetadata.UpdatePermissions != nil {
		updatePermissions = *hasuraTableMetadata.UpdatePermissions
	}
	for _, updatePermission := range updatePermissions {
		checks, permissionCheck := addRlsToUpdatePermission(updatePermission, shortestRelationship)
		updatePermission.Permission.Check = &checks
		updatePermission.Permission.Filter = &checks

		*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
			Name:   updatePermission.Role,
			Check:  &permissionCheck,
			Filter: &permissionCheck,
		})
	}
	hasuraTableMetadata.UpdatePermissions = getNilIfEmptyArr(updatePermissions)
}

func addRlsToDeletePermission(deletePermission HasuraDeletePermissions, shortestRelationship string) (interface{}, interface{}) {
	permissionWrite := permissionPrefix + writeSuffix
	locationFilter := getLocationPermissionQuery(table, permissionWrite)
	filters := getAllDeletePermission([]HasuraDeletePermissions{deletePermission})
	return buildRls(shortestRelationship, locationFilter, filters)
}

func addRLSDeletePermissionToAllRole(hasuraTableMetadata *HasuraTable, shortestRelationship string, permissionHasura *[]TemplateHasuraRole) {
	deletePermissions := []HasuraDeletePermissions{}
	if hasuraTableMetadata.DeletePermissions != nil {
		deletePermissions = *hasuraTableMetadata.DeletePermissions
	}
	for _, deletePermission := range deletePermissions {
		checks, permissionCheck := addRlsToDeletePermission(deletePermission, shortestRelationship)
		deletePermission.Permission.Check = &checks
		deletePermission.Permission.Filter = &checks
		*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
			Name:   deletePermission.Role,
			Check:  &permissionCheck,
			Filter: &permissionCheck,
		})
	}
	hasuraTableMetadata.DeletePermissions = getNilIfEmptyArr(deletePermissions)
}

func getFirstLevelQuery(shortestRelationship string, templateVersion string, ownerCol string, relationship string) string {
	tables := strings.Split(shortestRelationship, "/")
	result := ""
	if templateVersion != "4" && len(tables) > 1 {
		result = strings.ReplaceAll(tables[1], "o:", "")
		result = strings.ReplaceAll(result, "a:", "")
	} else if templateVersion == "1" {
		result = relationship
	} else if templateVersion == "4" {
		result = ownerCol
	} else if templateVersion == "3" {
		result = "_exists"
	}
	return result
}

func setDefaultInputIfNotExisted() {
	if templateVersion == "" {
		templateVersion = "1"
	}
	if accessPathLocationCol != "" && !mapHasuraDirectly {
		accessPathTableKey = accessPathLocationCol
	}
	if otherTemplateFilterName != "" {
		firstLvQuery = otherTemplateFilterName
	}

	if templateVersion == "4" && pkey == "" && ownerCol != "" {
		pkey = ownerCol
	}
}

func buildManualObjConfig(hasuraTable *HasuraTable) *[]HasuraObjectRelationships {
	relationshipName := table
	objectRelationships := []HasuraObjectRelationships{}
	if hasuraTable.ObjectRelationships != nil {
		objectRelationships = *hasuraTable.ObjectRelationships
	}
	relation := buildManualObjectRelationship(table, accessPathTableKey, pkey)
	objectRelationship := HasuraObjectRelationships{
		Name:  relationshipName,
		Using: &relation,
	}

	return updateObjRelationship(objectRelationships, objectRelationship)
}

func buildManualArrConfig(hasuraTable *HasuraTable) *[]HasuraArrayRelationships {
	relationshipName := accessPathTable
	arrRelationships := []HasuraArrayRelationships{}
	if hasuraTable.ArrayRelationships != nil {
		arrRelationships = *hasuraTable.ArrayRelationships
	}
	relation := buildManualArrRelationship(accessPathTable, pkey, accessPathTableKey)
	arrRelationship := HasuraArrayRelationships{
		Name:  relationshipName,
		Using: &relation,
	}

	return updateArrRelationship(arrRelationships, arrRelationship)
}

func updateHasuraTableMetadata(hasuraTableMetadata []HasuraTable, shortestRelationship, relationshipName string, columns, insertColumns, updateColumns []string, selectStagePermission, insertStagePermission, updateStagePermission, deleteStagePermission *[]TemplateHasuraRole) []HasuraTable {
	for i, hasuraTable := range hasuraTableMetadata {
		if hasuraTable.Table.Name == table {
			if addRLSToAllPermissionHasura {
				addRLSSelectPermissionToAllRole(&hasuraTableMetadata[i], shortestRelationship, selectStagePermission)
				addRLSInsertPermissionToAllRole(&hasuraTableMetadata[i], shortestRelationship, insertStagePermission)
				addRLSUpdatePermissionToAllRole(&hasuraTableMetadata[i], shortestRelationship, updateStagePermission)
				addRLSDeletePermissionToAllRole(&hasuraTableMetadata[i], shortestRelationship, deleteStagePermission)
			} else {
				buildManabieRole(&hasuraTableMetadata[i], shortestRelationship, columns, insertColumns, updateColumns)
			}

			if mapHasuraDirectly && accessPathTableKey != "" && templateVersion != "3" {
				hasuraTableMetadata[i].ArrayRelationships = buildManualArrConfig(&hasuraTableMetadata[i])
			}
		}
		if hasuraTable.Table.Name == accessPathTable && templateVersion != "4" && templateVersion != "3" {
			granted := buildRelationShip(accessPathTableKey)
			objectRelationships := []HasuraObjectRelationships{}
			if hasuraTable.ObjectRelationships != nil {
				objectRelationships = *hasuraTable.ObjectRelationships
			}
			objectRelationship := HasuraObjectRelationships{
				Name:  relationshipName,
				Using: &granted,
			}

			hasuraTableMetadata[i].ObjectRelationships = updateObjRelationship(objectRelationships, objectRelationship)

			if mapHasuraDirectly && accessPathTableKey != "" {
				hasuraTableMetadata[i].ObjectRelationships = buildManualObjConfig(&hasuraTableMetadata[i])
			}

			fmt.Println("access table:", hasuraTable.Table.Name)
		}
		if hasuraTable.Table.Name == grantedTableName && templateVersion != "4" && templateVersion != "3" {
			relation := buildGrantedPermission(accessPathTable, accessPathTableKey)
			objectRelationships := []HasuraObjectRelationships{}
			if hasuraTable.ObjectRelationships != nil {
				objectRelationships = *hasuraTable.ObjectRelationships
			}
			objectRelationship := HasuraObjectRelationships{
				Name:  relationshipName,
				Using: &relation,
			}
			hasuraTableMetadata[i].ObjectRelationships = updateObjRelationship(objectRelationships, objectRelationship)
			fmt.Println("granted table:", hasuraTable.Table.Name)
		}
	}
	return hasuraTableMetadata
}

func detectRelationshipColumnAndView(hasuraTableMetadata *[]HasuraTable) (relationships []string, columns []string, insertColumns []string, updateColumns []string, hasGrantedView bool) {
	for _, hasuraTable := range *hasuraTableMetadata {
		if hasuraTable.Table.Name == grantedTableName {
			hasGrantedView = true
		}
		if hasuraTable.Table.Name == table {
			if mapHasuraDirectly && accessPathTableKey != "" {
				relationships = []string{fmt.Sprintf("%s/a:%s", table, accessPathTable)}
			} else {
				relationships = findAllRefAccessPath(hasuraTable, *hasuraTableMetadata, table)
			}
			columns = getAllowColumns(hasuraTable, SelectColumn)
			insertColumns = getAllowColumns(hasuraTable, InsertColumn)
			updateColumns = getAllowColumns(hasuraTable, UpdateColumn)
		}
	}
	return relationships, columns, insertColumns, updateColumns, hasGrantedView
}

func (h *Hasura) addGrantedViewIfNotExisted(hasGrantedView bool, hasuraTableMetadata *[]HasuraTable) error {
	if !hasGrantedView {
		*hasuraTableMetadata = append(*hasuraTableMetadata, getGrantedView())
		if hasuraVersion == "2" {
			err := h.addIncludedTableToHasuraMetadataV2(databaseName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Hasura) genRLSMetadata() (*HasuraTemplateStage, error) {
	fmt.Println("Running generate RLS to metadata file")

	setDefaultInputIfNotExisted()

	errMsg := checkHasuraArgs()
	if errMsg != "" {
		return nil, fmt.Errorf(errMsg)
	}

	if accessPathTable == "" {
		accessPathTable = table
	}

	hasuraTableMetadata, err := h.getTableMetadataContent(databaseName)
	if err != nil {
		return nil, fmt.Errorf("error when get content from metadata file %v", err)
	}

	relationships, columns, insertColumns, updateColumns, hasGrantedView := detectRelationshipColumnAndView(&hasuraTableMetadata)
	err = h.addGrantedViewIfNotExisted(hasGrantedView, &hasuraTableMetadata)
	shortestRelationship := getShortestRelationShip(relationships)

	selectStagePermission := &[]TemplateHasuraRole{}
	insertStagePermission := &[]TemplateHasuraRole{}
	updateStagePermission := &[]TemplateHasuraRole{}
	deleteStagePermission := &[]TemplateHasuraRole{}

	fmt.Printf("shortest relation ship from %s to %s is %s\n", table, accessPathTable, shortestRelationship)
	relationshipName := table + "_location_permission"

	hasuraTableMetadata = updateHasuraTableMetadata(hasuraTableMetadata, shortestRelationship, relationshipName, columns, insertColumns, updateColumns, selectStagePermission, insertStagePermission, updateStagePermission, deleteStagePermission)

	if err != nil {
		return nil, err
	}

	err = h.updateHasuraMetadata(hasuraVersion, databaseName, hasuraTableMetadata)

	if err != nil {
		return nil, err
	}

	permissions := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}

	hasuraStage := &HasuraTemplateStage{
		FileDir:      fmt.Sprintf(hasuraTemplateTableMetadataPath, databaseName),
		Permissions:  permissions,
		Relationship: relationshipName,
		FirstLvQuery: getFirstLevelQuery(shortestRelationship, templateVersion, ownerCol, relationshipName),
		HasuraPolicy: &TemplateHasuraPolicy{
			SelectPermission: selectStagePermission,
			InsertPermission: insertStagePermission,
			UpdatePermission: updateStagePermission,
			DeletePermission: deleteStagePermission,
		},
	}

	return hasuraStage, nil
}
