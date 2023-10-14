package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/multierr"
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
	user := &entity.LegacyUser{}
	if err := multierr.Combine(
		user.ID.Set(currentUserID),
		user.LastLoginDate.Set(req.LastLoginDate.AsTime()),
	); err != nil {
		return nil, fmt.Errorf("failed to set user last login date: %w", err)
	}

	if err := s.UserRepo.UpdateLastLoginDate(ctx, s.DB, user); err != nil {
		return nil, fmt.Errorf("failed to update user last login date: %w", err)
	}

	return &pb.UpdateUserLastLoginDateResponse{
		Successful: true,
	}, nil
}
