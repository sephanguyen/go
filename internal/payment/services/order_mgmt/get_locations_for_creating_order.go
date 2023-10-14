package ordermgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	locationService "github.com/manabie-com/backend/internal/payment/services/domain_service/location"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type GetLocationsForCreatingOrder struct {
	DB database.Ext

	LocationService ILocationServiceForCreatingOrder
}

type ILocationServiceForCreatingOrder interface {
	GetLowestGrantedLocationsForCreatingOrder(ctx context.Context, db database.Ext, req *pb.GetLocationsForCreatingOrderRequest) (locationInfos []*pb.LocationInfo, err error)
}

func (s *GetLocationsForCreatingOrder) GetLocationsForCreatingOrder(ctx context.Context, req *pb.GetLocationsForCreatingOrderRequest) (res *pb.GetLocationsForCreatingOrderResponse, err error) {
	res = &pb.GetLocationsForCreatingOrderResponse{}
	locationInfos, err := s.LocationService.GetLowestGrantedLocationsForCreatingOrder(ctx, s.DB, req)
	if err != nil {
		return
	}
	res.LocationInfos = locationInfos
	return
}

func NewGetLocationsForCreatingOrder(db database.Ext) *GetLocationsForCreatingOrder {
	return &GetLocationsForCreatingOrder{
		DB:              db,
		LocationService: locationService.NewLocationService(),
	}
}
