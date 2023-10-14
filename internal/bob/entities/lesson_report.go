package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type (
	ReportType              string
	ReportSubmittingStatus  string
	StudentAttendStatus     string
	DomainType              string
	StudentAttendanceNotice string
	StudentAttendanceReason string
)

const (
	ReportTypeIndividual ReportType = "LESSON_REPORT_INDIVIDUAL"
	ReportTypeGroup      ReportType = "LESSON_REPORT_GROUP"

	ReportSubmittingStatusSaved     ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_SAVED"
	ReportSubmittingStatusSubmitted ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_SUBMITTED"
	ReportSubmittingStatusApproved  ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_APPROVED"

	StudentAttendStatusEmpty      StudentAttendStatus = "STUDENT_ATTEND_STATUS_EMPTY"
	StudentAttendStatusAttend     StudentAttendStatus = "STUDENT_ATTEND_STATUS_ATTEND"
	StudentAttendStatusAbsent     StudentAttendStatus = "STUDENT_ATTEND_STATUS_ABSENT"
	StudentAttendStatusLate       StudentAttendStatus = "STUDENT_ATTEND_STATUS_LATE"
	StudentAttendStatusLeaveEarly StudentAttendStatus = "STUDENT_ATTEND_STATUS_LEAVE_EARLY"

	DomainTypeBo      DomainType = "DOMAIN_TYPE_BO"
	DomainTypeTeacher DomainType = "DOMAIN_TYPE_TEACHER"
	DomainTypeLearner DomainType = "DOMAIN_TYPE_LEARNER"

	StudentAttendanceNoticeEmpty     StudentAttendanceNotice = "NOTICE_EMPTY"
	StudentAttendanceNoticeInAdvance StudentAttendanceNotice = "IN_ADVANCE"
	StudentAttendanceNoticeOnTheDay  StudentAttendanceNotice = "ON_THE_DAY"
	StudentAttendanceNoticeNoContact StudentAttendanceNotice = "NO_CONTACT"

	StudentAttendanceReasonEmpty             StudentAttendanceReason = "REASON_EMPTY"
	StudentAttendanceReasonPhysicalCondition StudentAttendanceReason = "PHYSICAL_CONDITION"
	StudentAttendanceReasonSchoolEvent       StudentAttendanceReason = "SCHOOL_EVENT"
	StudentAttendanceReasonFamilyReason      StudentAttendanceReason = "FAMILY_REASON"
	StudentAttendanceReasonOther             StudentAttendanceReason = "REASON_OTHER"
)

// Lesson Report reflect lesson_reports table
type LessonReport struct {
	LessonReportID         pgtype.Text
	ReportSubmittingStatus pgtype.Text
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	FormConfigID           pgtype.Text
	LessonID               pgtype.Text
}

// FieldMap Lesson Report table data fields
func (lr *LessonReport) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_report_id",
			"report_submitting_status",
			"created_at",
			"updated_at",
			"deleted_at",
			"form_config_id",
			"lesson_id",
		}, []interface{}{
			&lr.LessonReportID,
			&lr.ReportSubmittingStatus,
			&lr.CreatedAt,
			&lr.UpdatedAt,
			&lr.DeletedAt,
			&lr.FormConfigID,
			&lr.LessonID,
		}
}

// TableName returns "lesson_reports"
func (lr *LessonReport) TableName() string {
	return "lesson_reports"
}

func (lr *LessonReport) PreUpdate() error {
	return lr.UpdatedAt.Set(time.Now())
}

type LessonReports []*LessonReport

func (lr *LessonReports) Add() database.Entity {
	e := &LessonReport{}
	*lr = append(*lr, e)

	return e
}
