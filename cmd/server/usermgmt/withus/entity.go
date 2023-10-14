package withus

import (
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

const (
	defaultContactPreference      = "STUDENT_PHONE_NUMBER"
	emailDomainManagaraBase       = "@managara-base.jp"
	emailDomainManagaraHighSchool = "@managara.nsf-h.ed.jp"
)

var (
	startTime, _ = time.Parse(constant.DateLayout, "2020/04/01")
	endTime, _   = time.Parse(constant.DateLayout, "2099/03/31")
	G2Prefix     = "2_"
	G3Prefix     = "3_"
	G4Prefix     = "4_"
)

type ManagaraStudent struct {
	entity.NullDomainStudent
	Parent Parent

	UserIDAttr             string
	CustomerNumber         field.String
	StudentNumber          field.String
	Name                   field.String
	PasswordRaw            field.String
	StudentEmail           field.String
	Locations              field.String
	TagG2                  field.String
	TagG3                  field.String
	TagG4                  field.String
	TagG5                  field.String
	Courses                field.String
	DeleteFlag             field.String
	GraduationExpectedDate field.String

	// temporary fix for email domain, need to fix the way we embed interface
	emailDomain string
}

type ManagaraStudents []ManagaraStudent

func (students ManagaraStudents) externalUserIDs() []string {
	studentIDs := []string{}
	for _, v := range students {
		studentIDs = append(studentIDs, v.ExternalUserID().String())
	}
	return studentIDs
}

type ManagaraBaseStudent struct {
	Parent         Parent
	UserIDAttr     string
	CustomerNumber field.String `csv:"顧客番号"`
	StudentNumber  field.String `csv:"生徒番号"`
	Name           field.String `csv:"氏名（ニックネーム）"`
	PasswordRaw    field.String `csv:"パスワード"`
	StudentEmail   field.String `csv:"生徒メール"`
	Locations      field.String `csv:"G1（所属）"`
	TagG2          field.String `csv:"G2（セグメント）"`
	TagG3          field.String `csv:"G3（生徒区分）"`
	TagG4          field.String `csv:"G4（本校）"`
	TagG5          field.String `csv:"G5（学年）"`
	Courses        field.String `csv:"所持商品"`
	DeleteFlag     field.String `csv:"削除フラグ"`
}

type ManagaraHSStudent struct {
	Parent                 Parent
	UserIDAttr             string
	CustomerNumber         field.String `csv:"顧客番号"`
	StudentNumber          field.String `csv:"生徒番号"`
	Name                   field.String `csv:"氏名"`
	PasswordRaw            field.String `csv:"パスワード"`
	StudentEmail           field.String `csv:"生徒メール"`
	Locations              field.String `csv:"G1（所属）"`
	TagG2                  field.String `csv:"G2（コース）"`
	TagG3                  field.String `csv:"G3（高校生区分）"`
	TagG4                  field.String `csv:"G4（本校）"`
	TagG5                  field.String `csv:"G5（学年）"`
	Courses                field.String `csv:"所持商品"`
	DeleteFlag             field.String `csv:"削除フラグ"`
	GraduationExpectedDate field.String `csv:"卒業予定日"`
}

func (student ManagaraBaseStudent) toManagaraStudent() ManagaraStudent {
	student.Parent.emailDomain = emailDomainManagaraBase

	return ManagaraStudent{
		Parent:         student.Parent,
		UserIDAttr:     student.UserIDAttr,
		CustomerNumber: student.CustomerNumber,
		StudentNumber:  student.StudentNumber,
		Name:           student.Name,
		PasswordRaw:    student.PasswordRaw,
		StudentEmail:   student.StudentEmail,
		Locations:      student.Locations,
		TagG2:          student.TagG2,
		TagG3:          student.TagG3,
		TagG4:          student.TagG4,
		TagG5:          student.TagG5,
		Courses:        student.Courses,
		DeleteFlag:     student.DeleteFlag,
		emailDomain:    emailDomainManagaraBase,
	}
}

func (student ManagaraHSStudent) toManagaraStudent() ManagaraStudent {
	student.Parent.emailDomain = emailDomainManagaraHighSchool

	return ManagaraStudent{
		Parent:                 student.Parent,
		UserIDAttr:             student.UserIDAttr,
		CustomerNumber:         student.CustomerNumber,
		StudentNumber:          student.StudentNumber,
		Name:                   student.Name,
		PasswordRaw:            student.PasswordRaw,
		StudentEmail:           student.StudentEmail,
		Locations:              student.Locations,
		TagG2:                  student.TagG2,
		TagG3:                  student.TagG3,
		TagG4:                  student.TagG4,
		TagG5:                  student.TagG5,
		Courses:                student.Courses,
		DeleteFlag:             student.DeleteFlag,
		GraduationExpectedDate: student.GraduationExpectedDate,
		emailDomain:            emailDomainManagaraHighSchool,
	}
}

func (student ManagaraStudent) partnerTagIDs() []string {
	partnerTagIDs := make([]string, 0)

	if field.IsPresent(student.TagG2) {
		partnerTagIDs = append(partnerTagIDs, G2Prefix+student.TagG2.String())
	}
	if field.IsPresent(student.TagG3) {
		partnerTagIDs = append(partnerTagIDs, G3Prefix+student.TagG3.String())
	}
	if field.IsPresent(student.TagG4) {
		partnerTagIDs = append(partnerTagIDs, G4Prefix+student.TagG4.String())
	}
	if field.IsPresent(student.GraduationExpectedDate) {
		partnerTagIDs = append(partnerTagIDs, student.GraduationExpectedDate.String())
	}

	return partnerTagIDs
}

func (student ManagaraStudent) partnerLocationIDs() []string {
	locationIDs := make([]string, 0)
	if field.IsPresent(student.Locations) {
		locationIDs = strings.Split(student.Locations.String(), ",")
	}
	return locationIDs
}

func (student ManagaraStudent) partnerGradeID() field.String {
	return student.TagG5
}

func (student ManagaraStudent) partnerCourseIDs() []string {
	courseIDs := make([]string, 0)
	if field.IsPresent(student.Courses) {
		courseIDs = strings.Split(student.Courses.String(), ",")
	}
	return courseIDs
}

func (student ManagaraStudent) FullName() field.String {
	return student.Name
}

func (student ManagaraStudent) FirstName() field.String {
	fullName := strings.Split(student.Name.String(), " ")
	if len(fullName) > 1 {
		return field.NewString(strings.Join(fullName[1:], " "))
	}
	return field.NewNullString()
}

func (student ManagaraStudent) LastName() field.String {
	fullName := strings.Split(student.Name.String(), " ")
	if len(fullName) > 1 {
		return field.NewString(fullName[0])
	}
	return field.NewNullString()
}

func (student ManagaraStudent) GivenName() field.String {
	return student.Name
}

// UserName currently withus will use email as username
func (student ManagaraStudent) UserName() field.String {
	return field.NewString(student.CustomerNumber.String() + student.emailDomain).ToLower()
}

func (student ManagaraStudent) Email() field.String {
	return field.NewString(student.CustomerNumber.String() + student.emailDomain)
}

func (student ManagaraStudent) LoginEmail() field.String {
	return field.NewString(student.CustomerNumber.String() + student.emailDomain)
}

func (student ManagaraStudent) Password() field.String {
	return student.PasswordRaw
}

var enrollmentStatusMapping = map[string]string{
	"":  constant.StudentEnrollmentStatusEnrolled,
	"1": constant.StudentEnrollmentStatusWithdrawn,
	"2": constant.StudentEnrollmentStatusLOA,
}

func (student ManagaraStudent) EnrollmentStatus() field.String {
	return field.NewString(enrollmentStatusMapping[student.DeleteFlag.String()])
}

func (student ManagaraStudent) StudentNote() field.String {
	return field.NewString(student.StudentEmail.String())
}

func (student ManagaraStudent) Remarks() field.String {
	return student.StudentEmail
}

func (student ManagaraStudent) ContactPreference() field.String {
	return field.NewString(defaultContactPreference)
}

func (student ManagaraStudent) UserID() field.String {
	if student.UserIDAttr != "" {
		return field.NewString(student.UserIDAttr)
	}

	return field.NewNullString()
}

func (student ManagaraStudent) ExternalUserID() field.String {
	return student.StudentNumber
}

func (student ManagaraStudent) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}

type Parent struct {
	entity.NullDomainParent
	UserIDAttr        string
	ParentNumber      field.String `csv:"保護者番号"`
	ParentName        field.String `csv:"保護者氏名"`
	ParentRawPassword field.String `csv:"保護者パスワード"`
	ParentEmail       field.String `csv:"保護者メール"`

	// temporary fix for email domain, need to fix the way we embed interface
	emailDomain string
}

func (p Parent) UserID() field.String {
	if p.UserIDAttr != "" {
		return field.NewString(p.UserIDAttr)
	}

	return field.NewNullString()
}

// UserName currently withus will use email as username
func (p Parent) UserName() field.String {
	return field.NewString(p.ParentNumber.String() + p.emailDomain).ToLower()
}

func (p Parent) Email() field.String {
	return field.NewString(p.ParentNumber.String() + p.emailDomain)
}

func (p Parent) LoginEmail() field.String {
	return field.NewString(p.ParentNumber.String() + p.emailDomain)
}

func (p Parent) Remarks() field.String {
	switch {
	case field.IsPresent(p.ParentNumber) && field.IsPresent(p.ParentEmail):
		return field.NewString(fmt.Sprintf("%s,%s", p.ParentNumber.String(), p.ParentEmail.String()))
	case field.IsPresent(p.ParentNumber):
		return p.ParentNumber
	case field.IsPresent(p.ParentEmail):
		return p.ParentEmail
	default:
		return field.NewNullString()
	}
}

func (p Parent) FullName() field.String {
	return p.ParentName
}

func (p Parent) FirstName() field.String {
	fullName := strings.Split(p.ParentName.String(), " ")
	if len(fullName) > 1 {
		return field.NewString(strings.Join(fullName[1:], " "))
	}
	return field.NewNullString()
}

func (p Parent) LastName() field.String {
	fullName := strings.Split(p.ParentName.String(), " ")
	if len(fullName) > 1 {
		return field.NewString(fullName[0])
	}
	return field.NewNullString()
}

func (p Parent) GivenName() field.String {
	return p.ParentName
}

func (p Parent) Password() field.String {
	return p.ParentRawPassword
}

func (p Parent) ExternalUserID() field.String {
	return p.ParentNumber
}

func (p Parent) UserRole() field.String {
	return field.NewString(string(constant.UserRoleParent))
}

type withUsCourse struct {
	studentPackageID field.String
	courseID         field.String
	startAt          field.Time
	endAt            field.Time
}

func (s withUsCourse) StudentPackageID() field.String {
	return s.studentPackageID
}

func (s withUsCourse) CourseID() field.String {
	return s.courseID
}

func (s withUsCourse) StartAt() field.Time {
	return s.startAt
}

func (s withUsCourse) EndAt() field.Time {
	return s.endAt
}

type withUsEnrollmentStatusImpl struct {
	entity.DefaultDomainEnrollmentStatusHistory

	LocationAttr         field.String
	EnrollmentStatusAttr field.String
	StartDateAttr        field.Time
}

func (e *withUsEnrollmentStatusImpl) EnrollmentStatus() field.String {
	return field.NewString(e.EnrollmentStatusAttr.String())
}

func (e *withUsEnrollmentStatusImpl) LocationID() field.String {
	return e.LocationAttr
}

func (e *withUsEnrollmentStatusImpl) StartDate() field.Time {
	return e.StartDateAttr
}
