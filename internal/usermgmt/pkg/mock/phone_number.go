package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserPhoneNumber struct {
	entity.DefaultDomainUserPhoneNumber

	phoneNumber     field.String
	phoneNumberType field.String
}

type RandomUserPhoneNumber struct {
	entity.DefaultDomainUserPhoneNumber
	PrefectureID   field.String
	PrefectureCode field.String
}

func NewUserPhoneNumber(phoneNumber, phoneNumberType string) *UserPhoneNumber {
	return &UserPhoneNumber{
		phoneNumber:     field.NewString(phoneNumber),
		phoneNumberType: field.NewString(phoneNumberType),
	}
}

func (u *UserPhoneNumber) PhoneNumber() field.String {
	return u.phoneNumber
}

func (u *UserPhoneNumber) Type() field.String {
	return u.phoneNumberType
}
