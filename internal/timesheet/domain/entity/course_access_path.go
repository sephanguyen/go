package entity

import (
	"github.com/jackc/pgtype"
)

type CourseAccessPath struct {
	LocationID pgtype.Text
	CourseID   pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (c *CourseAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_id", "course_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.LocationID, &c.CourseID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*CourseAccessPath) TableName() string {
	return "course_access_paths"
}
