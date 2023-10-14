package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) UpdateUserDeviceToken(ctx context.Context, req *npb.UpdateUserDeviceTokenRequest) (*npb.UpdateUserDeviceTokenResponse, error) {
	if req.UserId == "" || req.DeviceToken == "" {
		return nil, status.Error(codes.InvalidArgument, "userID or device token empty")
	}

	userDeviceToken := &entities.UserDeviceToken{}
	err := multierr.Combine(
		userDeviceToken.UserID.Set(req.UserId),
		userDeviceToken.DeviceToken.Set(req.DeviceToken),
		userDeviceToken.AllowNotification.Set(req.AllowNotification),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("multierr userDeviceToken %v", err))
	}
	err = svc.UserDeviceTokenRepo.UpsertUserDeviceToken(ctx, svc.DB, userDeviceToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("UpdateUserDeviceToken.UpsertUserDeviceToken: %v", err))
	}

	// users, err := svc.UserRepo.Retrieve(ctx, svc.DB, database.TextArray([]string{req.UserId}))
	findUserFilter := repositories.NewFindUserFilter()
	_ = findUserFilter.UserIDs.Set([]string{req.UserId})
	users, _, err := svc.UserRepo.FindUser(ctx, svc.DB, findUserFilter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("UserRepo.Retrieve: %v", err))
	}

	if len(users) == 0 {
		return nil, status.Errorf(codes.Internal, "UserRepo found no users")
	}

	data := &pb.EvtUserInfo{
		UserId:            req.UserId,
		DeviceToken:       req.DeviceToken,
		AllowNotification: req.AllowNotification,
		Name:              users[0].Name.String,
	}

	msg, err := data.Marshal()
	if err != nil {
		return nil, fmt.Errorf("UpdateUserDeviceToken: %v", err)
	}

	// publish EvtUserInfo to Tom chat
	msgID, err := svc.JSM.PublishAsyncContext(ctx, constants.SubjectUserDeviceTokenUpdated, msg)
	if err != nil {
		ctxzap.Extract(ctx).Error("UpdateUserDeviceToken s.JSM.PublishAsyncContext failed", zap.String("msg-id", msgID), zap.Error(err))
	}

	return &npb.UpdateUserDeviceTokenResponse{
		Successful: true,
	}, nil
}
