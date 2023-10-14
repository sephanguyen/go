package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LocationController struct {
	LocationService interface {
		GetListGrantedLocationOfStaff(ctx context.Context, staffID, name string, limit int32) ([]*dto.Location, error)
	}
}

func (c *LocationController) GetGrantedLocationsOfStaff(ctx context.Context, request *pb.GetGrantedLocationsOfStaffRequest) (*pb.GetGrantedLocationsOfStaffResponse, error) {
	err := dto.ValidateGetGrantedLocation(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	locationsDto, err := c.LocationService.GetListGrantedLocationOfStaff(ctx, request.StaffId, request.GetName(), request.GetLimit())

	if err != nil {
		return nil, err
	}

	return dto.NewListLocationsToRPCResponse(locationsDto), nil
}
