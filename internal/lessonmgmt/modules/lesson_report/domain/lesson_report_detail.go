package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_report_consts "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

// Lesson Report reflect lesson_reports table
type LessonReportDetail struct {
	LessonReportDetailID string
	LessonReportID       string
	StudentID            string
	AttendanceStatus     lesson_report_consts.StudentAttendStatus
	AttendanceRemark     string
	Fields               LessonReportFields
	CreatedAt            time.Time
	UpdatedAt            time.Time
	AttendanceNotice     lesson_report_consts.StudentAttendanceNotice
	AttendanceReason     lesson_report_consts.StudentAttendanceReason
	AttendanceNote       string
	ReportVersion        int
	// For Group Lesson Report
	StudentIDs []string
}

type LessonReportDetailBuilder struct {
	lessonReportDetail *LessonReportDetail
}

func NewLessonReportDetailBuilder() *LessonReportDetailBuilder {
	return &LessonReportDetailBuilder{
		lessonReportDetail: &LessonReportDetail{},
	}
}

func (l *LessonReportDetailBuilder) Build() (*LessonReportDetail, error) {
	if err := l.lessonReportDetail.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid lesson report detail: %w", err)
	}
	return l.lessonReportDetail, nil
}

func (l *LessonReportDetailBuilder) WithLessonReportDetailID(id string) *LessonReportDetailBuilder {
	l.lessonReportDetail.LessonReportDetailID = id
	return l
}

func (l *LessonReportDetailBuilder) WithLessonReportID(id string) *LessonReportDetailBuilder {
	l.lessonReportDetail.LessonReportID = id
	return l
}

func (l *LessonReportDetailBuilder) WithStudentID(id string) *LessonReportDetailBuilder {
	l.lessonReportDetail.StudentID = id
	return l
}

func (l *LessonReportDetailBuilder) WithStudentIDs(ids []string) *LessonReportDetailBuilder {
	l.lessonReportDetail.StudentIDs = ids
	return l
}

func (l *LessonReportDetailBuilder) WithAttendanceStatus(status lesson_report_consts.StudentAttendStatus) *LessonReportDetailBuilder {
	l.lessonReportDetail.AttendanceStatus = status
	return l
}

func (l *LessonReportDetailBuilder) WithAttendanceRemark(remark string) *LessonReportDetailBuilder {
	l.lessonReportDetail.AttendanceRemark = remark
	return l
}

func (l *LessonReportDetailBuilder) WithFields(fields LessonReportFields) *LessonReportDetailBuilder {
	l.lessonReportDetail.Fields = fields
	return l
}

func (l *LessonReportDetailBuilder) WithModificationTime(createdAt, updatedAt time.Time) *LessonReportDetailBuilder {
	l.lessonReportDetail.CreatedAt = createdAt
	l.lessonReportDetail.UpdatedAt = updatedAt
	return l
}

func (l *LessonReportDetailBuilder) WithReportVersion(reportVersion int) *LessonReportDetailBuilder {
	l.lessonReportDetail.ReportVersion = reportVersion
	return l
}

func (l *LessonReportDetail) IsValid() error {
	if len(l.StudentID) == 0 {
		return fmt.Errorf("student_id could not be empty")
	}

	if err := l.Fields.IsValid(); err != nil {
		return fmt.Errorf("invalid fields: %v", err)
	}

	return nil
}

type LessonReportDetails []*LessonReportDetail

func (ls LessonReportDetails) ToLessonMembersEntity(lessonID string) domain.LessonMembers {
	now := time.Now()
	res := make(domain.LessonMembers, 0, len(ls))
	for _, l := range ls {
		e := &domain.LessonMember{
			LessonID:         lessonID,
			StudentID:        l.StudentID,
			AttendanceStatus: string(l.AttendanceStatus),
			AttendanceRemark: l.AttendanceRemark,
			AttendanceNotice: string(l.AttendanceNotice),
			AttendanceReason: string(l.AttendanceReason),
			AttendanceNote:   l.AttendanceNote,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		res = append(res, e)
	}

	return res
}
func (ls LessonReportDetails) AddFieldValues(fields map[string]LessonReportFields) {
	for i := range ls {
		if v, ok := fields[ls[i].LessonReportDetailID]; ok {
			ls[i].Fields = v
		}
	}
}

func (ls LessonReportDetails) CheckStudentsAttendance() bool {
	for _, l := range ls {
		if len(l.AttendanceStatus) == 0 || l.AttendanceStatus == lesson_report_consts.StudentAttendStatusEmpty {
			return false
		}
	}
	return true
}

func (ls LessonReportDetails) IsValid() error {
	studentIDs := make(map[string]bool)
	for _, l := range ls {
		if err := l.IsValid(); err != nil {
			return err
		}

		if _, ok := studentIDs[l.StudentID]; ok {
			return fmt.Errorf("lesson report detail's student id %s be duplicated", l.StudentID)
		}
		studentIDs[l.StudentID] = true
	}

	return nil
}

func (ls LessonReportDetails) OnlyHaveLearnerIDs(learnerIDs []string) error {
	learnerIDsMap := make(map[string]bool)
	for _, id := range learnerIDs {
		learnerIDsMap[id] = true
	}
	for _, detail := range ls {
		if _, ok := learnerIDsMap[detail.StudentID]; !ok {
			return fmt.Errorf("learner %s doesn't belong to lesson", detail.StudentID)
		}
	}

	return nil
}

func (ls LessonReportDetails) OnlyHaveAllowFields(allowFields map[string]*FormConfigField) error {
	// validate field id
	for _, detail := range ls {
		for _, field := range detail.Fields {
			// check field id exist or not in form config
			if _, ok := allowFields[field.FieldID]; !ok {
				return fmt.Errorf("field id %s of user %s not exist in form config", field.FieldID, detail.StudentID)
			}

			// check dynamic field not allow same system defined fields
			if field.FieldID == string(lesson_report_consts.SystemDefinedFieldAttendanceStatus) ||
				field.FieldID == string(lesson_report_consts.SystemDefinedFieldAttendanceRemark) {
				return fmt.Errorf("field id %s of user %s is not a dynamic field", field.FieldID, detail.StudentID)
			}
		}

		// check system defined fields exist or not in form config
		if _, ok := allowFields[string(lesson_report_consts.SystemDefinedFieldAttendanceStatus)]; len(detail.AttendanceStatus) != 0 && detail.AttendanceStatus != lesson_report_consts.StudentAttendStatusEmpty && !ok {
			return fmt.Errorf("field id %s of user %s not exist in form config", lesson_report_consts.SystemDefinedFieldAttendanceStatus, detail.StudentID)
		}
		if _, ok := allowFields[string(lesson_report_consts.SystemDefinedFieldAttendanceRemark)]; len(detail.AttendanceRemark) != 0 && !ok {
			return fmt.Errorf("field id %s of user %s not exist in form config", lesson_report_consts.SystemDefinedFieldAttendanceRemark, detail.StudentID)
		}
	}

	return nil
}

func (ls LessonReportDetails) ValidateRequiredFieldsValue(requiredFields map[string]*FormConfigField) error {
	for _, detail := range ls {
		if detail.AttendanceStatus == lesson_report_consts.StudentAttendStatusAbsent ||
			detail.AttendanceStatus == lesson_report_consts.StudentAttendStatusReallocate {
			continue
		}
		inputFieldsByID := make(map[string]*LessonReportField)
		for i, field := range detail.Fields {
			inputFieldsByID[field.FieldID] = detail.Fields[i]
		}
		for id, requiredField := range requiredFields {
			// check system defined fields
			if id == string(lesson_report_consts.SystemDefinedFieldAttendanceStatus) {
				if len(detail.AttendanceStatus) == 0 || detail.AttendanceStatus == lesson_report_consts.StudentAttendStatusEmpty {
					return fmt.Errorf("field %s is required", lesson_report_consts.SystemDefinedFieldAttendanceStatus)
				}
				continue
			}
			if id == string(lesson_report_consts.SystemDefinedFieldAttendanceRemark) {
				if len(detail.AttendanceRemark) == 0 {
					return fmt.Errorf("field %s is required", lesson_report_consts.SystemDefinedFieldAttendanceRemark)
				}
				continue
			}

			v, ok := inputFieldsByID[id]
			if !ok || v.Value == nil {
				return fmt.Errorf("field %s is required", id)
			}
			switch requiredField.ValueType {
			case lesson_report_consts.FieldValueTypeInt:
			case lesson_report_consts.FieldValueTypeString:
				if len(v.Value.String) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case lesson_report_consts.FieldValueTypeBool:
			case lesson_report_consts.FieldValueTypeIntArray:
				if len(v.Value.IntArray) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case lesson_report_consts.FieldValueTypeStringArray:
				if len(v.Value.StringArray) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case lesson_report_consts.FieldValueTypeIntSet:
				if len(v.Value.IntSet) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case lesson_report_consts.FieldValueTypeStringSet:
				if len(v.Value.StringSet) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			}
		}
	}

	return nil
}

func (ls LessonReportDetails) GetByStudentIDs(studentIDs []string) LessonReportDetails {
	if len(ls) == 0 {
		return nil
	}

	studentIDsMap := make(map[string]bool)
	for _, id := range studentIDs {
		studentIDsMap[id] = true
	}

	details := make(LessonReportDetails, 0, len(ls))
	for i, detail := range ls {
		if _, ok := studentIDsMap[detail.StudentID]; ok {
			details = append(details, ls[i])
			delete(studentIDsMap, detail.StudentID)
		}
	}

	return details
}

// Normalize will remove duplicated StudentID items and normalize Fields attribute
func (ls *LessonReportDetails) Normalize() {
	if len(*ls) == 0 {
		return
	}

	studentIDs := make([]string, 0, len(*ls))
	for _, detail := range *ls {
		studentIDs = append(studentIDs, detail.StudentID)
	}
	*ls = ls.GetByStudentIDs(studentIDs)

	for i := range *ls {
		(*ls)[i].Fields.Normalize()
	}
}

func (ls LessonReportDetails) ToLessonReportDetailsDomain(lessonReportID string) (LessonReportDetails, error) {
	now := time.Now()
	res := make(LessonReportDetails, 0, len(ls))
	for _, l := range ls {
		reportVersion := support.Max(1, l.ReportVersion)
		e := &LessonReportDetail{
			LessonReportDetailID: idutil.ULIDNow(),
			LessonReportID:       lessonReportID,
			StudentID:            l.StudentID,
			StudentIDs:           l.StudentIDs,
			CreatedAt:            now,
			UpdatedAt:            now,
			ReportVersion:        reportVersion,
		}
		res = append(res, e)
	}

	return res, nil
}

func (ls *LessonReportDetails) RemoveAttendanceInfo() error {
	for _, l := range *ls {
		newReportDetail, err := NewLessonReportDetailBuilder().
			WithLessonReportDetailID(l.LessonReportDetailID).
			WithLessonReportID(l.LessonReportID).
			WithStudentID(l.StudentID).
			WithModificationTime(l.CreatedAt, l.UpdatedAt).
			WithStudentIDs(l.StudentIDs).
			WithFields(l.Fields).
			WithReportVersion(l.ReportVersion).
			Build()

		if err != nil {
			return err
		}
		*l = *newReportDetail
	}
	return nil
}
