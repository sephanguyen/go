package http

import (
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

var (
	StudentEnrollmentStatusMap = map[int]string{
		1: constant.StudentEnrollmentStatusPotential,
		2: constant.StudentEnrollmentStatusEnrolled,
		3: constant.StudentEnrollmentStatusWithdrawn,
		4: constant.StudentEnrollmentStatusGraduated,
		5: constant.StudentEnrollmentStatusLOA,
		6: constant.StudentEnrollmentStatusTemporary,
		7: constant.StudentEnrollmentStatusNonPotential,
	}

	StudentContactPreference = map[int]string{
		1: entity.UserPhoneNumberTypeStudentPhoneNumber,
		2: entity.UserPhoneNumberTypeStudentHomePhoneNumber,
		3: entity.UserPhoneNumberTypeParentPrimaryPhoneNumber,
		4: entity.UserPhoneNumberTypeParentSecondaryPhoneNumber,
	}
)

type UpsertStudentsRequest struct {
	Students []StudentProfile `json:"students"`
}

type StudentProfile struct {
	UserID            field.String `json:"-"`
	GradeID           field.String `json:"-"`
	StudentExternalID field.String `json:"-"`
	FullName          field.String `json:"-"`
	FullNamePhonetic  field.String `json:"-"`
	LoginEmail        field.String `json:"-"`

	ExternalUserID    field.String   `json:"external_user_id"`
	UserName          field.String   `json:"username"`
	FirstName         field.String   `json:"first_name"`
	LastName          field.String   `json:"last_name"`
	FirstNamePhonetic field.String   `json:"first_name_phonetic"`
	LastNamePhonetic  field.String   `json:"last_name_phonetic"`
	Email             field.String   `json:"email"`
	Password          field.String   `json:"password"`
	EnrollmentStatus  field.Int32    `json:"enrollment_status"`
	Grade             field.String   `json:"grade"`
	Birthday          field.Date     `json:"birthday"`
	Gender            field.Int32    `json:"gender"`
	Note              field.String   `json:"note"`
	Tags              []field.String `json:"tags"`
	Locations         []field.String `json:"locations"`

	SchoolHistories           []SchoolHistoryPayload           `json:"school_histories"`
	EnrollmentStatusHistories []EnrollmentStatusHistoryPayload `json:"enrollment_status_histories"`
	UserPhoneNumber           *PhoneNumberPayload              `json:"phone_number"`
	Address                   *AddressPayload                  `json:"address"`
}

type SchoolHistoryPayload struct {
	School       field.String `json:"school"`
	SchoolCourse field.String `json:"school_course"`
	StartDate    field.Date   `json:"start_date"`
	EndDate      field.Date   `json:"end_date"`
}

type DomainStudentImpl struct {
	entity.NullDomainStudent

	StudentProfile
}

func (p DomainStudentImpl) UserID() field.String {
	return p.StudentProfile.UserID
}
func (p DomainStudentImpl) GradeID() field.String {
	return p.StudentProfile.GradeID
}
func (p DomainStudentImpl) ExternalUserID() field.String {
	return p.StudentProfile.ExternalUserID
}
func (p DomainStudentImpl) UserName() field.String {
	username := strings.ToLower(p.StudentProfile.UserName.TrimSpace().String())
	return field.NewString(username)
}
func (p DomainStudentImpl) FirstName() field.String {
	return p.StudentProfile.FirstName
}
func (p DomainStudentImpl) LastName() field.String {
	return p.StudentProfile.LastName
}
func (p DomainStudentImpl) FullName() field.String {
	return p.StudentProfile.FullName
}
func (p DomainStudentImpl) FirstNamePhonetic() field.String {
	return p.StudentProfile.FirstNamePhonetic
}
func (p DomainStudentImpl) LastNamePhonetic() field.String {
	return p.StudentProfile.LastNamePhonetic
}
func (p DomainStudentImpl) FullNamePhonetic() field.String {
	return p.StudentProfile.FullNamePhonetic
}
func (p DomainStudentImpl) Email() field.String {
	return p.StudentProfile.Email
}
func (p DomainStudentImpl) Password() field.String {
	return p.StudentProfile.Password
}
func (p DomainStudentImpl) EnrollmentStatus() field.String {
	enrollmentStatus := p.StudentProfile.EnrollmentStatus
	// backward compatible: use first enrollment status history
	if field.IsPresent(enrollmentStatus) {
		return field.NewString(StudentEnrollmentStatusMap[int(enrollmentStatus.Int32())])
	}
	if len(p.StudentProfile.EnrollmentStatusHistories) != 0 {
		enrollmentStatusBackward := p.StudentProfile.EnrollmentStatusHistories[0].EnrollmentStatus
		return field.NewString(StudentEnrollmentStatusMap[int(enrollmentStatusBackward.Int16())])
	}

	return field.NewNullString()
}
func (p DomainStudentImpl) Birthday() field.Date {
	return p.StudentProfile.Birthday
}
func (p DomainStudentImpl) Gender() field.String {
	gender := p.StudentProfile.Gender
	if field.IsPresent(gender) {
		return field.NewString(constant.UserGenderMap[int(gender.Int32())])
	}

	return field.NewNullString()
}
func (p DomainStudentImpl) StudentNote() field.String {
	return field.NewString(p.StudentProfile.Note.String())
}
func (p DomainStudentImpl) ContactPreference() field.String {
	if p.UserPhoneNumber != nil {
		contactPreference := p.UserPhoneNumber.ContactPreference
		if field.IsPresent(contactPreference) {
			return field.NewString(StudentContactPreference[int(contactPreference.Int32())])
		}
	}

	return field.NewNullString()
}
func (p DomainStudentImpl) LoginEmail() field.String {
	return p.StudentProfile.LoginEmail
}

func (p DomainStudentImpl) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}

type DomainSchoolHistoryImpl struct {
	entity.DefaultDomainSchoolHistory

	SchoolHistoryPayload
}

func (p DomainSchoolHistoryImpl) StartDate() field.Time {
	if field.IsPresent(p.SchoolHistoryPayload.StartDate) {
		return field.NewTime(p.SchoolHistoryPayload.StartDate.Date())
	}
	return field.NewNullTime()
}
func (p DomainSchoolHistoryImpl) EndDate() field.Time {
	if field.IsPresent(p.SchoolHistoryPayload.EndDate) {
		return field.NewTime(p.SchoolHistoryPayload.EndDate.Date())
	}
	return field.NewNullTime()
}

type AddressPayload struct {
	PostalCode   field.String `json:"postal_code"`
	Prefecture   field.String `json:"prefecture"`
	City         field.String `json:"city"`
	FirstStreet  field.String `json:"first_street"`
	SecondStreet field.String `json:"second_street"`
}

type DomainUserAddressImpl struct {
	entity.DefaultDomainUserAddress
	prefectureID field.String

	AddressPayload
}

func (p DomainUserAddressImpl) PostalCode() field.String {
	return p.AddressPayload.PostalCode
}
func (p DomainUserAddressImpl) City() field.String {
	return p.AddressPayload.City
}
func (p DomainUserAddressImpl) FirstStreet() field.String {
	return p.AddressPayload.FirstStreet
}
func (p DomainUserAddressImpl) SecondStreet() field.String {
	return p.AddressPayload.SecondStreet
}
func (p DomainUserAddressImpl) PrefectureID() field.String {
	return p.prefectureID
}

type DomainPrefectureImpl struct {
	entity.DefaultDomainPrefecture

	prefectureCode field.String
}

func (e DomainPrefectureImpl) PrefectureCode() field.String {
	return e.prefectureCode
}

type PhoneNumberPayload struct {
	PhoneNumber       field.String `json:"student_phone_number"`
	HomePhoneNumber   field.String `json:"student_home_phone_number"`
	ContactPreference field.Int32  `json:"contact_preference"`
}

type DomainUserPhoneNumberImpl struct {
	entity.DefaultDomainUserPhoneNumber

	userPhoneNumberID field.String
	phoneNumberType   field.String
	phoneNumber       field.String
}

func (p DomainUserPhoneNumberImpl) UserPhoneNumberID() field.String {
	return p.userPhoneNumberID
}
func (p DomainUserPhoneNumberImpl) PhoneNumber() field.String {
	return p.phoneNumber
}
func (p DomainUserPhoneNumberImpl) Type() field.String {
	return p.phoneNumberType
}

type EnrollmentStatusHistoryPayload struct {
	EnrollmentStatus field.Int16  `json:"enrollment_status"`
	Location         field.String `json:"location"`
	StartDate        field.Date   `json:"start_date"`
	EndDate          field.Date   `json:"end_date"`
}

type DomainEnrollmentStatusHistoryImpl struct {
	entity.DefaultDomainEnrollmentStatusHistory

	EnrollmentStatusHistoryPayload
}

func (e DomainEnrollmentStatusHistoryImpl) EnrollmentStatus() field.String {
	enrollmentStatus := e.EnrollmentStatusHistoryPayload.EnrollmentStatus
	if field.IsPresent(enrollmentStatus) {
		return field.NewString(StudentEnrollmentStatusMap[int(enrollmentStatus.Int16())])
	}
	return field.NewNullString()
}

func (e DomainEnrollmentStatusHistoryImpl) StartDate() field.Time {
	return field.NewTime(e.EnrollmentStatusHistoryPayload.StartDate.Date())
}

func (e DomainEnrollmentStatusHistoryImpl) EndDate() field.Time {
	if e.EnrollmentStatusHistoryPayload.EndDate.Date().IsZero() {
		return field.NewNullTime()
	}
	return field.NewTime(e.EnrollmentStatusHistoryPayload.EndDate.Date())
}
