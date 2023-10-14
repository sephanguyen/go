package entity

import (
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type User struct {
	ID                pgtype.Text
	Avatar            pgtype.Text
	Group             pgtype.Text
	LastName          pgtype.Text
	GivenName         pgtype.Text
	Country           pgtype.Text
	PhoneNumber       pgtype.Text
	Email             pgtype.Text
	DeviceToken       pgtype.Text
	AllowNotification pgtype.Bool
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	IsTester          pgtype.Bool
	FacebookID        pgtype.Text
	PhoneVerified     pgtype.Bool
	EmailVerified     pgtype.Bool
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
	LastLoginDate     pgtype.Timestamptz
	Birthday          pgtype.Date
	Gender            pgtype.Text

	UserAdditionalInfo `sql:"-"`
}

// FieldMap returns field in users table
func (e *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"country",
			"name",
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
			"deleted_at",
			"resource_path",
			"last_login_date",
			"birthday",
			"gender",
		}, []interface{}{
			&e.ID,
			&e.Group,
			&e.Country,
			&e.LastName,
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
			&e.DeletedAt,
			&e.ResourcePath,
			&e.LastLoginDate,
			&e.Birthday,
			&e.Gender,
		}
}

func (e *User) TableName() string {
	return "users"
}

func (e *User) GetName() string {
	if e.GivenName.String == "" {
		return e.LastName.String
	}

	return e.GivenName.String + " " + e.LastName.String
}

type Users []*User

func (u *Users) Add() database.Entity {
	e := &User{}
	*u = append(*u, e)

	return e
}

func (u Users) Append(user ...*User) Users {
	newUsers := make(Users, 0, len(u)+len(user))
	newUsers = append(newUsers, u...)
	newUsers = append(newUsers, user...)
	return newUsers
}

func (u Users) Emails() []string {
	emails := make([]string, 0, len(u))
	for _, user := range u {
		emails = append(emails, user.Email.String)
	}
	return emails
}

func (u Users) PhoneNumbers() []string {
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
}

// GetUID is auth.User implementation
func (e *User) GetUID() string {
	return e.ID.String
}

// GetEmail is auth.User implementation
func (e *User) GetEmail() string {
	return strings.ToLower(e.Email.String)
}

// GetPhoneNumber is auth.User implementation
func (e *User) GetPhoneNumber() string {
	return e.PhoneNumber.String
}

// GetDisplayName is auth.User implementation
func (e *User) GetDisplayName() string {
	return e.GetName()
}

// GetPhotoURL is auth.User implementation
func (e *User) GetPhotoURL() string {
	return e.Avatar.String
}

// GetCustomClaims is auth.User implementation
func (e *User) GetCustomClaims() map[string]interface{} {
	return e.UserAdditionalInfo.CustomClaims
}

// GetRawPassword is auth.User implementation
func (e *User) GetRawPassword() string {
	return e.UserAdditionalInfo.Password
}

// GetPasswordHash is auth.User implementation
func (e *User) GetPasswordHash() []byte {
	return e.UserAdditionalInfo.PasswordHash
}

// GetPasswordSalt is auth.User implementation
func (e *User) GetPasswordSalt() []byte {
	return e.UserAdditionalInfo.PasswordSalt
}
