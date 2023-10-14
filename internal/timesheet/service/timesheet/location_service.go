package timesheet

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type LocationServiceImpl struct {
	DB database.Ext

	LocationRepo interface {
		GetGrantedLocationOfStaff(ctx context.Context, db database.QueryExecer, staffID pgtype.Text) ([]*entity.Location, error)
		GetListChildLocations(ctx context.Context, db database.QueryExecer, locationID pgtype.Text) ([]*entity.Location, error)
	}
}

func (s *LocationServiceImpl) GetListGrantedLocationOfStaff(ctx context.Context, staffID, searchName string, limit int32) ([]*dto.Location, error) {
	staffGrantedLocationEntities, err := s.LocationRepo.GetGrantedLocationOfStaff(ctx, s.DB, database.Text(staffID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("location service get granted location of staff %s got error: %s", staffID, err.Error())
	}

	locationDtos := dto.ListLocations{}
	if len(staffGrantedLocationEntities) > 0 {
		for _, grantedLocationE := range staffGrantedLocationEntities {
			listSecondLevelLocation, err := s.LocationRepo.GetListChildLocations(ctx, s.DB, database.Text(grantedLocationE.LocationID.String))
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}

			if len(listSecondLevelLocation) > 0 {
				for _, secondLevelLocation := range listSecondLevelLocation {
					listThirdLevelLocation, err := s.LocationRepo.GetListChildLocations(ctx, s.DB, database.Text(secondLevelLocation.LocationID.String))
					if err != nil && !errors.Is(err, pgx.ErrNoRows) {
						return nil, err
					}

					if len(listThirdLevelLocation) > 0 {
						for _, thirdLevelLocation := range listThirdLevelLocation {
							locationDtos = append(locationDtos, dto.NewLocationDtoFromEntity(thirdLevelLocation))
						}
					} else {
						locationDtos = append(locationDtos, dto.NewLocationDtoFromEntity(secondLevelLocation))
					}
				}
			} else {
				locationDtos = append(locationDtos, dto.NewLocationDtoFromEntity(grantedLocationE))
			}
		}
	}
	// sometimes user belong more than one group and this will duplicate locations
	if len(locationDtos) > 0 {
		locationDtos = unique(locationDtos)
		locationDtos = search(locationDtos, searchName)
		locationDtos = getLimit(locationDtos, limit)
	}

	return locationDtos, nil
}

func unique(locations []*dto.Location) []*dto.Location {
	inResult := make(map[string]bool)
	var result []*dto.Location
	for _, location := range locations {
		if _, ok := inResult[location.LocationID]; !ok {
			inResult[location.LocationID] = true
			result = append(result, location)
		}
	}
	return result
}

func getLimit(locations []*dto.Location, limit int32) []*dto.Location {
	if len(locations) > int(limit) {
		return locations[:limit]
	}
	return locations
}

func search(locations []*dto.Location, searchName string) []*dto.Location {
	if searchName == "" {
		return locations
	}
	result := []*dto.Location{}
	for _, location := range locations {
		if strings.Contains(strings.ToLower(location.Name), strings.ToLower(searchName)) {
			result = append(result, location)
		}
	}

	return result
}
