package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LocationReaderServices struct {
	DB                         database.Ext
	LocationRepo               infrastructure.LocationRepo
	LocationTypeRepo           infrastructure.LocationTypeRepo
	GetLocationQueryHandler    queries.GetLocationQueryHandler
	ExportLocationQueryHandler queries.ExportLocationQueryHandler
}

func (l *LocationReaderServices) RetrieveLocations(ctx context.Context, req *mpb.RetrieveLocationsRequest) (*mpb.RetrieveLocationsResponse, error) {
	locations, err := l.GetLocationQueryHandler.GetBaseLocationsByQuery(ctx, &queries.GetLocations{
		FilterLocation: domain.FilterLocation{
			IncludeIsArchived: true,
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`GetLocationQueryHandler.GetLocationsTree: %v`, err))
	}
	return &mpb.RetrieveLocationsResponse{
		Locations: convertDomainToLocationRes(locations),
	}, nil
}

func convertDomainToLocationRes(locations []*domain.Location) (res []*mpb.RetrieveLocationsResponse_Location) {
	// sort by updated_at, created_at asc
	slices.SortFunc(locations, func(l1, l2 *domain.Location) bool {
		if l1.UpdatedAt.Equal(l2.UpdatedAt) {
			return l1.CreatedAt.Before(l2.CreatedAt)
		}
		return l1.UpdatedAt.Before(l2.UpdatedAt)
	})

	locationsResp := make([]*mpb.RetrieveLocationsResponse_Location, 0, len(locations))
	for _, l := range locations {
		location := &mpb.RetrieveLocationsResponse_Location{
			LocationId:       l.LocationID,
			Name:             l.Name,
			LocationType:     l.LocationType,
			ParentLocationId: l.ParentLocationID,
			CreatedAt:        timestamppb.New(l.CreatedAt),
			AccessPath:       l.AccessPath,
			IsUnauthorized:   l.IsUnauthorized,
		}
		locationsResp = append(locationsResp, location)
	}
	return locationsResp
}

func (l *LocationReaderServices) RetrieveLocationTypes(ctx context.Context, req *mpb.RetrieveLocationTypesRequest) (*mpb.RetrieveLocationTypesResponse, error) {
	locationTypes, err := l.LocationTypeRepo.RetrieveLocationTypes(ctx, l.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`locationTypeRepo.RetrieveLocationTypes: %v`, err))
	}
	locationsResp := make([]*mpb.RetrieveLocationTypesResponse_LocationType, 0, len(locationTypes))
	for _, lt := range locationTypes {
		locationType := &mpb.RetrieveLocationTypesResponse_LocationType{
			LocationTypeId:       lt.LocationTypeID,
			Name:                 lt.Name,
			DisplayName:          lt.DisplayName,
			ParentName:           lt.ParentName,
			ParentLocationTypeId: lt.ParentLocationTypeID,
			Level:                int32(lt.Level),
		}
		locationsResp = append(locationsResp, locationType)
	}
	return &mpb.RetrieveLocationTypesResponse{
		LocationTypes: locationsResp,
	}, nil
}

func (l *LocationReaderServices) RetrieveLocationTypesV2(ctx context.Context, req *mpb.RetrieveLocationTypesV2Request) (*mpb.RetrieveLocationTypesV2Response, error) {
	locationTypes, err := l.LocationTypeRepo.RetrieveLocationTypesV2(ctx, l.DB)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`locationTypeRepo.RetrieveLocationTypesV2: %v`, err))
	}

	locTypeResp := sliceutils.Map(locationTypes, func(lt *domain.LocationType) *mpb.RetrieveLocationTypesV2Response_LocationType {
		return &mpb.RetrieveLocationTypesV2Response_LocationType{
			LocationTypeId: lt.LocationTypeID,
			Name:           lt.Name,
			DisplayName:    lt.DisplayName,
			Level:          int32(lt.Level),
		}
	})

	return &mpb.RetrieveLocationTypesV2Response{
		LocationTypes: locTypeResp,
	}, nil
}

func (l *LocationReaderServices) RetrieveLowestLevelLocations(ctx context.Context, req *mpb.RetrieveLowestLevelLocationsRequest) (*mpb.RetrieveLowestLevelLocationsResponse, error) {
	params := &repo.GetLowestLevelLocationsParams{
		Name:        req.Name,
		Limit:       req.Limit,
		Offset:      req.Offset,
		LocationIDs: req.LocationIds,
	}
	var locations []*domain.Location

	locations, err := l.LocationRepo.GetLowestLevelLocationsV2(ctx, l.DB, params)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`locationRepo.GetLowestLevelLocationsV2: %v`, err))
	}

	locationsResp := sliceutils.Map(locations, func(l *domain.Location) *mpb.RetrieveLowestLevelLocationsResponse_Location {
		return &mpb.RetrieveLowestLevelLocationsResponse_Location{
			LocationId: l.LocationID,
			Name:       l.Name,
		}
	})

	return &mpb.RetrieveLowestLevelLocationsResponse{
		Locations: locationsResp,
	}, nil
}

func (l *LocationReaderServices) ExportLocations(ctx context.Context, req *mpb.ExportLocationsRequest) (res *mpb.ExportLocationsResponse, err error) {
	bytes, err := l.ExportLocationQueryHandler.ExportLocation(ctx)
	if err != nil {
		return &mpb.ExportLocationsResponse{}, err
	}
	res = &mpb.ExportLocationsResponse{
		Data: bytes,
	}
	return res, nil
}

func (l *LocationReaderServices) ExportLocationTypes(ctx context.Context, req *mpb.ExportLocationTypesRequest) (res *mpb.ExportLocationTypesResponse, err error) {
	bytes, err := l.ExportLocationQueryHandler.ExportLocationType(ctx)
	if err != nil {
		return &mpb.ExportLocationTypesResponse{}, err
	}
	res = &mpb.ExportLocationTypesResponse{
		Data: bytes,
	}
	return res, nil
}

func (l *LocationReaderServices) GenerateUnauthorizedLocation(locations []*domain.Location, lt []*domain.LocationType) ([]*domain.Location, error) {
	var rs []*domain.Location

	if len(locations) == 0 {
		return locations, nil
	}

	locationType, err := l.SortLocationType(lt)
	if err != nil {
		return nil, err
	}

	levels := make([]map[string]*domain.Location, len(locationType))

	for i := 0; i < len(locationType); i++ {
		levels[i] = make(map[string]*domain.Location)
	}

	for _, location := range locations {
		path := strings.Split(location.AccessPath, "/")
		if len(path) == 1 {
			return locations, nil
		}

		location.IsUnauthorized = false
		levels[len(path)-1][location.LocationID] = location
	}

	orgLocationID := strings.Split(locations[0].AccessPath, "/")[0]
	unauthLocationOrg := &domain.Location{
		LocationID:       orgLocationID,
		Name:             "UnAuthorized",
		LocationType:     locationType[0].LocationTypeID,
		ParentLocationID: "",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		AccessPath:       orgLocationID,
		IsUnauthorized:   true,
	}
	rs = append(rs, unauthLocationOrg)

	for i := 1; i < len(levels); i++ {
		// for each existing location at level[i] remove all its children
		for _, location := range levels[i] {
			l.ReplaceByParent(i, levels, location, &rs)
		}
	}

	for i := len(levels) - 1; i > 0; i-- {
		// for each existing location at level[i] create its parent
		for _, location := range levels[i] {
			if _, ok := levels[i][location.ParentLocationID]; !ok {
				parentAccessPath := strings.ReplaceAll(location.AccessPath, "/"+location.LocationID, "")
				path := strings.Split(parentAccessPath, "/")
				grandparentID := ""
				if len(path) > 1 {
					grandparentID = path[len(path)-2]
				}
				unauthLocation := &domain.Location{
					LocationID:       location.ParentLocationID,
					Name:             "UnAuthorized",
					LocationType:     locationType[i-1].LocationTypeID,
					ParentLocationID: grandparentID,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					AccessPath:       parentAccessPath,
					IsUnauthorized:   true,
				}
				l.ReplaceByParent(i-1, levels, unauthLocation, &rs)
				levels[i-1][unauthLocation.LocationID] = unauthLocation
			}
		}
	}

	return rs, nil
}

func (l *LocationReaderServices) SortLocationType(locationType []*domain.LocationType) ([]*domain.LocationType, error) {
	var rs []*domain.LocationType

	for _, locType := range locationType {
		if locType.ParentLocationTypeID == "" {
			rs = append(rs, locType)
			break
		}
	}

	for range locationType {
		for _, lt := range locationType {
			if lt.ParentLocationTypeID == rs[len(rs)-1].LocationTypeID {
				rs = append(rs, lt)
				break
			}
		}
	}

	if len(rs) < len(locationType) {
		return nil, errors.New("wrong location_type")
	}

	return rs, nil
}

func (l *LocationReaderServices) ReplaceByParent(currentLevelCheck int, levels []map[string]*domain.Location, parentLocation *domain.Location, rs *[]*domain.Location) {
	for i := currentLevelCheck + 1; i < len(levels); i++ {
		for _, location := range levels[i] {
			if strings.Contains(location.AccessPath, parentLocation.AccessPath) {
				*rs = append(*rs, location)
				delete(levels[i], location.LocationID)
			}
		}
	}
}

func (l *LocationReaderServices) GetLocationTree(ctx context.Context, req *mpb.GetLocationTreeRequest) (*mpb.GetLocationTreeResponse, error) {
	q := &queries.GetLocations{
		FilterLocation: domain.FilterLocation{
			IncludeIsArchived: true,
		},
	}
	jsonTree, err := l.GetLocationQueryHandler.GetLocationsTree(ctx, q)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf(`GetLocationQueryHandler.GetLocationsTree: %v`, err))
	}

	return &mpb.GetLocationTreeResponse{Tree: jsonTree}, nil
}
