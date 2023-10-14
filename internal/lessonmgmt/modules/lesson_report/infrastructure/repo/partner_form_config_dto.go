package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type (
	FeatureName           string
	PartnerFormConfigDTOs []*PartnerFormConfigDTO
)

func (pfs *PartnerFormConfigDTOs) Add() database.Entity {
	pf := &PartnerFormConfigDTO{}
	*pfs = append(*pfs, pf)

	return pf
}

type PartnerFormConfigDTO struct {
	FormConfigID   pgtype.Text
	PartnerID      pgtype.Int4
	FeatureName    pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	FormConfigData pgtype.JSONB
}

func (p *PartnerFormConfigDTO) FieldMap() ([]string, []interface{}) {
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

func (p *PartnerFormConfigDTO) TableName() string {
	return "partner_form_configs"
}

type PartnerDynamicFormFieldValueDTO struct {
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

func (p *PartnerDynamicFormFieldValueDTO) FieldMap() ([]string, []interface{}) {
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

func (p *PartnerDynamicFormFieldValueDTO) TableName() string {
	return "partner_dynamic_form_field_values"
}

type PartnerDynamicFormFieldValueDTOs []*PartnerDynamicFormFieldValueDTO

func (p *PartnerDynamicFormFieldValueDTOs) Add() database.Entity {
	e := &PartnerDynamicFormFieldValueDTO{}
	*p = append(*p, e)

	return e
}

type PartnerDynamicFormFieldValueWithStudentIdDTO struct {
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
	StudentID               pgtype.Text
}

func (p *PartnerDynamicFormFieldValueWithStudentIdDTO) FieldMap() ([]string, []interface{}) {
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
			"student_id",
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
			&p.StudentID,
		}
}

func (p *PartnerDynamicFormFieldValueWithStudentIdDTO) TableName() string {
	return "partner_dynamic_form_field_values"
}

type PartnerDynamicFormFieldValueWithStudentIdDTOs []*PartnerDynamicFormFieldValueWithStudentIdDTO

func (p *PartnerDynamicFormFieldValueWithStudentIdDTOs) Add() database.Entity {
	e := &PartnerDynamicFormFieldValueWithStudentIdDTO{}
	*p = append(*p, e)

	return e
}

func (p *PartnerDynamicFormFieldValueWithStudentIdDTO) ToDomain() *domain.LessonReportField {
	return &domain.LessonReportField{
		FieldID:          p.FieldID.String,
		FieldRenderGuide: p.FieldRenderGuide.Bytes,
		ValueType:        p.ValueType.String,
		Value: &domain.AttributeValue{
			Int:         int(p.IntValue.Int),
			String:      p.StringValue.String,
			Bool:        p.BoolValue.Bool,
			IntArray:    database.Int4ArrayToInt32Array(p.IntArrayValue),
			StringArray: database.FromTextArray(p.StringArrayValue),
			IntSet:      database.Int4ArrayToInt32Array(p.IntSetValue),
			StringSet:   database.FromTextArray(p.StringSetValue),
		},
	}
}

func NewPartnerFormConfigFromEntity(f *domain.PartnerFormConfig) (*PartnerFormConfigDTO, error) {
	dto := &PartnerFormConfigDTO{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.FormConfigID.Set(f.FormConfigID),
		dto.PartnerID.Set(f.PartnerID),
		dto.FeatureName.Set(f.FeatureName),
		dto.FormConfigData.Set(f.FormConfigData),
		dto.CreatedAt.Set(f.CreatedAt),
		dto.UpdatedAt.Set(f.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from lesson entity to lesson dto: %w", err)
	}
	return dto, nil
}
