package user

import (
	"fmt"
	"reflect"
	"strings"
)

type UserField string

const (
	UserFieldUID          UserField = "uid"
	UserFieldEmail        UserField = "email"
	UserFieldPhoneNumber  UserField = "phoneNumber"
	UserFieldDisplayName  UserField = "displayName"
	UserFieldPhotoURL     UserField = "photoURL"
	UserFieldCustomClaims UserField = "customClaims"
	UserFieldRawPassword  UserField = "rawPassword"
	UserFieldPasswordHash UserField = "passwordHash"
	UserFieldPasswordSalt UserField = "passwordSalt"

	UserMinimumPasswordLength = 6
	UserMaximumPasswordLength = 1024
)

var AllUserFields = []UserField{
	UserFieldUID,
	UserFieldEmail,
	UserFieldPhoneNumber,
	UserFieldDisplayName,
	UserFieldPhotoURL,
	UserFieldCustomClaims,
	UserFieldRawPassword,
}

var DefaultUserFieldsToCreate = []UserField{
	UserFieldUID,
	UserFieldEmail,
	UserFieldPhoneNumber,
	UserFieldDisplayName,
	UserFieldPhotoURL,
	UserFieldRawPassword,
}

var DefaultUserFieldsToImport = []UserField{
	UserFieldUID,
	UserFieldEmail,
	UserFieldPhoneNumber,
	UserFieldDisplayName,
	UserFieldPhotoURL,
	UserFieldCustomClaims,
	UserFieldPasswordHash,
	UserFieldPasswordSalt,
}

var DefaultUserFieldsToUpdate = []UserField{
	UserFieldUID,
	UserFieldEmail,
	UserFieldPhoneNumber,
	UserFieldDisplayName,
	UserFieldPhotoURL,
	UserFieldCustomClaims,
	UserFieldRawPassword,
}

type User interface {
	GetUID() string
	GetEmail() string
	GetPhoneNumber() string
	GetPhotoURL() string
	GetDisplayName() string
	GetCustomClaims() map[string]interface{}
	GetRawPassword() string
	GetPasswordHash() []byte
	GetPasswordSalt() []byte
}

// IsUserInfoValid checks user's fields is valid or not
// If fieldsToValid is empty, func will check all fields of user
func IsUserInfoValid(user User, fieldsToValid ...UserField) error {
	if len(fieldsToValid) == 0 {
		fieldsToValid = AllUserFields
	}

	for _, fieldToValid := range fieldsToValid {
		switch fieldToValid {
		case UserFieldUID:
			if strings.TrimSpace(user.GetUID()) == "" {
				return ErrUserUIDEmpty
			}
			if len(user.GetUID()) > 128 {
				return ErrUIDMaxLength
			}
		case UserFieldEmail:
			if strings.TrimSpace(user.GetEmail()) == "" {
				return ErrUserEmailEmpty
			}
		case UserFieldPhoneNumber:
			// Identity platform allowed to import a user with empty phone number
			if strings.TrimSpace(user.GetPhoneNumber()) == "" {
				continue
			}
			// If phone number is not empty, check format, it should be E.164
			// ...
		case UserFieldRawPassword:
			if len(user.GetRawPassword()) < UserMinimumPasswordLength {
				return ErrUserPasswordMinLength
			}
			if len(user.GetRawPassword()) > UserMaximumPasswordLength {
				return ErrUserPasswordMaxLength
			}
		}
	}
	return nil
}

type Users []User

func (users Users) Len() int {
	return len(users)
}

func (users Users) Cap() int {
	return cap(users)
}

func (users Users) IsEmpty() bool {
	return users.Len() < 1
}

func (users Users) Append(usersToAppend ...User) Users {
	newUsers := make([]User, 0, len(users)+len(usersToAppend))
	newUsers = append(newUsers, users...)
	newUsers = append(newUsers, usersToAppend...)
	return newUsers
}

func (users Users) IDAndUserMap() map[string]User {
	m := make(map[string]User, len(users))
	for _, user := range users {
		m[user.GetUID()] = user
	}
	return m
}

func (users Users) UserIDs() []string {
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.GetUID())
	}
	return ids
}

func (users Users) Emails() []string {
	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, user.GetEmail())
	}
	return emails
}

type user struct {
	uid          string
	email        string
	phoneNumber  string
	photoURL     string
	displayName  string
	customClaims map[string]interface{}
	rawPassword  string
	passwordHash []byte
	passwordSalt []byte
}

type Option func(*user)

func NewUser(opt ...Option) User {
	u := user{}
	for _, o := range opt {
		o(&u)
	}
	return &u
}

func WithUID(uid string) Option {
	return func(u *user) {
		u.uid = uid
	}
}

func WithEmail(email string) Option {
	return func(u *user) {
		u.email = email
	}
}

func WithPhoneNumber(phoneNumber string) Option {
	return func(u *user) {
		u.phoneNumber = phoneNumber
	}
}

func WithPhotoURL(photoURL string) Option {
	return func(u *user) {
		u.photoURL = photoURL
	}
}

func WithDisplayName(displayName string) Option {
	return func(u *user) {
		u.displayName = displayName
	}
}

func WithCustomClaims(customClaims map[string]interface{}) Option {
	return func(u *user) {
		u.customClaims = customClaims
	}
}

func WithRawPassword(rawPassword string) Option {
	return func(u *user) {
		u.rawPassword = rawPassword
	}
}

func WithPasswordHash(passwordHash []byte) Option {
	return func(u *user) {
		u.passwordHash = passwordHash
	}
}

func WithPasswordSalt(passwordSalt []byte) Option {
	return func(u *user) {
		u.passwordSalt = passwordSalt
	}
}

func (u *user) GetUID() string {
	return u.uid
}

func (u *user) GetEmail() string {
	return u.email
}

func (u *user) GetPhoneNumber() string {
	return u.phoneNumber
}

func (u *user) GetPhotoURL() string {
	return u.photoURL
}

func (u *user) GetDisplayName() string {
	return u.displayName
}

func (u *user) GetCustomClaims() map[string]interface{} {
	return u.customClaims
}

func (u *user) GetRawPassword() string {
	return u.rawPassword
}

func (u *user) GetPasswordHash() []byte {
	return u.passwordHash
}

func (u *user) GetPasswordSalt() []byte {
	return u.passwordSalt
}

func IsUserValueEqual(usr User, otherUsr User) bool {
	switch {
	case usr.GetUID() != otherUsr.GetUID():
		return false
	case strings.ToLower(usr.GetEmail()) != strings.ToLower(otherUsr.GetEmail()):
		return false
	case usr.GetPhoneNumber() != otherUsr.GetPhoneNumber():
		return false
	case usr.GetDisplayName() != otherUsr.GetDisplayName():
		return false
	case usr.GetPhotoURL() != otherUsr.GetPhotoURL():
		return false
	case len(usr.GetCustomClaims()) != len(otherUsr.GetCustomClaims()):
		fmt.Println("usr.GetCustomClaims()) != len(otherUsr.GetCustomClaims()")
		return false
	}
	if len(usr.GetCustomClaims()) > 0 || len(otherUsr.GetCustomClaims()) > 0 {
		if !reflect.DeepEqual(usr.GetCustomClaims(), otherUsr.GetCustomClaims()) {
			return false
		}
	}
	return true
}

type ImportUsersResult struct {
	TenantID            string
	UsersSuccessImport  []User
	UsersFailedToImport UsersFailedToImport
}

type UserFailedToImport struct {
	User User
	Err  string
}

type UsersFailedToImport []*UserFailedToImport

func (users UsersFailedToImport) IDs() []string {
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.User.GetUID())
	}
	return ids
}

func (users UsersFailedToImport) Emails() []string {
	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, user.User.GetEmail())
	}
	return emails
}
