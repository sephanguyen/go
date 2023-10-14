package usermgmt

import (
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Organization struct {
	organizationID string
	schoolID       int32
}

func (org *Organization) OrganizationID() field.String {
	return field.NewString(org.organizationID)
}

func (org *Organization) SchoolID() field.Int32 {
	return field.NewInt32(org.schoolID)
}

type User struct {
	userID string
}

func (user *User) Avatar() field.String {
	return field.NewString(fmt.Sprintf("https://%s.com", strings.ToLower(newID())))
}

func (user *User) Group() field.String {
	return field.NewString(constant.UserGroupSchoolAdmin)
}

func (user *User) FullName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) FirstName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) LastName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) GivenName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) FullNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) FirstNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) LastNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (user *User) Country() field.String {
	return field.NewString("abc")
}

func (user *User) PhoneNumber() field.String {
	return field.NewString(user.userID)
}

func (user *User) Email() field.String {
	return field.NewString(fmt.Sprintf("%s@manabie.com", user.userID))
}

func (user *User) Password() field.String {
	return field.NewString(fmt.Sprintf("pwd-%s", user.userID))
}

func (user *User) DeviceToken() field.String {
	return field.NewNullString()
}

func (user *User) AllowNotification() field.Boolean {
	return field.NewBoolean(false)
}

func (user *User) LastLoginDate() field.Time {
	return field.NewTime(time.Now())
}

func (user *User) Birthday() field.Date {
	return field.NewDate(time.Now())
}

func (user *User) Gender() field.String {
	return field.NewString("MALE")
}

func (user *User) IsTester() field.Boolean {
	return field.NewBoolean(false)
}

func (user *User) FacebookID() field.String {
	return field.NewNullString()
}

func (user *User) PhoneVerified() field.Boolean {
	return field.NewBoolean(false)
}

func (user *User) EmailVerified() field.Boolean {
	return field.NewBoolean(false)
}

func (user *User) UserID() field.String {
	return field.NewString(user.userID)
}

func (user *User) OrganizationID() field.String {
	return field.NewUndefinedString()
}

func (user *User) SchoolID() field.Int32 {
	return field.NewUndefinedInt32()
}

type SchoolAdmin struct {
	randomID string
}

func (schoolAdmin *SchoolAdmin) Avatar() field.String {
	return field.NewString(fmt.Sprintf("https://%s.com", strings.ToLower(newID())))
}

func (schoolAdmin *SchoolAdmin) Group() field.String {
	return field.NewString(constant.UserGroupSchoolAdmin)
}

func (schoolAdmin *SchoolAdmin) FullName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) FirstName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) LastName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) GivenName() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) FullNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) FirstNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) LastNamePhonetic() field.String {
	return field.NewString(strings.ToLower(newID()))
}

func (schoolAdmin *SchoolAdmin) Country() field.String {
	return field.NewString("abc")
}

func (schoolAdmin *SchoolAdmin) PhoneNumber() field.String {
	return field.NewString(schoolAdmin.randomID)
}

func (schoolAdmin *SchoolAdmin) Email() field.String {
	return field.NewString(fmt.Sprintf("%s@example.com", schoolAdmin.randomID))
}

func (schoolAdmin *SchoolAdmin) Password() field.String {
	return field.NewString(fmt.Sprintf("pwd-%s", schoolAdmin.randomID))
}

func (schoolAdmin *SchoolAdmin) DeviceToken() field.String {
	return field.NewNullString()
}

func (schoolAdmin *SchoolAdmin) AllowNotification() field.Boolean {
	return field.NewBoolean(false)
}

func (schoolAdmin *SchoolAdmin) LastLoginDate() field.Time {
	return field.NewTime(time.Now())
}

func (schoolAdmin *SchoolAdmin) Birthday() field.Date {
	return field.NewDate(time.Now())
}

func (schoolAdmin *SchoolAdmin) Gender() field.String {
	return field.NewString("MALE")
}

func (schoolAdmin *SchoolAdmin) IsTester() field.Boolean {
	return field.NewBoolean(false)
}

func (schoolAdmin *SchoolAdmin) FacebookID() field.String {
	return field.NewNullString()
}

func (schoolAdmin *SchoolAdmin) PhoneVerified() field.Boolean {
	return field.NewBoolean(false)
}

func (schoolAdmin *SchoolAdmin) EmailVerified() field.Boolean {
	return field.NewBoolean(false)
}

func (schoolAdmin *SchoolAdmin) UserID() field.String {
	return field.NewString(schoolAdmin.randomID)
}

func (schoolAdmin *SchoolAdmin) OrganizationID() field.String {
	return field.NewUndefinedString()
}

func (schoolAdmin *SchoolAdmin) SchoolID() field.Int32 {
	return field.NewUndefinedInt32()
}

type UserAccessPath struct {
	userID     field.String
	locationID field.String
}

func (u *UserAccessPath) LocationID() field.String {
	return u.locationID
}

func (u *UserAccessPath) UserID() field.String {
	return u.userID
}
