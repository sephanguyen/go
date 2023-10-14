package rls

import "fmt"

func getCustomFilter(hasuraRoles *[]TemplateHasuraRole, roleName string) *TemplateHasuraRole {
	if hasuraRoles == nil {
		return nil
	}
	for _, hasuraRole := range *hasuraRoles {
		if hasuraRole.Name == roleName {
			return &hasuraRole
		}
	}
	return nil
}

func buildCustomFilter(customFilter *TemplateHasuraRole, fullQueries []interface{}, isInsert bool) (interface{}, interface{}) {
	var filter interface{}
	if customFilter != nil && (customFilter.Filter != nil || customFilter.Check != nil) {
		if isInsert {
			filter = *customFilter.Check
		} else {
			filter = *customFilter.Filter
		}
		switch v := filter.(type) {
		case []interface{}:
			fullQueries = append(fullQueries, v...)
		default:
			fullQueries = append(fullQueries, v)
		}
	}

	andQuery := make(map[string]interface{})
	and := "_and"
	andQuery[and] = fullQueries

	return andQuery, filter
}

func addCustomRlsToSelectPermission(selectPermission HasuraSelectPermissions, customPerHasura *[]TemplateHasuraRole, roleName string) (interface{}, interface{}) {
	customFilter := getCustomFilter(customPerHasura, roleName)
	if customFilter == nil {
		return nil, nil
	}

	fullQueries := getAllSelectPermission([]HasuraSelectPermissions{selectPermission})

	return buildCustomFilter(customFilter, fullQueries, false)
}

func addCustomRlsToInsertPermission(insertPermission HasuraInsertPermissions, customPerHasura *[]TemplateHasuraRole, roleName string) (interface{}, interface{}) {
	customFilter := getCustomFilter(customPerHasura, roleName)
	if customFilter == nil {
		return nil, nil
	}
	fullQueries := getAllInsertPermission([]HasuraInsertPermissions{insertPermission})
	return buildCustomFilter(customFilter, fullQueries, true)
}

func addCustomRlsToUpdatePermission(updatePermission HasuraInsertPermissions, customPerHasura *[]TemplateHasuraRole, roleName string) (interface{}, interface{}) {
	customFilter := getCustomFilter(customPerHasura, roleName)
	if customFilter == nil {
		return nil, nil
	}
	fullQueries := getAllInsertPermission([]HasuraInsertPermissions{updatePermission})
	return buildCustomFilter(customFilter, fullQueries, true)
}

func addCustomRlsToDeletePermission(deletePermission HasuraDeletePermissions, customPerHasura *[]TemplateHasuraRole, roleName string) (interface{}, interface{}) {
	customFilter := getCustomFilter(customPerHasura, roleName)
	if customFilter == nil {
		return nil, nil
	}
	fullQueries := getAllDeletePermission([]HasuraDeletePermissions{deletePermission})
	return buildCustomFilter(customFilter, fullQueries, true)
}

func addCustomRlsToSelectPermissions(hasuraTableMetadata *HasuraTable, hasuraPolicy *TemplateHasuraPolicy, permissionHasura *[]TemplateHasuraRole) {
	if hasuraPolicy.SelectPermission != nil && hasuraTableMetadata.SelectPermissions != nil {
		for _, selectPermission := range *hasuraTableMetadata.SelectPermissions {
			filters, permissionCheck := addCustomRlsToSelectPermission(selectPermission, hasuraPolicy.SelectPermission, selectPermission.Role)
			if permissionCheck == nil {
				continue
			}
			selectPermission.Permission.Filter = &filters

			*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
				Name:   selectPermission.Role,
				Filter: &permissionCheck,
			})
		}
	}
}

func addCustomRlsToInsertPermissions(hasuraTableMetadata *HasuraTable, hasuraPolicy *TemplateHasuraPolicy, permissionHasura *[]TemplateHasuraRole) {
	if hasuraPolicy.InsertPermission != nil && hasuraTableMetadata.InsertPermissions != nil {
		for _, insertPermission := range *hasuraTableMetadata.InsertPermissions {
			filters, permissionCheck := addCustomRlsToInsertPermission(insertPermission, hasuraPolicy.InsertPermission, insertPermission.Role)
			if permissionCheck == nil {
				continue
			}
			insertPermission.Permission.Check = &filters

			*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
				Name:   insertPermission.Role,
				Filter: &permissionCheck,
			})
		}
	}
}

func addCustomRlsToUpdatePermissions(hasuraTableMetadata *HasuraTable, hasuraPolicy *TemplateHasuraPolicy, permissionHasura *[]TemplateHasuraRole) {
	if hasuraPolicy.UpdatePermission != nil && hasuraTableMetadata.UpdatePermissions != nil {
		for _, updatePermission := range *hasuraTableMetadata.UpdatePermissions {
			filters, permissionCheck := addCustomRlsToUpdatePermission(updatePermission, hasuraPolicy.UpdatePermission, updatePermission.Role)
			if permissionCheck == nil {
				continue
			}
			updatePermission.Permission.Filter = &filters
			updatePermission.Permission.Check = &filters

			*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
				Name:   updatePermission.Role,
				Filter: &permissionCheck,
			})
		}
	}
}

func addCustomRlsToDeletePermissions(hasuraTableMetadata *HasuraTable, hasuraPolicy *TemplateHasuraPolicy, permissionHasura *[]TemplateHasuraRole) {
	if hasuraPolicy.DeletePermission != nil && hasuraTableMetadata.DeletePermissions != nil {
		for _, deletePermission := range *hasuraTableMetadata.DeletePermissions {
			filters, permissionCheck := addCustomRlsToDeletePermission(deletePermission, hasuraPolicy.DeletePermission, deletePermission.Role)
			if permissionCheck == nil {
				continue
			}
			deletePermission.Permission.Check = &filters
			deletePermission.Permission.Filter = &filters

			*permissionHasura = append(*permissionHasura, TemplateHasuraRole{
				Name:   deletePermission.Role,
				Filter: &permissionCheck,
			})
		}
	}
}

func getCustomArrRelationship(hasuraPolicy *TemplateHasuraPolicy, tableName string) []HasuraArrayRelationships {
	arrs := []HasuraArrayRelationships{}
	if hasuraPolicy.ArrayCustomRelationship == nil {
		return arrs
	}

	for _, arr := range *hasuraPolicy.ArrayCustomRelationship {
		if arr.TableName == tableName {
			arrs = append(arrs, arr.ManualConfig)
		}
	}

	return arrs
}

func getCustomObjRelationship(hasuraPolicy *TemplateHasuraPolicy, tableName string) []HasuraObjectRelationships {
	arrs := []HasuraObjectRelationships{}
	if hasuraPolicy.ObjectCustomRelationship == nil {
		return arrs
	}

	for _, arr := range *hasuraPolicy.ObjectCustomRelationship {
		if arr.TableName == tableName {
			arrs = append(arrs, arr.ManualConfig)
		}
	}

	return arrs
}

func buildCustomRelationship(hasuraTableMetadata *HasuraTable, hasuraPolicy *TemplateHasuraPolicy) {
	if hasuraTableMetadata == nil {
		return
	}
	objRe := getCustomObjRelationship(hasuraPolicy, hasuraTableMetadata.Table.Name)
	arrRe := getCustomArrRelationship(hasuraPolicy, hasuraTableMetadata.Table.Name)

	for _, obj := range objRe {
		if hasuraTableMetadata.ObjectRelationships == nil {
			hasuraTableMetadata.ObjectRelationships = &[]HasuraObjectRelationships{}
		}
		*hasuraTableMetadata.ObjectRelationships = append(*hasuraTableMetadata.ObjectRelationships, obj)
	}
	for _, arr := range arrRe {
		if hasuraTableMetadata.ArrayRelationships == nil {
			hasuraTableMetadata.ArrayRelationships = &[]HasuraArrayRelationships{}
		}
		*hasuraTableMetadata.ArrayRelationships = append(*hasuraTableMetadata.ArrayRelationships, arr)
	}
}

func addGrantedPermission(hasuraTableMetadata *[]HasuraTable) {
	if hasuraTableMetadata == nil {
		return
	}
	grantedView := findByTableName(grantedTableName, *hasuraTableMetadata, "")
	if grantedView == nil {
		*hasuraTableMetadata = append(*hasuraTableMetadata, getGrantedView())
	}
}

func (h *Hasura) getCustomPolicy(hasuraPolicy *TemplateHasuraPolicy) (*HasuraTemplateStage, error) {
	fmt.Println("Running generate custom RLS to metadata file")

	if hasuraPolicy == nil {
		return nil, fmt.Errorf("error when get TemplateHasuraPolicy")
	}

	hasuraTableMetadata, err := h.getTableMetadataContent(databaseName)
	if err != nil {
		return nil, fmt.Errorf("error when get content from metadata file %v", err)
	}

	selectStagePermission := &[]TemplateHasuraRole{}
	insertStagePermission := &[]TemplateHasuraRole{}
	updateStagePermission := &[]TemplateHasuraRole{}
	deleteStagePermission := &[]TemplateHasuraRole{}

	addGrantedPermission(&hasuraTableMetadata)

	for i, hasuraTable := range hasuraTableMetadata {
		if hasuraTable.Table.Name == table {
			addCustomRlsToSelectPermissions(&hasuraTableMetadata[i], hasuraPolicy, selectStagePermission)
			addCustomRlsToInsertPermissions(&hasuraTableMetadata[i], hasuraPolicy, insertStagePermission)
			addCustomRlsToUpdatePermissions(&hasuraTableMetadata[i], hasuraPolicy, updateStagePermission)
			addCustomRlsToDeletePermissions(&hasuraTableMetadata[i], hasuraPolicy, deleteStagePermission)
		}
		buildCustomRelationship(&hasuraTableMetadata[i], hasuraPolicy)
	}
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
		Relationship: "",
		FirstLvQuery: getFirstLevelQuery("", templateVersion, ownerCol, ""),
		HasuraPolicy: &TemplateHasuraPolicy{
			SelectPermission: selectStagePermission,
			InsertPermission: insertStagePermission,
			UpdatePermission: updateStagePermission,
			DeletePermission: deleteStagePermission,
		},
	}

	return hasuraStage, nil
}
