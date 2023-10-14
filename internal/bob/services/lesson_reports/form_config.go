package services

import (
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
)

type FieldValueType string
type SystemDefinedField string

const (
	FieldValueTypeInt         FieldValueType = "VALUE_TYPE_INT"
	FieldValueTypeString      FieldValueType = "VALUE_TYPE_STRING"
	FieldValueTypeBool        FieldValueType = "VALUE_TYPE_BOOL"
	FieldValueTypeIntArray    FieldValueType = "VALUE_TYPE_INT_ARRAY"
	FieldValueTypeStringArray FieldValueType = "VALUE_TYPE_STRING_ARRAY"
	FieldValueTypeIntSet      FieldValueType = "VALUE_TYPE_INT_SET"
	FieldValueTypeStringSet   FieldValueType = "VALUE_TYPE_STRING_SET"
	FieldValueTypeNull        FieldValueType = "VALUE_TYPE_NULL"

	SystemDefinedFieldAttendanceStatus SystemDefinedField = "attendance_status"
	SystemDefinedFieldAttendanceRemark SystemDefinedField = "attendance_remark"
	SystemDefinedFieldAttendanceNotice SystemDefinedField = "attendance_notice"
	SystemDefinedFieldAttendanceReason SystemDefinedField = "attendance_reason"
	SystemDefinedFieldAttendanceNote   SystemDefinedField = "attendance_note"
)

type FormConfig struct {
	FormConfigID   string
	FormConfigData *FormConfigData
}

func NewFormConfigByPartnerFormConfig(cf *entities.PartnerFormConfig) (*FormConfig, error) {
	res := &FormConfig{
		FormConfigID: cf.FormConfigID.String,
	}

	formConfigData := &FormConfigData{}
	if err := cf.FormConfigData.AssignTo(formConfigData); err != nil {
		return nil, fmt.Errorf("could not unmarshal form config data: %v", err)
	}
	res.FormConfigData = formConfigData

	return res, nil
}

func (f *FormConfig) IsValid() error {
	if len(f.FormConfigID) == 0 {
		return fmt.Errorf("form_config_id could not be empty")
	}

	if f.FormConfigData != nil {
		currentSectionIDs := make(map[string]bool)
		currentFieldIDs := make(map[string]bool)
		for _, section := range f.FormConfigData.Sections {
			if len(section.SectionID) == 0 {
				return fmt.Errorf("section id could not be empty")
			}
			if ok := currentSectionIDs[section.SectionID]; ok {
				return fmt.Errorf("section id %s be duplicated", section.SectionID)
			} else {
				currentSectionIDs[section.SectionID] = true
			}

			// check fields of a section
			for _, field := range section.Fields {
				if len(field.FieldID) == 0 {
					return fmt.Errorf("field's id of section %s could not be empty", section.SectionID)
				}
				if ok := currentFieldIDs[field.FieldID]; ok {
					return fmt.Errorf("field id %s of section %s be duplicated", field.FieldID, section.SectionID)
				} else {
					currentFieldIDs[field.FieldID] = true
				}

				if len(field.ValueType) == 0 {
					return fmt.Errorf("%s's value type of section %s could not be empty", field.FieldID, section.SectionID)
				}

				if field.ValueType != FieldValueTypeInt &&
					field.ValueType != FieldValueTypeString &&
					field.ValueType != FieldValueTypeBool &&
					field.ValueType != FieldValueTypeIntArray &&
					field.ValueType != FieldValueTypeStringArray &&
					field.ValueType != FieldValueTypeIntSet &&
					field.ValueType != FieldValueTypeStringSet &&
					field.ValueType != FieldValueTypeNull {
					return fmt.Errorf("field.ValueType was not defined")
				}

				// check system defined fields
				switch field.FieldID {
				case string(SystemDefinedFieldAttendanceStatus),
					string(SystemDefinedFieldAttendanceRemark),
					string(SystemDefinedFieldAttendanceNotice),
					string(SystemDefinedFieldAttendanceReason),
					string(SystemDefinedFieldAttendanceNote):
					if field.ValueType != FieldValueTypeString {
						return fmt.Errorf("expected %s's value type of section %s is %s but got %s", field.FieldID, section.SectionID, FieldValueTypeString, field.ValueType)
					}
				}
			}
		}
	}

	return nil
}

func (f *FormConfig) GetFieldIDs() []string {
	fieldIDs := make(map[string]bool)
	for _, section := range f.FormConfigData.Sections {
		for _, field := range section.Fields {
			fieldIDs[field.FieldID] = true
		}
	}

	ids := make([]string, 0, len(fieldIDs))
	for id := range fieldIDs {
		ids = append(ids, id)
	}

	return ids
}

func (f *FormConfig) GetFieldsMap() map[string]*FormConfigField {
	if f.FormConfigData == nil || f.FormConfigData.Sections == nil {
		return nil
	}
	fields := make(map[string]*FormConfigField)
	for _, section := range f.FormConfigData.Sections {
		if section.Fields == nil {
			continue
		}
		for i, field := range section.Fields {
			fields[field.FieldID] = section.Fields[i]
		}
	}

	return fields
}

func (f *FormConfig) GetRequiredFieldsMap() map[string]*FormConfigField {
	if f.FormConfigData == nil || f.FormConfigData.Sections == nil {
		return nil
	}
	fields := make(map[string]*FormConfigField)
	for _, section := range f.FormConfigData.Sections {
		if section.Fields == nil {
			continue
		}
		for i, field := range section.Fields {
			if field.IsRequired {
				fields[field.FieldID] = section.Fields[i]
			}
		}
	}

	return fields
}

type FormConfigData struct {
	Sections []*FormConfigSection `json:"sections"`
}

func (f *FormConfigData) GetFieldByID(id string) (*FormConfigField, bool) {
	for _, section := range f.Sections {
		if v, ok := section.Fields.GetFieldByID(id); ok {
			return v, true
		}
	}

	return nil, false
}

type FormConfigSection struct {
	SectionID string           `json:"section_id"`
	Fields    FormConfigFields `json:"fields"`
}

type FormConfigFields []*FormConfigField

func (f FormConfigFields) GetFieldByID(id string) (*FormConfigField, bool) {
	for _, field := range f {
		if field.FieldID == id {
			return field, true
		}
	}
	return nil, false
}

type FormConfigField struct {
	FieldID    string         `json:"field_id"`
	ValueType  FieldValueType `json:"value_type"`
	IsRequired bool           `json:"is_required"`
}

type AttributeValue struct {
	Int         int
	String      string
	Bool        bool
	IntArray    []int
	StringArray []string
	IntSet      []int
	StringSet   []string
}

func (l *AttributeValue) SetInt(v int) {
	l.Int = v
}

func (l *AttributeValue) SetString(v string) {
	l.String = v
}

func (l *AttributeValue) SetBool(v bool) {
	l.Bool = v
}

func (l *AttributeValue) SetIntArray(v []int) {
	l.IntArray = v
}

func (l *AttributeValue) SetStringArray(v []string) {
	l.StringArray = v
}

func (l *AttributeValue) SetIntSet(v []int) {
	intMap := make(map[int]bool)
	l.IntSet = make([]int, 0, len(v))
	for i := range v {
		if _, ok := intMap[v[i]]; ok {
			continue
		}
		l.IntSet = append(l.IntSet, v[i])
		intMap[v[i]] = true
	}
}

func (l *AttributeValue) SetStringSet(v []string) {
	stringMap := make(map[string]bool)
	l.StringSet = make([]string, 0, len(v))
	for i := range v {
		if _, ok := stringMap[v[i]]; ok {
			continue
		}
		l.StringSet = append(l.StringSet, v[i])
		stringMap[v[i]] = true
	}
}
