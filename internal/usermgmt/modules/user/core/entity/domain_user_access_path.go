package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainUserAccessPath interface {
	valueobj.HasUserID
	valueobj.HasLocationID
	valueobj.HasOrganizationID
}

type UserAccessPathWillBeDelegated struct {
	valueobj.HasUserID
	valueobj.HasLocationID
	valueobj.HasOrganizationID
}

type DefaultUserAccessPath struct{}

func (uap DefaultUserAccessPath) UserID() field.String {
	return field.NewNullString()
}

func (uap DefaultUserAccessPath) LocationID() field.String {
	return field.NewNullString()
}

func (uap DefaultUserAccessPath) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainUserAccessPaths []DomainUserAccessPath

func (userAccessPaths DomainUserAccessPaths) LocationIDs() []string {
	locationIDs := make([]string, 0, len(userAccessPaths))
	for _, uap := range userAccessPaths {
		locationIDs = append(locationIDs, uap.LocationID().String())
	}
	return locationIDs
}

func (userAccessPaths DomainUserAccessPaths) UserIDs() []string {
	userIDs := []string{}
	for _, uap := range userAccessPaths {
		userIDs = append(userIDs, uap.UserID().String())
	}
	return userIDs
}
