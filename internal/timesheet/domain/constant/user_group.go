package constant

const (
	UserGroupStudent     = "USER_GROUP_STUDENT"
	UserGroupAdmin       = "USER_GROUP_ADMIN"
	UserGroupTeacher     = "USER_GROUP_TEACHER"
	UserGroupParent      = "USER_GROUP_PARENT"
	UserGroupSchoolAdmin = "USER_GROUP_SCHOOL_ADMIN"

	RoleSchoolAdmin   = "School Admin"
	RoleHQStaff       = "HQ Staff"
	RoleCentreManager = "Centre Manager"
	RoleCentreLead    = "Centre Lead"
	RoleCentreStaff   = "Centre Staff"
	RoleTeacherLead   = "Teacher Lead"
	RoleTeacher       = "Teacher"
	RoleParent        = "Parent"
	RoleStudent       = "Student"
)

var RolesWriteOtherMemberTimesheet = map[string]struct{}{
	RoleSchoolAdmin:   {},
	RoleHQStaff:       {},
	RoleCentreManager: {},
}
