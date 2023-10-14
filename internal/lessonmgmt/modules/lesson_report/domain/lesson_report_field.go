package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type LessonReportField struct {
	FieldID          string
	Value            *AttributeValue
	FieldRenderGuide []byte
	ValueType        string

	// For Group Lesson Report
	IsRequired bool
	IsInternal bool
}

func (l *LessonReportField) IsValid() error {
	if len(l.FieldID) == 0 {
		return fmt.Errorf("field_id could not be empty")
	}
	// For lesson group
	if l.IsRequired && l.Value == nil {
		return fmt.Errorf("LessonReportFields.IsValid(): Field with ID %s is required but no data was inputted", l.FieldID)
	}
	return nil
}

type LessonReportFields []*LessonReportField

func (ls LessonReportFields) IsValid() error {
	fieldIDs := make(map[string]bool)
	for _, l := range ls {
		if err := l.IsValid(); err != nil {
			return err
		}

		if _, ok := fieldIDs[l.FieldID]; ok {
			return fmt.Errorf("LessonReportFields.IsValid(): Field with ID %s is duplicated", l.FieldID)
		}
		fieldIDs[l.FieldID] = true
	}

	return nil
}

func (ls LessonReportFields) GetFieldsByIDs(ids []string) LessonReportFields {
	idsMap := make(map[string]bool)
	for _, id := range ids {
		idsMap[id] = true
	}

	fields := make(LessonReportFields, 0, len(ids))
	for i, field := range ls {
		if _, ok := idsMap[field.FieldID]; ok {
			fields = append(fields, ls[i])
			delete(idsMap, field.FieldID)
		}
	}

	return fields
}
func (ls *LessonReportFields) Normalize() {
	if len(*ls) == 0 {
		return
	}

	fieldIDs := make(map[string]bool)
	notDuplicated := make(LessonReportFields, 0, len(*ls))
	for i, field := range *ls {
		if _, ok := fieldIDs[field.FieldID]; !ok {
			notDuplicated = append(notDuplicated, (*ls)[i])
			fieldIDs[field.FieldID] = true
		}
	}
	*ls = notDuplicated
}

func (lr LessonReportField) ToPartnerDynamicFormFieldValue(lessonReportDetailID string) (*PartnerDynamicFormFieldValue, error) {
	now := time.Now()
	partnerDynamicFormFieldValueEntity, err := NewPartnerDynamicFormFieldValueBuilder().
		WithDynamicFormValueID(idutil.ULIDNow()).
		WithFieldID(lr.FieldID).WithLessonDetailReportID(lessonReportDetailID).
		WithFieldRenderGuide(lr.FieldRenderGuide).
		WithIntValue(lr.Value.Int).
		WithStringValue(lr.Value.String).WithBoolValue(lr.Value.Bool).
		WithStringArrayValue(lr.Value.StringArray).
		WithIntArrayValue(lr.Value.IntArray).
		WithStringSetValue(lr.Value.StringSet).
		WithIntSetValue(lr.Value.IntSet).
		WithValueType(lr.ValueType).
		WithModificationTime(now, now).Build()
	if err != nil {
		return nil, err
	}
	return partnerDynamicFormFieldValueEntity, nil
}

func (ls LessonReportFields) ToPartnerDynamicFormFieldValueEntities(lessonReportDetailID string) ([]*PartnerDynamicFormFieldValue, error) {
	now := time.Now()
	res := make([]*PartnerDynamicFormFieldValue, 0, len(ls))
	for _, l := range ls {
		partnerDynamicFormFieldValueEntity, err := NewPartnerDynamicFormFieldValueBuilder().
			WithDynamicFormValueID(idutil.ULIDNow()).
			WithFieldID(l.FieldID).WithLessonDetailReportID(lessonReportDetailID).
			WithFieldRenderGuide(l.FieldRenderGuide).
			WithIntValue(l.Value.Int).
			WithStringValue(l.Value.String).WithBoolValue(l.Value.Bool).
			WithStringArrayValue(l.Value.StringArray).
			WithIntArrayValue(l.Value.IntArray).
			WithStringSetValue(l.Value.StringSet).
			WithIntSetValue(l.Value.IntSet).
			WithValueType(l.ValueType).
			WithModificationTime(now, now).Build()
		if err != nil {
			return nil, err
		}
		res = append(res, partnerDynamicFormFieldValueEntity)
	}
	return res, nil
}

func CreateNewPartnerDynamicFormField(fieldID string, lessonReportDetailID string, valueType string) (*PartnerDynamicFormFieldValue, error) {
	now := time.Now()
	partnerDynamicFormFieldValueEntity, err := NewPartnerDynamicFormFieldValueBuilder().
		WithDynamicFormValueID(idutil.ULIDNow()).
		WithFieldID(fieldID).WithLessonDetailReportID(lessonReportDetailID).
		WithValueType(valueType).
		WithModificationTime(now, now).Build()
	if err != nil {
		return nil, err
	}
	return partnerDynamicFormFieldValueEntity, nil
}

func LessonReportFieldsFromDynamicFieldValueGRPC(fields ...*lpb.DynamicFieldValue) (LessonReportFields, error) {
	res := make(LessonReportFields, 0, len(fields))
	for _, field := range fields {
		value := &AttributeValue{}
		switch v := field.Value.(type) {
		case *lpb.DynamicFieldValue_IntValue:
			value.SetInt(int(v.IntValue))
		case *lpb.DynamicFieldValue_StringValue:
			value.SetString(v.StringValue)
		case *lpb.DynamicFieldValue_BoolValue:
			value.SetBool(v.BoolValue)
		case *lpb.DynamicFieldValue_IntArrayValue_:
			intArray := make([]int32, 0, len(v.IntArrayValue.ArrayValue))
			intArray = append(intArray, v.IntArrayValue.GetArrayValue()...)
			value.SetIntArray(intArray)
		case *lpb.DynamicFieldValue_StringArrayValue_:
			value.SetStringArray(v.StringArrayValue.GetArrayValue())
		case *lpb.DynamicFieldValue_IntSetValue_:
			intSet := make([]int32, 0, len(v.IntSetValue.ArrayValue))
			intSet = append(intSet, v.IntSetValue.GetArrayValue()...)
			value.SetIntSet(intSet)
		case *lpb.DynamicFieldValue_StringSetValue_:
			value.SetStringSet(v.StringSetValue.GetArrayValue())
		default:
			return nil, fmt.Errorf("unimplement handler for type %T", field.Value)
		}

		res = append(res, &LessonReportField{
			FieldID:          field.DynamicFieldId,
			FieldRenderGuide: field.FieldRenderGuide,
			Value:            value,
			// for Lesson Group
			IsRequired: field.IsRequired,
			IsInternal: field.IsInternal,
			ValueType:  field.ValueType.String(),
		})
	}

	return res, nil
}
