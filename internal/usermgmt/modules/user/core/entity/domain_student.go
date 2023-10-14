package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
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

const (
	StudentFieldCurrentGrade              = "current_grade"
	StudentFieldEnrollmentStatus          = "enrollment_status"
	StudentFieldEnrollmentStatusStartDate = "status_start_date"
	StudentFieldContactPreference         = "phone_number.contact_preference"
	StudentFieldStudentPhoneNumber        = "student_phone_number"
	StudentFieldHomePhoneNumber           = "home_phone_number"
	StudentFieldPrimaryPhoneNumber        = "primary_phone_number"
	StudentFieldSecondaryPhoneNumber      = "secondary_phone_number"
	StudentSchoolHistoryField             = "school_histories"
	StudentSchoolField                    = "school"
	StudentGradeField                     = "grade"
	StudentSchoolCourseField              = "school_course"
	StudentSchoolHistoryStartDateField    = "start_date"
	StudentSchoolHistoryEndDateField      = "end_date"
	StudentSchoolHistorySchoolLevel       = "school_level"
	StudentUserAddressPrefectureField     = "prefecture"
	StudentLocationsField                 = "location"
	StudentLocationTypeField              = "location_type"
	StudentTagsField                      = "student_tag"
	StudentGenderField                    = "gender"
	StudentBirthdayField                  = "birthday"
)

type DomainStudentProfile interface {
	UserProfile
	ExternalStudentID() field.String
	CurrentGrade() field.Int16
	EnrollmentStatus() field.String
	StudentNote() field.String
	ContactPreference() field.String
}

type DomainStudent interface {
	DomainStudentProfile
	valueobj.HasCountry
	valueobj.HasOrganizationID
	valueobj.HasSchoolID
	valueobj.HasGradeID
	valueobj.HasUserID
	valueobj.HasLoginEmail
}

type StudentWillBeDelegated struct {
	DomainStudentProfile
	valueobj.HasCountry
	valueobj.HasOrganizationID
	valueobj.HasSchoolID
	valueobj.HasGradeID
	valueobj.HasUserID
	valueobj.HasLoginEmail
}

type NullDomainStudent struct {
	EmptyUser
}

func (student NullDomainStudent) SchoolID() field.Int32 {
	return field.NewNullInt32()
}

func (student NullDomainStudent) CurrentGrade() field.Int16 {
	return field.NewNullInt16()
}

func (student NullDomainStudent) EnrollmentStatus() field.String {
	return field.NewNullString()
}

func (student NullDomainStudent) StudentNote() field.String {
	return field.NewNullString()
}

func (student NullDomainStudent) GradeID() field.String {
	return field.NewNullString()
}

func (student NullDomainStudent) Group() field.String {
	return field.NewString(constant.UserGroupStudent)
}

func (student NullDomainStudent) ContactPreference() field.String {
	return field.NewNullString()
}

func (student NullDomainStudent) ExternalStudentID() field.String {
	return field.NewNullString()
}

func (student NullDomainStudent) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}
