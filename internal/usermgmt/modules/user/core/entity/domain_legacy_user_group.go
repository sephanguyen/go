package entity

import (
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type LegacyUserGroupAttribute interface {
	GroupID() field.String
	IsOrigin() field.Boolean
	Status() field.String
}

type LegacyUserGroupAttributes []LegacyUserGroupAttribute

type LegacyUserGroup interface {
	LegacyUserGroupAttribute
	valueobj.HasOrganizationID
	valueobj.HasUserID
}

type LegacyUserGroupWillBeDelegated struct {
	LegacyUserGroupAttribute
	valueobj.HasOrganizationID
	valueobj.HasUserID
}

func DelegateToLegacyUserGroup(legacyUserGroup LegacyUserGroupAttribute, organization valueobj.HasOrganizationID, user valueobj.HasUserID) LegacyUserGroup {
	delegatedLegacyUserGroup := &LegacyUserGroupWillBeDelegated{
		LegacyUserGroupAttribute: legacyUserGroup,
		HasOrganizationID:        organization,
		HasUserID:                user,
	}
	return delegatedLegacyUserGroup
}

/*func DelegateToLegacyUserGroups(legacyUserGroups LegacyUserGroupAttributes, organization valueobj.HasOrganizationID, user valueobj.HasUserID) LegacyUserGroups {
	delegatedLegacyUserGroups := make(LegacyUserGroups, 0, len(legacyUserGroups))
	for _, legacyUserGroup := range legacyUserGroups {
		delegatedLegacyUserGroups = append(delegatedLegacyUserGroups, DelegateToLegacyUserGroup(legacyUserGroup, organization, user))
	}
	return delegatedLegacyUserGroups
}*/

type LegacyUserGroups []LegacyUserGroup

type EmptyLegacyUserGroup struct{}

func (legacyUserGroup EmptyLegacyUserGroup) GroupID() field.String {
	return field.NewString(idutil.ULIDNow())
}

func (legacyUserGroup EmptyLegacyUserGroup) IsOrigin() field.Boolean {
	return field.NewBoolean(false)
}

func (legacyUserGroup EmptyLegacyUserGroup) Status() field.String {
	return field.NewNullString()
}

func (legacyUserGroup EmptyLegacyUserGroup) OrganizationID() field.String {
	return field.NewNullString()
}

func (legacyUserGroup EmptyLegacyUserGroup) UserID() field.String {
	return field.NewNullString()
}

type ActiveAndOriginatedLegacyUserGroup struct {
	EmptyLegacyUserGroup
}

func (legacyUserGroup ActiveAndOriginatedLegacyUserGroup) IsOrigin() field.Boolean {
	return field.NewBoolean(true)
}

func (legacyUserGroup ActiveAndOriginatedLegacyUserGroup) Status() field.String {
	return field.NewString("USER_GROUP_STATUS_ACTIVE")
}

type SchoolAdminLegacyUserGroup struct {
	ActiveAndOriginatedLegacyUserGroup
}

func (schoolAdminLegacyUserGroup SchoolAdminLegacyUserGroup) GroupID() field.String {
	return field.NewString(constant.UserGroupSchoolAdmin)
}

type StudentLegacyUserGroup struct {
	ActiveAndOriginatedLegacyUserGroup
}

func (studentLegacyUserGroup StudentLegacyUserGroup) GroupID() field.String {
	return field.NewString(constant.UserGroupStudent)
}

type ParentLegacyUserGroup struct {
	ActiveAndOriginatedLegacyUserGroup
}

func (parentLegacyUserGroup ParentLegacyUserGroup) GroupID() field.String {
	return field.NewString(constant.UserGroupParent)
}

func (parentLegacyUserGroup ParentLegacyUserGroup) IsOrigin() field.Boolean {
	return field.NewBoolean(true)
}
func (parentLegacyUserGroup ParentLegacyUserGroup) Status() field.String {
	return field.NewString(UserGroupStatusActive)
}
