package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserAccessPath struct {
	RandomUserAccessPath
}

type RandomUserAccessPath struct {
	entity.DefaultUserAccessPath
	LocationID field.String
	UserID     field.String
}

func (u UserAccessPath) LocationID() field.String {
	return u.RandomUserAccessPath.LocationID
}
func (u UserAccessPath) UserID() field.String {
	return u.RandomUserAccessPath.UserID
}
