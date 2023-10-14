package entity

import (
	"github.com/jackc/pgtype"
)

type Course struct {
	ID                pgtype.Text
	Name              pgtype.Text
	Country           pgtype.Text
	Subject           pgtype.Text
	Grade             pgtype.Int2
	DisplayOrder      pgtype.Int2
	SchoolID          pgtype.Int4
	TeacherIDs        pgtype.TextArray
	CourseType        pgtype.Text
	Icon              pgtype.Text
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	StartDate         pgtype.Timestamptz
	EndDate           pgtype.Timestamptz
	PresetStudyPlanID pgtype.Text
	Status            pgtype.Text
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"name",
		"country",
		"subject",
		"grade",
		"display_order",
		"school_id",
		"teacher_ids",
		"course_type",
		"updated_at",
		"created_at",
		"deleted_at",
		"start_date",
		"end_date",
		"preset_study_plan_id",
		"icon",
		"status",
	}
	values = []interface{}{
		&c.ID,
		&c.Name,
		&c.Country,
		&c.Subject,
		&c.Grade,
		&c.DisplayOrder,
		&c.SchoolID,
		&c.TeacherIDs,
		&c.CourseType,
		&c.UpdatedAt,
		&c.CreatedAt,
		&c.DeletedAt,
		&c.StartDate,
		&c.EndDate,
		&c.PresetStudyPlanID,
		&c.Icon,
		&c.Status,
	}
	return
}

func (*Course) TableName() string {
	return "courses"
}
