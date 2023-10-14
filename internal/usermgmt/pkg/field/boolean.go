package field

type Boolean struct {
	status Status
	value  bool
}

func NewUndefinedBoolean() Boolean {
	return Boolean{
		status: StatusUndefined,
	}
}

func NewNullBoolean() Boolean {
	return Boolean{
		status: StatusNull,
	}
}

func NewBoolean(value bool) Boolean {
	return Boolean{
		status: StatusPresent,
		value:  value,
	}
}

func (field Boolean) Status() Status {
	return field.status
}

func (field Boolean) Ptr() *Boolean {
	ptr := &field
	return ptr
}

func (field Boolean) Boolean() bool {
	switch field.Status() {
	case StatusUndefined, StatusNull:
		return false
	default:
		return field.value
	}
}

type Booleans []Boolean

func (booleans Booleans) Booleans() []bool {
	booleanValues := make([]bool, 0, len(booleans))
	for _, booleanField := range booleans {
		booleanValues = append(booleanValues, booleanField.Boolean())
	}
	return booleanValues
}
