package field

import (
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

type Date struct {
	status Status
	value  time.Time
}

func NewUndefinedDate() Date {
	return Date{
		status: StatusUndefined,
	}
}

func NewNullDate() Date {
	return Date{
		status: StatusNull,
	}
}

func NewDate(value time.Time) Date {
	return Date{
		status: StatusPresent,
		value:  value,
	}
}

func (field Date) Status() Status {
	return field.status
}

func (field Date) Ptr() *Date {
	ptr := &field
	return ptr
}

func (field Date) Date() time.Time {
	switch field.Status() {
	case StatusUndefined, StatusNull:
		return time.Time{}
	default:
		return field.value
	}
}

func (field Date) MarshalJSON() ([]byte, error) {
	if IsUndefined(field) {
		return json.Marshal(nil)
	}
	return json.Marshal(field.Date().Format(constant.DateLayout))
}

func (field *Date) UnmarshalJSON(data []byte) error {
	var value *string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "Date",
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
			FieldName:    "Date",
			Reason:       err.Error(),
			InvalidValue: *value,
		}
	}
	field.value = date
	field.status = StatusPresent
	return nil
}

func (field *Date) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	date, err := time.Parse(constant.DateLayout, data)
	if err != nil {
		return InvalidDataError{
			Method:       "time.Parse",
			FieldName:    "Date",
			Reason:       err.Error(),
			InvalidValue: data,
		}
	}

	field.value = date
	field.status = StatusPresent
	return nil
}

func (field *Date) SetNull() {
	*field = NewNullDate()
}

type Dates []Date

func (dates Dates) Dates() []time.Time {
	times := make([]time.Time, 0, len(dates))
	for _, date := range dates {
		times = append(times, date.Date())
	}
	return times
}
