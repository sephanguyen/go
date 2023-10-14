package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	logger_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	logger_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	vc_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/controller"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLiveRoomModifierService_JoinLiveRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	config := configurations.Config{
		Agora:      configurations.AgoraConfig{},
		Whiteboard: configs.WhiteboardConfig{AppID: "app-id"},
	}
	videoTokenSuffix := "samplevideosuffix"

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	whiteboardSvc := new(mock_whiteboard.MockService)
	agoraTokenSvc := &vc_controller.AgoraTokenService{
		AgoraCfg: config.Agora,
	}
	liveRoomRepo := &mock_liveroom_repo.MockLiveRoomRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	request := &vpb.JoinLiveRoomRequest{
		ChannelName: "channel-name1",
	}
	teacherID := "user-id1"
	studentID := "user-id2"
	channelID := "channel-id1"
	whiteboardToken := "whiteboard-token"
	roomUUID := "sample-room-uuid1"
	logID := "log-id1"
	now := time.Now()

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.JoinLiveRoomRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher join live room and create new channel",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
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

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&logger_repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, teacherID).
					Return(nil).Once()
			},
		},
		{
			name:      "student join live room and create new channel",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
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

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&logger_repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, studentID).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher join live room with existing channel",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
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

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&logger_repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, teacherID).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher join live room and tries to create new channel but got created first",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				liveRoomRepo.On("GetLiveRoomByChannelName", ctx, mock.Anything, request.ChannelName).Once().
					Return(nil, domain.ErrChannelNotFound)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				liveRoomRepo.On("CreateLiveRoom", ctx, mock.Anything, mock.AnythingOfType("string"), request.ChannelName, roomUUID).Once().
					Return(domain.ErrNoChannelCreated)

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

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&logger_repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, teacherID).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher join live room with existing channel but blank whiteboard room ID",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
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

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&logger_repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, teacherID).
					Return(nil).Once()
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
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				VideoTokenSuffix:    videoTokenSuffix,
				WhiteboardSvc:       whiteboardSvc,
				AgoraTokenSvc:       agoraTokenSvc,
				LiveRoomRepo:        liveRoomRepo,
				StudentsRepo:        studentsRepo,
			}

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				JSM:                 jsm,
				Cfg:                 config,
				LiveRoomLogService:  liveRoomLogService,
				LiveRoomCommand:     command,
			}

			tc.req.RtmUserId = tc.reqUserID
			response, err := service.JoinLiveRoom(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, channelID, response.ChannelId)
				assert.EqualValues(t, whiteboardToken, response.WhiteboardToken)
				assert.EqualValues(t, roomUUID, response.RoomId)
				assert.EqualValues(t, config.Whiteboard.AppID, response.WhiteboardAppId)
				assert.NotEmpty(t, response.StmToken)
				assert.NotEmpty(t, response.StreamToken)

				if tc.reqUserID == teacherID {
					assert.NotEmpty(t, response.VideoToken)
					assert.NotEmpty(t, response.ScreenRecordingToken)
				}
			}

			mock.AssertExpectationsForObjects(t, db, liveRoomRepo, studentsRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateAnnotation(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	studentIDs := []string{studentID, "student-id2"}
	teacherID := "teacher-id1"
	stateType := vc_domain.LearnerStateTypeAnnotation
	state := &vc_domain.StateValue{
		BoolValue:        true,
		StringArrayValue: []string{},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher enables annotation in live room",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_AnnotationEnable{
					AnnotationEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					studentIDs,
					stateType,
					state,
				).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to enable annotation in live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_AnnotationEnable{
					AnnotationEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher disables annotation in live room",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_AnnotationDisable{
					AnnotationDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				state.BoolValue = false
				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					studentIDs,
					stateType,
					state,
				).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to disable annotation in live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_AnnotationDisable{
					AnnotationDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher disables annotation for all in live room",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command:   &vpb.ModifyLiveRoomStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					stateType,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, false)
				}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to disable annotation for all in live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command:   &vpb.ModifyLiveRoomStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				LiveRoomLogService:      liveRoomLogService,
				StudentsRepo:            studentsRepo,
				LiveRoomMemberStateRepo: liveRoomMemberRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomMemberRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateChatPermission(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	studentIDs := []string{studentID, "student-id2"}
	teacherID := "teacher-id1"
	stateType := vc_domain.LearnerStateTypeChat
	state := &vc_domain.StateValue{
		BoolValue:        true,
		StringArrayValue: []string{},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher enables chat permission in live room",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_ChatEnable{
					ChatEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					studentIDs,
					stateType,
					state,
				).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to enable chat permission in live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_ChatEnable{
					ChatEnable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher disables chat permission in live room",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_ChatDisable{
					ChatDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				state.BoolValue = false
				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					studentIDs,
					stateType,
					state,
				).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to disable chat permission in live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_ChatDisable{
					ChatDisable: &vpb.ModifyLiveRoomStateRequest_Learners{
						Learners: studentIDs,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				LiveRoomLogService:      liveRoomLogService,
				StudentsRepo:            studentsRepo,
				LiveRoomMemberStateRepo: liveRoomMemberRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomMemberRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_Polling(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	liveRoomPollRepo := &mock_liveroom_repo.MockLiveRoomPollRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	studentIDs := []string{studentID, "student-id2", "student-id3"}
	teacherID := "teacher-id1"

	liveRoomState := &domain.LiveRoomState{
		ChannelID: channelID,
	}

	now := time.Now()
	currentPolling := &vc_domain.CurrentPolling{
		Question:  "Question...?",
		CreatedAt: now,
		UpdatedAt: now,
		EndedAt:   &now,
		StoppedAt: &now,
		Options: vc_domain.CurrentPollingOptions{
			{
				Answer:    "A",
				IsCorrect: true,
				Content:   "content-A",
			},
			{
				Answer:    "B",
				IsCorrect: false,
			},
			{
				Answer:    "C",
				IsCorrect: false,
			},
		},
	}

	liveRoomMemberStates := domain.LiveRoomMemberStates{
		{
			ChannelID:        channelID,
			UserID:           studentID,
			StateType:        string(vc_domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"A", "B"},
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ChannelID:        channelID,
			UserID:           studentIDs[1],
			StateType:        string(vc_domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"A", "C"},
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ChannelID:        channelID,
			UserID:           studentIDs[2],
			StateType:        string(vc_domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"B"},
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	stateValue := &vc_domain.StateValue{
		BoolValue:        false,
		StringArrayValue: []string{},
	}

	// requests
	startPollingReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_StartPolling{
			StartPolling: &vpb.ModifyLiveRoomStateRequest_PollingOptions{
				Options: []*vpb.ModifyLiveRoomStateRequest_PollingOption{
					{
						Answer:    "A",
						IsCorrect: true,
						Content:   "content-A",
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
			},
		},
	}

	stopPollingReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_StopPolling{},
	}

	endPollingReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_EndPolling{},
	}

	submitPollingAnsReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_SubmitPollingAnswer{
			SubmitPollingAnswer: &vpb.ModifyLiveRoomStateRequest_PollingAnswer{
				StringArrayValue: []string{"A"},
			},
		},
	}

	sharePollingReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_SharePolling{
			SharePolling: true,
		},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher starts polling in a live room",
			reqUserID: teacherID,
			req:       startPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomStateRepo.
					On("UpsertLiveRoomCurrentPollingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentPolling)
						assert.Equal(t, vc_domain.CurrentPollingStatusStarted, state.Status)
						assert.False(t, state.CreatedAt.IsZero())
					}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher starts polling in a live room but has existing poll",
			reqUserID: teacherID,
			req:       startPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStarted
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner tries to start polling in a live room",
			reqUserID: studentID,
			req:       startPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher stops polling in a live room",
			reqUserID: teacherID,
			req:       stopPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStarted
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomStateRepo.
					On("UpsertLiveRoomCurrentPollingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentPolling)
						assert.Equal(t, vc_domain.CurrentPollingStatusStopped, state.Status)
						assert.False(t, state.StoppedAt.IsZero())
					}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher stops polling in a live room but poll is already stopped",
			reqUserID: teacherID,
			req:       stopPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStopped
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher stops polling in a live room but poll is empty",
			reqUserID: teacherID,
			req:       stopPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner tries to stop polling in a live room",
			reqUserID: studentID,
			req:       stopPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher ends polling in a live room",
			reqUserID: teacherID,
			req:       endPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStopped
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomMemberRepo.On("GetLiveRoomMemberStatesWithParams", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						params := args.Get(2).(*domain.SearchLiveRoomMemberStateParams)
						assert.Equal(t, params.ChannelID, channelID)
						assert.Equal(t, params.StateType, string(vc_domain.LearnerStateTypePollingAnswer))
					}).
					Return(liveRoomMemberStates, nil).Once()

				liveRoomPollRepo.On("CreateLiveRoomPoll", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						poll := args.Get(2).(*domain.LiveRoomPoll)
						assert.Equal(t, poll.ChannelID, channelID)
						assert.False(t, poll.EndedAt.IsZero())
					}).
					Return(nil).Once()

				liveRoomStateRepo.
					On("UpsertLiveRoomCurrentPollingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentPolling)
						assert.Nil(t, state)
					}).
					Return(nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					studentIDs,
					vc_domain.LearnerStateTypePollingAnswer,
					stateValue,
				).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher ends polling in a live room but polling is empty",
			reqUserID: teacherID,
			req:       endPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher ends polling in a live room but polling is still on started status",
			reqUserID: teacherID,
			req:       endPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStarted
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner tries to end polling in a live room",
			reqUserID: studentID,
			req:       endPollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "learner submits polling answer in live room",
			reqUserID: studentID,
			req:       submitPollingAnsReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStarted
				liveRoomState.CurrentPolling = currentPolling
				userIDs := []string{studentID}

				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomMemberRepo.On("GetLiveRoomMemberStatesWithParams", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						params := args.Get(2).(*domain.SearchLiveRoomMemberStateParams)
						assert.Equal(t, params.ChannelID, channelID)
						assert.Equal(t, params.UserIDs, userIDs)
						assert.Equal(t, params.StateType, string(vc_domain.LearnerStateTypePollingAnswer))
					}).
					Return(domain.LiveRoomMemberStates{}, nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					userIDs,
					vc_domain.LearnerStateTypePollingAnswer,
					mock.Anything,
				).Run(func(args mock.Arguments) {
					stateValue := args.Get(5).(*vc_domain.StateValue)
					assert.Equal(t, stateValue.StringArrayValue, submitPollingAnsReq.GetSubmitPollingAnswer().GetStringArrayValue())
				}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner submits polling answer in live room but polling is already stopped",
			reqUserID: studentID,
			req:       submitPollingAnsReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStopped
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner submits polling answer in live room but polling is empty",
			reqUserID: studentID,
			req:       submitPollingAnsReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher tries to submit polling answer in a live room",
			reqUserID: teacherID,
			req:       submitPollingAnsReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher shares polling in a live room",
			reqUserID: teacherID,
			req:       sharePollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStopped
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomStateRepo.
					On("UpsertLiveRoomCurrentPollingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentPolling)
						assert.True(t, state.IsShared)
					}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher shares polling in a live room but polling is in started status",
			reqUserID: teacherID,
			req:       sharePollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				currentPolling.Status = vc_domain.CurrentPollingStatusStarted
				liveRoomState.CurrentPolling = currentPolling

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher shares polling in a live room but polling is empty",
			reqUserID: teacherID,
			req:       sharePollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner tries to share polling in a live room",
			reqUserID: studentID,
			req:       sharePollingReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				LiveRoomLogService:      liveRoomLogService,
				LiveRoomStateRepo:       liveRoomStateRepo,
				LiveRoomMemberStateRepo: liveRoomMemberRepo,
				LiveRoomPoll:            liveRoomPollRepo,
				StudentsRepo:            studentsRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, liveRoomMemberRepo, liveRoomPollRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateHandsUp(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	studentID2 := "student-id2"
	teacherID := "teacher-id1"
	stateType := vc_domain.LearnerStateTypeHandsUp
	state := &vc_domain.StateValue{
		StringArrayValue: []string{},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "learner raises hand in the live room",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command:   &vpb.ModifyLiveRoomStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				state.BoolValue = true

				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					[]string{studentID},
					stateType,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(5).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, true)
				}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher fold hand of another learner",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_FoldUserHand{
					FoldUserHand: studentID,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				state.BoolValue = false

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("BulkUpsertLiveRoomMembersState", ctx, tx,
					channelID,
					[]string{studentID},
					stateType,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(5).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, false)
				}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to fold hand of another learner",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_FoldUserHand{
					FoldUserHand: studentID2,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher fold all hands of the available learners",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command:   &vpb.ModifyLiveRoomStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				state.BoolValue = false

				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomMemberRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					stateType,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, false)
				}).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to fold all hands of the available learners",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command:   &vpb.ModifyLiveRoomStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				LiveRoomLogService:      liveRoomLogService,
				StudentsRepo:            studentsRepo,
				LiveRoomMemberStateRepo: liveRoomMemberRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomMemberRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateSpotlight(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	teacherID := "teacher-id1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher spotlights a user",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_Spotlight_{
					Spotlight: &vpb.ModifyLiveRoomStateRequest_Spotlight{
						UserId:      studentID,
						IsSpotlight: true,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomSpotlightState", ctx, tx, channelID, studentID).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to spotlight a user",
			reqUserID: studentID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_Spotlight_{
					Spotlight: &vpb.ModifyLiveRoomStateRequest_Spotlight{
						UserId:      studentID,
						IsSpotlight: true,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher removes spotlight",
			reqUserID: teacherID,
			req: &vpb.ModifyLiveRoomStateRequest{
				ChannelId: channelID,
				Command: &vpb.ModifyLiveRoomStateRequest_Spotlight_{
					Spotlight: &vpb.ModifyLiveRoomStateRequest_Spotlight{
						IsSpotlight: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UnSpotlight", ctx, tx, channelID).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				LiveRoomLogService:  liveRoomLogService,
				StudentsRepo:        studentsRepo,
				LiveRoomStateRepo:   liveRoomStateRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateWhiteboardZoomState(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	teacherID := "teacher-id1"

	wbZoomStateReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_WhiteboardZoomState_{
			WhiteboardZoomState: &vpb.ModifyLiveRoomStateRequest_WhiteboardZoomState{
				PdfScaleRatio: 80.0,
				PdfWidth:      1440.0,
				PdfHeight:     2160.0,
				CenterX:       2.0,
				CenterY:       3.0,
			},
		},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher modifies the whiteboard zoom state",
			reqUserID: teacherID,
			req:       wbZoomStateReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomWhiteboardZoomState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.WhiteboardZoomState)
						req := wbZoomStateReq.GetWhiteboardZoomState()
						assert.Equal(t, state.PdfScaleRatio, req.PdfScaleRatio)
						assert.Equal(t, state.PdfWidth, req.PdfWidth)
						assert.Equal(t, state.PdfHeight, req.PdfHeight)
						assert.Equal(t, state.CenterX, req.CenterX)
						assert.Equal(t, state.CenterY, req.CenterY)
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner tries to modify the whiteboard zoom state",
			reqUserID: studentID,
			req:       wbZoomStateReq,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				LiveRoomLogService:  liveRoomLogService,
				StudentsRepo:        studentsRepo,
				LiveRoomStateRepo:   liveRoomStateRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpdateCurrentMaterial(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomMemberStateRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	studentID := "student-id1"
	teacherID := "teacher-id1"
	mediaID := "media-id1"

	currentMaterialVideoReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: mediaID,
				State: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(1 * time.Second),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PAUSE,
					},
				},
			},
		},
	}

	currentMaterialAudioReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: mediaID,
				State: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_AudioState{
					AudioState: &vpb.VirtualClassroomState_CurrentMaterial_AudioState{
						CurrentTime: durationpb.New(2 * time.Second),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	currentMaterialPdfReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: mediaID,
			},
		},
	}

	stopShareMaterialReq := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command:   &vpb.ModifyLiveRoomStateRequest_StopSharingMaterial{},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher shares a video material",
			reqUserID: teacherID,
			req:       currentMaterialVideoReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Equal(t, mediaID, state.MediaID)
						assert.Equal(t, vc_domain.Duration(1*time.Second), state.VideoState.CurrentTime)
						assert.Equal(t, vc_domain.PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher shares an audio material",
			reqUserID: teacherID,
			req:       currentMaterialAudioReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Equal(t, mediaID, state.MediaID)
						assert.Equal(t, vc_domain.Duration(2*time.Second), state.AudioState.CurrentTime)
						assert.Equal(t, vc_domain.PlayerStatePlaying, state.AudioState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher shares a pdf material",
			reqUserID: teacherID,
			req:       currentMaterialPdfReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Equal(t, mediaID, state.MediaID)
						assert.Nil(t, state.AudioState)
						assert.Nil(t, state.VideoState)
						assert.False(t, state.UpdatedAt.IsZero())
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner shares a video material",
			reqUserID: studentID,
			req:       currentMaterialVideoReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Equal(t, mediaID, state.MediaID)
						assert.Equal(t, vc_domain.Duration(1*time.Second), state.VideoState.CurrentTime)
						assert.Equal(t, vc_domain.PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher stops sharing material",
			reqUserID: teacherID,
			req:       stopShareMaterialReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Nil(t, state)
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
		{
			name:      "learner stops sharing material",
			reqUserID: studentID,
			req:       stopShareMaterialReq,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Nil(t, state)
					}).Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				LiveRoomLogService:      liveRoomLogService,
				StudentsRepo:            studentsRepo,
				LiveRoomStateRepo:       liveRoomStateRepo,
				LiveRoomMemberStateRepo: liveRoomMemberStateRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, liveRoomMemberStateRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_ModifyLiveRoomState_UpsertSessionTime(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	channelID := "channel-id1"
	teacherID := "teacher-id1"

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: channelID,
		Command: &vpb.ModifyLiveRoomStateRequest_UpsertSessionTime{
			UpsertSessionTime: true,
		},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyLiveRoomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher upserts session time",
			reqUserID: teacherID,
			req:       req,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				liveRoomStateRepo.On("UpsertLiveRoomSessionTime", ctx, tx, channelID, mock.Anything).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID, logger_repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				LiveRoomLogService:  liveRoomLogService,
				StudentsRepo:        studentsRepo,
				LiveRoomStateRepo:   liveRoomStateRepo,
			}

			_, err := service.ModifyLiveRoomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, studentsRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_LeaveLiveRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	liveRoomRepo := &mock_liveroom_repo.MockLiveRoomRepo{}

	teacherID := "user-id1"
	channelID := "channel-id1"
	channelName := "channel name"

	request := &vpb.LeaveLiveRoomRequest{
		ChannelId: channelID,
		UserId:    teacherID,
	}

	liveRoom := &domain.LiveRoom{
		ChannelID:   channelID,
		ChannelName: channelName,
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.LeaveLiveRoomRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher leave live room",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				liveRoomRepo.On("GetLiveRoomByChannelID", ctx, mock.Anything, request.ChannelId).Once().
					Return(liveRoom, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Return("", nil)
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
				LessonmgmtDB: db,
				LiveRoomRepo: liveRoomRepo,
			}

			service := &controller.LiveRoomModifierService{
				JSM:                jsm,
				LiveRoomCommand:    &commands.LiveRoomCommand{},
				LiveRoomStateQuery: query,
			}

			_, err := service.LeaveLiveRoom(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, liveRoomRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_EndLiveRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveRoomRepo := &mock_liveroom_repo.MockLiveRoomRepo{}
	liveRoomMemberStateRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	studentsRepo := &mock_virtual_repo.MockStudentsRepo{}
	lessonRepo := &mock_virtual_repo.MockVirtualLessonRepo{}

	logRepo := &mock_liveroom_repo.MockLiveRoomLogRepo{}
	liveRoomLogService := &logger_controller.LiveRoomLogService{
		DB:              db,
		LiveRoomLogRepo: logRepo,
	}

	teacherID := "user-id1"
	channelID := "channel-id1"
	channelName := "channel name"

	request := &vpb.EndLiveRoomRequest{
		ChannelId: channelID,
		LessonId:  "lesson-id1",
	}

	liveRoom := &domain.LiveRoom{
		ChannelID:   channelID,
		ChannelName: channelName,
	}
	state := &vc_domain.StateValue{
		BoolValue:        false,
		StringArrayValue: []string{},
	}
	stateTrue := &vc_domain.StateValue{
		BoolValue:        true,
		StringArrayValue: []string{},
	}

	now := time.Now()
	currentPolling := &vc_domain.CurrentPolling{
		Question:  "Question...?",
		CreatedAt: now,
		UpdatedAt: now,
		EndedAt:   &now,
		StoppedAt: &now,
		Options: vc_domain.CurrentPollingOptions{
			{
				Answer:    "A",
				IsCorrect: true,
				Content:   "content-A",
			},
			{
				Answer:    "B",
				IsCorrect: false,
			},
			{
				Answer:    "C",
				IsCorrect: false,
			},
		},
	}

	liveRoomState := &domain.LiveRoomState{
		ChannelID:      channelID,
		CurrentPolling: currentPolling,
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.EndLiveRoomRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher end live room successfully with lesson ID",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomRepo.On("GetLiveRoomByChannelID", ctx, mock.Anything, channelID).Once().
					Return(liveRoom, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				// clear share material
				liveRoomStateRepo.On("UpsertLiveRoomCurrentMaterialState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentMaterial)
						assert.Nil(t, state)
					}).Return(nil).Once()

				// reset annotation
				liveRoomMemberStateRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					vc_domain.LearnerStateTypeAnnotation,
					stateTrue,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, true)
				}).Return(nil).Once()

				// reset hands
				liveRoomMemberStateRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					vc_domain.LearnerStateTypeHandsUp,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, false)
				}).Return(nil).Once()

				// reset polling
				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, tx, channelID).
					Return(liveRoomState, nil).Once()

				liveRoomStateRepo.
					On("UpsertLiveRoomCurrentPollingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CurrentPolling)
						assert.Nil(t, state)
					}).
					Return(nil).Once()

				liveRoomMemberStateRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					vc_domain.LearnerStateTypePollingAnswer,
					state,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Empty(t, state.StringArrayValue)
				}).Return(nil).Once()

				// reset whiteboard zoom state
				liveRoomStateRepo.On("UpsertLiveRoomWhiteboardZoomState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.WhiteboardZoomState)
						req := new(vc_domain.WhiteboardZoomState).SetDefault()
						assert.Equal(t, state.PdfScaleRatio, req.PdfScaleRatio)
						assert.Equal(t, state.PdfWidth, req.PdfWidth)
						assert.Equal(t, state.PdfHeight, req.PdfHeight)
						assert.Equal(t, state.CenterX, req.CenterX)
						assert.Equal(t, state.CenterY, req.CenterY)
					}).Return(nil).Once()

				// reset spotlight
				liveRoomStateRepo.On("UnSpotlight", ctx, tx, channelID).
					Return(nil).Once()

				// reset chat
				liveRoomMemberStateRepo.On("UpdateAllLiveRoomMembersState", ctx, tx,
					channelID,
					vc_domain.LearnerStateTypeChat,
					stateTrue,
				).Run(func(args mock.Arguments) {
					state := args.Get(4).(*vc_domain.StateValue)
					assert.Equal(t, state.BoolValue, true)
				}).Return(nil).Once()

				// clear recording state
				liveRoomStateRepo.On("UpsertRecordingState", ctx, tx, channelID, mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*vc_domain.CompositeRecordingState)
						assert.Nil(t, state)
					}).Return(nil).Once()

				// end live room
				liveRoomRepo.On("EndLiveRoom", ctx, db, channelID, mock.Anything).
					Return(nil).Once()

				// end live lesson
				lessonRepo.On("EndLiveLesson", ctx, db, "lesson-id1", mock.Anything).
					Return(nil).Once()

				logRepo.On("CompleteLogByChannelID", ctx, db, channelID).
					Return(nil).Once()

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)
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
				LessonmgmtDB:        db,
				WrapperDBConnection: wrapperConnection,
				LiveRoomRepo:        liveRoomRepo,
				StudentsRepo:        studentsRepo,
				LessonRepo:          lessonRepo,
			}

			query := queries.LiveRoomStateQuery{
				LessonmgmtDB: db,
				LiveRoomRepo: liveRoomRepo,
			}

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:            db,
				WrapperDBConnection:     wrapperConnection,
				JSM:                     jsm,
				LiveRoomLogService:      liveRoomLogService,
				LiveRoomCommand:         command,
				LiveRoomStateQuery:      query,
				LiveRoomStateRepo:       liveRoomStateRepo,
				LiveRoomMemberStateRepo: liveRoomMemberStateRepo,
				StudentsRepo:            studentsRepo,
			}

			_, err := service.EndLiveRoom(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, liveRoomMemberStateRepo, studentsRepo, logRepo, lessonRepo, mockUnleashClient)
		})
	}
}

func TestLiveRoomModifierService_PreparePublishLiveRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	maximumLearnerStreamings := 5

	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomActivityLogRepo := &mock_liveroom_repo.MockLiveRoomActivityLogRepo{}

	teacherID := "user-id1"
	learnerID1 := "learner-id1"
	learnerID2 := "learner-id2"
	learnerID3 := "learner-id3"
	channelID := "channel-id1"
	learnerIDs := []string{learnerID2, learnerID3}

	request := &vpb.PreparePublishLiveRoomRequest{
		ChannelId: channelID,
	}

	tcs := []struct {
		name           string
		reqUserID      string
		req            *vpb.PreparePublishLiveRoomRequest
		expectedStatus vpb.PrepareToPublishStatus
		setup          func(ctx context.Context)
		hasError       bool
	}{
		{
			name:           "user prepare publish successfully",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(learnerIDs, nil)

				liveRoomStateRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1, maximumLearnerStreamings).Once().
					Return(nil)

				liveRoomActivityLogRepo.On("CreateLog", ctx, mock.Anything, channelID, learnerID1, constant.LogActionTypePublish).Once().
					Return(nil)
			},
		},
		{
			name:           "user prepare publish successfully but live room state not existing",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(nil, domain.ErrChannelNotFound)

				liveRoomStateRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1, maximumLearnerStreamings).Once().
					Return(nil)

				liveRoomActivityLogRepo.On("CreateLog", ctx, mock.Anything, channelID, learnerID1, constant.LogActionTypePublish).Once().
					Return(nil)
			},
		},
		{
			name:           "user gets prepared before status",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				request.LearnerId = learnerID2

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(learnerIDs, nil)
			},
		},
		{
			name:           "user gets reached maximum learner streamings limit status",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(learnerIDs, nil)

				liveRoomStateRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1, maximumLearnerStreamings).Once().
					Return(domain.ErrNoChannelUpdated)
			},
		},
		{
			name:           "failed to get streaming learners",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(nil, errors.New("err"))
			},
			hasError: true,
		},
		{
			name:           "failed to increase streaming learners",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(learnerIDs, nil)

				liveRoomStateRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1, maximumLearnerStreamings).Once().
					Return(errors.New("err"))
			},
			hasError: true,
		},
		{
			name:           "failed to create live room activity log",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				request.LearnerId = learnerID1

				liveRoomStateRepo.On("GetStreamingLearners", ctx, mock.Anything, channelID, true).Once().
					Return(learnerIDs, nil)

				liveRoomStateRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1, maximumLearnerStreamings).Once().
					Return(nil)

				liveRoomActivityLogRepo.On("CreateLog", ctx, mock.Anything, channelID, learnerID1, constant.LogActionTypePublish).Once().
					Return(errors.New("err"))
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			command := &commands.LiveRoomCommand{
				LessonmgmtDB:             db,
				MaximumLearnerStreamings: maximumLearnerStreamings,
				LiveRoomStateRepo:        liveRoomStateRepo,
				LiveRoomActivityLogRepo:  liveRoomActivityLogRepo,
			}

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:    db,
				LiveRoomCommand: command,
			}

			res, err := service.PreparePublishLiveRoom(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedStatus, res.Status)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, liveRoomActivityLogRepo)
		})
	}
}

func TestLiveRoomModifierService_UnpublishLiveRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	liveRoomActivityLogRepo := &mock_liveroom_repo.MockLiveRoomActivityLogRepo{}

	teacherID := "user-id1"
	learnerID1 := "learner-id1"
	channelID := "channel-id1"

	request := &vpb.UnpublishLiveRoomRequest{
		ChannelId: channelID,
		LearnerId: learnerID1,
	}

	tcs := []struct {
		name           string
		reqUserID      string
		req            *vpb.UnpublishLiveRoomRequest
		expectedStatus vpb.UnpublishStatus
		setup          func(ctx context.Context)
		hasError       bool
	}{
		{
			name:           "user unpublish successfully",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				liveRoomStateRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1).Once().
					Return(nil)

				liveRoomActivityLogRepo.On("CreateLog", ctx, mock.Anything, channelID, learnerID1, constant.LogActionTypeUnpublish).Once().
					Return(nil)
			},
		},
		{
			name:           "user unpublish but get unpublish before",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1).Once().
					Return(domain.ErrNoChannelUpdated)
			},
		},
		{
			name:           "user unpublish but get error",
			reqUserID:      teacherID,
			req:            request,
			expectedStatus: vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_NONE,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				liveRoomStateRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, channelID, learnerID1).Once().
					Return(errors.New("error"))
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			command := &commands.LiveRoomCommand{
				LessonmgmtDB:            db,
				LiveRoomStateRepo:       liveRoomStateRepo,
				LiveRoomActivityLogRepo: liveRoomActivityLogRepo,
			}

			service := &controller.LiveRoomModifierService{
				LessonmgmtDB:    db,
				LiveRoomCommand: command,
			}

			res, err := service.UnpublishLiveRoom(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedStatus, res.Status)
			}

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomStateRepo, liveRoomActivityLogRepo)
		})
	}
}
