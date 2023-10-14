package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserAddress struct {
	RandomUserAddress
}

type RandomUserAddress struct {
	entity.DefaultDomainUserAddress
	UserAddressID field.String
	PostalCode    field.String
	PrefectureID  field.String
}

func (s UserAddress) UserAddressID() field.String {
	return s.RandomUserAddress.UserAddressID
}
func (s UserAddress) PostalCode() field.String {
	return s.RandomUserAddress.PostalCode
}
func (s UserAddress) PrefectureID() field.String {
	return s.RandomUserAddress.PrefectureID
}
