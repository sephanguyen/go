package entities

import (
	"github.com/jackc/pgtype"
)

// User user ID and group permission
type User struct {
	ID                pgtype.Text `sql:"user_id,pk"`
	Avatar            pgtype.Text
	Group             pgtype.Text `sql:"user_group"`
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
	ResourcePath      pgtype.Text
	LastLoginDate     pgtype.Timestamptz
	Birthday          pgtype.Date
	Gender            pgtype.Text
}

// FieldMap returns field in users table
func (e *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"country",
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
			"deleted_at",
			"resource_path",
			"last_login_date",
			"birthday",
			"gender",
		}, []interface{}{
			&e.ID,
			&e.Group,
			&e.Country,
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
