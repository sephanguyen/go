package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

const (
	ClassMemberStatusActive   = "CLASS_MEMBER_STATUS_ACTIVE"
	ClassMemberStatusInactive = "CLASS_MEMBER_STATUS_INACTIVE"
)

type ClassMember struct {
	ID                    pgtype.Text `sql:"class_member_id,pk"`
	ClassID               pgtype.Int4
	UserID                pgtype.Text
	UserGroup             pgtype.Text
	IsOwner               pgtype.Bool `sql:",notnull"`
	StudentSubscriptionID pgtype.Text `sql:"student_subscription_id"`
	Status                pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
}

func (t *ClassMember) FieldMap() ([]string, []interface{}) {
	return []string{
			"class_member_id", "class_id", "user_id", "status", "user_group", "is_owner", "student_subscription_id", "updated_at", "created_at",
		}, []interface{}{
			&t.ID, &t.ClassID, &t.UserID, &t.Status, &t.UserGroup, &t.IsOwner, &t.StudentSubscriptionID, &t.UpdatedAt, &t.CreatedAt,
		}
}

func (t *ClassMember) TableName() string {
	return "class_members"
}

type ClassMembers []*ClassMember

func (ss *ClassMembers) Add() database.Entity {
	e := &ClassMember{}
	*ss = append(*ss, e)

	return e
}
