package domain

import "time"

type TreeLocation struct {
	LocationID        string    `json:"locationId"`
	Name              string    `json:"name"`
	LocationType      string    `json:"locationType"`
	ParentLocationID  string    `json:"parentLocationId"`
	PartnerInternalID string    `json:"partnerInternalId"`
	IsArchived        bool      `json:"isArchived"`
	AccessPath        string    `json:"accessPath"`
	IsUnauthorized    bool      `json:"isUnauthorized"`
	IsLowestLevel     bool      `json:"isLowestLevel"`
	UpdatedAt         time.Time `json:"createdAt"`
	CreatedAt         time.Time `json:"updatedAt"`

	Children []*TreeLocation `json:"children"`
}
