package domain

import (
	"time"
)

type PartnerDynamicFormFieldValue struct {
	DynamicFormFieldValueID string
	FieldID                 string
	LessonReportDetailID    string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               *time.Time
	ValueType               string
	IntValue                int
	StringValue             string
	BoolValue               bool
	StringArrayValue        []string
	IntArrayValue           []int32
	StringSetValue          []string
	IntSetValue             []int32
	FieldRenderGuide        []byte
}
type PartnerDynamicFormFieldValues []*PartnerDynamicFormFieldValue

type PartnerDynamicFormFieldValueBuilder struct {
	partnerDynamicFormFieldValue *PartnerDynamicFormFieldValue
}

func NewPartnerDynamicFormFieldValueBuilder() *PartnerDynamicFormFieldValueBuilder {
	return &PartnerDynamicFormFieldValueBuilder{
		partnerDynamicFormFieldValue: &PartnerDynamicFormFieldValue{},
	}
}

func (p *PartnerDynamicFormFieldValueBuilder) Build() (*PartnerDynamicFormFieldValue, error) {
	return p.partnerDynamicFormFieldValue, nil
}

func (p *PartnerDynamicFormFieldValueBuilder) WithDynamicFormValueID(id string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.DynamicFormFieldValueID = id
	return p
}

func (p *PartnerDynamicFormFieldValueBuilder) WithFieldID(id string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.FieldID = id
	return p
}

func (p *PartnerDynamicFormFieldValueBuilder) WithLessonDetailReportID(id string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.LessonReportDetailID = id
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithValueType(valueType string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.ValueType = valueType
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithIntValue(value int) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.IntValue = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithStringValue(value string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.StringValue = value
	return p
}

func (p *PartnerDynamicFormFieldValueBuilder) WithBoolValue(value bool) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.BoolValue = value
	return p
}

func (p *PartnerDynamicFormFieldValueBuilder) WithStringArrayValue(value []string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.StringArrayValue = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithIntArrayValue(value []int32) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.IntArrayValue = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithStringSetValue(value []string) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.StringSetValue = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithIntSetValue(value []int32) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.IntSetValue = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithFieldRenderGuide(value []byte) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.FieldRenderGuide = value
	return p
}
func (p *PartnerDynamicFormFieldValueBuilder) WithModificationTime(createdAt, updatedAt time.Time) *PartnerDynamicFormFieldValueBuilder {
	p.partnerDynamicFormFieldValue.CreatedAt = createdAt
	p.partnerDynamicFormFieldValue.UpdatedAt = updatedAt
	return p
}
