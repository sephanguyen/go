package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Location struct {
	entity.NullDomainLocation
	LocationIDAttr        field.String
	PartnerInternalIDAttr field.String
}

func NewLocation(locationIDAttr, partnerInternalIDAttr string) *Location {
	return &Location{
		LocationIDAttr:        field.NewString(locationIDAttr),
		PartnerInternalIDAttr: field.NewString(partnerInternalIDAttr),
	}
}

func (location Location) LocationID() field.String {
	return location.LocationIDAttr
}

func (location Location) PartnerInternalID() field.String {
	return location.PartnerInternalIDAttr
}
