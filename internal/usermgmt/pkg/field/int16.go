package field

import (
	"encoding/json"
	"strconv"
)

type Int16 struct {
	status Status
	value  int16
}

func NewUndefinedInt16() Int16 {
	return Int16{
		status: StatusUndefined,
	}
}

func NewNullInt16() Int16 {
	return Int16{
		status: StatusNull,
	}
}

func NewInt16(value int16) Int16 {
	return Int16{
		status: StatusPresent,
		value:  value,
	}
}

func (field Int16) Int16() int16 {
	switch field.status {
	case StatusUndefined, StatusNull:
		return 0
	default:
		return field.value
	}
}

func (field Int16) Status() Status {
	return field.status
}

func (field Int16) Ptr() *Int16 {
	ptr := &field
	return ptr
}

func (field Int16) MarshalJSON() ([]byte, error) {
	return json.Marshal(field.Int16())
}

func (field *Int16) UnmarshalJSON(data []byte) error {
	var value *int16
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "Int16",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}
	if value == nil {
		field.status = StatusNull
		return nil
	}
	field.value = *value
	field.status = StatusPresent
	return nil
}

func (field *Int16) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	value, err := strconv.ParseInt(data, 10, 16)
	if err != nil {
		field.SetNull()
		return InvalidDataError{
			Method:       "strconv.ParseInt",
			FieldName:    "Int16",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}

	field.value = int16(value)
	field.status = StatusPresent
	return nil
}

func (field *Int16) SetNull() {
	*field = NewNullInt16()
}

type Int16s []Int16

func (int16s Int16s) Int16s() []int16 {
	int16Values := make([]int16, 0, len(int16s))
	for _, int16Field := range int16s {
		int16Values = append(int16Values, int16Field.Int16())
	}
	return int16Values
}
