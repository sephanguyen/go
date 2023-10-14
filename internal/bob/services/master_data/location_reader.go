package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

type MasterDataReaderService struct {
	DB                    database.Ext
	LocationReaderService interface {
		RetrieveLocations(context.Context, *mpb.RetrieveLocationsRequest) (*mpb.RetrieveLocationsResponse, error)
		RetrieveLocationTypes(context.Context, *mpb.RetrieveLocationTypesRequest) (*mpb.RetrieveLocationTypesResponse, error)
		RetrieveLowestLevelLocations(context.Context, *mpb.RetrieveLowestLevelLocationsRequest) (*mpb.RetrieveLowestLevelLocationsResponse, error)
	}
}

func (m *MasterDataReaderService) RetrieveLocations(ctx context.Context, req *bpb.RetrieveLocationsRequest) (*bpb.RetrieveLocationsResponse, error) {
	res, err := m.LocationReaderService.RetrieveLocations(ctx, &mpb.RetrieveLocationsRequest{
		IsArchived: req.GetIsArchived(),
	})
	if err != nil {
		return nil, err
	}
	locations := make([]*bpb.RetrieveLocationsResponse_Location, 0, len(res.Locations))
	for _, l := range res.Locations {
		location := &bpb.RetrieveLocationsResponse_Location{
			LocationId:       l.LocationId,
			Name:             l.Name,
			LocationType:     l.LocationType,
			ParentLocationId: l.ParentLocationId,
			CreatedAt:        l.CreatedAt,
			AccessPath:       l.AccessPath,
		}
		locations = append(locations, location)
	}

	return &bpb.RetrieveLocationsResponse{Locations: locations}, nil
}

func (m *MasterDataReaderService) RetrieveLocationTypes(ctx context.Context, req *bpb.RetrieveLocationTypesRequest) (*bpb.RetrieveLocationTypesResponse, error) {
	res, err := m.LocationReaderService.RetrieveLocationTypes(ctx, &mpb.RetrieveLocationTypesRequest{})
	if err != nil {
		return nil, err
	}
	locationTypes := make([]*bpb.RetrieveLocationTypesResponse_LocationType, 0, len(res.LocationTypes))
	for _, lt := range res.LocationTypes {
		locationType := &bpb.RetrieveLocationTypesResponse_LocationType{
			LocationTypeId:       lt.LocationTypeId,
			Name:                 lt.Name,
			DisplayName:          lt.DisplayName,
			ParentName:           lt.ParentName,
			ParentLocationTypeId: lt.ParentLocationTypeId,
		}
		locationTypes = append(locationTypes, locationType)
	}
	return &bpb.RetrieveLocationTypesResponse{LocationTypes: locationTypes}, nil
}

func (m *MasterDataReaderService) RetrieveLowestLevelLocations(ctx context.Context, req *bpb.RetrieveLowestLevelLocationsRequest) (*bpb.RetrieveLowestLevelLocationsResponse, error) {
	res, err := m.LocationReaderService.RetrieveLowestLevelLocations(ctx, &mpb.RetrieveLowestLevelLocationsRequest{
		Name:        req.Name,
		Limit:       req.Limit,
		Offset:      req.Offset,
		LocationIds: req.LocationIds,
	})
	if err != nil {
		return nil, err
	}
	locations := make([]*bpb.RetrieveLowestLevelLocationsResponse_Location, 0, len(res.Locations))
	for _, l := range res.Locations {
		location := &bpb.RetrieveLowestLevelLocationsResponse_Location{
			LocationId: l.LocationId,
			Name:       l.Name,
		}
		locations = append(locations, location)
	}
	return &bpb.RetrieveLowestLevelLocationsResponse{Locations: locations}, nil
}
