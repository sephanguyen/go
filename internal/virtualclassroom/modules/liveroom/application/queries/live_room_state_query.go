package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vc_infrastructure "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type LiveRoomStateQuery struct {
	LessonmgmtDB database.Ext

	LiveRoomRepo            infrastructure.LiveRoomRepo
	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
	MediaModulePort         vc_infrastructure.MediaModulePort
}

func (l *LiveRoomStateQuery) GetLiveRoomState(ctx context.Context, channelID string) (*payloads.GetLiveRoomStateResponse, error) {
	// live room state
	liveRoomState, err := l.LiveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, l.LessonmgmtDB, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return nil, fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}
	if err == domain.ErrChannelNotFound {
		liveRoomState = &domain.LiveRoomState{
			SpotlightedUser:     "",
			Recording:           &vc_domain.CompositeRecordingState{},
			WhiteboardZoomState: new(vc_domain.WhiteboardZoomState).SetDefault(),
		}
	}

	// live room state - material
	var media *media_domain.Media
	if liveRoomState.CurrentMaterial != nil && liveRoomState.CurrentMaterial.MediaID != "" {
		mediaID := liveRoomState.CurrentMaterial.MediaID
		medias, err := l.MediaModulePort.RetrieveMediasByIDs(ctx, []string{mediaID})
		if err != nil {
			return nil, fmt.Errorf("error in MediaModulePort.RetrieveMediasByIDs, media %s: %w", mediaID, err)
		}
		if len(medias) == 0 {
			return nil, fmt.Errorf("media %s is not found but is part of the current material in channel %s", mediaID, channelID)
		}
		media = medias[0]
	}

	// live room member state
	liveRoomMemberStates, err := l.LiveRoomMemberStateRepo.GetLiveRoomMemberStatesByChannelID(ctx, l.LessonmgmtDB, channelID)
	if err != nil {
		return nil, fmt.Errorf("error in LiveRoomMemberStateRepo.GetLiveRoomMemberStateByChannelID, channel %s: %w", channelID, err)
	}
	if err := liveRoomMemberStates.ValidInChannel(channelID); err != nil {
		return nil, err
	}
	userStates := liveRoomMemberStates.ConvertToUserState()

	return &payloads.GetLiveRoomStateResponse{
		ChannelID:     channelID,
		LiveRoomState: liveRoomState,
		Media:         media,
		UserStates:    userStates,
	}, nil
}

func (l *LiveRoomStateQuery) GetLiveRoom(ctx context.Context, channelID string) (*domain.LiveRoom, error) {
	liveRoom, err := l.LiveRoomRepo.GetLiveRoomByChannelID(ctx, l.LessonmgmtDB, channelID)
	if err != nil {
		return nil, fmt.Errorf("error in LiveRoomRepo.GetLiveRoomByChannelID, channel %s: %w", channelID, err)
	}

	return liveRoom, nil
}

func (l *LiveRoomStateQuery) GetLiveRoomStateOnlyByChannelID(ctx context.Context, channelID string) (*domain.LiveRoomState, error) {
	liveRoomState, err := l.LiveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, l.LessonmgmtDB, channelID)

	if err == domain.ErrChannelNotFound {
		liveRoomState = &domain.LiveRoomState{
			SpotlightedUser:     "",
			Recording:           &vc_domain.CompositeRecordingState{},
			WhiteboardZoomState: new(vc_domain.WhiteboardZoomState).SetDefault(),
		}
		return liveRoomState, nil
	}

	return liveRoomState, err
}

func (l *LiveRoomStateQuery) GetLiveRoomStateOnlyByChannelIDWithoutCheck(ctx context.Context, channelID string) (*domain.LiveRoomState, error) {
	liveRoomState, err := l.LiveRoomStateRepo.GetLiveRoomStateByChannelID(ctx, l.LessonmgmtDB, channelID)
	return liveRoomState, err
}
