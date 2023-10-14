package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Location interface {
	LocationID() field.String
	Name() field.String
	IsArchived() field.Boolean
}

type DomainLocation interface {
	Location
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type LocationWillBeDelegated struct {
	Location
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type NullDomainLocation struct{}

func (location NullDomainLocation) LocationID() field.String {
	return field.NewNullString()
}

func (location NullDomainLocation) PartnerInternalID() field.String {
	return field.NewNullString()
}

func (location NullDomainLocation) OrganizationID() field.String {
	return field.NewNullString()
}
func (location NullDomainLocation) Name() field.String {
	return field.NewNullString()
}
func (location NullDomainLocation) IsArchived() field.Boolean {
	return field.NewNullBoolean()
}

type DomainLocations []DomainLocation

func (locations DomainLocations) LocationIDs() []string {
	locationIDs := []string{}
	for _, location := range locations {
		locationIDs = append(locationIDs, location.LocationID().String())
	}
	return locationIDs
}

func (locations DomainLocations) PartnerInternalIDs() []string {
	partnerInternalIDs := []string{}
	for _, location := range locations {
		partnerInternalIDs = append(partnerInternalIDs, location.PartnerInternalID().String())
	}
	return partnerInternalIDs
}

func (locations DomainLocations) ToUserAccessPath(user valueobj.HasUserID) DomainUserAccessPaths {
	userAccessPaths := make(DomainUserAccessPaths, 0, len(locations))
	for _, location := range locations {
		userAccessPaths = append(userAccessPaths, UserAccessPathWillBeDelegated{
			HasUserID:         user,
			HasLocationID:     location,
			HasOrganizationID: location,
		})
	}
	return userAccessPaths
}
