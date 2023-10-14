package importstudent

import (
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

var (
	mapStudentContactPreference = map[string]string{
		"1": constant.StudentPhoneNumber,
		"2": constant.StudentHomePhoneNumber,
		"3": constant.ParentPrimaryPhoneNumber,
		"4": constant.ParentSecondaryPhoneNumber,
	}

	studentEnrollmentStatusMap = map[int16]string{
		1: constant.StudentEnrollmentStatusPotential,
		2: constant.StudentEnrollmentStatusEnrolled,
		3: constant.StudentEnrollmentStatusWithdrawn,
		4: constant.StudentEnrollmentStatusGraduated,
		5: constant.StudentEnrollmentStatusLOA,
		6: constant.StudentEnrollmentStatusTemporary,
		7: constant.StudentEnrollmentStatusNonPotential,
	}
)

// enforce impl interface
var (
	_ entity.DomainStudent                 = (*StudentCSV)(nil)
	_ entity.DomainUserAddress             = (*UserAddress)(nil)
	_ entity.DomainSchoolHistory           = (*SchoolHistory)(nil)
	_ entity.DomainEnrollmentStatusHistory = (*EnrollmentStatusHistory)(nil)
)

type StudentCSV struct {
	entity.NullDomainStudent

	// internal fields
	birthday             field.Date
	FullNameAttr         field.String
	FullNamePhoneticAttr field.String
	LoginEmailAttr       field.String

	// csv fields
	UserNameAttr                 field.String `csv:"username"`
	FirstNameAttr                field.String `csv:"first_name"`
	LastNameAttr                 field.String `csv:"last_name"`
	FirstNamePhoneticAttr        field.String `csv:"first_name_phonetic"`
	LastNamePhoneticAttr         field.String `csv:"last_name_phonetic"`
	EmailAttr                    field.String `csv:"email"`
	EnrollmentStatusAttr         field.String `csv:"enrollment_status"`
	GradeAttr                    field.String `csv:"grade"`
	BirthdayAttr                 field.String `csv:"birthday"`
	GenderAttr                   field.String `csv:"gender"`
	LocationAttr                 field.String `csv:"location"`
	PostalCodeAttr               field.String `csv:"postal_code"`
	PrefectureAttr               field.String `csv:"prefecture"`
	CityAttr                     field.String `csv:"city"`
	FirstStreetAttr              field.String `csv:"first_street"`
	SecondStreetAttr             field.String `csv:"second_street"`
	StudentPhoneNumberAttr       field.String `csv:"student_phone_number"`
	StudentHomePhoneNumberAttr   field.String `csv:"home_phone_number"`
	StudentContactPreferenceAttr field.String `csv:"contact_preference"`
	SchoolAttr                   field.String `csv:"school"`
	SchoolCourseAttr             field.String `csv:"school_course"`
	StartDateAttr                field.String `csv:"start_date"`
	EndDateAttr                  field.String `csv:"end_date"`
	StudentTagAttr               field.String `csv:"student_tag"`
	StatusStartDateAttr          field.String `csv:"status_start_date"`
	IDAttr                       field.String `csv:"user_id"`
	ExternalUserIDAttr           field.String `csv:"external_user_id"`
	NoteAttr                     field.String `csv:"remarks"`
	PasswordAttr                 field.String `csv:"password"`
}

func (student StudentCSV) UserID() field.String {
	return student.IDAttr
}

func (student StudentCSV) ExternalUserID() field.String {
	return student.ExternalUserIDAttr
}

func (student StudentCSV) UserName() field.String {
	username := strings.ToLower(student.UserNameAttr.TrimSpace().String())
	return field.NewString(username)
}

func (student StudentCSV) FirstName() field.String {
	return student.FirstNameAttr
}

func (student StudentCSV) LastName() field.String {
	return student.LastNameAttr
}

func (student StudentCSV) FullName() field.String {
	return student.FullNameAttr
}

func (student StudentCSV) FirstNamePhonetic() field.String {
	return student.FirstNamePhoneticAttr
}

func (student StudentCSV) LastNamePhonetic() field.String {
	return student.LastNamePhoneticAttr
}

func (student StudentCSV) FullNamePhonetic() field.String {
	return student.FullNamePhoneticAttr
}

func (student StudentCSV) Email() field.String {
	return student.EmailAttr
}

func (student StudentCSV) EnrollmentStatus() field.String {
	return student.EnrollmentStatusAttr
}

func (student StudentCSV) Birthday() field.Date {
	return student.birthday
}

func (student StudentCSV) Gender() field.String {
	return student.GenderAttr
}

func (student StudentCSV) StudentNote() field.String {
	switch {
	// When "note" column does not exist
	case field.IsUndefined(student.NoteAttr):
		return field.NewString("")
	// When "note" column exists but value is null
	case field.IsNull(student.NoteAttr):
		return field.NewString("")
	// When "note" column exists and has value
	default:
		return student.NoteAttr
	}
}

func (student StudentCSV) ContactPreference() field.String {
	return student.StudentContactPreferenceAttr
}

func (student StudentCSV) Password() field.String {
	return student.PasswordAttr
}

func (student StudentCSV) GradeID() field.String {
	return student.GradeAttr
}

func (student StudentCSV) LoginEmail() field.String {
	return student.LoginEmailAttr
}
func (student StudentCSV) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}

type UserAddress struct {
	entity.DefaultDomainUserAddress

	AddressIDAttr    field.String
	AddressTypeAttr  field.String
	PostalCodeAttr   field.String
	PrefectureAttr   field.String
	CityAttr         field.String
	FirstStreetAttr  field.String
	SecondStreetAttr field.String
}

func (u UserAddress) UserAddressID() field.String {
	return u.AddressIDAttr
}

func (u UserAddress) AddressType() field.String {
	return u.AddressTypeAttr
}

func (u UserAddress) PostalCode() field.String {
	return u.PostalCodeAttr
}

func (u UserAddress) City() field.String {
	return u.CityAttr
}

func (u UserAddress) PrefectureID() field.String {
	return u.PrefectureAttr
}

func (u UserAddress) FirstStreet() field.String {
	return u.FirstStreetAttr
}

func (u UserAddress) SecondStreet() field.String {
	return u.SecondStreetAttr
}

type SchoolHistory struct {
	entity.DefaultDomainSchoolHistory

	SchoolIDAttr       field.String
	SchoolCourseIDAttr field.String
	StartDateAttr      field.Time
	EndDateAttr        field.Time
}

func (s SchoolHistory) StartDate() field.Time {
	switch {
	case field.IsUndefined(s.StartDateAttr):
		return s.DefaultDomainSchoolHistory.StartDate()
	default:
		return s.StartDateAttr
	}
}

func (s SchoolHistory) EndDate() field.Time {
	switch {
	case field.IsUndefined(s.EndDateAttr):
		return s.DefaultDomainSchoolHistory.EndDate()
	default:
		return s.EndDateAttr
	}
}

func (s SchoolHistory) SchoolID() field.String {
	return s.SchoolIDAttr
}

func (s SchoolHistory) SchoolCourseID() field.String {
	switch {
	case field.IsUndefined(s.SchoolCourseIDAttr):
		return s.DefaultDomainSchoolHistory.SchoolCourseID()
	default:
		return s.SchoolCourseIDAttr
	}
}

type SchoolInfoImpl struct {
	entity.DefaultDomainSchool

	SchoolIDAttr                field.String
	SchoolPartnerInternalIDAttr field.String
}

func (s SchoolInfoImpl) PartnerInternalID() field.String {
	return s.SchoolPartnerInternalIDAttr
}

type SchoolCourseImpl struct {
	entity.DefaultDomainSchoolCourse

	SchoolCourseIDAttr                field.String
	SchoolCoursePartnerInternalIDAttr field.String
}

func (s SchoolCourseImpl) PartnerInternalID() field.String {
	return s.SchoolCoursePartnerInternalIDAttr
}

type EnrollmentStatusHistory struct {
	entity.DefaultDomainEnrollmentStatusHistory

	EnrollmentStatusAttr field.String
	LocationIDAttr       field.String
	UserIDAttr           field.String
	ResourcePathAttr     field.String
	StartDateAttr        field.Time
}

func (e EnrollmentStatusHistory) UserID() field.String {
	return e.UserIDAttr
}

func (e EnrollmentStatusHistory) EnrollmentStatus() field.String {
	return e.EnrollmentStatusAttr
}

func (e EnrollmentStatusHistory) StartDate() field.Time {
	switch {
	case field.IsUndefined(e.StartDateAttr):
		return e.DefaultDomainEnrollmentStatusHistory.StartDate()
	default:
		return e.StartDateAttr
	}
}

func (e EnrollmentStatusHistory) OrganizationID() field.String {
	return e.ResourcePathAttr
}

func (e EnrollmentStatusHistory) LocationID() field.String {
	return e.LocationIDAttr
}

type UserPhoneNumber struct {
	entity.DefaultDomainUserPhoneNumber

	PhoneIDAttr     field.String
	PhoneTypeAttr   field.String
	PhoneNumberAttr field.String
}

func (p UserPhoneNumber) UserPhoneNumberID() field.String {
	return p.PhoneIDAttr
}

func (p UserPhoneNumber) PhoneNumber() field.String {
	return p.PhoneNumberAttr
}

func (p UserPhoneNumber) Type() field.String {
	return p.PhoneTypeAttr
}

type LocationImpl struct {
	entity.NullDomainLocation
	LocationAttr                field.String
	LocationPartnerInternalAttr field.String
}

func (location LocationImpl) PartnerInternalID() field.String {
	return location.LocationPartnerInternalAttr
}

type TagImpl struct {
	entity.EmptyDomainTaggedUser

	StudentTagAttr field.String
}

func (t TagImpl) PartnerInternalID() field.String {
	return t.StudentTagAttr
}

type GradeImpl struct {
	entity.NullDomainGrade

	GradeAttr field.String
}

func (g GradeImpl) PartnerInternalID() field.String {
	return g.GradeAttr
}

type PrefectureImpl struct {
	entity.DefaultDomainPrefecture

	PrefectureCodeAttr field.String
}

func (u PrefectureImpl) PrefectureCode() field.String {
	return u.PrefectureCodeAttr
}
