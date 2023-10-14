package entity

import "github.com/manabie-com/backend/internal/usermgmt/pkg/field"

type UserField string

const (
	UserFieldUserID UserField = "user_id"
	UserFieldEmail  UserField = "email"
)

type User interface {
	UserID() field.String
	Email() field.String
}

type Users []User

func (users Users) UserIDs() map[string]User {
	m := make(map[string]User, len(users))
	for _, user := range users {
		m[user.UserID().String()] = user
	}
	return m
}

func ValidateUsers(users Users) error {
	for i, user := range users {
		if user.Email().String() == "" {
			return InvalidFieldError{
				EntityName: "user",
				Index:      i,
				FieldName:  string(UserFieldEmail),
			}
		}
	}
	return nil
}

/*func ValidateUsers(users Users) error {
	for i, user := range users {
		ValidateUser(user)
	}
}*/
