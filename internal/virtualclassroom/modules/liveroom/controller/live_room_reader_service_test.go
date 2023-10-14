package controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	logger_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLiveRoomReaderService_GetLiveRoomState(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomMemberStateRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	request := &vpb.GetLiveRoomStateRequest{
		ChannelId: "channel-id1",
	}

	now := time.Now()
	teacherID := "teacher-id1"
	studentID := "student-id1"
	studentIDs := []string{studentID, "student-id2"}
	mediaID := "media-id1"
	media := media_domain.Media{
		ID:        mediaID,
		Name:      "media-name",
		Resource:  "object",
		Type:      media_domain.MediaTypeVideo,
		CreatedAt: now,
		UpdatedAt: now,
		Comments: []media_domain.Comment{
			{
				Comment:  "hello",
				Duration: int64(12),
			},
			{
				Comment:  "test",
				Duration: int64(25),
			},
		},
		ConvertedImages: []media_domain.ConvertedImage{
			{
				Width:    int32(1920),
				Height:   int32(1080),
				ImageURL: "link",
			},
		},
		FileSizeBytes: int64(1234567),
		Duration:      time.Duration(int64(60)),
	}

	liveRoomState := domain.LiveRoomState{
		CurrentPolling: &vc_domain.CurrentPolling{
			Question: "sample question",
			IsShared: true,
			Options: vc_domain.CurrentPollingOptions{
				&vc_domain.CurrentPollingOption{
					Answer:    "A",
					IsCorrect: false,
					Content:   "sample content",
				},
				&vc_domain.CurrentPollingOption{
					Answer:    "B",
					IsCorrect: true,
					Content:   "sample content",
				},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		Recording: &vc_domain.CompositeRecordingState{
			IsRecording: true,
			Creator:     teacherID,
		},
		WhiteboardZoomState: new(vc_domain.WhiteboardZoomState).SetDefault(),
		CurrentMaterial: &vc_domain.CurrentMaterial{
			MediaID:   mediaID,
			UpdatedAt: now,
			VideoState: &vc_domain.VideoState{
				CurrentTime: vc_domain.Duration(time.Duration(int64(25))),
				PlayerState: vc_domain.PlayerStatePlaying,
			},
		},
		SessionTime: &now,
	}

	liveRoomStateAudio := liveRoomState
	liveRoomStateAudio.CurrentMaterial = &vc_domain.CurrentMaterial{
		MediaID:   mediaID,
		UpdatedAt: now,
		AudioState: &vc_domain.AudioState{
			CurrentTime: vc_domain.Duration(time.Duration(int64(25))),
			PlayerState: vc_domain.PlayerStatePlaying,
		},
	}

	liveRoomMemberStates := domain.LiveRoomMemberStates{
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[0],
			StateType: string(vc_domain.LearnerStateTypeHandsUp),
			BoolValue: true,
		},
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[0],
			StateType: string(vc_domain.LearnerStateTypeAnnotation),
			BoolValue: true,
		},
		&domain.LiveRoomMemberState{
			ChannelID:        request.ChannelId,
			UserID:           studentIDs[0],
			StateType:        string(vc_domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"A"},
		},
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[0],
			StateType: string(vc_domain.LearnerStateTypeChat),
			BoolValue: true,
		},
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[1],
			StateType: string(vc_domain.LearnerStateTypeHandsUp),
			BoolValue: false,
		},
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[1],
			StateType: string(vc_domain.LearnerStateTypeAnnotation),
			BoolValue: true,
		},
		&domain.LiveRoomMemberState{
			ChannelID:        request.ChannelId,
			UserID:           studentIDs[1],
			StateType:        string(vc_domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"B"},
		},
		&domain.LiveRoomMemberState{
			ChannelID: request.ChannelId,
			UserID:    studentIDs[1],
			StateType: string(vc_domain.LearnerStateTypeChat),
			BoolValue: false,
		},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher gets live room state",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(&liveRoomState, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				liveRoomMemberStateRepo.On("GetLiveRoomMemberStatesByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(liveRoomMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, mock.Anything, request.ChannelId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher gets live room state with current material audio",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(&liveRoomStateAudio, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				liveRoomMemberStateRepo.On("GetLiveRoomMemberStatesByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(liveRoomMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, mock.Anything, request.ChannelId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher gets live room state but not found",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(nil, domain.ErrChannelNotFound)

				liveRoomMemberStateRepo.On("GetLiveRoomMemberStatesByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(liveRoomMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, mock.Anything, request.ChannelId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			query := queries.LiveRoomStateQuery{
				LessonmgmtDB:            db,
				LiveRoomStateRepo:       liveRoomStateRepo,
				LiveRoomMemberStateRepo: liveRoomMemberStateRepo,
				MediaModulePort:         mediaModulePort,
			}

			service := &controller.LiveRoomReaderService{
				LiveRoomStateQuery: query,
				LiveRoomLogService: liveRoomLogService,
			}

			response, err := service.GetLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
			}

			mock.AssertExpectationsForObjects(t, db, liveRoomStateRepo, liveRoomMemberStateRepo, mediaModulePort, logRepo)
		})
	}
}

func TestLiveRoomReaderService_GetWhiteboardToken(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	videoTokenSuffix := "samplevideosuffix"
	whiteboardAppID := "app-id"
	whiteboardSvc := new(mock_whiteboard.MockService)
	liveRoomRepo := &mock_liveroom_repo.MockLiveRoomRepo{}

	request := &vpb.GetWhiteboardTokenRequest{
		ChannelName: "channel-name1",
	}

	teacherID := "user-id1"
	channelID := "channel-id1"
	whiteboardToken := "whiteboard-token"
	roomUUID := "sample-room-uuid1"
	now := time.Now()

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetWhiteboardTokenRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher gets whiteboard token with channel not yet existing",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomRepo.On("GetLiveRoomByChannelName", ctx, mock.Anything, request.ChannelName).Once().
					Return(nil, domain.ErrChannelNotFound)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				liveRoomRepo.On("CreateLiveRoom", ctx, mock.Anything, mock.AnythingOfType("string"), request.ChannelName, roomUUID).Once().
					Return(nil)

				liveRoomRepo.On("GetLiveRoomByChannelName", ctx, mock.Anything, request.ChannelName).Once().
					Return(&domain.LiveRoom{
						ChannelID:        channelID,
						ChannelName:      request.ChannelName,
						WhiteboardRoomID: roomUUID,
						EndedAt:          nil,
						CreatedAt:        now,
						UpdatedAt:        now,
						DeletedAt:        nil,
					}, nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
		{
			name:      "teacher gets whiteboard token with channel existing",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomRepo.On("GetLiveRoomByChannelName", ctx, mock.Anything, request.ChannelName).Once().
					Return(&domain.LiveRoom{
						ChannelID:        channelID,
						ChannelName:      request.ChannelName,
						WhiteboardRoomID: roomUUID,
						EndedAt:          nil,
						CreatedAt:        now,
						UpdatedAt:        now,
						DeletedAt:        nil,
					}, nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
		{
			name:      "teacher gets whiteboard token with existing channel but blank whiteboard room ID",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomRepo.On("GetLiveRoomByChannelName", ctx, mock.Anything, request.ChannelName).Once().
					Return(&domain.LiveRoom{
						ChannelID:        channelID,
						ChannelName:      request.ChannelName,
						WhiteboardRoomID: "",
						EndedAt:          nil,
						CreatedAt:        now,
						UpdatedAt:        now,
						DeletedAt:        nil,
					}, nil)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				liveRoomRepo.On("UpdateChannelRoomID", ctx, mock.Anything, channelID, roomUUID).Once().
					Return(nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			command := &commands.LiveRoomCommand{
				LessonmgmtDB:     db,
				VideoTokenSuffix: videoTokenSuffix,
				WhiteboardAppID:  whiteboardAppID,
				WhiteboardSvc:    whiteboardSvc,
				LiveRoomRepo:     liveRoomRepo,
			}

			service := &controller.LiveRoomReaderService{
				LiveRoomCommand: command,
			}

			response, err := service.GetWhiteboardToken(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, channelID, response.ChannelId)
				assert.EqualValues(t, roomUUID, response.RoomId)
				assert.EqualValues(t, whiteboardToken, response.WhiteboardToken)
				assert.EqualValues(t, whiteboardAppID, response.WhiteboardAppId)
			}

			mock.AssertExpectationsForObjects(t, db, liveRoomRepo)
		})
	}
}
