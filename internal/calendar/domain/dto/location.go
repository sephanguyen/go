package dto

import (
	"time"
)

type Location struct {
	LocationID              string
	Name                    string
	LocationType            string
	ParentLocationID        string
	PartnerInternalID       string
	PartnerInternalParentID string
	IsArchived              bool
	AccessPath              string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               time.Time
}
