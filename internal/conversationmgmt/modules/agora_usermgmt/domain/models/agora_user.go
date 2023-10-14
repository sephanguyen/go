package models

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AgoraUser struct {
	UserID      pgtype.Text
	AgoraUserID pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (*AgoraUser) TableName() string {
	return "agora_user"
}

func (e *AgoraUser) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_id",
		"agora_user_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.UserID,
		&e.AgoraUserID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type AgoraUsers []*AgoraUser

func (u *AgoraUsers) Add() database.Entity {
	e := &AgoraUser{}
	*u = append(*u, e)

	return e
}
