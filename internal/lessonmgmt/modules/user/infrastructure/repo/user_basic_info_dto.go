package repo

import (
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
	GradeID           pgtype.Text
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
			"grade_id",
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
			&u.GradeID,
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

func (u *UserBasicInfo) TableName() string {
	return "user_basic_info"
}
