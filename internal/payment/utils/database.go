package utils

func GetScanFields(fieldNames []string, fieldValues []interface{}, reqFieldNames []string) []interface{} {
	n := len(fieldValues)
	if len(reqFieldNames) < n {
		n = len(reqFieldNames)
	}

	result := make([]interface{}, 0, n)
	for _, reqname := range reqFieldNames {
		for i, name := range fieldNames {
			if name == reqname {
				result = append(result, fieldValues[i])
				break
			}
		}
	}

	return result
}
