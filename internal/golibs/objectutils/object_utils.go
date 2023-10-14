package objectutils

import (
	"fmt"
	"reflect"
	"strings"
)

func SafeGetObject[T any](fGet func() *T) *T {
	val := fGet()
	if val == nil {
		val = new(T)
	}
	return val
}

func ExtractFieldMapWithSuffix[T any](config interface{}, fieldSuffix string) (map[string]T, error) {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be a struct")
	}
	t := reflect.TypeOf(config)
	ret := map[string]T{}
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		if strings.HasSuffix(fieldName, fieldSuffix) {
			key := fieldName[:len(fieldName)-len(fieldSuffix)]
			fv := v.FieldByName(fieldName)
			if !fv.IsValid() {
				return nil, fmt.Errorf("field %s not found", fieldName)
			}
			out, ok := fv.Interface().(T)
			if !ok {
				tv := fv.Type()
				return nil, fmt.Errorf("expected field %q to be of \"%T\" type, got \"%s.%s\"", fieldName, out, tv.PkgPath(), tv.Name())
			}
			ret[key] = out
		}
	}
	return ret, nil
}
