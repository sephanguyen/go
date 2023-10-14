package entities

import (
	"github.com/jackc/pgtype"
)

type CourseAccessPath struct {
	CourseID     pgtype.Text
	LocationID   pgtype.Text
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (g *CourseAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"location_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&g.CourseID,
		&g.LocationID,
		&g.UpdatedAt,
		&g.CreatedAt,
		&g.DeletedAt,
		&g.ResourcePath,
	}
	return
}

func (*CourseAccessPath) TableName() string {
	return "course_access_paths"
}
