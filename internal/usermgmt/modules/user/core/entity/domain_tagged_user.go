package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainTaggedUser interface {
	valueobj.HasTagID
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type TaggedUserWillBeDelegated struct {
	valueobj.HasTagID
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

func DelegateToTaggedUser(user valueobj.HasUserID, tag valueobj.HasTagID, org valueobj.HasOrganizationID) DomainTaggedUser {
	return &TaggedUserWillBeDelegated{
		HasTagID:          tag,
		HasUserID:         user,
		HasOrganizationID: org,
	}
}

type DomainTaggedUsers []DomainTaggedUser

func (taggedUsers DomainTaggedUsers) UserIDs() []string {
	userIDs := []string{}
	for _, taggedUser := range taggedUsers {
		userIDs = append(userIDs, taggedUser.UserID().String())
	}

	return userIDs
}

func (taggedUsers DomainTaggedUsers) TagIDs() []string {
	tagIDs := []string{}
	for _, taggedUser := range taggedUsers {
		tagIDs = append(tagIDs, taggedUser.TagID().String())
	}

	return tagIDs
}

type EmptyDomainTaggedUser struct{}

func (e *EmptyDomainTaggedUser) TagID() field.String {
	return field.NewUndefinedString()
}

func (e *EmptyDomainTaggedUser) UserID() field.String {
	return field.NewUndefinedString()
}

func (e *EmptyDomainTaggedUser) OrganizationID() field.String {
	return field.NewNullString()
}
