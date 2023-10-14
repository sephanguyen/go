package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ClassStudent struct {
	BaseEntity
	StudentID pgtype.Text
	ClassID   pgtype.Text
}

func (rcv *ClassStudent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_id", "class_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&rcv.StudentID, &rcv.ClassID, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.DeletedAt}
	return
}

func (rcv *ClassStudent) TableName() string {
	return "class_students"
}

type ClassStudents []*ClassStudent

func (rcv *ClassStudents) Add() database.Entity {
	e := &ClassStudent{}
	*rcv = append(*rcv, e)

	return e
}
