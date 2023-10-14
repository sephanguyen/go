package rls

func parse2ArrayInterface(field interface{}) []interface{} {
	nestedFields, ok := field.([]interface{})
	if ok {
		return nestedFields
	}
	return []interface{}{}
}

func getUniqueConditions(conditions []interface{}) []interface{} {
	var uniqueConditions []interface{}
	for _, condition := range conditions {
		if !checkQueryExisted(uniqueConditions, condition) {
			uniqueConditions = append(uniqueConditions, condition)
		}
	}
	return uniqueConditions
}

func parseInterfaceConditions(filter interface{}) []interface{} {
	var conditions []interface{}
	fields, _ := filter.(map[interface{}]interface{})
	for k, field := range fields {
		key, _ := k.(string)
		if len(fields) == 1 && key != "_and" && key != "_or" {
			conditions = append(conditions, fields)
			break
		} else {
			conditions = append(conditions, parse2ArrayInterface(field)...)
		}
	}
	return conditions
}

func getAllUniqueConditions(filters []interface{}) []interface{} {
	var conditions []interface{}
	for _, filter := range filters {
		conditions = append(conditions, parseInterfaceConditions(filter)...)
	}

	return getUniqueConditions(conditions)
}

func getAllSelectPermission(selectPermissions []HasuraSelectPermissions) []interface{} {
	var filters []interface{}
	for _, selectPermission := range selectPermissions {
		if filter := selectPermission.Permission.Filter; filter != nil {
			filters = append(filters, *filter)
		}
	}
	return getAllUniqueConditions(filters)
}

func getAllInsertPermission(insertPermissions []HasuraInsertPermissions) []interface{} {
	var checks []interface{}
	for _, insertPermission := range insertPermissions {
		if check := insertPermission.Permission.Check; check != nil {
			checks = append(checks, *check)
		}
	}
	return getAllUniqueConditions(checks)
}

func getAllDeletePermission(deletePermissions []HasuraDeletePermissions) []interface{} {
	var checks []interface{}
	for _, deletePermission := range deletePermissions {
		if check := deletePermission.Permission.Check; check != nil {
			checks = append(checks, *check)
		}
	}
	return getAllUniqueConditions(checks)
}
