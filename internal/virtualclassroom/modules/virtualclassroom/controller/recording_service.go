package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	recording "github.com/manabie-com/backend/internal/golibs/recording"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	lr_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonRecordingService struct {
	Cfg                  configurations.Config
	Logger               *zap.Logger
	RecordingCommand     commands.RecordingCommand
	RecordingQuery       queries.RecordedVideoQuery
	LessonRoomStateQuery queries.LessonRoomStateQuery
	LiveRoomStateQuery   lr_queries.LiveRoomStateQuery
	OrganizationQuery    queries.OrganizationQuery
	FileStore            infrastructure.FileStore
}

func (l *LessonRecordingService) getRecordingConfig() recording.Config {
	ag := l.Cfg.Agora
	return recording.Config{
		AppID:           ag.AppID,
		Cert:            ag.Cert,
		CustomerID:      ag.CustomerID,
		CustomerSecret:  ag.CustomerSecret,
		BucketID:        ag.BucketName,
		BucketAccessKey: ag.BucketAccessKey,
		BucketSecretKey: ag.BucketSecretKey,
		Endpoint:        ag.Endpoint,
		MaxIdleTime:     ag.MaxIdleTime,
	}
}

func (l *LessonRecordingService) getRecordingInformation(ctx context.Context, channelID, lessonID string) (recordingChannel string, recordingRef constant.RecordingReference, recordingState *domain.CompositeRecordingState, err error) {
	channelIDLength := len(strings.TrimSpace(channelID))
	lessonIDLength := len(strings.TrimSpace(lessonID))

	switch {
	case channelIDLength > 0:
		recordingChannel = channelID
		recordingRef = constant.LiveRoomRecordingRef

		state, err := l.LiveRoomStateQuery.GetLiveRoomStateOnlyByChannelID(ctx, recordingChannel)
		if err != nil {
			return "", "", nil, status.Error(codes.Internal, fmt.Sprintf("error on LiveRoomStateQuery.GetLiveRoomStateOnlyByChannelID: %s", err.Error()))
		}
		recordingState = state.Recording
	case channelIDLength == 0 && lessonIDLength > 0:
		recordingChannel = lessonID
		recordingRef = constant.LessonRecordingRef

		state, err := l.LessonRoomStateQuery.GetLessonRoomStateByLessonID(ctx, queries.LessonRoomStateQueryPayload{LessonID: recordingChannel})
		if err != nil {
			return "", "", nil, status.Error(codes.Internal, fmt.Sprintf("error on LessonRoomStateQuery.GetLessonRoomStateByLessonID, lesson %s: %s", recordingChannel, err.Error()))
		}
		recordingState = state.Recording
	default:
		return "", "", nil, status.Error(codes.InvalidArgument, "channel ID or lesson ID should not be empty in the request")
	}

	return recordingChannel, recordingRef, recordingState, nil
}

func (l *LessonRecordingService) getSaveRecordingReference(channelID, lessonID string) (recordingChannel string, recordingRef constant.RecordingReference) {
	channelIDLength := len(strings.TrimSpace(channelID))
	lessonIDLength := len(strings.TrimSpace(lessonID))

	// use lesson ID to save recording as long as its present
	// else use channel ID
	switch {
	case lessonIDLength > 0:
		recordingChannel = lessonID
		recordingRef = constant.LessonRecordingRef
	case lessonIDLength == 0 && channelIDLength > 0:
		recordingChannel = channelID
		recordingRef = constant.LiveRoomRecordingRef
	}

	return
}

func (l *LessonRecordingService) StartRecording(ctx context.Context, req *vpb.StartRecordingRequest) (*vpb.StartRecordingResponse, error) {
	recordingChannel, recordingRef, recordingState, err := l.getRecordingInformation(ctx, req.GetChannelId(), req.GetLessonId())
	if err != nil {
		return nil, err
	}
	if recordingState.IsRecording {
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("the state found for %s %s is already recording", recordingRef, recordingChannel))
	}

	orgMap, err := l.OrganizationQuery.GetOrganizationMap(context.Background())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error in fetching organizations map: %s", err))
	}

	cfg := l.getRecordingConfig()
	rec, err := recording.NewRecorder(ctx, cfg, l.Logger, recordingChannel, orgMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when initialize NewRecorder using %s %s: %s", recordingRef, recordingChannel, err.Error()))
	}
	if _, err := rec.Acquire(); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call Acquire API using %s %s: %s", recordingRef, recordingChannel, err.Error()))
	}

	tc := req.TranscodingConfig
	transcodingConfig := fmt.Sprintf(`
	{
		"height": %d,
		"width": %d,
		"bitrate": %d,
		"fps": %d,
		"mixedVideoLayout": %d,
		"backgroundColor": "%s"
	}`, tc.GetHeight(), tc.GetWidth(), tc.GetBitrate(), tc.GetFps(), tc.GetMixedVideoLayout(), tc.GetBackgroundColor())
	sc := &recording.StartCall{
		SubscribeVideoUids:    req.SubscribeVideoUids,
		SubscribeAudioUids:    req.SubscribeAudioUids,
		FileNamePrefix:        req.FileNamePrefix,
		TranscodingConfigJSON: transcodingConfig,
	}
	if _, err := rec.Start(sc); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call Start Recording API using %s %s: %s", recordingRef, recordingChannel, err.Error()))
	}
	_, err = rec.CallStatusAPI()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call Query Status Recording API using %s %s: %s", recordingRef, recordingChannel, err.Error()))
	}

	user := interceptors.UserIDFromContext(ctx)
	upsertPayload := &commands.UpsertRecordingStatePayload{
		RecordingRef:     recordingRef,
		RecordingChannel: recordingChannel,
		Recording: &domain.CompositeRecordingState{
			ResourceID:  rec.RID,
			SID:         rec.SID,
			UID:         rec.UID,
			IsRecording: true,
			Creator:     user,
		},
	}
	if err := l.RecordingCommand.UpsertRecordingState(ctx, upsertPayload); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on RecordingCommand.UpsertRecordingState: %s", err.Error()))
	}

	return &vpb.StartRecordingResponse{
		UserId: user,
	}, nil
}

func (l *LessonRecordingService) StopRecording(ctx context.Context, req *vpb.StopRecordingRequest) (*vpb.StopRecordingResponse, error) {
	recordingChannel, recordingRef, recordingState, err := l.getRecordingInformation(ctx, req.GetChannelId(), req.GetLessonId())
	if err != nil {
		return nil, err
	}
	if !recordingState.IsRecording {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("the state found for %s %s is not recording", recordingRef, recordingChannel))
	}

	cfg := l.getRecordingConfig()
	rec := recording.GetExistingRecorder(ctx, cfg, l.Logger, recordingState.UID, recordingChannel, recordingState.ResourceID, recordingState.SID)
	videoDetails, err := rec.Stop()
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when call Stop Recording API using %s %s: %s", recordingRef, recordingChannel, err.Error()))
	}

	upsertPayload := &commands.UpsertRecordingStatePayload{
		RecordingRef:     recordingRef,
		RecordingChannel: recordingChannel,
		Recording:        nil,
	}
	if err := l.RecordingCommand.UpsertRecordingState(ctx, upsertPayload); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on RecordingCommand.UpsertRecordingState: %s", err.Error()))
	}

	now := time.Now()
	recordingChannel, recordingRef = l.getSaveRecordingReference(req.GetChannelId(), req.GetLessonId())
	rs, err := domain.ToRecordedVideos(ctx, videoDetails.ServerResponse.FileList, now, recordingChannel, recordingState.Creator, l.Cfg.Agora.BucketName, l.FileStore.GetObjectInfo)
	if err != nil {
		return nil, status.Error(codes.Internal, "error when convert result of agora to Recording Video domain: "+err.Error())
	}

	if err := l.RecordingCommand.NewRecordedVideos(ctx,
		&commands.NewRecordingVideoPayload{
			RecordingRef:     recordingRef,
			RecordingChannel: recordingChannel,
			RecordedVideos:   rs,
		}); err != nil {
		return nil, status.Error(codes.Internal, "error when RecordingCommand.NewRecordedVideos: "+err.Error())
	}

	return &vpb.StopRecordingResponse{}, nil
}

func (l *LessonRecordingService) GetRecordingByLessonID(ctx context.Context, req *vpb.GetRecordingByLessonIDRequest) (*vpb.GetRecordingByLessonIDResponse, error) {
	if req.GetLessonId() == "" {
		return nil, status.Error(codes.Internal, "missing lessonId")
	}
	if req.GetPaging() == nil {
		return nil, status.Error(codes.Internal, "missing paging info")
	}
	result := l.RecordingQuery.RetrieveRecordedVideosByLessonID(ctx, &payloads.RetrieveRecordedVideosByLessonIDPayload{
		LessonID:        req.GetLessonId(),
		Limit:           req.GetPaging().GetLimit(),
		RecordedVideoID: req.GetPaging().GetOffsetString(),
	})
	if result.Err != nil {
		return nil, result.Err
	}
	items := []*vpb.GetRecordingByLessonIDResponse_RecordingItem{}
	for _, v := range result.Recs {
		items = append(items, &vpb.GetRecordingByLessonIDResponse_RecordingItem{
			Id:        v.ID,
			StartTime: timestamppb.New(v.DateTimeRecorded),
			Duration:  durationpb.New(v.Media.Duration),
			FileSize:  float32(v.Media.FileSizeBytes),
		})
	}
	lastItem := ""
	if len(result.Recs) > 0 {
		lastItem = result.Recs[len(result.Recs)-1].ID
	}
	return &vpb.GetRecordingByLessonIDResponse{
		Items: items,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: result.PrePageID,
			},
		},
		TotalItems: result.Total,
	}, nil
}

func (l *LessonRecordingService) GetRecordingDownloadLinkByID(ctx context.Context, req *vpb.GetRecordingDownloadLinkByIDRequest) (*vpb.GetRecordingDownloadLinkByIDResponse, error) {
	if err := l.validateRequestGetRecordingDownloadLinkByID(req); err != nil {
		return nil, err
	}
	record, err := l.RecordingQuery.GetRecordingByID(ctx, &payloads.GetRecordingByIDPayload{
		RecordedVideoID: req.GetRecordingId(),
	})
	if err != nil {
		return nil, err
	}

	url, err := l.FileStore.GenerateGetObjectURL(ctx, record.Media.Resource, req.GetFileName(), req.Expiry.AsDuration())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error when FileStore.GenerateGetObjectURL: %s", err))
	}
	return &vpb.GetRecordingDownloadLinkByIDResponse{
		Url: url.String(),
	}, nil
}

func (l *LessonRecordingService) validateRequestGetRecordingDownloadLinkByID(req *vpb.GetRecordingDownloadLinkByIDRequest) error {
	if req.GetRecordingId() == "" {
		return status.Error(codes.InvalidArgument, "missing recordingId")
	}
	expiry := req.GetExpiry()
	if expiry == nil {
		return status.Error(codes.InvalidArgument, "missing expiry")
	}
	if expiry.Seconds == 0 {
		return status.Error(codes.InvalidArgument, "expiry is not zero")
	}
	if req.GetFileName() == "" {
		return status.Error(codes.InvalidArgument, "missing file name")
	}
	return nil
}
