package constants

type BookType string

const (
	Staging = "stag"
	Local   = "local"
)

const (
	RoleSchoolAdmin = "School Admin"
	RoleTeacher     = "Teacher"
	RoleParent      = "Parent"
	RoleStudent     = "Student"

	RoleHQStaff       = "HQ Staff"
	RoleCentreManager = "Centre Manager"
	RoleCentreLead    = "Centre Lead"
	RoleCentreStaff   = "Centre Staff"
	RoleTeacherLead   = "Teacher Lead"
)

// limitParamOfQuery: postgresql limited parameters are 65535 for each query
const LimitParamOfQuery = 65535
