package helpers

import "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

const (
	OrganizationTypeNew     = "organization type new"
	OrganizationTypeJPREP   = "JPREP organization"
	OrganizationTypeManabie = "Manabie organization"
	OrganizationTypeDefault = "Manabie organization"

	StaffGrantedRoleSchoolAdmin = "staff granted role school admin"
	StaffGrantedRoleTeacher     = "staff granted role teacher"
	StaffGrantedRoleTeacherLead = "staff granted role teacher lead"
	// StaffGrantedRoleSchoolAdminAndTeacher = "staff granted role school admin and teacher"
	StaffGrantedRoleHQStaff       = "staff granted role hq staff"
	StaffGrantedRoleCentreLead    = "staff granted role centre lead"
	StaffGrantedRoleCentreManager = "staff granted role centre manager"
	StaffGrantedRoleCentreStaff   = "staff granted role centre staff"

	NumberOfNewCenterLocationCreated = 5

	OrganizationLocationTypeName = "org"
	BrandLocationTypeName        = "brand"
	CenterLocationTypeName       = "center"

	ManabieOrgLocation     = "01FR4M51XJY9E77GSN4QZ1Q9N1"
	ManabieOrgLocationType = "01FR4M51XJY9E77GSN4QZ1Q9M1"
	JPREPOrgLocation       = "01FR4M51XJY9E77GSN4QZ1Q9N2"
	JPREPOrgLocationType   = "01FR4M51XJY9E77GSN4QZ1Q9M2"

	JPREPResourcePath = -2147483647
)

var (
	OrganizationRoles = []string{
		constant.RoleTeacher,
		constant.RoleSchoolAdmin,
		constant.RoleStudent,
		constant.RoleParent,
		constant.RoleHQStaff,
		constant.RoleCentreManager,
		constant.RoleCentreStaff,
		constant.RoleCentreLead,
		constant.RoleTeacherLead,
	}
)
