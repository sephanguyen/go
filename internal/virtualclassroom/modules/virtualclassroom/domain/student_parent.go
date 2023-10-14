package domain

import "time"

type StudentParent struct {
	StudentID    string
	ParentID     string
	Relationship string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
