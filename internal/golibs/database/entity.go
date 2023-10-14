package database

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

// setter requires a Set method, which is implemented by pgtype.Value
type setter interface {
	Set(src interface{}) error
}

// Entity generic interface
type Entity interface {
	FieldMap() ([]string, []interface{})
	TableName() string
}

// Entities ...
type Entities interface {
	Add() Entity
}

// AllNullEntity sets null value of all fields in entity e if the field has a Set() method.
// The affected fields will have their status set to pgtype.Null. AllNullEntity is mainly used
// to initialize data to be inserted into database, a situation in which pytype.Undefined status
// is not accepted.
//
// Note that user-defined entities (e.g. City, District) are not affected.
func AllNullEntity(e Entity) {
	_, fields := e.FieldMap()
	for _, field := range fields {
		f, ok := field.(setter)
		if ok {
			f.Set(nil)
		}
	}
}

func AllRandomEntity(e Entity) {
	fnames, fields := e.FieldMap()
	for idx, field := range fields {
		var val interface{}
		switch field.(type) {
		case *pgtype.Text:
			val = idutil.ULIDNow()
		case *pgtype.TextArray:
			val = []string{idutil.ULIDNow()}
		case *pgtype.Int4:
			val = int32(rand.Int())
		case *pgtype.Int8:
			val = int64(rand.Int())
		case *pgtype.Timestamptz:
			val = time.Now()
		case *pgtype.Bool:
			val = rand.Intn(2) < 1
		case *pgtype.JSONB:
			val = fmt.Sprintf("{\"id\": \"%s\"}", idutil.ULIDNow())
		case *pgtype.Int2:
			val = int16(rand.Int())
		case *pgtype.Date:
			val = time.Now()
		default:
			panic(fmt.Sprintf("add your implementation for type %T here", field))
		}
		f, ok := field.(setter)
		if ok {
			err := f.Set(val)
			if err != nil {
				panic(fmt.Sprintf("cannot set value for field %s with type %T", fnames[idx], field))
			}
		}
	}
}

// GetFieldNames returns all field names from entity e.
func GetFieldNames(e Entity) []string {
	fieldName, _ := e.FieldMap()
	return fieldName
}

// GetFieldNamesExcepts returns all field names from entity e excepts exceptedFieldNames.
func GetFieldNamesExcepts(e Entity, ignoredFieldNames []string) []string {
	numberIgnoredFieldNames := len(ignoredFieldNames)
	fieldNames, _ := e.FieldMap()
	if numberIgnoredFieldNames == 0 {
		return fieldNames
	}
	mapIgnoredFieldNames := make(map[string]bool)
	for _, exceptedFieldName := range ignoredFieldNames {
		mapIgnoredFieldNames[exceptedFieldName] = true
	}
	result := make([]string, 0, len(fieldNames)-numberIgnoredFieldNames)
	for _, fieldName := range fieldNames {
		if mapIgnoredFieldNames[fieldName] {
			continue
		}
		result = append(result, fieldName)
	}
	return result
}

// GetScanFields returns field values requested in reqlist. Matched fields are returned
// in the same order as in reqlist. Unmatched fields are ignored.
func GetScanFields(e Entity, reqlist []string) []interface{} {
	allNames, allValues := e.FieldMap()

	// Allocate enough capacity for result slice
	n := len(allValues)
	if len(reqlist) < n {
		n = len(reqlist)
	}

	result := make([]interface{}, 0, n)
	for _, reqname := range reqlist {
		for i, name := range allNames {
			if name == reqname {
				result = append(result, allValues[i])
				break
			}
		}
	}

	return result
}

// countFieldNumber returns the number of pgtype fields/subfields in v.
func countFieldNumber(v reflect.Value) int {
	count := 0
	if v.Type().PkgPath() == "github.com/jackc/pgtype" {
		return 1
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		count += countFieldNumber(f)
	}
	return count
}

// CheckEntityDefinition checks if e.FieldMap() returns the correct number of fields and the values
// returned are pointer type. It also checks if there are any duplicated fields in e.FieldMap().
func CheckEntityDefinition(e Entity) error {
	v := reflect.ValueOf(e).Elem()
	typeOfT := v.Type()

	dbFields := 0
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if !f.CanSet() {
			continue
		}
		if f.Type().PkgPath() == "github.com/jackc/pgtype" {
			dbFields++
		}
		if f.Type().PkgPath() == "github.com/manabie-com/backend/internal/eureka/entities" {
			dbFields += countFieldNumber(f)
		}
	}
	fieldNames, fieldValues := e.FieldMap()
	if dbFields != len(fieldNames) {
		return fmt.Errorf("%s.FieldMap() returned %d field names, expected %d", typeOfT.Name(), len(fieldNames), dbFields)
	}
	if dbFields != len(fieldValues) {
		return fmt.Errorf("%s.FieldMap() returned %d field values, expected %d", typeOfT.Name(), len(fieldValues), dbFields)
	}
	if len(e.TableName()) == 0 {
		return fmt.Errorf("%s.TableName() returned empty", typeOfT.Name())
	}

	for k, f := range fieldValues {
		if reflect.ValueOf(f).Kind() != reflect.Ptr {
			return fmt.Errorf("field %s of %s is not a pointer", fieldNames[k], typeOfT.Name())
		}
	}

	for i := 0; i < dbFields; i++ {
		for j := i + 1; j < dbFields; j++ {
			if fieldNames[i] == fieldNames[j] {
				return fmt.Errorf("duplicated field name %q (index %d and %d) in entity %s", fieldNames[i], i, j, typeOfT.Name())
			}

			if fieldValues[i] == fieldValues[j] {
				return fmt.Errorf("duplicated pointer value (index %d and %d) in entity %s", i, j, typeOfT.Name())
			}
		}
	}

	return nil
}

// CheckEntitiesDefinition verifies that es is a slice type and correctly implements Add.
func CheckEntitiesDefinition(es Entities) error {
	esVal := reflect.ValueOf(es).Elem()
	if esVal.Kind() != reflect.Slice {
		return fmt.Errorf("%s's underlying type must be a slice", esVal.Type())
	}

	oldLen := esVal.Len()
	_ = es.Add()
	newLen := esVal.Len()
	if oldLen+1 != newLen {
		return fmt.Errorf("%s.Add fails to add a new element (oldLen: %d, newLen: %d)", esVal.Type(), oldLen, newLen)
	}
	return nil
}

// GetFieldMapExcept returns all field names from entity e excepts exceptedFieldNames.
func GetFieldMapExcept(e Entity, fieldNamesToIgnore ...string) ([]string, []interface{}) {
	ignoredFieldNames := make(map[string]bool, len(fieldNamesToIgnore))
	for _, fieldNameToIgnore := range fieldNamesToIgnore {
		ignoredFieldNames[fieldNameToIgnore] = true
	}

	fieldNamesToFilter, fieldValuesToFilter := e.FieldMap()

	filteredFieldNames := make([]string, 0, len(fieldNamesToFilter)-len(fieldNamesToIgnore))
	filteredFieldValues := make([]interface{}, 0, len(fieldNamesToFilter)-len(fieldNamesToIgnore))

	for i := range fieldNamesToFilter {
		fieldNameToFilter := fieldNamesToFilter[i]
		fieldValueToFilter := fieldValuesToFilter[i]

		if ignoredFieldNames[fieldNameToFilter] {
			continue
		}
		filteredFieldNames = append(filteredFieldNames, fieldNameToFilter)
		filteredFieldValues = append(filteredFieldValues, fieldValueToFilter)
	}

	return filteredFieldNames, filteredFieldValues
}

func SerializeFields(e Entity, prefix string) string {
	fieldNames, _ := e.FieldMap()

	prefixedFieldNames := make([]string, 0)
	for _, name := range fieldNames {
		prefixedFieldNames = append(prefixedFieldNames, fmt.Sprintf("%s.%s", prefix, name))
	}

	return strings.Join(prefixedFieldNames, ", ")
}
