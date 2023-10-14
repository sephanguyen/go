package grpc

import (
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
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
		0: entity.UserPhoneNumberTypeStudentPhoneNumber,
		1: entity.UserPhoneNumberTypeStudentHomePhoneNumber,
		2: entity.UserPhoneNumberTypeParentPrimaryPhoneNumber,
		3: entity.UserPhoneNumberTypeParentSecondaryPhoneNumber,
	}
	MapStudentPhoneNumberType = map[string]string{
		"PHONE_NUMBER":      constant.StudentPhoneNumber,
		"HOME_PHONE_NUMBER": constant.StudentHomePhoneNumber,
	}
)

type UpdateStudentRequest struct {
	Body *pb.UpdateStudentRequest

	entity.EmptyUser
}

func NewUpdateStudentRequest(body *pb.UpdateStudentRequest) entity.User {
	return &UpdateStudentRequest{Body: body}
}

func (req *UpdateStudentRequest) UserID() field.String {
	return field.NewString(req.Body.StudentProfile.GetId())
}

type DomainSchoolHistoryImpl struct {
	entity.DefaultDomainSchoolHistory

	SchoolIDAttr       field.String
	SchoolCourseIDAttr field.String
	StartDateAttr      field.Time
	EndDateAttr        field.Time
}

func (s *DomainSchoolHistoryImpl) StartDate() field.Time {
	if s.StartDateAttr.Time().Unix() == 0 {
		return s.DefaultDomainSchoolHistory.StartDate()
	}
	return field.NewTime(utils.TruncateToDay(s.StartDateAttr.Time()))
}

func (s *DomainSchoolHistoryImpl) EndDate() field.Time {
	if s.EndDateAttr.Time().Unix() == 0 {
		return s.DefaultDomainSchoolHistory.EndDate()
	}
	return field.NewTime(utils.TruncateToDay(s.EndDateAttr.Time()))
}

func (s *DomainSchoolHistoryImpl) SchoolID() field.String {
	if s.SchoolIDAttr.String() == "" {
		return s.DefaultDomainSchoolHistory.SchoolID()
	}
	return s.SchoolIDAttr
}

func (s *DomainSchoolHistoryImpl) SchoolCourseID() field.String {
	if s.SchoolCourseIDAttr.String() == "" {
		return s.DefaultDomainSchoolHistory.SchoolCourseID()
	}
	return s.SchoolCourseIDAttr
}

type DomainEnrollmentStatusHistoryImpl struct {
	entity.DefaultDomainEnrollmentStatusHistory

	StudentIDAttr        field.String
	LocationAttr         field.String
	EnrollmentStatusAttr field.Int32
	StartDateAttr        field.Time
	EndDateAttr          field.Time
}

func (e *DomainEnrollmentStatusHistoryImpl) EnrollmentStatus() field.String {
	enrollmentStatus := e.EnrollmentStatusAttr
	if field.IsPresent(enrollmentStatus) {
		return field.NewString(StudentEnrollmentStatusMap[int(enrollmentStatus.Int32())])
	}
	return field.NewNullString()
}

func (e *DomainEnrollmentStatusHistoryImpl) LocationID() field.String {
	return e.LocationAttr
}

func (e *DomainEnrollmentStatusHistoryImpl) UserID() field.String {
	return e.StudentIDAttr
}

func (e *DomainEnrollmentStatusHistoryImpl) StartDate() field.Time {
	return e.StartDateAttr
}

func (e *DomainEnrollmentStatusHistoryImpl) EndDate() field.Time {
	if e.EndDateAttr.Time().Unix() == 0 || e.EndDateAttr.Time().Year() == 1 {
		return e.DefaultDomainEnrollmentStatusHistory.EndDate()
	}
	return e.EndDateAttr
}

type PhoneNumberAttr struct {
	entity.DefaultDomainUserPhoneNumber

	PhoneIDAttr     field.String
	PhoneTypeAttr   field.String
	PhoneNumberAttr field.String
}

func (p *PhoneNumberAttr) UserPhoneNumberID() field.String {
	if p.PhoneIDAttr.String() == "" {
		return field.NewString(idutil.ULIDNow())
	}
	return p.PhoneIDAttr
}

func (p *PhoneNumberAttr) PhoneNumber() field.String {
	if p.PhoneNumberAttr.String() == "" {
		return p.DefaultDomainUserPhoneNumber.PhoneNumber()
	}
	return p.PhoneNumberAttr
}

func (p *PhoneNumberAttr) Type() field.String {
	return p.PhoneTypeAttr
}

type DomainPhoneNumberImpl struct {
	PhoneNumberAttr   []PhoneNumberAttr
	ContactPreference field.Int32
}

type DomainUserAddressImpl struct {
	entity.DefaultDomainUserAddress

	AddressIDAttr    field.String
	AddressTypeAttr  field.String
	PostalCodeAttr   field.String
	PrefectureAttr   field.String
	CityAttr         field.String
	FirstStreetAttr  field.String
	SecondStreetAttr field.String
}

func (d *DomainUserAddressImpl) UserAddressID() field.String {
	if d.AddressIDAttr.String() == "" {
		return field.NewString(idutil.ULIDNow())
	}
	return d.AddressIDAttr
}

func (d *DomainUserAddressImpl) AddressType() field.String {
	return d.AddressTypeAttr
}

func (d *DomainUserAddressImpl) PostalCode() field.String {
	return d.PostalCodeAttr
}

func (d *DomainUserAddressImpl) City() field.String {
	return d.CityAttr
}

func (d *DomainUserAddressImpl) PrefectureID() field.String {
	if d.PrefectureAttr.String() == "" {
		return d.DefaultDomainUserAddress.PrefectureID()
	}
	return d.PrefectureAttr
}

func (d *DomainUserAddressImpl) FirstStreet() field.String {
	return d.FirstStreetAttr
}

func (d *DomainUserAddressImpl) SecondStreet() field.String {
	return d.SecondStreetAttr
}

type DomainStudentImpl struct {
	entity.NullDomainStudent

	UserIDAttr            field.String
	GradeIDAttr           field.String
	ExternalUserIDAttr    field.String
	UserNameAttr          field.String
	FirstNameAttr         field.String
	LastNameAttr          field.String
	FullNameAttr          field.String
	FirstNamePhoneticAttr field.String
	LastNamePhoneticAttr  field.String
	FullNamePhoneticAttr  field.String
	EmailAttr             field.String
	EnrollmentStatusAttr  field.String
	GradeAttr             field.String
	BirthdayAttr          field.Date
	GenderAttr            field.Int32
	NoteAttr              field.String
	TagsAttr              []field.String
	PasswordAttr          field.String
	ExternalStudentIDAttr field.String
	LoginEmailAttr        field.String

	SchoolHistoriesAttr           []DomainSchoolHistoryImpl
	EnrollmentStatusHistoriesAttr []DomainEnrollmentStatusHistoryImpl
	UserPhoneNumberAttr           *DomainPhoneNumberImpl
	ContactPreferenceAttr         field.Int32
	AddressAttr                   *DomainUserAddressImpl
}

type DomainTaggedUserImpl struct {
	entity.EmptyDomainTaggedUser

	TagIDAttr  field.String
	UserIDAttr field.String
}

func (e *DomainTaggedUserImpl) TagID() field.String {
	return e.TagIDAttr
}

func (e *DomainTaggedUserImpl) UserID() field.String {
	return e.UserIDAttr
}

type DomainUserAccessPathImpl struct {
	entity.DefaultUserAccessPath

	UserIDAttr     field.String
	LocationIDAttr field.String
}

func (d *DomainUserAccessPathImpl) UserID() field.String {
	return d.UserIDAttr
}

func (d *DomainUserAccessPathImpl) LocationID() field.String {
	return d.LocationIDAttr
}

func (p DomainStudentImpl) UserID() field.String {
	return p.UserIDAttr
}

func (p DomainStudentImpl) GradeID() field.String {
	return p.GradeIDAttr
}

func (p DomainStudentImpl) ExternalUserID() field.String {
	if p.ExternalUserIDAttr.String() == "" {
		return p.NullDomainStudent.ExternalUserID()
	}
	return p.ExternalUserIDAttr
}

func (p DomainStudentImpl) ExternalStudentID() field.String {
	if p.ExternalStudentIDAttr.String() == "" {
		return p.NullDomainStudent.ExternalStudentID()
	}
	return p.ExternalStudentIDAttr
}

func (p DomainStudentImpl) FirstName() field.String {
	return p.FirstNameAttr
}

func (p DomainStudentImpl) UserName() field.String {
	username := strings.ToLower(p.UserNameAttr.TrimSpace().String())
	return field.NewString(username)
}

func (p DomainStudentImpl) LastName() field.String {
	return p.LastNameAttr
}

func (p DomainStudentImpl) FullName() field.String {
	return p.FullNameAttr
}

func (p DomainStudentImpl) FirstNamePhonetic() field.String {
	return p.FirstNamePhoneticAttr
}

func (p DomainStudentImpl) LastNamePhonetic() field.String {
	return p.LastNamePhoneticAttr
}

func (p DomainStudentImpl) FullNamePhonetic() field.String {
	return p.FullNamePhoneticAttr
}

func (p DomainStudentImpl) Email() field.String {
	return p.EmailAttr
}

func (p DomainStudentImpl) EnrollmentStatus() field.String {
	if field.IsPresent(p.EnrollmentStatusAttr) {
		return p.EnrollmentStatusAttr
	}
	return field.NewNullString()
}

func (p DomainStudentImpl) Birthday() field.Date {
	if p.BirthdayAttr.Date().Unix() == 0 {
		return p.NullDomainStudent.Birthday()
	}
	return field.NewDate(utils.TruncateToDay(p.BirthdayAttr.Date()))
}

func (p DomainStudentImpl) Gender() field.String {
	gender := p.GenderAttr
	if p.GenderAttr.Int32() == 0 {
		return p.NullDomainStudent.Gender()
	}
	return field.NewString(constant.UserGenderMap[int(gender.Int32())])
}

func (p DomainStudentImpl) StudentNote() field.String {
	return field.NewString(p.NoteAttr.String())
}

func (p DomainStudentImpl) ContactPreference() field.String {
	return field.NewString(StudentContactPreference[int(p.ContactPreferenceAttr.Int32())])
}

func (p DomainStudentImpl) Password() field.String {
	return field.NewString(p.PasswordAttr.String())
}

func (p DomainStudentImpl) LoginEmail() field.String {
	return p.LoginEmailAttr
}
func (p DomainStudentImpl) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}
