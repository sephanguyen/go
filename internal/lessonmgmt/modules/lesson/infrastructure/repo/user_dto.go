package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

type User struct {
	AppleUser `sql:"-"`

	ID                pgtype.Text `sql:"user_id,pk"`
	Avatar            pgtype.Text
	Group             pgtype.Text `sql:"user_group"`
	LastName          pgtype.Text `sql:"name"`
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
	FirstName         pgtype.Text
}

type AppleUser struct {
	ID        pgtype.Text
	UserID    pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

// FieldMap returns field in users table
func (u *User) FieldMap() ([]string, []interface{}) {
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
			"first_name",
		}, []interface{}{
			&u.ID,
			&u.Group,
			&u.Country,
			&u.LastName,
			&u.GivenName,
			&u.Avatar,
			&u.PhoneNumber,
			&u.Email,
			&u.DeviceToken,
			&u.AllowNotification,
			&u.UpdatedAt,
			&u.CreatedAt,
			&u.IsTester,
			&u.FacebookID,
			&u.PhoneVerified,
			&u.EmailVerified,
			&u.DeletedAt,
			&u.ResourcePath,
			&u.LastLoginDate,
			&u.FirstName,
		}
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) GetName() string {
	if u.GivenName.String == "" {
		return u.LastName.String
	}

	return u.GivenName.String + " " + u.LastName.String
}

type Users []*User

func (u *Users) Add() database.Entity {
	e := &User{}
	*u = append(*u, e)

	return e
}

func (u *User) ToUserEntity() *domain.User {
	user := &domain.User{
		ID:                u.ID.String,
		Avatar:            u.Avatar.String,
		Group:             u.Group.String,
		LastName:          u.LastName.String,
		GivenName:         u.GivenName.String,
		Country:           u.Country.String,
		PhoneNumber:       u.PhoneNumber.String,
		Email:             u.Email.String,
		DeviceToken:       u.DeviceToken.String,
		AllowNotification: u.AllowNotification.Bool,
		ResourcePath:      u.ResourcePath.String,
		FacebookID:        u.FacebookID.String,
		PhoneVerified:     u.PhoneVerified.Bool,
		EmailVerified:     u.EmailVerified.Bool,
		CreatedAt:         u.CreatedAt.Time,
		UpdatedAt:         u.UpdatedAt.Time,
		LastLoginDate:     u.LastLoginDate.Time,
		FirstName:         u.FirstName.String,
	}
	if u.DeletedAt.Status == pgtype.Present {
		user.DeletedAt = &u.DeletedAt.Time
	}
	return user
}
