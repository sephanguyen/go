package domain

import (
	"time"
)

type (
	EnrollmentStatus string
)

const (
	EnrollmentStatusEnrolled  EnrollmentStatus = "STUDENT_ENROLLMENT_STATUS_ENROLLED"
	EnrollmentStatusPotential EnrollmentStatus = "STUDENT_ENROLLMENT_STATUS_POTENTIAL"
)

type User struct {
	ID                string
	Avatar            string
	Group             string
	LastName          string
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
	FirstName         string

	AppleUser AppleUser
}

func (e *User) GetName() string {
	if e.GivenName == "" {
		return e.LastName
	}

	return e.GivenName + " " + e.LastName
}

type AppleUser struct {
	ID        string
	UserID    string
	UpdatedAt time.Time
	CreatedAt time.Time
}

type UserPermissions struct {
	UserGroup        string
	Permissions      []string
	LocationIDs      []string
	GrantedLocations map[string][]string
}
