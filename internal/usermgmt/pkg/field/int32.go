package field

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Int32 struct {
	status Status
	value  int32
}

func NewUndefinedInt32() Int32 {
	return Int32{
		status: StatusUndefined,
	}
}

func (field *Int32) MarshalCSV() (string, error) {
	if field.status == StatusPresent {
		return fmt.Sprint(field.value), nil
	}
	return "", nil
}

func NewNullInt32() Int32 {
	return Int32{
		status: StatusNull,
	}
}

func NewInt32(value int32) Int32 {
	return Int32{
		status: StatusPresent,
		value:  value,
	}
}

func (field Int32) Int32() int32 {
	switch field.status {
	case StatusUndefined, StatusNull:
		return 0
	default:
		return field.value
	}
}

func (field Int32) Status() Status {
	return field.status
}

func (field Int32) Ptr() *Int32 {
	ptr := &field
	return ptr
}

func (field Int32) MarshalJSON() ([]byte, error) {
	if IsUndefined(field) {
		return json.Marshal(nil)
	}
	return json.Marshal(field.Int32())
}

func (field *Int32) UnmarshalJSON(data []byte) error {
	var value *int32
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "Int32",
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

func (field *Int32) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	value, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		field.SetNull()
		return InvalidDataError{
			Method:       "strconv.ParseInt",
			FieldName:    "Int32",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}

	field.value = int32(value)
	field.status = StatusPresent
	return nil
}

func (field *Int32) SetNull() {
	*field = NewNullInt32()
}

type Int32s []Int32

func (int32s Int32s) Int32s() []int32 {
	int32Values := make([]int32, 0, len(int32s))
	for _, int32Field := range int32s {
		int32Values = append(int32Values, int32Field.Int32())
	}
	return int32Values
}
