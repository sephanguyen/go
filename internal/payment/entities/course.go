package entities

import (
	"github.com/jackc/pgtype"
)

type Course struct {
	CourseID       pgtype.Text
	Name           pgtype.Text
	Grade          pgtype.Int2
	TeachingMethod pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
}

func (g *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"name",
		"grade",
		"teaching_method",
		"updated_at",
		"created_at",
		"resource_path",
	}
	values = []interface{}{
		&g.CourseID,
		&g.Name,
		&g.Grade,
		&g.TeachingMethod,
		&g.UpdatedAt,
		&g.CreatedAt,
		&g.ResourcePath,
	}
	return
}

func (*Course) TableName() string {
	return "courses"
}
