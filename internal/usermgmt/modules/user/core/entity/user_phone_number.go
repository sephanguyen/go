package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

const (
	StudentPhoneNumber         = "STUDENT_PHONE_NUMBER"
	StudentHomePhoneNumber     = "STUDENT_HOME_PHONE_NUMBER"
	ParentPrimaryPhoneNumber   = "PARENT_PRIMARY_PHONE_NUMBER"
	ParentSecondaryPhoneNumber = "PARENT_SECONDARY_PHONE_NUMBER"
	StaffPrimaryPhoneNumber    = "STAFF_PRIMARY_PHONE_NUMBER"
	StaffSecondaryPhoneNumber  = "STAFF_SECONDARY_PHONE_NUMBER"
)

type UserPhoneNumber struct {
	ID              pgtype.Text
	UserID          pgtype.Text
	PhoneNumber     pgtype.Text
	PhoneNumberType pgtype.Text
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
	ResourcePath    pgtype.Text
}

type UserPhoneNumbers []*UserPhoneNumber

func (e *UserPhoneNumber) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_phone_number_id",
			"user_id",
			"phone_number",
			"type",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.ID,
			&e.UserID,
			&e.PhoneNumber,
			&e.PhoneNumberType,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *UserPhoneNumber) TableName() string {
	return "user_phone_number"
}

func (u *UserPhoneNumbers) Add() database.Entity {
	e := &UserPhoneNumber{}
	*u = append(*u, e)

	return e
}
