package models

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type UserBasicInfo struct {
	UserID            pgtype.Text
	FullName          pgtype.Text
	FirstName         pgtype.Text
	LastName          pgtype.Text
	FullNamePhonetic  pgtype.Text
	FirstNamePhonetic pgtype.Text
	LastNamePhonetic  pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (u *UserBasicInfo) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"name",
			"first_name",
			"last_name",
			"full_name_phonetic",
			"first_name_phonetic",
			"last_name_phonetic",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&u.UserID,
			&u.FullName,
			&u.FirstName,
			&u.LastName,
			&u.FullNamePhonetic,
			&u.FirstNamePhonetic,
			&u.LastNamePhonetic,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		}
}
func (u *UserBasicInfo) GetName() string {
	if u.FullName.String == "" {
		if u.FirstName.String == "" {
			return u.LastName.String
		}
		return u.FirstName.String + " " + u.LastName.String
	}
	return u.FullName.String
}

// TODO: Update it to `user_basic_info` table when sync it to tom. And remember to add `grade_id` field
func (u *UserBasicInfo) TableName() string {
	return "users"
}

type UserBasicInfos []*UserBasicInfo

func (u *UserBasicInfos) Add() database.Entity {
	e := &UserBasicInfo{}
	*u = append(*u, e)

	return e
}
