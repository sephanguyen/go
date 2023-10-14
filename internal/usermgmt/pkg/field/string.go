package field

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
)

type String struct {
	status Status
	value  string
}

func NewUndefinedString() String {
	return String{
		status: StatusUndefined,
	}
}

func NewNullString() String {
	return String{
		status: StatusNull,
	}
}

func NewString(value string) String {
	return String{
		status: StatusPresent,
		value:  value,
	}
}

func (field String) Status() Status {
	return field.status
}

func (field String) RawValue() string {
	return field.value
}

func (field String) String() string {
	switch field.Status() {
	case StatusUndefined, StatusNull:
		return ""
	default:
		return field.RawValue()
	}
}

func (field String) IsEmpty() bool {
	return field.String() == ""
}

func (field String) Equal(otherField String) bool {
	if field.Status() != otherField.Status() {
		return false
	}
	if field.String() != otherField.String() {
		return false
	}
	return true
}

func (field String) Ptr() *String {
	return &field
}

func (field String) MarshalJSON() ([]byte, error) {
	if IsUndefined(field) {
		return json.Marshal(nil)
	}
	return json.Marshal(field.String())
}

func (field *String) UnmarshalJSON(data []byte) error {
	var value *string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return InvalidDataError{
			Method:       "json.Unmarshal",
			FieldName:    "String",
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

func (field *String) UnmarshalCSV(data string) error {
	if field == nil {
		return nil
	}

	if data == "" {
		field.SetNull()
		return nil
	}

	field.value = data
	field.status = StatusPresent
	return nil
}

func (field *String) SetNull() {
	*field = NewNullString()
}

func (field String) ToDate() (Date, error) {
	if field.IsEmpty() {
		return NewNullDate(), nil
	}

	date, err := time.Parse(constant.DateLayout, field.String())
	if err != nil {
		return NewNullDate(), err
	}

	return NewDate(date), nil
}

func (field String) TrimSpace() String {
	if !IsPresent(field) {
		return field
	}

	trimmedString := strings.TrimSpace(field.String())
	return NewString(trimmedString)
}

func (field String) ToLower() String {
	if !IsPresent(field) {
		return field
	}

	loweredString := strings.ToLower(field.String())
	return NewString(loweredString)
}

func ToSliceString(data []String) []string {
	result := make([]string, 0, len(data))
	for _, element := range data {
		result = append(result, element.String())
	}

	return result
}

type Strings []String

func (strings Strings) Strings() []string {
	stringValues := make([]string, 0, len(strings))
	for _, stringField := range strings {
		stringValues = append(stringValues, stringField.String())
	}
	return stringValues
}
