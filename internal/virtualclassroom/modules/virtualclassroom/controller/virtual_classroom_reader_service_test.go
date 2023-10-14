package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	logger_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetrieveWhiteboardToken(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	config := configurations.Config{
		Whiteboard: configs.WhiteboardConfig{AppID: "app-id"},
	}
	whiteboardSvc := new(mock_whiteboard.MockService)
	agoraTokenSvc := &controller.AgoraTokenService{
		AgoraCfg: config.Agora,
	}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}

	request := &vpb.RetrieveWhiteboardTokenRequest{
		LessonId: "lesson-id1",
	}
	teacherID := "user-id1"
	studentID := "user-id2"
	courseID := "course-id1"
	whiteboardToken := "whiteboard-token"
	roomUUID := "sample-room-uuid1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.RetrieveWhiteboardTokenRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher retrieve whiteboard token and create new room id",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, request.LessonId, roomUUID).Once().
					Return(nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
		{
			name:      "teacher retrieve whiteboard token but existing room ID",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   roomUUID,
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
		{
			name:      "student retrieve whiteboard token and create new room id",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				lessonMemberRepo.On("GetCourseAccessible", ctx, mock.Anything, studentID).Once().
					Return([]string{courseID}, nil)

				lessonRepo.On("GetVirtualLessonByLessonIDsAndCourseIDs", ctx, mock.Anything, []string{request.LessonId}, []string{courseID}).Once().
					Return([]*domain.VirtualLesson{{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}}, nil)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, request.LessonId, roomUUID).Once().
					Return(nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)
			},
		},
		{
			name:      "student retrieve whiteboard token but existing room ID",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				lessonMemberRepo.On("GetCourseAccessible", ctx, mock.Anything, studentID).Once().
					Return([]string{courseID}, nil)

				lessonRepo.On("GetVirtualLessonByLessonIDsAndCourseIDs", ctx, mock.Anything, []string{request.LessonId}, []string{courseID}).Once().
					Return([]*domain.VirtualLesson{{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}}, nil)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, request.LessonId, roomUUID).Once().
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

			command := &commands.LiveLessonCommand{
				WrapperDBConnection: wrapperConnection,
				WhiteboardSvc:       whiteboardSvc,
				AgoraTokenSvc:       agoraTokenSvc,
				VirtualLessonRepo:   lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				StudentsRepo:        studentsRepo,
				CourseRepo:          courseRepo,
			}

			service := &controller.VirtualClassroomReaderService{
				Cfg:               config,
				LiveLessonCommand: command,
			}

			response, err := service.RetrieveWhiteboardToken(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, whiteboardToken, response.WhiteboardToken)
				assert.EqualValues(t, roomUUID, response.RoomId)
				assert.EqualValues(t, config.Whiteboard.AppID, response.WhiteboardAppId)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, studentsRepo, courseRepo, mockUnleashClient)
		})
	}
}

func TestGetLiveLessonState(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	lessonRoomStateRepo := &mock_repositories.MockLessonRoomStateRepo{}

	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	virtualClassroomLogService := &logger_controller.VirtualClassRoomLogService{
		WrapperConnection: wrapperConnection,
		Repo:              logRepo,
	}

	request := &vpb.GetLiveLessonStateRequest{
		LessonId: "lesson-id1",
	}

	now := time.Now()
	teacherID := "teacher-id1"
	teacherIDs := []string{teacherID, "teacher-id2"}
	teacherIDExcluded := "teacher-id3"
	studentID := "student-id1"
	studentIDs := []string{studentID, "student-id2"}
	studentIDExcluded := "student-id3"
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

	virtualLesson := domain.VirtualLesson{
		LessonID: request.LessonId,
		RoomID:   "",
		TeacherIDs: domain.TeacherIDs{
			TeacherIDs: teacherIDs,
		},
		LearnerIDs: domain.LearnerIDs{
			LearnerIDs: studentIDs,
		},
	}

	lessonRoomState := domain.LessonRoomState{
		CurrentPolling: &domain.CurrentPolling{
			Question: "sample question",
			IsShared: true,
			Options: domain.CurrentPollingOptions{
				&domain.CurrentPollingOption{
					Answer:    "A",
					IsCorrect: false,
					Content:   "sample content",
				},
				&domain.CurrentPollingOption{
					Answer:    "B",
					IsCorrect: true,
					Content:   "sample content",
				},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		Recording: &domain.CompositeRecordingState{
			IsRecording: true,
			Creator:     teacherID,
		},
		WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
		CurrentMaterial: &domain.CurrentMaterial{
			MediaID:   mediaID,
			UpdatedAt: now,
			VideoState: &domain.VideoState{
				CurrentTime: domain.Duration(time.Duration(int64(25))),
				PlayerState: domain.PlayerStatePlaying,
			},
		},
		SessionTime: &now,
	}

	lessonRoomStateAudio := lessonRoomState
	lessonRoomStateAudio.CurrentMaterial = &domain.CurrentMaterial{
		MediaID:   mediaID,
		UpdatedAt: now,
		AudioState: &domain.AudioState{
			CurrentTime: domain.Duration(time.Duration(int64(25))),
			PlayerState: domain.PlayerStatePlaying,
		},
	}

	lessonMemberStates := domain.LessonMemberStates{
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[0],
			StateType: string(domain.LearnerStateTypeHandsUp),
			BoolValue: true,
		},
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[0],
			StateType: string(domain.LearnerStateTypeAnnotation),
			BoolValue: true,
		},
		&domain.LessonMemberState{
			LessonID:         request.LessonId,
			UserID:           studentIDs[0],
			StateType:        string(domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"A"},
		},
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[0],
			StateType: string(domain.LearnerStateTypeChat),
			BoolValue: true,
		},
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[1],
			StateType: string(domain.LearnerStateTypeHandsUp),
			BoolValue: false,
		},
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[1],
			StateType: string(domain.LearnerStateTypeAnnotation),
			BoolValue: true,
		},
		&domain.LessonMemberState{
			LessonID:         request.LessonId,
			UserID:           studentIDs[1],
			StateType:        string(domain.LearnerStateTypePollingAnswer),
			StringArrayValue: []string{"B"},
		},
		&domain.LessonMemberState{
			LessonID:  request.LessonId,
			UserID:    studentIDs[1],
			StateType: string(domain.LearnerStateTypeChat),
			BoolValue: false,
		},
	}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetLiveLessonStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher gets live lesson state",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&virtualLesson, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(&lessonRoomState, nil)

				lessonMemberRepo.On("GetLessonMemberStatesByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(lessonMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByLessonID", ctx, mock.Anything, request.LessonId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher gets live lesson state with current material audio",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&virtualLesson, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(&lessonRoomStateAudio, nil)

				lessonMemberRepo.On("GetLessonMemberStatesByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(lessonMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByLessonID", ctx, mock.Anything, request.LessonId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher not part of lesson gets live lesson state",
			reqUserID: teacherIDExcluded,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&virtualLesson, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherIDExcluded).Once().
					Return(false, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(&lessonRoomState, nil)

				lessonMemberRepo.On("GetLessonMemberStatesByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(lessonMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByLessonID", ctx, mock.Anything, request.LessonId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "student gets live lesson state",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&virtualLesson, nil)

				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).Once().
					Return(media_domain.Medias{&media}, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(&lessonRoomState, nil)

				lessonMemberRepo.On("GetLessonMemberStatesByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(lessonMemberStates, nil)

				logRepo.On("IncreaseTotalTimesByLessonID", ctx, mock.Anything, request.LessonId, repo.TotalTimesGettingRoomState).Once().
					Return(nil)
			},
		},
		{
			name:      "student not part of lesson unable to get live lesson state",
			reqUserID: studentIDExcluded,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&virtualLesson, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentIDExcluded).Once().
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

			query := queries.LessonRoomStateQuery{
				WrapperDBConnection: wrapperConnection,
				VirtualLessonRepo:   lessonRepo,
				LessonRoomStateRepo: lessonRoomStateRepo,
				LessonMemberRepo:    lessonMemberRepo,
				MediaModulePort:     mediaModulePort,
				StudentsRepo:        studentsRepo,
			}

			service := &controller.VirtualClassroomReaderService{
				LessonRoomStateQuery:       query,
				VirtualClassRoomLogService: virtualClassroomLogService,
			}

			response, err := service.GetLiveLessonState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, lessonMemberRepo, lessonRoomStateRepo, studentsRepo, mediaModulePort, logRepo, mockUnleashClient)
		})
	}
}

func TestGetUserInformation(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userBasicInfoRepo := &mock_repositories.MockUserBasicInfoRepo{}
	userIDs := []string{"user-id1", "user-id2", "user-id3"}
	userInfos := []domain.UserBasicInfo{
		{
			UserID: "user-id1",
			Name:   "user name 1",
		},
		{
			UserID: "user-id2",
			Name:   "user name 2",
		},
		{
			UserID: "user-id3",
			Name:   "user name 3",
		},
	}

	request := &vpb.GetUserInformationRequest{
		UserIds: userIDs,
	}

	tcs := []struct {
		name     string
		req      *vpb.GetUserInformationRequest
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "user gets user information",
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetUserInfosByIDs", ctx, db, userIDs).
					Return(userInfos, nil).Once()
			},
		},
		{
			name: "user gets error",
			req:  request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetUserInfosByIDs", ctx, db, userIDs).
					Return(nil, errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			query := queries.UserInfoQuery{
				WrapperDBConnection: wrapperConnection,
				UserBasicInfoRepo:   userBasicInfoRepo,
			}

			service := &controller.VirtualClassroomReaderService{
				UserInfoQuery: query,
			}

			res, err := service.GetUserInformation(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
			}

			mock.AssertExpectationsForObjects(t, db, userBasicInfoRepo, mockUnleashClient)
		})
	}
}
