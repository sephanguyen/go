package rls

import (
	"fmt"
	"reflect"
)

func checkMigrateRoleArgs() string {
	errMsg := ""
	if databaseName == "" {
		errMsg += "databaseName arg is missing. "
	}

	return errMsg
}

func getNewRoleFilters(conditions interface{}) interface{} {
	if isZero(conditions) {
		empty := make(map[string]interface{})
		return empty
	}
	andQuery := make(map[string]interface{})
	and := "_and"
	andQuery[and] = conditions
	return andQuery
}

func getNewRole(columns []string, conditions interface{}) HasuraSelectPermissions {
	filters := getNewRoleFilters(conditions)
	permission := &HasuraPermission{
		Columns:           columns,
		AllowAggregations: true,
	}
	permission.Filter = &filters
	return HasuraSelectPermissions{
		Role:       defaultRoleName,
		Permission: permission,
	}
}

func isZero(v interface{}) bool {
	if v == nil {
		return true
	}
	return reflect.ValueOf(v).IsZero()
}

func updateSelectPermissionFor(newSelectPermission HasuraSelectPermissions, selectPermissions []HasuraSelectPermissions) *[]HasuraSelectPermissions {
	isExisted := false
	for i, permission := range selectPermissions {
		if permission.Role == newSelectPermission.Role {
			selectPermissions[i] = newSelectPermission
			isExisted = true
			if removeOldRole == "true" {
				return &[]HasuraSelectPermissions{permission}
			}
		}
	}
	if !isExisted {
		selectPermissions = append(selectPermissions, newSelectPermission)
		if removeOldRole == "true" {
			return &[]HasuraSelectPermissions{newSelectPermission}
		}
	}

	return &selectPermissions
}

func (h *Hasura) genNewRole() error {
	fmt.Println("Running genNewRole")

	errMsg := checkMigrateRoleArgs()
	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}

	hasuraTableMetadata, err := h.getTableMetadataContent(databaseName)
	if err != nil {
		return fmt.Errorf("error when get content from metadata file %v", err)
	}
	for i, hasuraTable := range hasuraTableMetadata {
		fmt.Println("updating table:", hasuraTable.Table.Name)

		if hasuraTable.Table.Name == grantedTableName {
			continue
		}

		if hasuraTable.SelectPermissions == nil {
			continue
		}
		selectPermissions := *hasuraTable.SelectPermissions
		filters := getAllSelectPermission(selectPermissions)
		columns := getAllowColumns(hasuraTable, SelectColumn)
		newRole := getNewRole(columns, filters)
		hasuraTableMetadata[i].SelectPermissions = updateSelectPermissionFor(newRole, selectPermissions)
	}
	err = h.updateHasuraMetadata(hasuraVersion, databaseName, hasuraTableMetadata)

	if err != nil {
		return err
	}

	return nil
}
