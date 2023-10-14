package field

import (
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

type Time struct {
	status Status
	value  time.Time
}

func NewUndefinedTime() Time {
	return Time{
		status: StatusUndefined,
	}
}

func NewNullTime() Time {
	return Time{
		status: StatusNull,
	}
}

func (field Time) Time() time.Time {
	switch field.Status() {
	case StatusUndefined, StatusNull:
		return time.Time{}
	default:
		return field.value
	}
}

func NewTime(value time.Time) Time {
	return Time{
		status: StatusPresent,
		value:  value,
	}
}

func (field Time) Status() Status {
	return field.status
}

func (field Time) Ptr() *Time {
	ptr := &field
	return ptr
}

func (field Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(field.Time())
}

func (field *Time) UnmarshalJSON(data []byte) error {
	var value *string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "Time",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}
	if value == nil || *value == "" {
		field.status = StatusNull
		return nil
	}
	date, err := time.Parse(constant.DateLayout, *value)
	if err != nil {
		return InvalidDataError{
			Method:       "time.Parse",
			FieldName:    "Time",
			Reason:       err.Error(),
			InvalidValue: *value,
		}
	}
	field.value = date
	field.status = StatusPresent
	return nil
}

func (field *Time) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	value, err := time.Parse(constant.DateLayout, data)
	if err != nil {
		field.SetNull()
		return InvalidDataError{
			Method:       "time.Parse",
			FieldName:    "Time",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}

	field.value = value
	field.status = StatusPresent
	return nil
}

func (field *Time) SetNull() {
	*field = NewNullTime()
}

type Times []Time

func (times Times) Times() []time.Time {
	timeValues := make([]time.Time, 0, len(times))
	for _, timeField := range times {
		timeValues = append(timeValues, timeField.Time())
	}
	return timeValues
}
