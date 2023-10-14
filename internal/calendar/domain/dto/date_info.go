package dto

import "time"

type UpsertDateInfoParams struct {
	DateInfo *DateInfo
}

type DuplicateDateInfoParams struct {
	DateInfo *DateInfo
	Dates    []time.Time
}

type DateInfo struct {
	Date                time.Time
	LocationID          string
	DateTypeID          string
	DateTypeDisplayName string
	OpeningTime         string
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
	TimeZone            string
}
