package repositories

import (
	"github.com/manabie-com/backend/internal/calendar/domain/dto"

	"github.com/jackc/pgtype"
)

type User struct {
	ID       pgtype.Text `sql:"user_id,pk"`
	FullName pgtype.Text `sql:"name"`
	Email    pgtype.Text `sql:"email"`
}

func (u *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"name",
			"email",
		}, []interface{}{
			&u.ID,
			&u.FullName,
			&u.Email,
		}
}

func (u *User) ConvertDTO() *dto.User {
	return &dto.User{
		UserID: u.ID.String,
		Name:   u.FullName.String,
		Email:  u.Email.String,
	}
}
