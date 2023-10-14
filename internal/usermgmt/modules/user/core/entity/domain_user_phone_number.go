package entity

import (
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

const (
	UserPhoneNumberTypeStudentPhoneNumber         = "STUDENT_PHONE_NUMBER"
	UserPhoneNumberTypeStudentHomePhoneNumber     = "STUDENT_HOME_PHONE_NUMBER"
	UserPhoneNumberTypeParentPrimaryPhoneNumber   = "PARENT_PRIMARY_PHONE_NUMBER"
	UserPhoneNumberTypeParentSecondaryPhoneNumber = "PARENT_SECONDARY_PHONE_NUMBER"
)

type UserPhoneNumberAttribute interface {
	UserPhoneNumberID() field.String
	PhoneNumber() field.String
	Type() field.String
}

type DomainUserPhoneNumber interface {
	UserPhoneNumberAttribute
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type UserPhoneNumberWillBeDelegated struct {
	UserPhoneNumberAttribute
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type DefaultDomainUserPhoneNumber struct{}

func (userPhoneNumber DefaultDomainUserPhoneNumber) UserPhoneNumberID() field.String {
	return field.NewString(idutil.ULIDNow())
}
func (userPhoneNumber DefaultDomainUserPhoneNumber) PhoneNumber() field.String {
	return field.NewNullString()
}
func (userPhoneNumber DefaultDomainUserPhoneNumber) Type() field.String {
	return field.NewString(UserPhoneNumberTypeStudentPhoneNumber)
}
func (userPhoneNumber DefaultDomainUserPhoneNumber) UserID() field.String {
	return field.NewNullString()
}
func (userPhoneNumber DefaultDomainUserPhoneNumber) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainUserPhoneNumbers []DomainUserPhoneNumber

func (userPhoneNumbers DomainUserPhoneNumbers) UserIDs() []string {
	userIDs := []string{}
	for _, userPhoneNumber := range userPhoneNumbers {
		userIDs = append(userIDs, userPhoneNumber.UserID().String())
	}

	return userIDs
}

func (userPhoneNumbers DomainUserPhoneNumbers) PhoneNumbers() []string {
	phoneNumbers := []string{}
	for _, userPhoneNumber := range userPhoneNumbers {
		phoneNumbers = append(phoneNumbers, userPhoneNumber.PhoneNumber().String())
	}

	return phoneNumbers
}
