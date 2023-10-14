package entities

import (
	"github.com/jackc/pgtype"
)

const (
	CourseAcademicYearStatusActive   = "ACADEMIC_YEAR_STATUS_ACTIVE"
	CourseAcademicYearStatusInActive = "ACADEMIC_YEAR_STATUS_INACTIVE"
)

type CourseAcademicYear struct {
	CourseID       pgtype.Text
	AcademicYearID pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (rcv *CourseAcademicYear) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"academic_year_id",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&rcv.CourseID,
		&rcv.AcademicYearID,
		&rcv.UpdatedAt,
		&rcv.CreatedAt,
		&rcv.DeletedAt,
	}
	return
}

func (*CourseAcademicYear) TableName() string {
	return "courses_academic_years"
}
