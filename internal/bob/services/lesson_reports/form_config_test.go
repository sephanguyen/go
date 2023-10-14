package services

import (
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormConfig(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                     string
		partnerFormConfig        *entities.PartnerFormConfig
		hasError                 bool
		isValid                  bool
		expectedConfig           *FormConfig
		expectedFieldIDs         []string
		expectedRequiredFieldIDs []string
	}{
		{
			name: "full fields",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-0",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "attendance_status",
									"label": "attendance status",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_remark",
									"label": "attendance remark",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_notice",
									"label": "attendance notice",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_reason",
									"label": "attendance reason",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "attendance_note",
									"label": "attendance note",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						},
						{
							"section_id": "section-id-3",
							"section_name": "section-name-3",
							"fields": [
								{
									"field_id": "field-id-5",
									"label": "display name 5",
									"value_type": "VALUE_TYPE_BOOL"
								},
								{
									"field_id": "field-id-6",
									"label": "display name 6",
									"value_type": "VALUE_TYPE_STRING_ARRAY",
									"is_required": true
								},
								{
									"field_id": "field-id-7",
									"label": "display name 7",
									"value_type": "VALUE_TYPE_INT_SET",
									"is_required": true
								},
								{
									"field_id": "field-id-8",
									"label": "display name 8",
									"value_type": "VALUE_TYPE_NULL",
									"is_required": false
								}
							]
						}
					]
				}
			`),
			},
			isValid: true,
			expectedConfig: &FormConfig{
				FormConfigID: "form-config-id-1",
				FormConfigData: &FormConfigData{
					Sections: []*FormConfigSection{
						{
							SectionID: "section-id-0",
							Fields: []*FormConfigField{
								{
									FieldID:    "attendance_status",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: false,
								},
								{
									FieldID:    "attendance_remark",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: false,
								},
								{
									FieldID:    "attendance_notice",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: false,
								},
								{
									FieldID:    "attendance_reason",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: false,
								},
								{
									FieldID:    "attendance_note",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: true,
								},
							},
						},
						{
							SectionID: "section-id-1",
							Fields: []*FormConfigField{
								{
									FieldID:    "field-id-1",
									ValueType:  "VALUE_TYPE_INT",
									IsRequired: true,
								},
								{
									FieldID:    "field-id-2",
									ValueType:  "VALUE_TYPE_STRING",
									IsRequired: true,
								},
							},
						},
						{
							SectionID: "section-id-2",
							Fields: []*FormConfigField{
								{
									FieldID:    "field-id-3",
									ValueType:  "VALUE_TYPE_STRING_SET",
									IsRequired: false,
								},
								{
									FieldID:    "field-id-4",
									ValueType:  "VALUE_TYPE_INT_ARRAY",
									IsRequired: false,
								},
							},
						},
						{
							SectionID: "section-id-3",
							Fields: []*FormConfigField{
								{
									FieldID:    "field-id-5",
									ValueType:  "VALUE_TYPE_BOOL",
									IsRequired: false,
								},
								{
									FieldID:    "field-id-6",
									ValueType:  "VALUE_TYPE_STRING_ARRAY",
									IsRequired: true,
								},
								{
									FieldID:    "field-id-7",
									ValueType:  "VALUE_TYPE_INT_SET",
									IsRequired: true,
								},
								{
									FieldID:    "field-id-8",
									ValueType:  "VALUE_TYPE_NULL",
									IsRequired: false,
								},
							},
						},
					},
				},
			},
			expectedFieldIDs: []string{
				"attendance_status",
				"attendance_remark",
				"field-id-1",
				"field-id-2",
				"field-id-3",
				"field-id-4",
				"field-id-5",
				"field-id-6",
				"field-id-7",
				"field-id-8",
				"attendance_notice",
				"attendance_reason",
				"attendance_note",
			},
			expectedRequiredFieldIDs: []string{"field-id-1", "field-id-2", "field-id-6", "field-id-7", "attendance_note"},
		},
		{
			name: "form config data is null",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID:   database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(nil),
			},
			isValid: true,
			expectedConfig: &FormConfig{
				FormConfigID:   "form-config-id-1",
				FormConfigData: &FormConfigData{},
			},
		},
		{
			name: "form config data is empty",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID:   database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`{}`),
			},
			isValid: true,
		},
		{
			name: "there one empty section id",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "duplicated section id",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "there one empty field id",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "duplicated field id",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-1",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "form config id empty",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_INT",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "missing value type",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "non-existing value type",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_UNDEFINED",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "wrong value type of attendance_status field",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-0",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "attendance_status",
									"label": "attendance status",
									"value_type": "VALUE_TYPE_INT",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
		{
			name: "wrong value type of attendance_remark field",
			partnerFormConfig: &entities.PartnerFormConfig{
				FormConfigID: database.Text("form-config-id-1"),
				FormConfigData: database.JSONB(`
				{
					"sections": [
						{
							"section_id": "section-id-0",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "attendance_remark",
									"label": "attendance remark",
									"value_type": "VALUE_TYPE_STRING_SET",
									"is_required": false,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-1",
							"section_name": "section-name",
							"fields": [
								{
									"field_id": "field-id-1",
									"label": "display name 1",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-2",
									"label": "display name 2",
									"value_type": "VALUE_TYPE_STRING",
									"is_required": true,
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								}
							]
						},
						{
							"section_id": "section-id-2",
							"section_name": "section-name-2",
							"fields": [
								{
									"field_id": "field-id-3",
									"label": "display name 3",
									"value_type": "VALUE_TYPE_STRING_SET",
									"component_props": {},
									"component_config": {
										"type": "DynamicFieldsComponentType.AUTOCOMPLETE"
									},
									"display_config": {
										"is_label": true,
										"size": {}
									}
								},
								{
									"field_id": "field-id-4",
									"label": "display name 4",
									"value_type": "VALUE_TYPE_INT_ARRAY",
									"is_required": false,
									"component_props": {}
								}
							]
						}
					]
				}
			`),
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewFormConfigByPartnerFormConfig(tc.partnerFormConfig)
			if tc.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if !tc.isValid {
				assert.Error(t, actual.IsValid())
				return
			}
			assert.NoError(t, actual.IsValid())
			if tc.expectedConfig != nil {
				assert.EqualValues(t, *tc.expectedConfig, *actual)
				actualIDs := actual.GetFieldIDs()
				assert.Len(t, actualIDs, len(tc.expectedFieldIDs))
				assert.ElementsMatch(t, tc.expectedFieldIDs, actualIDs)

				actualRequiredFields := actual.GetRequiredFieldsMap()
				assert.Len(t, actualRequiredFields, len(tc.expectedRequiredFieldIDs))
				for _, expectedID := range tc.expectedRequiredFieldIDs {
					assert.NotNil(t, actualRequiredFields[expectedID])
				}
			}
		})
	}
}
