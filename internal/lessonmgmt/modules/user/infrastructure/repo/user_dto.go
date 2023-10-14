package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
)

type User struct {
	ID                pgtype.Text `sql:"user_id,pk"`
	Avatar            pgtype.Text
	Group             pgtype.Text `sql:"user_group"`
	FullName          pgtype.Text `sql:"name"`
	FirstName         pgtype.Text
	LastName          pgtype.Text
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
}

// FieldMap returns field in users table
func (u *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"country",
			"name",
			"first_name",
			"last_name",
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
		}, []interface{}{
			&u.ID,
			&u.Group,
			&u.Country,
			&u.FullName,
			&u.FirstName,
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
		}
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) GetName() string {
	if u.FullName.String == "" {
		if u.FirstName.String == "" {
			return u.LastName.String
		}
		return u.FirstName.String + " " + u.LastName.String
	}
	return u.FullName.String
}

type Users []*User

func (u *Users) Add() database.Entity {
	e := &User{}
	*u = append(*u, e)

	return e
}

type Teacher struct {
	UserBasicInfo `sql:"-"`
	ID            pgtype.Text `sql:"staff_id,pk"`
	UpdatedAt     pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
}

func (t *Teacher) FieldMap() ([]string, []interface{}) {
	return []string{
			"staff_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.UpdatedAt,
			&t.CreatedAt,
			&t.DeletedAt,
		}
}

func (t *Teacher) TableName() string {
	return "staff"
}

func (t *Teacher) ToTeacherEntity() *domain.Teacher {
	return domain.NewTeacher(t.ID.String, t.UserBasicInfo.GetName(), t.CreatedAt.Time, t.UpdatedAt.Time)
}

type Teachers []*Teacher

func (t *Teachers) Add() database.Entity {
	e := &Teacher{}
	*t = append(*t, e)

	return e
}

func (t Teachers) ToTeachersEntity() domain.Teachers {
	res := make(domain.Teachers, 0, len(t))
	for i := range t {
		res = append(res, t[i].ToTeacherEntity())
	}

	return res
}

// UserFindFilter for filtering users in DB
type UserFindFilter struct {
	IDs       pgtype.TextArray
	Email     pgtype.Text
	Phone     pgtype.Text
	UserGroup pgtype.Text
}

func (u *User) ToUserEntity() *domain.User {
	user := &domain.User{
		ID:                u.ID.String,
		Avatar:            u.Avatar.String,
		Group:             u.Group.String,
		LastName:          u.LastName.String,
		GivenName:         u.GivenName.String,
		FirstName:         u.FirstName.String,
		FullName:          u.FullName.String,
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
	}
	if u.DeletedAt.Status == pgtype.Present {
		user.DeletedAt = &u.DeletedAt.Time
	}
	return user
}

type Students []*Student

type Student struct {
	ID    pgtype.Text
	Name  pgtype.Text
	Email pgtype.Text
}

func (s *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"name",
			"email",
		}, []interface{}{
			&s.ID,
			&s.Name,
			&s.Email,
		}
}

func (s *Student) TableName() string {
	return "user_basic_info"
}

func (s *Student) ToStudentEntity() *domain.Student {
	return domain.NewStudent(s.ID.String, s.Name.String, s.Email.String)
}

func (t Students) ToStudentsEntity() domain.Students {
	res := make(domain.Students, 0, len(t))
	for i := range t {
		res = append(res, t[i].ToStudentEntity())
	}

	return res
}
