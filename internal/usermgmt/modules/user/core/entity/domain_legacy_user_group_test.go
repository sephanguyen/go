package entity

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

type mockUser struct{}

func (mockUser *mockUser) UserID() field.String {
	return field.NewString("example-uid")
}

func TestDelegateToLegacyUserGroup(t *testing.T) {
	legacyUserGroup := EmptyLegacyUserGroup{}
	organization := &mockOrganization{}
	user := &mockUser{}

	delegatedLegacyUserGroup := DelegateToLegacyUserGroup(legacyUserGroup, organization, user)

	assert.Equal(t, delegatedLegacyUserGroup.OrganizationID(), organization.OrganizationID())
	assert.Equal(t, delegatedLegacyUserGroup.UserID(), user.UserID())
}
