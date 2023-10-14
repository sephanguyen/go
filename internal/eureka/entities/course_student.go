package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type CourseStudent struct {
	BaseEntity
	ID        pgtype.Text
	CourseID  pgtype.Text
	StudentID pgtype.Text
	StartAt   pgtype.Timestamptz
	EndAt     pgtype.Timestamptz
}

func (rcv *CourseStudent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_student_id", "course_id", "student_id", "created_at", "updated_at", "deleted_at", "start_at", "end_at"}
	values = []interface{}{&rcv.ID, &rcv.CourseID, &rcv.StudentID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt, &rcv.StartAt, &rcv.EndAt}
	return
}

func (rcv *CourseStudent) TableName() string {
	return "course_students"
}

type CourseStudents []*CourseStudent

func (c *CourseStudents) Add() database.Entity {
	e := &CourseStudent{}
	*c = append(*c, e)

	return e
}

type StudentTag struct {
	ID   pgtype.Text
	Name pgtype.Text
}

func (st *StudentTag) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_tag_id", "user_tag_name"}
	values = []interface{}{&st.ID, &st.Name}
	return
}

func (st *StudentTag) TableName() string {
	return "user_tag"
}

type StudentTags []*StudentTag

func (st *StudentTags) Add() database.Entity {
	e := &StudentTag{}
	*st = append(*st, e)

	return e
}
