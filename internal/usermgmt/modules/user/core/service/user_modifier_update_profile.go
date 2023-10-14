package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// UpdateUserProfile updates a user's profile
func (s *UserModifierService) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	profile := req.Profile
	if profile.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid profile")
	}
	currentUserID := interceptors.UserIDFromContext(ctx)
	if profile.Id != "" && profile.Id != currentUserID {
		return nil, status.Error(codes.PermissionDenied, "user can only update own profile")
	}

	profile.Id = currentUserID
	_, err := s.UserRepo.Get(ctx, s.DB, database.Text(profile.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %s", err.Error()))
	}

	userEnt, err := toUserEntity(req)
	if err != nil {
		return nil, err
	}

	err = s.UserRepo.UpdateProfileV1(ctx, s.DB, userEnt)
	if err != nil {
		return nil, err
	}
	data := &pb.EvtUserInfo{
		UserId: profile.Id,
		Name:   userEnt.GetName(),
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}

	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		ctxzap.Extract(ctx).Error("UpdateUserProfile s.BusFactory.PublishAsync failed", zap.String("msg-id", msgID), zap.Error(err))
	}

	return &pb.UpdateUserProfileResponse{
		Successful: true,
	}, nil
}

func toUserEntity(src *pb.UpdateUserProfileRequest) (*entity.LegacyUser, error) {
	user := new(entity.LegacyUser)
	database.AllNullEntity(user)
	firstName, lastName := SplitNameToFirstNameAndLastName(src.Profile.Name)
	if err := multierr.Combine(
		user.ID.Set(src.Profile.Id),
		user.FullName.Set(src.Profile.Name),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Avatar.Set(src.Profile.Avatar),
		user.Group.Set(src.Profile.Group),
		user.Country.Set(src.Profile.Country.String()),
		user.Email.Set(src.Profile.Email),
		user.DeviceToken.Set(src.Profile.DeviceToken),
		user.IsTester.Set(nil),
		user.UpdatedAt.Set(time.Now()),
	); err != nil {
		return nil, err
	}

	return user, nil
}
