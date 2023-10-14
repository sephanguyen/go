package multitenant

import (
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
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

//User has the same signature with the domain entity of the user
//We don't directly import business logic entity but define a clone here
//to avoid dependencies since this pkg can be used by multiple modules
type User interface {
	UserID() field.String
	Email() field.String
	PhoneNumber() field.String
	Password() field.String
	Avatar() field.String
	FullName() field.String
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
			if strings.TrimSpace(user.UserID().String()) == "" {
				return ErrUserUIDEmpty
			}
			if len(user.UserID().String()) > 128 {
				return ErrUIDMaxLength
			}
		case UserFieldEmail:
			if strings.TrimSpace(user.Email().String()) == "" {
				return ErrUserEmailEmpty
			}
		case UserFieldPhoneNumber:
			// Identity platform allowed to import a user with empty phone number
			if strings.TrimSpace(user.PhoneNumber().String()) == "" {
				continue
			}
			// If phone number is not empty, check format, it should be E.164
			// ...
		case UserFieldRawPassword:
			if len(user.Password().String()) < UserMinimumPasswordLength {
				return ErrUserPasswordMinLength
			}
			if len(user.Password().String()) > UserMaximumPasswordLength {
				return ErrUserPasswordMaxLength
			}
		}
	}
	return nil
}

type Users []User

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
		ids = append(ids, user.User.UserID().String())
	}
	return ids
}

func (users UsersFailedToImport) Emails() []string {
	emails := make([]string, 0, len(users))
	for _, user := range users {
		emails = append(emails, user.User.Email().String())
	}
	return emails
}

/*type User interface {
	GetUID() string
	GetEmail() string
	GetPhoneNumber() string
	GetPhotoURL() string
	GetDisplayName() string
	GetCustomClaims() map[string]interface{}
	GetRawPassword() string
	GetPasswordHash() []byte
	GetPasswordSalt() []byte
}*/

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

func (user *user) UserID() field.String {
	return field.NewString(user.uid)
}
func (user *user) Avatar() field.String {
	return field.NewString(user.photoURL)
}
func (user *user) FullName() field.String {
	return field.NewString(user.displayName)
}
func (user *user) PhoneNumber() field.String {
	return field.NewString(user.phoneNumber)
}
func (user *user) Email() field.String {
	return field.NewString(user.email)
}
func (user *user) Password() field.String {
	return field.NewString(user.rawPassword)
}

//need to refactor one more time after update integration test
/*func IsUserValueEqual(usr User, otherUsr User) bool {
	switch {
	case usr.UserID().String() != otherUsr.UserID().String():
		return false
	case strings.ToLower(usr.Email().String()) != strings.ToLower(otherUsr.Email().String()):
		return false
	case usr.PhoneNumber().String() != otherUsr.PhoneNumber().String():
		return false
	case usr.FullName().String() != otherUsr.FullName().String():
		return false
	case usr.Avatar().String() != otherUsr.Avatar().String():
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
}*/
