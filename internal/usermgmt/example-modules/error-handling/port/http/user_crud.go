package http

import (
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/gin-gonic/gin"
)

type User struct {
	UserIDAttr string `json:"user_id"`
	EmailAttr  string `json:"email"`
}

func (user User) UserID() field.String {
	return field.NewString(user.UserIDAttr)
}
func (user User) Email() field.String {
	return field.NewString(user.EmailAttr)
}

type UpsertUserRequest struct {
	Users []User `json:"users"`
}

func (service *UserService) upsertUser(c *gin.Context) (interface{}, error) {
	// Parse request data
	request := new(UpsertUserRequest)
	if err := JSONDecode(c.Request.Body, request); err != nil {
		return nil, err
	}

	usersToUpsert := make(entity.Users, len(request.Users))
	for i := range usersToUpsert {
		usersToUpsert[i] = request.Users[i]
	}

	// Use domain service to create user
	err := service.UserDomainService.UpsertUsers(c.Request.Context(), usersToUpsert)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

type GetUserRequest struct {
	UserIDs field.Strings `json:"user_ids"`
}

func (service *UserService) getUser(c *gin.Context) (interface{}, error) {
	request := new(GetUserRequest)
	if err := JSONDecode(c.Request.Body, request); err != nil {
		return nil, err
	}

	users, err := service.UserDomainService.GetUsers(c.Request.Context(), request.UserIDs)
	if err != nil {
		return nil, err
	}

	return users, nil
}
