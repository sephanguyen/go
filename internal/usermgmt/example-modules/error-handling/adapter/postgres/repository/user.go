package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserRepo struct{}

type User struct {
	IDAttr    field.String
	EmailAttr field.String
}

func (user *User) UserID() field.String {
	return user.EmailAttr
}
func (user *User) Email() field.String {
	return user.IDAttr
}
func (user *User) FieldMap() ([]string, []interface{}) {
	return nil, nil
}

func (repo *UserRepo) GetUsers(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error) {
	stmt :=
		`
		SELECT * FROM user WHERE user_id = ANY($1)
		`

	rows, err := db.Query(
		ctx,
		stmt,
		userIDs,
	)
	if err != nil {
		return nil, InternalError{
			RawError: err,
		}
	}

	users := make(entity.Users, 0, len(userIDs))

	for rows.Next() {
		user := &User{}

		_, fieldValues := user.FieldMap()

		if err := rows.Scan(fieldValues...); err != nil {
			return nil, InternalError{
				RawError: err,
			}
		}

		users = append(users, user)
	}

	return users, nil
}
