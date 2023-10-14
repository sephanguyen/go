package form_partner

import (
	"fmt"
	"strings"

	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
)

type ConverterFormReport struct {
	EvictionPartner EvictionPartner
}

func (c *ConverterFormReport) Convert(lessonReport *domain.LessonReport, lessonReportDetail *domain.LessonReportDetail, fields domain.LessonReportFields) (*domain.LessonReportMigrateData, error) {
	attendance_notice := ""
	attendance_note := ""
	attendance_status := string(lessonReportDetail.AttendanceStatus)
	mapNewField := c.EvictionPartner.getMapNewField()

	switch string(lessonReportDetail.AttendanceStatus) {
	case string(lesson_domain.StudentAttendStatusInformedAbsent):
		attendance_notice = string(lesson_domain.InAdvance)
		attendance_status = string(constant.StudentAttendStatusAbsent)

	case string(lesson_domain.StudentAttendStatusAbsent), string(lesson_domain.StudentAttendStatusLate):
		attendance_status = string(constant.StudentAttendStatusAbsent)
		attendance_notice = string(lesson_domain.NoContact)

	case string(lesson_domain.StudentAttendStatusInformedLate):
		attendance_status = string(constant.StudentAttendStatusLate)
		attendance_notice = string(lesson_domain.InAdvance)

	case string(lesson_domain.StudentAttendStatusLeaveEarly):
		attendance_status = string(constant.StudentAttendStatusLeaveEarly)
		attendance_notice = string(lesson_domain.NoContact)
	}
	newFields := make(domain.PartnerDynamicFormFieldValues, 0, len(fields))
	mapFields := make(map[string]*domain.PartnerDynamicFormFieldValue)
	for _, field := range fields {
		newField, err := field.ToPartnerDynamicFormFieldValue(lessonReportDetail.LessonReportDetailID)
		if err != nil {
			return nil, err
		}
		mapFields[field.FieldID] = newField
	}

	for newKey, oldKeys := range mapNewField {
		if len(oldKeys) > 1 {
			// multi field
			newField, err := domain.CreateNewPartnerDynamicFormField(newKey,
				lessonReportDetail.LessonReportDetailID, "VALUE_TYPE_STRING")
			if err != nil {
				return nil, err
			}
			strValue := ""
			for _, key := range oldKeys {
				field, ok := mapFields[key]
				if ok {
					strValue = fmt.Sprintf("%s %s", strValue, field.StringValue)
				}
			}
			newField.StringValue = strings.TrimSpace(strValue)
			newFields = append(newFields, newField)
		} else {
			// only field
			oldField, ok := mapFields[oldKeys[0]]
			if ok {
				strValue := oldField.StringValue
				newField := oldField
				newField.StringValue = strValue
				newField.FieldID = newKey
				newFields = append(newFields, newField)
			}
		}
	}

	attendance_note = lessonReportDetail.AttendanceRemark

	lessonMemberUpdate := &lesson_domain.UpdateLessonMemberReport{
		LessonID:         lessonReport.LessonID,
		StudentID:        lessonReportDetail.StudentID,
		AttendanceNotice: attendance_notice,
		AttendanceStatus: attendance_status,
		AttendanceNote:   attendance_note,
	}
	return &domain.LessonReportMigrateData{
		LessonMemberUpdate: lessonMemberUpdate,
		Fields:             newFields,
	}, nil
}
