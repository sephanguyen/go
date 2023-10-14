package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// Lesson Report Detail reflect lesson_report_details table
type LessonReportDetailDTO struct {
	LessonReportDetailID pgtype.Text
	LessonReportID       pgtype.Text
	StudentID            pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	ReportVersion        pgtype.Int4
}

// FieldMap Lesson Report Detail table data fields
func (lrd *LessonReportDetailDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_report_id",
			"student_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"lesson_report_detail_id",
			"report_version",
		}, []interface{}{
			&lrd.LessonReportID,
			&lrd.StudentID,
			&lrd.CreatedAt,
			&lrd.UpdatedAt,
			&lrd.DeletedAt,
			&lrd.LessonReportDetailID,
			&lrd.ReportVersion,
		}
}

// TableName returns "lesson_report_details"
func (lrd *LessonReportDetailDTO) TableName() string {
	return "lesson_report_details"
}

type LessonReportDetailDTOs []*LessonReportDetailDTO

func (lrd *LessonReportDetailDTOs) Add() database.Entity {
	e := &LessonReportDetailDTO{}
	*lrd = append(*lrd, e)

	return e
}

type LessonReportDetailWithAttendanceStatusDTO struct {
	LessonReportDetailID pgtype.Text
	LessonReportID       pgtype.Text
	StudentID            pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	AttendanceStatus     pgtype.Text
	AttendanceRemark     pgtype.Text
	ReportVersion        pgtype.Int4
}

func (lrd *LessonReportDetailWithAttendanceStatusDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_report_id",
			"student_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"lesson_report_detail_id",
			"attendance_status",
			"attendance_remark",
			"report_version",
		}, []interface{}{
			&lrd.LessonReportID,
			&lrd.StudentID,
			&lrd.CreatedAt,
			&lrd.UpdatedAt,
			&lrd.DeletedAt,
			&lrd.LessonReportDetailID,
			&lrd.AttendanceStatus,
			&lrd.AttendanceRemark,
			&lrd.ReportVersion,
		}
}
func (lrd *LessonReportDetailWithAttendanceStatusDTO) TableName() string {
	return "lesson_report_details"
}

type LessonReportDetailWithAttendanceStatusDTOs []*LessonReportDetailWithAttendanceStatusDTO

func (lrd *LessonReportDetailWithAttendanceStatusDTOs) Add() database.Entity {
	e := &LessonReportDetailWithAttendanceStatusDTO{}
	*lrd = append(*lrd, e)

	return e
}
