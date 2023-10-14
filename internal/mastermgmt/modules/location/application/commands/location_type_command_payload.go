package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

type UpsertLocationTypes struct {
	LocationTypes map[int]*domain.LocationType
	Parents       map[int]*domain.LocationType
}
type ImportLocationTypeV2Payload struct {
	LocationTypes []*domain.LocationType
}
type LocationTypeLog struct {
	DeletedNames  []string
	UpsertedNames []string
}
