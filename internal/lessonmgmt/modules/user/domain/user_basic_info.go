package domain

import "time"

type UserBasicInfo struct {
	UserID            string
	FullName          string
	FirstName         string
	LastName          string
	FullNamePhonetic  string
	FirstNamePhonetic string
	LastNamePhonetic  string
	Email             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UsersBasicInfo []*UserBasicInfo
