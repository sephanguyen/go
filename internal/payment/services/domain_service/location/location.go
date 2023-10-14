package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocationService struct {
	locationRepo interface {
		GetByIDForUpdate(ctx context.Context, db database.QueryExecer, locationID string) (entities.Location, error)
		GetByIDs(ctx context.Context, db database.QueryExecer, locationIDs []string) ([]entities.Location, error)
		GetByID(ctx context.Context, db database.QueryExecer, entitiesID string) (entities.Location, error)
		GetLowestGrantedLocationIDsByUserIDAndPermissions(ctx context.Context, db database.QueryExecer, args repositories.GetGrantedLowestLevelLocationsParams) (locationIDs []string, err error)
	}
}

func (s *LocationService) GetLocationNameByID(ctx context.Context, db database.QueryExecer, locationID string) (locationName string, err error) {
	var location entities.Location
	location, err = s.locationRepo.GetByIDForUpdate(ctx, db, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when checking location id: %v", err.Error())
		return
	}
	if location.LocationID.Status != pgtype.Present {
		err = status.Errorf(codes.FailedPrecondition, "This location with id %s does not exist in the system", locationID)
	}
	locationName = location.Name.String
	return
}

func (s *LocationService) GetLocationsByIDs(ctx context.Context, db database.Ext, locationIDs []string) (locations []entities.Location, err error) {
	locations, err = s.locationRepo.GetByIDs(ctx, db, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get locations by ids: %v", err.Error())
	}
	return
}

func (s *LocationService) GetLocationInfoByID(ctx context.Context, db database.Ext, locationID string) (locationInfo *pb.LocationInfo, err error) {
	location, err := s.locationRepo.GetByID(ctx, db, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get location by id: %v", err.Error())
	}
	locationInfo = &pb.LocationInfo{
		LocationId:   location.LocationID.String,
		LocationName: location.Name.String,
	}
	return
}

func (s *LocationService) GetLowestGrantedLocationsForCreatingOrder(ctx context.Context, db database.Ext, req *pb.GetLocationsForCreatingOrderRequest) (locationInfos []*pb.LocationInfo, err error) {
	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("cannot get userID from context")
	}

	params := repositories.GetGrantedLowestLevelLocationsParams{
		Name:   req.Name,
		Limit:  req.Limit,
		UserID: userID,
		PermissionNames: []string{
			constant.LocationReadPermission,
			constant.OrderWritePermission,
		},
		LocationIDs: req.LocationIds,
	}
	locationIDs, err := s.locationRepo.GetLowestGrantedLocationIDsByUserIDAndPermissions(ctx, db, params)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get lowest granted locations by user_id and permissions: %v", err.Error())
		return
	}

	locations, err := s.locationRepo.GetByIDs(ctx, db, locationIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get locations by ids: %v", err.Error())
		return
	}

	locationInfos = make([]*pb.LocationInfo, 0, len(locations))

	for _, location := range locations {
		locationInfos = append(locationInfos, &pb.LocationInfo{
			LocationId:   location.LocationID.String,
			LocationName: location.Name.String,
		})
	}
	return
}

func NewLocationService() *LocationService {
	return &LocationService{
		locationRepo: &repositories.LocationRepo{},
	}
}
