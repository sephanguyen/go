package dto

import (
	"errors"

	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type Location struct {
	LocationID string
	Name       string
}

func NewLocationDtoFromEntity(locationE *entity.Location) *Location {
	return &Location{
		LocationID: locationE.LocationID.String,
		Name:       locationE.Name.String,
	}
}

type ListLocations []*Location

func ValidateGetGrantedLocation(req *tpb.GetGrantedLocationsOfStaffRequest) error {
	if req.GetStaffId() == "" {
		return errors.New("staff id must be not empty")
	}
	if req.GetLimit() == 0 {
		return errors.New("limit number must be not empty")
	}

	return nil
}

func NewListLocationsToRPCResponse(locations []*Location) *tpb.GetGrantedLocationsOfStaffResponse {
	responseLocation := []*tpb.Location{}
	for _, location := range locations {
		tempLocation := &tpb.Location{
			LocationId: location.LocationID,
			Name:       location.Name,
		}
		responseLocation = append(responseLocation, tempLocation)
	}

	return &tpb.GetGrantedLocationsOfStaffResponse{
		Locations: responseLocation,
	}
}
