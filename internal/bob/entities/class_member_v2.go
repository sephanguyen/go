package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ClassMemberV2 struct {
	ClassMemberID pgtype.Text `sql:"class_member_id,pk"`
	ClassID       pgtype.Text
	UserID        pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
}

func (t *ClassMemberV2) FieldMap() ([]string, []interface{}) {
	return []string{
			"class_member_id", "class_id", "user_id", "updated_at", "created_at",
		}, []interface{}{
			&t.ClassMemberID, &t.ClassID, &t.UserID, &t.UpdatedAt, &t.CreatedAt,
		}
}

func (t *ClassMemberV2) TableName() string {
	return "class_member"
}

type ClassMembersV2 []*ClassMemberV2

func (ss *ClassMembersV2) Add() database.Entity {
	e := &ClassMemberV2{}
	*ss = append(*ss, e)

	return e
}
