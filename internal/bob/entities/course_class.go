package entities

import (
	"github.com/jackc/pgtype"
)

const (
	CourseClassStatusActive   = "COURSE_CLASS_STATUS_ACTIVE"
	CourseClassStatusInActive = "COURSE_CLASS_STATUS_INACTIVE"
)

type CourseClass struct {
	CourseID  pgtype.Text `sql:"course_id,pk"`
	ClassID   pgtype.Int4 `sql:"class_id,pk"`
	Status    pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *CourseClass) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "class_id", "status", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.CourseID, &c.ClassID, &c.Status, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*CourseClass) TableName() string {
	return "courses_classes"
}
