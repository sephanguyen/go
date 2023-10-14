package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/repositories"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	userRepo interface {
		GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer, roleNames []string, locationID string) (userIDs []string, err error)
	}
}

func NewUserService() *UserService {
	return &UserService{
		userRepo: &repositories.UserRepo{},
	}
}

func (s *UserService) GetUserIDsForLoaNotification(ctx context.Context, db database.QueryExecer, locationID string) (userIDs []string, err error) {
	userIDs, err = s.userRepo.GetUserIDsByRoleNamesAndLocationID(ctx, db,
		[]string{
			constant.RoleCentreManager,
			constant.RoleCentreStaff,
		}, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get user_ids by role names and location id: %v", err.Error())
	}
	return
}
