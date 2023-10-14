package entity

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

type mockOrganization struct{}

func (mockOrganization *mockOrganization) OrganizationID() field.String {
	return field.NewString("1")
}

func (mockOrganization *mockOrganization) SchoolID() field.Int32 {
	return field.NewInt32(1)
}

func TestDelegateSchoolAdmin(t *testing.T) {
	schoolAdmin := NullDomainSchoolAdmin{}
	organization := &mockOrganization{}

	delegatedSchoolAdmin := &SchoolAdminToDelegate{
		DomainSchoolAdminProfile: schoolAdmin,
		HasOrganizationID:        organization,
		HasSchoolID:              organization,
	}

	assert.Equal(t, delegatedSchoolAdmin.OrganizationID(), organization.OrganizationID())
	assert.Equal(t, delegatedSchoolAdmin.SchoolID(), organization.SchoolID())
}
