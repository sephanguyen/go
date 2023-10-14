package constant

import (
	"time"
)

const (
	UserGroupStudent             = "USER_GROUP_STUDENT"
	UserGroupAdmin               = "USER_GROUP_ADMIN"
	UserGroupTeacher             = "USER_GROUP_TEACHER"
	UserGroupParent              = "USER_GROUP_PARENT"
	UserGroupSchoolAdmin         = "USER_GROUP_SCHOOL_ADMIN"
	UserGroupOrganizationManager = "USER_GROUP_ORGANIZATION_MANAGER"

	RoleSchoolAdmin         = "School Admin"
	RoleHQStaff             = "HQ Staff"
	RoleCentreManager       = "Centre Manager"
	RoleCentreLead          = "Centre Lead"
	RoleCentreStaff         = "Centre Staff"
	RoleTeacherLead         = "Teacher Lead"
	RoleTeacher             = "Teacher"
	RoleParent              = "Parent"
	RoleStudent             = "Student"
	RoleOpenAPI             = "OpenAPI"
	RoleUsermgmtScheduleJob = "UsermgmtScheduleJob"
	RoleReportReviewer      = "Report Reviewer"
	RolePaymentScheduleJob  = "PaymentScheduleJob"

	UserDeviceTokenNats            = "user_device_token"
	UserDeviceTokenNatsQueueGroup  = "user_device_token_queue"
	UserDeviceTokenNatsDurableName = "user_device_token_durable"

	InvalidLocations    = "INVALID_LOCATIONS"
	InvalidRemoveParent = "invalidRemoveParent"

	// permissions
	MasterLocationRead = "master.location.read"

	// config key
	KeyEnrollmentStatusHistoryConfig = "user.enrollment.update_status_manual"
	KeyIPRestrictionFeatureConfig    = "user.authentication.ip_address_restriction"
	KeyIPRestrictionWhitelistConfig  = "user.authentication.allowed_ip_address"
	KeyAuthUsernameConfig            = "user.auth.username"

	// config value on/off
	ConfigValueOff = "off"
	ConfigValueOn  = "on"

	IgnoreUpdateEmail = true
	AllowUpdateEmail  = false

	StatusSuccess = "success"
	StatusFailed  = "failed"

	LoginEmailPostfix = "@manabie.com"
	// Timezone
	JpTimeZone = "Asia/Tokyo"

	AppName             = "Manabie"
	JapanLanguageCode   = "ja"
	EnglishLanguageCode = "en"
)

const IdentityToolkitURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="

const (
	FeatureToggleAllowCombinationMultipleRoles = "User_Auth_AllowCombinationMultipleRoles"
)

var MapLegacyUserGroupWithRoles = map[string][]string{
	UserGroupSchoolAdmin: {RoleSchoolAdmin},
	UserGroupTeacher:     {RoleTeacher},
	UserGroupParent:      {RoleParent},
	UserGroupStudent:     {RoleStudent},
}

var MapRoleWithLegacyUserGroup = map[string]string{
	RoleSchoolAdmin:         UserGroupSchoolAdmin,
	RoleHQStaff:             UserGroupSchoolAdmin,
	RoleCentreLead:          UserGroupSchoolAdmin,
	RoleCentreManager:       UserGroupSchoolAdmin,
	RoleCentreStaff:         UserGroupSchoolAdmin,
	RoleUsermgmtScheduleJob: UserGroupSchoolAdmin,

	RoleTeacher:     UserGroupTeacher,
	RoleTeacherLead: UserGroupTeacher,

	RoleParent:  UserGroupParent,
	RoleStudent: UserGroupStudent,
}

var MapCombinationRole = map[string][]string{
	RoleSchoolAdmin:   nil,
	RoleHQStaff:       nil,
	RoleCentreLead:    {RoleCentreLead, RoleCentreManager},
	RoleCentreManager: {RoleCentreManager, RoleCentreLead},
	RoleCentreStaff:   nil,

	RoleTeacher:     {RoleTeacher, RoleTeacherLead},
	RoleTeacherLead: {RoleTeacherLead, RoleTeacher},
}

var AllowListRoles = []string{
	RoleSchoolAdmin,
	RoleTeacher,
	RoleParent,
	RoleStudent,
	RoleHQStaff,
	RoleCentreLead,
	RoleCentreManager,
	RoleCentreStaff,
	RoleTeacherLead,
	RoleOpenAPI,
	RoleUsermgmtScheduleJob,
	RoleReportReviewer,
	RolePaymentScheduleJob,
}

// MapRoleToLegacyUserGroup to allow staff use feature on teacher web
var MapRoleToLegacyUserGroup = map[string][]string{
	RoleSchoolAdmin:   {UserGroupTeacher, UserGroupSchoolAdmin},
	RoleHQStaff:       {UserGroupTeacher, UserGroupSchoolAdmin},
	RoleCentreLead:    {UserGroupTeacher, UserGroupSchoolAdmin},
	RoleCentreManager: {UserGroupTeacher, UserGroupSchoolAdmin},
	RoleCentreStaff:   {UserGroupTeacher, UserGroupSchoolAdmin},

	RoleTeacher:     {UserGroupTeacher},
	RoleTeacherLead: {UserGroupTeacher},

	RoleParent:  {UserGroupParent},
	RoleStudent: {UserGroupStudent},
}

var ConversationStaffRoles = []string{
	UserGroupAdmin,
	UserGroupSchoolAdmin,
	UserGroupOrganizationManager,
	UserGroupTeacher,
}

type FamilyRelationship string

const (
	FamilyRelationshipNone        FamilyRelationship = "FAMILY_RELATIONSHIP_NONE"
	FamilyRelationshipFather      FamilyRelationship = "FAMILY_RELATIONSHIP_FATHER"
	FamilyRelationshipMother      FamilyRelationship = "FAMILY_RELATIONSHIP_MOTHER"
	FamilyRelationshipGrandfather FamilyRelationship = "FAMILY_RELATIONSHIP_GRANDFATHER"
	FamilyRelationshipGrandmother FamilyRelationship = "FAMILY_RELATIONSHIP_GRANDMOTHER"
	FamilyRelationshipUncle       FamilyRelationship = "FAMILY_RELATIONSHIP_UNCLE"
	FamilyRelationshipAunt        FamilyRelationship = "FAMILY_RELATIONSHIP_AUNT"
	FamilyRelationshipOther       FamilyRelationship = "FAMILY_RELATIONSHIP_OTHER"
)

const JSMMaxDeliver = 10
const JSMAckWait = 30 * time.Second
const NumberOfGoroutines = 10

// regex pattern
const (
	// The FE team is using the pattern to validate user email. So we also use it for compatibility.
	EmailPattern                    = `^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`
	PhoneNumberPattern              = `^\d{7,20}$`
	ExtractTextBetweenQuotesPattern = `["']([^"']*)["']`
	UsernamePattern                 = `^[a-zA-Z0-9]+$`
)

const (
	DateLayout = "2006/01/02"
)

const (
	// LimitRowsCSV limit number of rows in csv for import parent, student, ...
	LimitRowsCSV      = 1000
	ArraySeparatorCSV = ";"
)

var (
	StaffGroupRole = []string{RoleTeacher, RoleSchoolAdmin, RoleHQStaff, RoleCentreStaff, RoleCentreLead, RoleCentreManager}

	UserGenderMap = map[int]string{
		1: UserGenderMale,
		2: UserGenderFemale,
	}
)

const (
	StudentPhoneNumber         = "STUDENT_PHONE_NUMBER"
	StudentHomePhoneNumber     = "STUDENT_HOME_PHONE_NUMBER"
	ParentPrimaryPhoneNumber   = "PARENT_PRIMARY_PHONE_NUMBER"
	ParentSecondaryPhoneNumber = "PARENT_SECONDARY_PHONE_NUMBER"
	StaffPrimaryPhoneNumber    = "STAFF_PRIMARY_PHONE_NUMBER"
	StaffSecondaryPhoneNumber  = "STAFF_SECONDARY_PHONE_NUMBER"
)

const (
	UserGenderMale   = "MALE"
	UserGenderFemale = "FEMALE"
)

const (
	StudentEnrollmentStatusPotential    = "STUDENT_ENROLLMENT_STATUS_POTENTIAL"
	StudentEnrollmentStatusEnrolled     = "STUDENT_ENROLLMENT_STATUS_ENROLLED"
	StudentEnrollmentStatusWithdrawn    = "STUDENT_ENROLLMENT_STATUS_WITHDRAWN"
	StudentEnrollmentStatusGraduated    = "STUDENT_ENROLLMENT_STATUS_GRADUATED"
	StudentEnrollmentStatusLOA          = "STUDENT_ENROLLMENT_STATUS_LOA"
	StudentEnrollmentStatusTemporary    = "STUDENT_ENROLLMENT_STATUS_TEMPORARY"
	StudentEnrollmentStatusNonPotential = "STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL"
)

type UserRole string

const (
	UserRoleSystem  UserRole = "system"
	UserRoleStaff   UserRole = "staff"
	UserRoleStudent UserRole = "student"
	UserRoleParent  UserRole = "parent"
)
