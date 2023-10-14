package infrastructure

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type LiveRoomRepo interface {
	CreateLiveRoom(ctx context.Context, db database.QueryExecer, channelID, channelName, roomUUID string) error
	GetLiveRoomByChannelName(ctx context.Context, db database.QueryExecer, channelName string) (*domain.LiveRoom, error)
	GetLiveRoomByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*domain.LiveRoom, error)
	EndLiveRoom(ctx context.Context, db database.QueryExecer, channelID string, endTime time.Time) error
	UpdateChannelRoomID(ctx context.Context, db database.QueryExecer, channelID, roomUUID string) error
}

type LiveRoomStateRepo interface {
	GetLiveRoomStateByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*domain.LiveRoomState, error)
	UpsertLiveRoomState(ctx context.Context, db database.QueryExecer, channelID string, value interface{}, fieldName string) error
	UpsertLiveRoomCurrentPollingState(ctx context.Context, db database.QueryExecer, channelID string, polling *vc_domain.CurrentPolling) error
	UpsertLiveRoomSpotlightState(ctx context.Context, db database.QueryExecer, channelID, spotlightedUser string) error
	UpsertLiveRoomWhiteboardZoomState(ctx context.Context, db database.QueryExecer, channelID string, wbZoomState *vc_domain.WhiteboardZoomState) error
	UpsertLiveRoomCurrentMaterialState(ctx context.Context, db database.QueryExecer, channelID string, currentMaterial *vc_domain.CurrentMaterial) error
	UpsertLiveRoomSessionTime(ctx context.Context, db database.QueryExecer, channelID string, sessionTime time.Time) error
	UpsertRecordingState(ctx context.Context, db database.QueryExecer, channelID string, recording *vc_domain.CompositeRecordingState) error
	UnSpotlight(ctx context.Context, db database.QueryExecer, channelID string) error
	GetStreamingLearners(ctx context.Context, db database.QueryExecer, channelID string, lockForUpdate bool) ([]string, error)
	IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, channelID, learnerID string, maximumLearnerStreamings int) error
	DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, channelID, learnerID string) error
}

type LiveRoomMemberStateRepo interface {
	GetLiveRoomMemberStatesByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (domain.LiveRoomMemberStates, error)
	GetLiveRoomMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *domain.SearchLiveRoomMemberStateParams) (domain.LiveRoomMemberStates, error)
	BulkUpsertLiveRoomMembersState(ctx context.Context, db database.QueryExecer, channelID string, userIDs []string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error
	UpdateAllLiveRoomMembersState(ctx context.Context, db database.QueryExecer, channelID string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error
	CreateLiveRoomMemberState(ctx context.Context, db database.QueryExecer, channelID, userID string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error
}

type LiveRoomPoll interface {
	CreateLiveRoomPoll(ctx context.Context, db database.QueryExecer, liveRoomPoll *domain.LiveRoomPoll) error
}

type LiveRoomRecordedVideos interface {
	InsertRecordedVideos(ctx context.Context, db database.QueryExecer, videos []*vc_domain.RecordedVideo) error
	GetLiveRoomRecordingsByChannelIDs(ctx context.Context, db database.QueryExecer, channelIDs []string) (vc_domain.RecordedVideos, error)
}

type LiveRoomActivityLogRepo interface {
	CreateLog(ctx context.Context, db database.Ext, channelID, userID, actionType string) error
}
