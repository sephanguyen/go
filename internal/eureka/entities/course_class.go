package entities

import (
	"github.com/jackc/pgtype"
)

type CourseClass struct {
	BaseEntity
	ID       pgtype.Text `sql:"course_class_id,pk"`
	CourseID pgtype.Text `sql:"course_id"`
	ClassID  pgtype.Text `sql:"class_id"`
}

func (rcv *CourseClass) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_class_id", "course_id", "class_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.ID, &rcv.CourseID, &rcv.ClassID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *CourseClass) TableName() string {
	return "course_classes"
}
