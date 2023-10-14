package services

import (
	"context"

	"github.com/manabie-com/backend/internal/discount/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	DB       database.Ext
	UserRepo interface {
		GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer,
			roleNames []string, locationID string) (userIDs []string, err error)
	}
}

func NewUserService() *UserService {
	return &UserService{
		UserRepo: &repositories.UserRepo{},
	}
}

func (s *UserService) GetUserIDsByRoleNamesAndLocationID(ctx context.Context, db database.QueryExecer,
	roleNames []string, locationID string) (userIDs []string, err error) {
	userIDs, err = s.UserRepo.GetUserIDsByRoleNamesAndLocationID(ctx, db,
		roleNames, locationID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Error when get user_ids by role names and location id: %v", err.Error())
	}
	return
}
