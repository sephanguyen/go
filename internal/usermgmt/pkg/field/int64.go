package field

import (
	"encoding/json"
	"strconv"
)

type Int64 struct {
	status Status
	value  int64
}

func NewUndefinedInt64() Int64 {
	return Int64{
		status: StatusUndefined,
	}
}

func NewNullInt64() Int64 {
	return Int64{
		status: StatusNull,
	}
}

func NewInt64(value int64) Int64 {
	return Int64{
		status: StatusPresent,
		value:  value,
	}
}

func (field Int64) Int64() int64 {
	switch field.status {
	case StatusUndefined, StatusNull:
		return 0
	default:
		return field.value
	}
}

func (field Int64) Status() Status {
	return field.status
}

func (field Int64) Ptr() *Int64 {
	ptr := &field
	return ptr
}

func (field Int64) MarshalJSON() ([]byte, error) {
	if IsUndefined(field) {
		return json.Marshal(nil)
	}
	return json.Marshal(field.Int64())
}

func (field *Int64) UnmarshalJSON(data []byte) error {
	var value *int64
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "Int64",
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

func (field *Int64) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	value, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		field.SetNull()
		return InvalidDataError{
			Method:       "strconv.ParseInt",
			FieldName:    "Int64",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}

	field.value = value
	field.status = StatusPresent
	return nil
}

func (field *Int64) SetNull() {
	*field = NewNullInt64()
}

type Int64s []Int64

func (int64s Int64s) Int64s() []int64 {
	int64Values := make([]int64, 0, len(int64s))
	for _, int64Field := range int64s {
		int64Values = append(int64Values, int64Field.Int64())
	}
	return int64Values
}
