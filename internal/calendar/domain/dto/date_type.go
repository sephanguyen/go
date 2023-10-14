package dto

import "time"

type DateType struct {
	DateTypeID  string
	DisplayName string
	IsArchived  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   time.Time
}
