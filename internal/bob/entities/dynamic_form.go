package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type (
	FeatureName        string
	PartnerFormConfigs []*PartnerFormConfig
)

const (
	FeatureNameIndividualLessonReport FeatureName = "FEATURE_NAME_INDIVIDUAL_LESSON_REPORT"
	FeatureNameGroupLessonReport      FeatureName = "FEATURE_NAME_GROUP_LESSON_REPORT"
)

func (pfs *PartnerFormConfigs) Add() database.Entity {
	pf := &PartnerFormConfig{}
	*pfs = append(*pfs, pf)

	return pf
}

type PartnerFormConfig struct {
	FormConfigID   pgtype.Text
	PartnerID      pgtype.Int4
	FeatureName    pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	FormConfigData pgtype.JSONB
}

func (p *PartnerFormConfig) FieldMap() ([]string, []interface{}) {
	return []string{
			"form_config_id",
			"partner_id",
			"feature_name",
			"created_at",
			"updated_at",
			"deleted_at",
			"form_config_data",
		}, []interface{}{
			&p.FormConfigID,
			&p.PartnerID,
			&p.FeatureName,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
			&p.FormConfigData,
		}
}

func (p *PartnerFormConfig) TableName() string {
	return "partner_form_configs"
}

type PartnerDynamicFormFieldValue struct {
	DynamicFormFieldValueID pgtype.Text
	FieldID                 pgtype.Text
	LessonReportDetailID    pgtype.Text
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
	ValueType               pgtype.Text
	IntValue                pgtype.Int4
	StringValue             pgtype.Text
	BoolValue               pgtype.Bool
	StringArrayValue        pgtype.TextArray
	IntArrayValue           pgtype.Int4Array
	StringSetValue          pgtype.TextArray
	IntSetValue             pgtype.Int4Array
	FieldRenderGuide        pgtype.JSONB
}

func (p *PartnerDynamicFormFieldValue) FieldMap() ([]string, []interface{}) {
	return []string{
			"dynamic_form_field_value_id",
			"field_id",
			"lesson_report_detail_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"value_type",
			"int_value",
			"string_value",
			"bool_value",
			"string_array_value",
			"int_array_value",
			"string_set_value",
			"int_set_value",
			"field_render_guide",
		}, []interface{}{
			&p.DynamicFormFieldValueID,
			&p.FieldID,
			&p.LessonReportDetailID,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.DeletedAt,
			&p.ValueType,
			&p.IntValue,
			&p.StringValue,
			&p.BoolValue,
			&p.StringArrayValue,
			&p.IntArrayValue,
			&p.StringSetValue,
			&p.IntSetValue,
			&p.FieldRenderGuide,
		}
}

func (p *PartnerDynamicFormFieldValue) TableName() string {
	return "partner_dynamic_form_field_values"
}

type PartnerDynamicFormFieldValues []*PartnerDynamicFormFieldValue

func (lrd *PartnerDynamicFormFieldValues) Add() database.Entity {
	e := &PartnerDynamicFormFieldValue{}
	*lrd = append(*lrd, e)

	return e
}
