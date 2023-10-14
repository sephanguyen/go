package field

type Entity interface {
	FieldMap() ([]string, []interface{})
}

type setter interface {
	SetNull()
}

type Field interface {
	Status() Status
}

func IsUndefined(field Field) bool {
	if IsNil(field) {
		return false
	}
	if field.Status() == StatusUndefined {
		return true
	}
	return false
}

func IsNull(field Field) bool {
	if IsNil(field) {
		return false
	}
	if field.Status() == StatusNull {
		return true
	}
	return false
}

func IsNil(field Field) bool {
	switch value := field.(type) {
	case *Boolean:
		return value == nil
	case *String:
		return value == nil
	case *Int16:
		return value == nil
	case *Int32:
		return value == nil
	case *Int64:
		return value == nil
	case *Date:
		return value == nil
	case *Time:
		return value == nil
	case *TimeWithoutTz:
		return value == nil
	}
	return field == nil
}

func IsPresent(field Field) bool {
	if IsNil(field) {
		return false
	}
	if field.Status() == StatusPresent {
		return true
	}
	return false
}

func SetUndefinedFieldsToNull(entity Entity) {
	_, fields := entity.FieldMap()
	for _, field := range fields {
		f, ok := field.(Field)
		if !ok {
			continue
		}
		if IsUndefined(f) {
			f, ok := field.(setter)
			if ok {
				f.SetNull()
			}
		}
	}
}
