package entity

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"golang.org/x/crypto/argon2"
)

const (
	UserGenderMale   = "MALE"
	UserGenderFemale = "FEMALE"
)

type UserField string

const (
	UserFieldUserID                    UserField = "user_id"
	UserFieldAvatar                    UserField = "avatar"
	UserFieldGroup                     UserField = "group"
	UserFieldUserName                  UserField = "username"
	UserFieldFullName                  UserField = "full_name"
	UserFieldFirstName                 UserField = "first_name"
	UserFieldLastName                  UserField = "last_name"
	UserFieldGivenName                 UserField = "given_name"
	UserFieldFullNamePhonetic          UserField = "full_name_phonetic"
	UserFieldFirstNamePhonetic         UserField = "first_name_phonetic"
	UserFieldLastNamePhonetic          UserField = "last_name_phonetic"
	UserFieldCountry                   UserField = "country"
	UserFieldPhoneNumber               UserField = "phone_number"
	UserFieldEmail                     UserField = "email"
	UserFieldPassword                  UserField = "password"
	UserFieldDeviceToken               UserField = "device_token"
	UserFieldAllowNotification         UserField = "allow_notification"
	UserFieldLastLoginDate             UserField = "last_login_date"
	UserFieldBirthday                  UserField = "birthday"
	UserFieldGender                    UserField = "gender"
	UserFieldIsTester                  UserField = "is_tester"
	UserFieldFacebookID                UserField = "facebook_id"
	UserFieldPhoneVerified             UserField = "phone_verified"
	UserFieldEmailVerified             UserField = "email_verified"
	UserFieldExternalUserID            UserField = "external_user_id"
	UserFieldRemarks                   UserField = "remark"
	UserFieldEncryptedUserIDByPassword UserField = "encrypted_user_id_by_password"
	UserFieldDeactivatedAt             UserField = "deactivated_at"
	UserFieldOrganizationID            UserField = "user_organization_id"
)

// UserProfile is deprecated and this will be merged in to below User interface soon
type UserProfile interface {
	Avatar() field.String
	Group() field.String
	UserName() field.String
	FullName() field.String
	FirstName() field.String
	LastName() field.String
	GivenName() field.String
	FullNamePhonetic() field.String
	FirstNamePhonetic() field.String
	LastNamePhonetic() field.String
	PhoneNumber() field.String
	Email() field.String
	Password() field.String
	DeviceToken() field.String
	AllowNotification() field.Boolean
	LastLoginDate() field.Time
	Birthday() field.Date
	Gender() field.String
	IsTester() field.Boolean
	FacebookID() field.String
	PhoneVerified() field.Boolean
	EmailVerified() field.Boolean
	ExternalUserID() field.String
	Remarks() field.String
	EncryptedUserIDByPassword() field.String
	DeactivatedAt() field.Time
	UserRole() field.String
}

// User represents a user in our business
//
// //go:generate hexagen ent-impl --type=User ../modules/user/core/valueobj .
type User interface {
	UserProfile
	valueobj.HasUserID
	valueobj.HasCountry
	valueobj.HasOrganizationID
	valueobj.HasLoginEmail
}

type Users []User

func (users Users) UserIDs() []string {
	userIDs := make([]string, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.UserID().String())
	}
	return userIDs
}

func (users Users) ExternalUserIDs() []string {
	userIDs := []string{}
	for _, user := range users {
		userIDs = append(userIDs, user.ExternalUserID().String())
	}
	return userIDs
}

func (users Users) Emails() []string {
	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, user.Email().String())
	}
	return emails
}

func (users Users) LowerCasedUserNames() []string {
	usernames := make([]string, 0, len(users))
	for _, user := range users {
		usernames = append(usernames, strings.ToLower(user.UserName().String()))
	}
	return usernames
}

func (users Users) LowerCaseEmails() []string {
	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, strings.ToLower(user.Email().String()))
	}
	return emails
}

// EncryptedUserIDByPasswordFromUser encrypt user info to return password-change-tag
func EncryptedUserIDByPasswordFromUser(user User) (field.String, error) {
	if user.Password().String() == "" {
		return field.NewNullString(), nil
	}

	// Use user's password as passphrase, uid as salt
	key := argon2.IDKey([]byte(user.Password().String()), []byte(user.UserID().String()), 2, 19*1024, 1, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return field.NewUndefinedString(), err
	}

	dataToEncrypt := []byte(user.UserID().String())

	encryptedUserIDByPassword := make([]byte, aes.BlockSize+len(dataToEncrypt))

	stream := cipher.NewCTR(block, encryptedUserIDByPassword[:aes.BlockSize])
	stream.XORKeyStream(encryptedUserIDByPassword[aes.BlockSize:], dataToEncrypt)

	encryptedUserIDByPassword = encryptedUserIDByPassword[aes.BlockSize:]

	return field.NewString(base64.RawStdEncoding.EncodeToString(encryptedUserIDByPassword)), nil
}

type EmptyUser struct{}

func (user EmptyUser) UserID() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Avatar() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Group() field.String {
	return field.NewNullString()
}
func (user EmptyUser) UserName() field.String {
	return field.NewNullString()
}
func (user EmptyUser) FullName() field.String {
	return field.NewNullString()
}
func (user EmptyUser) FirstName() field.String {
	return field.NewNullString()
}
func (user EmptyUser) LastName() field.String {
	return field.NewNullString()
}
func (user EmptyUser) GivenName() field.String {
	return field.NewNullString()
}
func (user EmptyUser) FullNamePhonetic() field.String {
	return field.NewNullString()
}
func (user EmptyUser) FirstNamePhonetic() field.String {
	return field.NewNullString()
}
func (user EmptyUser) LastNamePhonetic() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Country() field.String {
	return field.NewNullString()
}
func (user EmptyUser) PhoneNumber() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Email() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Password() field.String {
	return field.NewNullString()
}
func (user EmptyUser) DeviceToken() field.String {
	return field.NewNullString()
}
func (user EmptyUser) AllowNotification() field.Boolean {
	return field.NewBoolean(false)
}
func (user EmptyUser) LastLoginDate() field.Time {
	return field.NewNullTime()
}
func (user EmptyUser) Birthday() field.Date {
	return field.NewNullDate()
}
func (user EmptyUser) Gender() field.String {
	return field.NewNullString()
}
func (user EmptyUser) IsTester() field.Boolean {
	return field.NewBoolean(false)
}
func (user EmptyUser) FacebookID() field.String {
	return field.NewNullString()
}
func (user EmptyUser) ExternalUserID() field.String {
	return field.NewNullString()
}
func (user EmptyUser) PhoneVerified() field.Boolean {
	return field.NewBoolean(false)
}
func (user EmptyUser) EmailVerified() field.Boolean {
	return field.NewBoolean(false)
}
func (user EmptyUser) OrganizationID() field.String {
	return field.NewNullString()
}
func (user EmptyUser) Remarks() field.String {
	return field.NewNullString()
}
func (user EmptyUser) EncryptedUserIDByPassword() field.String {
	return field.NewNullString()
}
func (user EmptyUser) DeactivatedAt() field.Time {
	return field.NewNullTime()
}
func (user EmptyUser) LoginEmail() field.String {
	return field.NewNullString()
}

type UserProfileLoginEmailDelegate struct {
	Email string
}

func (u *UserProfileLoginEmailDelegate) LoginEmail() field.String {
	return field.NewString(u.Email)
}
func (user EmptyUser) UserRole() field.String {
	return field.NewNullString()
}
