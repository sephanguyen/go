package entity

import (
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

const (
	AddressTypeHomeAddress    = "HOME_ADDRESS"
	AddressTypeBillingAddress = "BILLING_ADDRESS"
)

const (
	UserAddressFieldAddressType = "address_type"
)

type UserAddressAttribute interface {
	UserAddressID() field.String
	AddressType() field.String
	PostalCode() field.String
	City() field.String
	FirstStreet() field.String
	SecondStreet() field.String
}

type DomainUserAddress interface {
	UserAddressAttribute
	valueobj.HasUserID
	valueobj.HasPrefectureID
	valueobj.HasOrganizationID
}

type UserAddressWillBeDelegated struct {
	UserAddressAttribute
	valueobj.HasUserID
	valueobj.HasPrefectureID
	valueobj.HasOrganizationID
}

type DefaultDomainUserAddress struct{}

func (userAddress DefaultDomainUserAddress) UserAddressID() field.String {
	return field.NewString(idutil.ULIDNow())
}

func (userAddress DefaultDomainUserAddress) AddressType() field.String {
	return field.NewString(AddressTypeHomeAddress)
}

func (userAddress DefaultDomainUserAddress) PostalCode() field.String {
	return field.NewNullString()
}

func (userAddress DefaultDomainUserAddress) City() field.String {
	return field.NewNullString()
}

func (userAddress DefaultDomainUserAddress) FirstStreet() field.String {
	return field.NewNullString()
}

func (userAddress DefaultDomainUserAddress) SecondStreet() field.String {
	return field.NewNullString()
}

func (userAddress DefaultDomainUserAddress) UserID() field.String {
	return field.NewUndefinedString()
}

func (userAddress DefaultDomainUserAddress) PrefectureID() field.String {
	return field.NewNullString()
}

func (userAddress DefaultDomainUserAddress) OrganizationID() field.String {
	return field.NewNullString()
}

func ValidUserAddress(userAddress DomainUserAddress) error {
	err := errorx.ReturnFirstErr(
		validateUserAddressAddressType(userAddress),
	)
	if err != nil {
		return err
	}
	return nil
}

func validateUserAddressAddressType(userAddress DomainUserAddress) error {
	switch userAddress.AddressType().Status() {
	case field.StatusUndefined, field.StatusNull:
		return errcode.Error{
			Code:      errcode.MissingMandatory,
			FieldName: UserAddressFieldAddressType,
		}
	}
	switch userAddress.AddressType().String() {
	case AddressTypeHomeAddress, AddressTypeBillingAddress:
		return nil
	default:
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: UserAddressFieldAddressType,
		}
	}
}

type DomainUserAddresses []DomainUserAddress

func (userAddresses DomainUserAddresses) UserIDs() []string {
	userIDs := []string{}
	for _, userAddress := range userAddresses {
		userIDs = append(userIDs, userAddress.UserID().String())
	}

	return userIDs
}
