package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

type UpsertLocation struct {
	Locations []*domain.Location
}

type LocationDataList struct {
	parentIDs       []string
	locationTypeIDs []string
	locationIDs     []string
}

type LocationLog struct {
	DeletedIds  []string
	UpsertedIds []string
}
