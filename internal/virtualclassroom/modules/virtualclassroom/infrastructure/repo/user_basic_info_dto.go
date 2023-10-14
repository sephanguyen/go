package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type UserBasicInfo struct {
	UserID            pgtype.Text
	Name              pgtype.Text
	FirstName         pgtype.Text
	LastName          pgtype.Text
	FullNamePhonetic  pgtype.Text
	FirstNamePhonetic pgtype.Text
	LastNamePhonetic  pgtype.Text
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
		}, []interface{}{
			&u.UserID,
			&u.Name,
			&u.FirstName,
			&u.LastName,
			&u.FullNamePhonetic,
			&u.FirstNamePhonetic,
			&u.LastNamePhonetic,
		}
}

func (u *UserBasicInfo) TableName() string {
	return "user_basic_info"
}

func (u *UserBasicInfo) ToUserBasicInfoDomain() domain.UserBasicInfo {
	return domain.UserBasicInfo{
		UserID:            u.UserID.String,
		Name:              u.Name.String,
		FirstName:         u.FirstName.String,
		LastName:          u.LastName.String,
		FullNamePhonetic:  u.FullNamePhonetic.String,
		FirstNamePhonetic: u.FirstNamePhonetic.String,
		LastNamePhonetic:  u.LastNamePhonetic.String,
	}
}
