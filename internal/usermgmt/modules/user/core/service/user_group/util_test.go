package user_group

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/stretchr/testify/assert"
)

func Test_combineRolesToLegacyUserGroup(t *testing.T) {
	t.Parallel()

	roleSchoolAdmin := &entity.Role{
		RoleName: database.Text(constant.RoleSchoolAdmin),
	}
	roleTeacher := &entity.Role{
		RoleName: database.Text(constant.RoleTeacher),
	}
	roleTeacherLead := &entity.Role{
		RoleName: database.Text(constant.RoleTeacherLead),
	}
	roleCentreManager := &entity.Role{
		RoleName: database.Text(constant.RoleCentreManager),
	}
	roleCentreLead := &entity.Role{
		RoleName: database.Text(constant.RoleCentreLead),
	}
	roleCentreStaff := &entity.Role{
		RoleName: database.Text(constant.RoleCentreStaff),
	}
	roleHQStaff := &entity.Role{
		RoleName: database.Text(constant.RoleHQStaff),
	}

	tests := []struct {
		name        string
		roles       []*entity.Role
		expected    string
		expectedErr error
	}{
		{
			name:        "1 role: school admin",
			roles:       []*entity.Role{roleSchoolAdmin},
			expectedErr: nil,
			expected:    constant.MapRoleWithLegacyUserGroup[roleSchoolAdmin.RoleName.String],
		},
		{
			name:        "1 role: teacher",
			roles:       []*entity.Role{roleTeacher},
			expectedErr: nil,
			expected:    constant.MapRoleWithLegacyUserGroup[roleTeacher.RoleName.String],
		},
		{
			name:        "combine teacher with teacher lead",
			roles:       []*entity.Role{roleTeacherLead, roleTeacher},
			expectedErr: nil,
			expected:    constant.MapRoleWithLegacyUserGroup[roleTeacher.RoleName.String],
		},
		{
			name:        "combine centre manager with centre lead",
			roles:       []*entity.Role{roleCentreLead, roleCentreManager},
			expectedErr: nil,
			expected:    constant.MapRoleWithLegacyUserGroup[roleCentreLead.RoleName.String],
		},
		{
			name:        "combine school admin with teacher",
			roles:       []*entity.Role{roleSchoolAdmin, roleTeacher},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine centre manager with teacher",
			roles:       []*entity.Role{roleCentreManager, roleTeacher},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine school admin with centre manager",
			roles:       []*entity.Role{roleCentreManager, roleSchoolAdmin},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine school admin with hq staff",
			roles:       []*entity.Role{roleHQStaff, roleSchoolAdmin},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine hq staff with centre manager",
			roles:       []*entity.Role{roleSchoolAdmin, roleTeacher},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine centre staff with centre manager",
			roles:       []*entity.Role{roleCentreStaff, roleCentreManager},
			expectedErr: errNotAllowedCombinationRole,
		},
		{
			name:        "combine centre staff with centre lead",
			roles:       []*entity.Role{roleCentreStaff, roleCentreManager},
			expectedErr: errNotAllowedCombinationRole,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {

			legacyUserGroup, err := combineRolesToLegacyUserGroup(testCase.roles)
			assert.Equal(t, testCase.expected, legacyUserGroup)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
