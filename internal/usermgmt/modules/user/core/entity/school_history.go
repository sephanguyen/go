package entity

import "github.com/jackc/pgtype"

type SchoolHistory struct {
	StudentID      pgtype.Text
	SchoolID       pgtype.Text
	SchoolCourseID pgtype.Text
	IsCurrent      pgtype.Bool
	StartDate      pgtype.Timestamptz
	EndDate        pgtype.Timestamptz

	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (s *SchoolHistory) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_id",
		"school_id",
		"school_course_id",
		"start_date",
		"end_date",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
		"is_current",
	}
	values = []interface{}{
		&s.StudentID,
		&s.SchoolID,
		&s.SchoolCourseID,
		&s.StartDate,
		&s.EndDate,
		&s.UpdatedAt,
		&s.CreatedAt,
		&s.DeletedAt,
		&s.ResourcePath,
		&s.IsCurrent,
	}
	return
}

func (*SchoolHistory) TableName() string {
	return "school_history"
}
