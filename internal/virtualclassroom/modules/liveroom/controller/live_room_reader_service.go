package controller

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	vc_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/controller"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LiveRoomReaderService struct {
	LiveRoomCommand    *commands.LiveRoomCommand
	LiveRoomStateQuery queries.LiveRoomStateQuery
	LiveRoomLogService *controller.LiveRoomLogService
}

func (l *LiveRoomReaderService) GetLiveRoomState(ctx context.Context, req *vpb.GetLiveRoomStateRequest) (*vpb.GetLiveRoomStateResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "channel ID can't empty")
	}

	response, err := l.LiveRoomStateQuery.GetLiveRoomState(ctx, req.ChannelId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	if err := l.LiveRoomLogService.LogWhenGetRoomState(ctx, req.ChannelId); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"LiveRoomLogService.LogWhenGetRoomState: could not log this activity",
			zap.String("channel_id", req.ChannelId),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	return toGetLiveRoomStatePb(response), nil
}

func (l *LiveRoomReaderService) GetWhiteboardToken(ctx context.Context, req *vpb.GetWhiteboardTokenRequest) (*vpb.GetWhiteboardTokenResponse, error) {
	if len(strings.TrimSpace(req.ChannelName)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "channel name can't empty")
	}

	response, err := l.LiveRoomCommand.CreateAndGetChannelInfo(ctx, req.ChannelName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.GetWhiteboardTokenResponse{
		ChannelId:       response.ChannelID,
		RoomId:          response.RoomID,
		WhiteboardAppId: response.WhiteboardAppID,
		WhiteboardToken: response.WhiteboardToken,
	}, nil
}

func toGetLiveRoomStatePb(liveRoomState *payloads.GetLiveRoomStateResponse) *vpb.GetLiveRoomStateResponse {
	res := &vpb.GetLiveRoomStateResponse{
		ChannelId:   liveRoomState.ChannelID,
		CurrentTime: timestamppb.New(time.Now()),
	}

	if liveRoomState.LiveRoomState.CurrentMaterial != nil {
		res.CurrentMaterial = vc_controller.ToCurrentMaterialPb(
			liveRoomState.LiveRoomState.CurrentMaterial,
			liveRoomState.Media,
		)
	}

	if len(liveRoomState.UserStates.LearnersState) > 0 {
		res.UsersState = vc_controller.ToUsersStatePb(liveRoomState.UserStates)
	}

	if liveRoomState.LiveRoomState.CurrentPolling != nil {
		res.CurrentPolling = vc_controller.ToCurrentPollingPb(liveRoomState.LiveRoomState.CurrentPolling)
	}

	if liveRoomState.LiveRoomState.Recording != nil {
		res.Recording = vc_controller.ToRecordingPb(liveRoomState.LiveRoomState.Recording)
	}

	if liveRoomState.LiveRoomState.SpotlightedUser != "" {
		res.Spotlight = vc_controller.ToSpotlightedUserPb(liveRoomState.LiveRoomState.SpotlightedUser)
	}

	if liveRoomState.LiveRoomState.WhiteboardZoomState != nil {
		res.WhiteboardZoomState = vc_controller.ToWhiteboardZoomStatePb(liveRoomState.LiveRoomState.WhiteboardZoomState)
	}

	if liveRoomState.LiveRoomState.SessionTime != nil {
		res.SessionTime = timestamppb.New(*liveRoomState.LiveRoomState.SessionTime)
	}

	return res
}
