package domain

import (
	"time"
)

type ClassMember struct {
	ClassMemberID string
	ClassID       string
	UserID        string
	UpdatedAt     time.Time
	CreatedAt     time.Time
	DeletedAt     *time.Time
	StartDate     time.Time
	EndDate       time.Time
}
