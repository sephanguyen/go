package models

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AgoraUserFailure struct {
	UserID      pgtype.Text
	AgoraUserID pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (*AgoraUserFailure) TableName() string {
	return "agora_user_failure"
}

func (e *AgoraUserFailure) FieldMap() (fields []string, values []interface{}) {
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

type AgoraUserFailures []*AgoraUserFailure

func (u *AgoraUserFailures) Add() database.Entity {
	e := &AgoraUserFailure{}
	*u = append(*u, e)

	return e
}
