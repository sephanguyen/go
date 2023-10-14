package entities

import (
	"github.com/jackc/pgtype"
)

type StudentParent struct {
	StudentID    pgtype.Text
	ParentID     pgtype.Text
	Relationship pgtype.Text
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

func (rcv *StudentParent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"student_id", "parent_id", "updated_at", "created_at", "deleted_at", "relationship"}
	values = []interface{}{&rcv.StudentID, &rcv.ParentID, &rcv.UpdatedAt, &rcv.CreatedAt, &rcv.DeletedAt, &rcv.Relationship}
	return
}

func (*StudentParent) TableName() string {
	return "student_parents"
}
