package app

import (
	tom_const "github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"golang.org/x/exp/slices"
)

func IsStaff(userGroup string) bool {
	staffList := []string{
		constant.UserGroupAdmin,
		constant.UserGroupTeacher,
		constant.UserGroupSchoolAdmin,
		constant.UserGroupOrganizationManager,
	}

	return slices.Contains(staffList, userGroup)
}

// return Student, Parent location config keys
func GetLocationConfigKeys(unleashEnabled bool) (string, string) {
	if unleashEnabled {
		return tom_const.ChatConfigKeyStudentV2, tom_const.ChatConfigKeyParentV2
	}

	return tom_const.ChatConfigKeyStudent, tom_const.ChatConfigKeyParent
}
