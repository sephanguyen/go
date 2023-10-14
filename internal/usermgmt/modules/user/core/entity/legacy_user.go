package entity

import (
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// UserGroup enum value in DB
var (
	UserGroupStudent     = "USER_GROUP_STUDENT"
	UserGroupAdmin       = "USER_GROUP_ADMIN"
	UserGroupTeacher     = "USER_GROUP_TEACHER"
	UserGroupParent      = "USER_GROUP_PARENT"
	UserGroupSchoolAdmin = "USER_GROUP_SCHOOL_ADMIN"
)

// LegacyUser represents a user entity
// Deprecated: no longer used, please avoid using.
type LegacyUser struct {
	AppleUser `sql:"-"`

	ID                pgtype.Text `sql:"user_id,pk"`
	Avatar            pgtype.Text
	Group             pgtype.Text `sql:"user_group"`
	UserName          pgtype.Text
	FullName          pgtype.Text `sql:"name"`
	FirstName         pgtype.Text
	LastName          pgtype.Text
	FirstNamePhonetic pgtype.Text
	LastNamePhonetic  pgtype.Text
	FullNamePhonetic  pgtype.Text
	GivenName         pgtype.Text
	Country           pgtype.Text
	PhoneNumber       pgtype.Text
	Email             pgtype.Text
	DeviceToken       pgtype.Text
	AllowNotification pgtype.Bool `sql:",notnull"`
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	IsTester          pgtype.Bool `sql:",notnull"`
	FacebookID        pgtype.Text `sql:"facebook_id"`
	PhoneVerified     pgtype.Bool `sql:",notnull"`
	EmailVerified     pgtype.Bool `sql:",notnull"`
	DeletedAt         pgtype.Timestamptz
	DeactivatedAt     pgtype.Timestamptz
	ResourcePath      pgtype.Text
	LastLoginDate     pgtype.Timestamptz
	Birthday          pgtype.Date
	Gender            pgtype.Text
	Remarks           pgtype.Text
	ExternalUserID    pgtype.Text
	LoginEmail        pgtype.Text
	UserRole          pgtype.Text

	UserAdditionalInfo `sql:"-"`
}

// FieldMap returns field in users table
func (e *LegacyUser) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"country",
			"username",
			"name",
			"first_name",
			"last_name",
			"first_name_phonetic",
			"last_name_phonetic",
			"full_name_phonetic",
			"given_name",
			"avatar",
			"phone_number",
			"email",
			"device_token",
			"allow_notification",
			"updated_at",
			"created_at",
			"is_tester",
			"facebook_id",
			"phone_verified",
			"email_verified",
			"deactivated_at",
			"deleted_at",
			"resource_path",
			"last_login_date",
			"birthday",
			"gender",
			"remarks",
			"user_external_id",
			"login_email",
			"user_role",
		}, []interface{}{
			&e.ID,
			&e.Group,
			&e.Country,
			&e.UserName,
			&e.FullName,
			&e.FirstName,
			&e.LastName,
			&e.FirstNamePhonetic,
			&e.LastNamePhonetic,
			&e.FullNamePhonetic,
			&e.GivenName,
			&e.Avatar,
			&e.PhoneNumber,
			&e.Email,
			&e.DeviceToken,
			&e.AllowNotification,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.IsTester,
			&e.FacebookID,
			&e.PhoneVerified,
			&e.EmailVerified,
			&e.DeactivatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
			&e.LastLoginDate,
			&e.Birthday,
			&e.Gender,
			&e.Remarks,
			&e.ExternalUserID,
			&e.LoginEmail,
			&e.UserRole,
		}
}

func (e *LegacyUser) TableName() string {
	return "users"
}

func (e *LegacyUser) GetName() string {
	return e.FullName.String
}

// LegacyUsers is a slice of LegacyUser
// Deprecated: no longer used, please avoid using.
type LegacyUsers []*LegacyUser

func ToUsers(users ...*LegacyUser) LegacyUsers {
	list := make(LegacyUsers, len(users))
	for i, user := range users {
		list[i] = user
	}
	return list
}

func (u *LegacyUsers) Add() database.Entity {
	e := &LegacyUser{}
	*u = append(*u, e)

	return e
}

func (u LegacyUsers) Append(user ...*LegacyUser) LegacyUsers {
	newUsers := make(LegacyUsers, 0, len(u)+len(user))
	newUsers = append(newUsers, u...)
	newUsers = append(newUsers, user...)
	return newUsers
}

func (u LegacyUsers) Limit(numberOfUsers int) LegacyUsers {
	if length := len(u); numberOfUsers > length {
		numberOfUsers = length
	}
	return u[:numberOfUsers]
}

func (u LegacyUsers) Emails() []string {
	emails := make([]string, 0, len(u))
	for _, user := range u {
		emails = append(emails, user.Email.String)
	}
	return emails
}

func (u LegacyUsers) ExternalUserIDs() []string {
	externalUserIDs := make([]string, 0, len(u))
	for _, user := range u {
		externalUserIDs = append(externalUserIDs, user.ExternalUserID.String)
	}
	return externalUserIDs
}

func (u LegacyUsers) IDs() []string {
	ids := make([]string, 0, len(u))
	for _, user := range u {
		ids = append(ids, user.ID.String)
	}
	return ids
}

func (u LegacyUsers) PhoneNumbers() []string {
	phoneNumbers := make([]string, 0, len(u))
	for _, user := range u {
		if user.PhoneNumber.Status == pgtype.Present {
			phoneNumbers = append(phoneNumbers, user.Email.String)
		}
	}
	return phoneNumbers
}

type UserAdditionalInfo struct {
	Password     string // RawPassword
	CustomClaims map[string]interface{}
	PasswordHash []byte
	PasswordSalt []byte
	TagIDs       []string
}

// GetUID is auth.User implementation
func (e *LegacyUser) GetUID() string {
	return e.ID.String
}

// GetEmail is auth.User implementation
func (e *LegacyUser) GetEmail() string {
	return strings.ToLower(e.LoginEmail.String)
}

// GetExternalUserID is auth.User implementation
func (e *LegacyUser) GetExternalUserID() string {
	return e.ExternalUserID.String
}

// GetPhoneNumber is auth.User implementation
func (e *LegacyUser) GetPhoneNumber() string {
	return e.PhoneNumber.String
}

// GetDisplayName is auth.User implementation
func (e *LegacyUser) GetDisplayName() string {
	return e.GetName()
}

// GetPhotoURL is auth.User implementation
func (e *LegacyUser) GetPhotoURL() string {
	return e.Avatar.String
}

// GetCustomClaims is auth.User implementation
func (e *LegacyUser) GetCustomClaims() map[string]interface{} {
	return e.UserAdditionalInfo.CustomClaims
}

// GetRawPassword is auth.User implementation
func (e *LegacyUser) GetRawPassword() string {
	return e.UserAdditionalInfo.Password
}

// GetPasswordHash is auth.User implementation
func (e *LegacyUser) GetPasswordHash() []byte {
	return e.UserAdditionalInfo.PasswordHash
}

// GetPasswordSalt is auth.User implementation
func (e *LegacyUser) GetPasswordSalt() []byte {
	return e.UserAdditionalInfo.PasswordSalt
}
