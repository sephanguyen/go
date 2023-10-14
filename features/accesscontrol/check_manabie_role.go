package accesscontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

type HasuraBody struct {
	Tables []HasuraTable `json:"tables"`
}

type HasuraTable struct {
	Table             Table                `json:"table"`
	SelectPermissions *[]SelectPermissions `json:"select_permissions"`
}
type Table struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type SelectPermissions struct {
	Role       string     `json:"role"`
	Permission Permission `json:"permission"`
}

type Permission struct {
	Columns []string     `json:"columns"`
	Filter  *interface{} `json:"filter"`
}

func (s *suite) accountAdminHasura(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.HasuraAdminAccount = "M@nabie123"
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) name(ctx context.Context, service string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch service {
	case "bob":
		stepState.HasuraURL = s.Cfg.BobHasuraAdminURL
	case "eureka":
		stepState.HasuraURL = s.Cfg.EurekaHasuraAdminURL
	case "fatima":
		stepState.HasuraURL = s.Cfg.FatimaHasuraAdminURL
	case "timesheet":
		stepState.HasuraURL = s.Cfg.TimesheetHasuraAdminURL
	case "entryexitmgmt":
		stepState.HasuraURL = s.Cfg.EntryexitmgmtHasuraAdminURL
	case "invoicemgmt":
		stepState.HasuraURL = s.Cfg.InvoicemgmtHasuraAdminURL
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminExportHasuraMetadata(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	body, err := s.ExportHasuraMetadata(stepState.HasuraURL, stepState.HasuraAdminAccount)

	if err != nil {
		return ctx, err
	}

	stepState.HasuraBody = body
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ignoreTables(ctx context.Context, table string) bool {
	stepState := StepStateFromContext(ctx)
	// In case this tables doesn't have old roles so we can ignore it
	// Only ignore tables outside bob
	ignoreTables := []string{"date_type", "student_parents", "invoice_schedule_student", "students", "users"}

	for _, a := range ignoreTables {
		if a == table && stepState.HasuraURL != s.Cfg.BobHasuraAdminURL {
			return true
		}
		if stepState.HasuraURL == s.Cfg.BobHasuraAdminURL && a == "date_type" {
			return true
		}
	}
	return false
}

func (s *suite) adminSeesTheExisted(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	body := &HasuraBody{}
	err := json.Unmarshal(stepState.HasuraBody, body)

	if err != nil {
		return ctx, err
	}

	for _, table := range body.Tables {
		if s.ignoreTables(ctx, table.Table.Name) {
			continue
		}
		isValidTable := false
		if table.SelectPermissions != nil {
			for _, permission := range *table.SelectPermissions {
				if permission.Role == role {
					isValidTable = true
					break
				}
			}
		}
		if !isValidTable {
			return ctx, fmt.Errorf("Table %s is missing %s role %s", table.Table.Name, role, stepState.HasuraURL)
		}
	}
	return ctx, nil
}

func (s *suite) adminSeesTheManabieRole(ctx context.Context) (context.Context, error) {
	return s.adminSeesTheExisted(ctx, "MANABIE")
}

func compareManabieColsAndOtherRoles(table HasuraTable) bool {
	otherRolesCols := make(map[string]bool)
	manabieCols := make(map[string]bool)
	if table.SelectPermissions != nil {
		for _, role := range *table.SelectPermissions {
			for _, col := range role.Permission.Columns {
				if role.Role != "MANABIE" {
					manabieCols[col] = true
				} else {
					otherRolesCols[col] = true
				}
			}
		}
	}

	otherRoleColsArrStr := sortMapStringType(otherRolesCols)
	manabieRoleColsArrStr := sortMapStringType(manabieCols)

	return check2ArrayStringSameOrder(otherRoleColsArrStr, manabieRoleColsArrStr)
}

func (s *suite) columnsIncludedAllColumnsFromOtherRoles(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	body := &HasuraBody{}
	err := json.Unmarshal(stepState.HasuraBody, body)

	if err != nil {
		return ctx, err
	}

	for _, table := range body.Tables {
		isEqual := compareManabieColsAndOtherRoles(table)
		if !isEqual {
			return ctx, fmt.Errorf("Table %s: MANABIE role columns is not correct in service %s", table.Table.Name, stepState.HasuraURL)
		}
	}
	return ctx, nil
}

func getFields2ndLvFilter(filter interface{}) []string {
	var rs = []string{}
	fields, ok := filter.([]interface{})
	if ok {
		for _, field := range fields {
			filters, ok := field.(map[string]interface{})
			if ok {
				for filter := range filters {
					rs = append(rs, filter)
				}
			}
		}
	}
	return rs
}

func getAllQueryName(filters interface{}) []string {
	fields := []string{}
	firstQueryLvType, _ := filters.(map[string]interface{})

	if andQuery := firstQueryLvType["_and"]; andQuery != nil {
		fields = append(fields, getFields2ndLvFilter(andQuery)...)
	} else if orQuery := firstQueryLvType["_or"]; orQuery != nil {
		fields = append(fields, getFields2ndLvFilter(orQuery)...)
	} else {
		for filter := range firstQueryLvType {
			fields = append(fields, filter)
		}
	}

	return fields
}

func sortMapStringType(filter map[string]bool) []string {
	keys := []string{}
	for k := range filter {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func check2ArrayStringSameOrder(keys []string, otherKeys []string) bool {
	if len(keys) != len(otherKeys) {
		return false
	}
	for index, key := range keys {
		if key != otherKeys[index] {
			return false
		}
	}
	return true
}

func compareManabieFilterAndOtherFilters(selectPermissions []SelectPermissions) bool {
	otherFilters := make(map[string]bool)
	manabieFilters := make(map[string]bool)

	for _, selectPermission := range selectPermissions {
		if selectPermission.Permission.Filter == nil {
			return false
		}
		if selectPermission.Role != "MANABIE" {
			fields := getAllQueryName(*selectPermission.Permission.Filter)
			for _, k := range fields {
				otherFilters[k] = true
			}
		} else {
			fields := getAllQueryName(*selectPermission.Permission.Filter)
			for _, k := range fields {
				manabieFilters[k] = true
			}
		}
	}

	otherRoleFieldsFilter := sortMapStringType(otherFilters)
	manabieRoleFieldsFilter := sortMapStringType(manabieFilters)

	return check2ArrayStringSameOrder(otherRoleFieldsFilter, manabieRoleFieldsFilter)
}

func (s *suite) filtersIncludedAllFiltersFromOtherRoles(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	body := &HasuraBody{}
	err := json.Unmarshal(stepState.HasuraBody, body)

	if err != nil {
		return ctx, err
	}

	for _, table := range body.Tables {
		if table.SelectPermissions == nil {
			continue
		}
		isEqual := compareManabieFilterAndOtherFilters(*table.SelectPermissions)
		if !isEqual {
			return ctx, fmt.Errorf("Table %s: MANABIE role filters is not correct in service %s", table.Table.Name, stepState.HasuraURL)
		}
	}
	return ctx, nil
}
