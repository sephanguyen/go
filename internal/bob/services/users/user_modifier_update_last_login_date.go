package users

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateUserLastLoginDate updates user last login date
func (s *UserModifierService) UpdateUserLastLoginDate(ctx context.Context, req *pb.UpdateUserLastLoginDateRequest) (*pb.UpdateUserLastLoginDateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid last login date request")
	}
	if req.LastLoginDate == nil || req.LastLoginDate.AsTime().IsZero() {
		return nil, status.Error(codes.InvalidArgument, "invalid last login date value")
	}

	currentUserID := interceptors.UserIDFromContext(ctx)
	e := &entities.User{}
	_ = e.ID.Set(currentUserID)
	_ = e.LastLoginDate.Set(req.LastLoginDate.AsTime())

	err := s.UserRepo.UpdateLastLoginDate(ctx, s.DB, e)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserLastLoginDateResponse{
		Successful: true,
	}, nil
}
