package domain

import (
	"encoding/json"
	"fmt"

	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
)

type FormConfig struct {
	FormConfigID   string
	FormConfigData *FormConfigData
}

func NewFormConfigByPartnerFormConfig(partnerCfg *PartnerFormConfig) (*FormConfig, error) {
	res := &FormConfig{
		FormConfigID: partnerCfg.FormConfigID,
	}

	formConfigData := &FormConfigData{}
	if err := json.Unmarshal(partnerCfg.FormConfigData, formConfigData); err != nil {
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
			}
			currentSectionIDs[section.SectionID] = true
			// check fields of a section
			for _, field := range section.Fields {
				if len(field.FieldID) == 0 {
					return fmt.Errorf("field's id of section %s could not be empty", section.SectionID)
				}
				if ok := currentFieldIDs[field.FieldID]; ok {
					return fmt.Errorf("field id %s of section %s be duplicated", field.FieldID, section.SectionID)
				}
				currentFieldIDs[field.FieldID] = true

				if len(field.ValueType) == 0 {
					return fmt.Errorf("%s's value type of section %s could not be empty", field.FieldID, section.SectionID)
				}

				if field.ValueType != lesson_report_consts.FieldValueTypeInt &&
					field.ValueType != lesson_report_consts.FieldValueTypeString &&
					field.ValueType != lesson_report_consts.FieldValueTypeBool &&
					field.ValueType != lesson_report_consts.FieldValueTypeIntArray &&
					field.ValueType != lesson_report_consts.FieldValueTypeStringArray &&
					field.ValueType != lesson_report_consts.FieldValueTypeIntSet &&
					field.ValueType != lesson_report_consts.FieldValueTypeStringSet &&
					field.ValueType != lesson_report_consts.FieldValueTypeNull {
					return fmt.Errorf("field.ValueType was not defined")
				}

				// check system defined fields
				switch field.FieldID {
				case string(lesson_report_consts.SystemDefinedFieldAttendanceStatus), string(lesson_report_consts.SystemDefinedFieldAttendanceRemark):
					if field.ValueType != lesson_report_consts.FieldValueTypeString {
						return fmt.Errorf("expected %s's value type of section %s is %s but got %s", field.FieldID, section.SectionID, lesson_report_consts.FieldValueTypeString, field.ValueType)
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
	FieldID    string                              `json:"field_id"`
	ValueType  lesson_report_consts.FieldValueType `json:"value_type"`
	IsRequired bool                                `json:"is_required"`
	IsInternal bool                                `json:"is_internal"`
}
