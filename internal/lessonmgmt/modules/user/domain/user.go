package domain

import (
	"fmt"
	"time"
)

// Teachers are not entity or aggregate, it's just is data type to contain list teaches
type Teachers []*Teacher

func (t Teachers) IsValid() error {
	for i := range t {
		if err := t[i].IsValid(); err != nil {
			return err
		}
	}
	return nil
}

func NewTeacher(
	id,
	name string,
	createdAt,
	updatedAt time.Time,
) *Teacher {
	return &Teacher{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

type Teacher struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Teacher) IsValid() error {
	if len(t.ID) == 0 {
		return fmt.Errorf("Teacher.ID could not be empty")
	}

	if len(t.Name) == 0 {
		return fmt.Errorf("Teacher.Name could not be empty")
	}

	if t.UpdatedAt.Before(t.CreatedAt) {
		return fmt.Errorf("updated time could not before created time")
	}

	return nil
}

type User struct {
	ID                string
	Avatar            string
	Group             string
	FullName          string
	LastName          string
	FirstName         string
	GivenName         string
	Country           string
	PhoneNumber       string
	Email             string
	DeviceToken       string
	AllowNotification bool
	FacebookID        string
	PhoneVerified     bool
	EmailVerified     bool
	ResourcePath      string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	LastLoginDate     time.Time
	AppleUser         AppleUser
}

func (u *User) GetName() string {
	if u.GivenName == "" {
		if u.FirstName == "" {
			return u.LastName
		}
		return u.FirstName + " " + u.LastName
	}
	return u.GivenName
}

type AppleUser struct {
	ID        string
	UserID    string
	UpdatedAt time.Time
	CreatedAt time.Time
}

type Users []*User

type Students []*Student

type Student struct {
	ID    string
	Name  string
	Email string
}

func NewStudent(
	id,
	name,
	email string,
) *Student {
	return &Student{
		ID:    id,
		Name:  name,
		Email: email,
	}
}
