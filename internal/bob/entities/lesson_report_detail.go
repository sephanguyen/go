package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type (
	LessonReportAttendance string
	LessonReportHomeWork   string
)

const (
	LessonReportAttendanceAttend  LessonReportAttendance = "LESSON_REPORT_ATTENDANCE_ATTEND"
	LessonReportAttendanceAbsence LessonReportAttendance = "LESSON_REPORT_ATTENDANCE_ABSENCE"
	LessonReportAttendanceLate    LessonReportAttendance = "LESSON_REPORT_ATTENDANCE_LATE"

	LessonReportHomeWorkDone    LessonReportHomeWork = "LESSON_REPORT_HOMEWORK_DONE"
	LessonReportHomeWorkNotDone LessonReportHomeWork = "LESSON_REPORT_HOMEWORK_NOT_DONE"
)

// Lesson Report Detail reflect lesson_report_details table
type LessonReportDetail struct {
	LessonReportDetailID pgtype.Text
	LessonReportID       pgtype.Text
	StudentID            pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
}

// FieldMap Lesson Report Detail table data fields
func (lrd *LessonReportDetail) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_report_id",
			"student_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"lesson_report_detail_id",
		}, []interface{}{
			&lrd.LessonReportID,
			&lrd.StudentID,
			&lrd.CreatedAt,
			&lrd.UpdatedAt,
			&lrd.DeletedAt,
			&lrd.LessonReportDetailID,
		}
}

// TableName returns "lesson_report_details"
func (lrd *LessonReportDetail) TableName() string {
	return "lesson_report_details"
}

type LessonReportDetails []*LessonReportDetail

func (lrd *LessonReportDetails) Add() database.Entity {
	e := &LessonReportDetail{}
	*lrd = append(*lrd, e)

	return e
}

func (lrd LessonReportDetails) ReportDetailIDs() pgtype.TextArray {
	list := make([]pgtype.Text, 0, len(lrd))
	for _, i := range lrd {
		list = append(list, i.LessonReportDetailID)
	}

	res := pgtype.TextArray{}
	res.Set(list)
	return res
}
