package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type User struct {
	RandomUser
}

type RandomUser struct {
	entity.EmptyUser
	UserID         field.String
	Email          field.String
	UserName       field.String
	ExternalUserID field.String
	DeactivatedAt  field.Time
	LoginEmail     field.String
}

func (user User) UserName() field.String {
	return user.RandomUser.UserName
}
func (user User) Email() field.String {
	return user.RandomUser.Email
}
func (user User) UserID() field.String {
	return user.RandomUser.UserID
}
func (user User) ExternalUserID() field.String {
	return user.RandomUser.ExternalUserID
}
func (user User) DeactivatedAt() field.Time {
	return user.RandomUser.DeactivatedAt
}
func (user User) LoginEmail() field.String {
	return user.RandomUser.LoginEmail
}
