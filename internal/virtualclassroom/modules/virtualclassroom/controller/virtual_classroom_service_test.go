package controller_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	logger_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	logger_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestModifyVirtualClassroomState(t *testing.T) {
	t.Parallel()

	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	virtualLessonPollingRepo := &mock_repositories.MockVirtualLessonPollingRepo{}
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)
	validPresentMaterialJSON := database.JSONB(`
	{
		"current_material": {
			"media_id": "media-1",
			"updated_at": "` + string(nowString) + `"
		}
	}`)
	jsm := &mock_nats.JetStreamManagement{}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.ModifyVirtualClassroomStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher command to share a material (video) in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
					ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
						State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState{
							VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
								CurrentTime: durationpb.New(12 * time.Second),
								PlayerState: vpb.PlayerState_PLAYER_STATE_PAUSE,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, tx, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentMaterial)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Equal(t, domain.Duration(12*time.Second), state.VideoState.CurrentTime)
						assert.Equal(t, domain.PlayerStatePause, state.VideoState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("GetByIDAndCourseID", ctx, tx, "lesson-group-1", "course-1").
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to share a material (pdf) in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
					ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, tx, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentMaterial)
						require.NoError(t, err)
						assert.Equal(t, "media-2", state.MediaID)
						assert.Nil(t, state.VideoState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("GetByIDAndCourseID", ctx, tx, "lesson-group-1", "course-1").
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to share a material (audio) in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
					ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
						MediaId: "media-3",
						State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_AudioState{
							AudioState: &vpb.VirtualClassroomState_CurrentMaterial_AudioState{
								CurrentTime: durationpb.New(13 * time.Second),
								PlayerState: vpb.PlayerState_PLAYER_STATE_PAUSE,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, tx, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentMaterial)
						require.NoError(t, err)
						assert.Equal(t, "media-3", state.MediaID)
						assert.Equal(t, domain.Duration(13*time.Second), state.AudioState.CurrentTime)
						assert.Equal(t, domain.PlayerStatePause, state.AudioState.PlayerState)
						assert.False(t, state.UpdatedAt.IsZero())
						assert.False(t, now.Equal(state.UpdatedAt))
					}).
					Return(nil).Once()
				lessonGroupRepo.
					On("GetByIDAndCourseID", ctx, tx, "lesson-group-1", "course-1").
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text("lesson-group-1"),
						CourseID:      database.Text("course-1"),
						MediaIDs:      database.TextArray([]string{"media-1", "media-2", "media-3"}),
					}, nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to share a material (pdf) in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
					ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop sharing current material in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, tx, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentMaterialState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentMaterial)
						require.NoError(t, err)
						assert.Nil(t, state)
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to stop sharing current material in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to fold hand all learner in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						db,
						"lesson-1",
						domain.LearnerStateTypeHandsUp,
						&repo.StateValueDTO{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold hand all learner in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_FoldHandAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to fold user's hand in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_FoldUserHand{
					FoldUserHand: "learner-2",
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*repo.LessonMemberStateDTO)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(domain.LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, false, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold self-hands state in a virtual classroom section",
			reqUserID: "learner-2",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_FoldUserHand{
					FoldUserHand: "learner-2",
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*repo.LessonMemberStateDTO)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(domain.LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, false, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to fold other learner's hands state in a virtual classroom section",
			reqUserID: "learner-2",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_FoldUserHand{
					FoldUserHand: "learner-3",
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(4)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()

			},
			hasError: true,
		},
		{
			name:      "teacher command to raise hand in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*repo.LessonMemberStateDTO)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "teacher-1", state.UserID.String)
						assert.Equal(t, string(domain.LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(fmt.Errorf("got 1 error")).
					Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to raise hand in a virtual classroom section",
			reqUserID: "learner-2",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						db,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*repo.LessonMemberStateDTO)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-2", state.UserID.String)
						assert.Equal(t, string(domain.LearnerStateTypeHandsUp), state.StateType.String)
						assert.Equal(t, true, state.BoolValue.Bool)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner who not belong to lesson command to raise hand in a virtual classroom section",
			reqUserID: "learner-5",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-5").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to enables annotation in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationEnable{
					AnnotationEnable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypeAnnotation,
						[]string{"learner-1", "learner-2"},
						&repo.StateValueDTO{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to enables annotation in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationEnable{
					AnnotationEnable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to disable annotation in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisable{
					AnnotationDisable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypeAnnotation,
						[]string{"learner-1", "learner-2"},
						&repo.StateValueDTO{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to disables annotation in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisable{
					AnnotationDisable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to disable annotation for all in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypeAnnotation,
						&repo.StateValueDTO{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "learner command to disable annotation for all in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_AnnotationDisableAll{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to start polling a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StartPolling{
					StartPolling: &vpb.ModifyVirtualClassroomStateRequest_PollingOptions{
						Options: []*vpb.ModifyVirtualClassroomStateRequest_PollingOption{
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
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentPollingState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentPolling)
						require.NoError(t, err)
						assert.Equal(t, domain.CurrentPollingStatusStarted, state.Status)
						assert.False(t, state.CreatedAt.IsZero())
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to start polling in a virtual classroom section when exists",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StartPolling{
					StartPolling: &vpb.ModifyVirtualClassroomStateRequest_PollingOptions{
						Options: []*vpb.ModifyVirtualClassroomStateRequest_PollingOption{
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
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStarted,
							CreatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to start polling in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StartPolling{
					StartPolling: &vpb.ModifyVirtualClassroomStateRequest_PollingOptions{
						Options: []*vpb.ModifyVirtualClassroomStateRequest_PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
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
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop polling in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStarted,
							CreatedAt: now,
						},
					}, nil).Once()

				lessonRoomStateRepo.
					On("UpsertCurrentPollingState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentPolling)
						require.NoError(t, err)
						assert.Equal(t, domain.CurrentPollingStatusStopped, state.Status)
						assert.False(t, now.Equal(*state.StoppedAt))
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to stop polling in a virtual classroom section when polling stopped",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStopped,
							CreatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop polling in a virtual classroom section when nothing polling",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{}, domain.ErrLessonRoomStateNotFound)
			},
			hasError: true,
		},
		{
			name:      "learner command to stop polling in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_StopPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "teacher command to end polling in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStopped,
							CreatedAt: now,
						},
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						tx,
						mock.Anything,
					).
					Return(
						repo.LessonMemberStateDTOs{
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue:        database.Bool(false),
								StringArrayValue: database.TextArray([]string{"A"}),
								DeletedAt:        database.Timestamptz(now),
							},
						},
						nil,
					).
					Once()
				virtualLessonPollingRepo.
					On("Create", ctx, tx, mock.Anything).
					Return(&repo.VirtualLessonPolling{
						PollID: database.Text("poll-1"),
					}, nil).Once()

				lessonRoomStateRepo.
					On("UpsertCurrentPollingState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentPolling)
						require.NoError(t, err)
						assert.Nil(t, state)
					}).
					Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertAllLessonMemberStateByStateType",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypePollingAnswer,
						&repo.StateValueDTO{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to end polling in a virtual classroom section when polling started",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStarted,
							CreatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to end polling in a virtual classroom section when nothing polling",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{}, nil)
			},
			hasError: true,
		},
		{
			name:      "learner command to end polling in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id:      "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_EndPolling{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				studentsRepo.On("IsUserIDAStudent", ctx, db, "learner-1").
					Once().
					Return(true, nil)
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in a virtual classroom section",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &vpb.ModifyVirtualClassroomStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(5)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStarted,
							CreatedAt: now,
						},
					}, nil).Once()
				lessonMemberRepo.
					On(
						"GetLessonMemberStatesWithParams",
						ctx,
						tx,
						mock.Anything,
					).
					Return(
						repo.LessonMemberStateDTOs{},
						nil,
					).
					Once()
				lessonMemberRepo.
					On(
						"UpsertLessonMemberState",
						ctx,
						tx,
						mock.Anything,
					).
					Run(func(args mock.Arguments) {
						state := args.Get(2).(*repo.LessonMemberStateDTO)
						assert.Equal(t, "lesson-1", state.LessonID.String)
						assert.Equal(t, "learner-1", state.UserID.String)
						assert.Equal(t, string(domain.LearnerStateTypePollingAnswer), state.StateType.String)
					}).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher try command to submit polling answer in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &vpb.ModifyVirtualClassroomStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(4)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in a virtual classroom section when polling stopped",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &vpb.ModifyVirtualClassroomStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(4)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStopped,
							CreatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "learner command to submit polling answer in a virtual classroom section when nothing polling",
			reqUserID: "learner-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer{
					SubmitPollingAnswer: &vpb.ModifyVirtualClassroomStateRequest_PollingAnswer{
						StringArrayValue: []string{"A"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(4)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{}, domain.ErrLessonRoomStateNotFound)
			},
			hasError: true,
		},
		{
			name:      "teacher command to share polling in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_SharePolling{
					SharePolling: true,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("GetLessonRoomStateByLessonID", ctx, tx, "lesson-1").
					Return(&domain.LessonRoomState{
						LessonID: "lesson-1",
						CurrentMaterial: &domain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
						CurrentPolling: &domain.CurrentPolling{
							Options: domain.CurrentPollingOptions{
								{
									Answer:    "A",
									IsCorrect: true,
									Content:   "content a. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
								{
									Answer:    "B",
									IsCorrect: false,
									Content:   "",
								},
								{
									Answer:    "C",
									IsCorrect: false,
									Content:   "content c. Curabitur aliquet quam id dui posuere blandit. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur arcu erat, accumsan id imperdiet et, porttitor at sem.",
								},
							},
							Status:    domain.CurrentPollingStatusStopped,
							CreatedAt: now,
						},
					}, nil).Once()
				lessonRoomStateRepo.
					On("UpsertCurrentPollingState", ctx, tx, "lesson-1", mock.Anything).
					Run(func(args mock.Arguments) {
						state := args.Get(3).(*domain.CurrentPolling)
						require.NoError(t, err)
						assert.True(t, state.IsShared)
					}).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to spotlight a user",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_Spotlight_{
					Spotlight: &vpb.ModifyVirtualClassroomStateRequest_Spotlight{
						UserId:      "learner-2",
						IsSpotlight: true,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UpsertSpotlightState", ctx, tx, "lesson-1", "learner-2").
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command remove spotlight",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_Spotlight_{
					Spotlight: &vpb.ModifyVirtualClassroomStateRequest_Spotlight{
						IsSpotlight: false,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UnSpotlight", ctx, tx, "lesson-1").
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command modify whiteboard zoom state",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_WhiteboardZoomState_{
					WhiteboardZoomState: &vpb.ModifyVirtualClassroomStateRequest_WhiteboardZoomState{
						PdfScaleRatio: 100.0,
						PdfWidth:      1920.0,
						PdfHeight:     1080.0,
						CenterX:       0.0,
						CenterY:       0.0,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRoomStateRepo.
					On("UpsertWhiteboardZoomState", ctx, tx, "lesson-1", mock.Anything).
					Return(nil).Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to enables chat in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ChatEnable{
					ChatEnable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypeChat,
						[]string{"learner-1", "learner-2"},
						&repo.StateValueDTO{
							BoolValue:        database.Bool(true),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to disable chat in a virtual classroom section",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_ChatDisable{
					ChatDisable: &vpb.ModifyVirtualClassroomStateRequest_Learners{
						Learners: []string{"learner-1", "learner-2"},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonMemberRepo.
					On(
						"UpsertMultiLessonMemberStateByState",
						ctx,
						tx,
						"lesson-1",
						domain.LearnerStateTypeChat,
						[]string{"learner-1", "learner-2"},
						&repo.StateValueDTO{
							BoolValue:        database.Bool(false),
							StringArrayValue: database.TextArray([]string{}),
						},
					).
					Return(nil).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-1",
					logger_repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
		},
		{
			name:      "teacher command to upsert session time",
			reqUserID: "teacher-1",
			req: &vpb.ModifyVirtualClassroomStateRequest{
				Id: "lesson-1",
				Command: &vpb.ModifyVirtualClassroomStateRequest_UpsertSessionTime{
					UpsertSessionTime: true,
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(6)
				lessonRepo.
					On("GetVirtualLessonByID", ctx, db, "lesson-1").
					Return(&domain.VirtualLesson{
						LessonID:      "lesson-1",
						LessonGroupID: "lesson-group-1",
						CourseID:      "course-1",
						RoomState:     domain.UnmarshalRoomStateJSON(validPresentMaterialJSON),
					},
						nil,
					).Once()

				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"teacher-1", "teacher-2"}, nil).Once()

				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, "lesson-1").
					Return([]string{"learner-1", "learner-2", "learner-3"}, nil).Once()

				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				lessonRoomStateRepo.On("UpsertLiveLessonSessionTime", ctx, tx, "lesson-1", mock.Anything).
					Return(nil).Once()

				logRepo.On("IncreaseTotalTimesByLessonID", ctx, db, "lesson-1", logger_repo.TotalTimesUpdatingRoomState).
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

			jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {}).Return("", nil)
			srv := &controller.VirtualClassroomModifierService{
				WrapperDBConnection: wrapperConnection,
				VirtualClassRoomLogService: &logger_controller.VirtualClassRoomLogService{
					WrapperConnection: wrapperConnection,
					Repo:              logRepo,
				},
				VirtualLessonRepo:        lessonRepo,
				LessonGroupRepo:          lessonGroupRepo,
				StudentsRepo:             studentsRepo,
				LessonMemberRepo:         lessonMemberRepo,
				VirtualLessonPollingRepo: virtualLessonPollingRepo,
				LessonRoomStateRepo:      lessonRoomStateRepo,
			}
			_, err := srv.ModifyVirtualClassroomState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonGroupRepo, lessonRoomStateRepo, studentsRepo, lessonMemberRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestJoinLiveLesson(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	videoTokenSuffix := "samplevideosuffix"
	config := configurations.Config{
		Agora:      configurations.AgoraConfig{},
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
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	virtualClassroomLogService := &logger_controller.VirtualClassRoomLogService{
		WrapperConnection: wrapperConnection,
		Repo:              logRepo,
	}
	jsm := &mock_nats.JetStreamManagement{}

	request := &vpb.JoinLiveLessonRequest{
		LessonId: "lesson-id1",
	}
	teacherID := "user-id1"
	studentID := "user-id2"
	courseID := "course-id1"
	whiteboardToken := "whiteboard-token"
	roomUUID := "sample-room-uuid1"
	logID := "log-id1"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.JoinLiveLessonRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher join live lesson and create new room id",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   "",
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				courseRepo.On("GetValidCoursesByCourseIDsAndStatus", ctx, mock.Anything, []string{courseID}, domain.StatusActive).Once().
					Return([]*domain.Course{{ID: courseID}}, nil)

				whiteboardSvc.On("CreateRoom", ctx, mock.Anything).Once().
					Return(&whiteboard.CreateRoomResponse{UUID: roomUUID}, nil)

				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, request.LessonId, roomUUID).Once().
					Return(nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByLessonID", ctx, db, request.LessonId, teacherID).
					Return(nil).Once()
			},
		},
		{
			name:      "teacher join live lesson with existing room id",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   roomUUID,
						CourseID: courseID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				courseRepo.On("GetValidCoursesByCourseIDsAndStatus", ctx, mock.Anything, []string{courseID}, domain.StatusActive).Once().
					Return([]*domain.Course{{ID: courseID}}, nil)

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByLessonID", ctx, db, request.LessonId, teacherID).
					Return(nil).Once()
			},
		},
		{
			name:      "student join live lesson and create new room id",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
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

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByLessonID", ctx, db, request.LessonId, studentID).
					Return(nil).Once()
			},
		},
		{
			name:      "student join live lesson with existing room id",
			reqUserID: studentID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						RoomID:   roomUUID,
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

				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().
					Return(whiteboardToken, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(false),
					}, nil).Once()

				logRepo.On("AddAttendeeIDByLessonID", ctx, db, request.LessonId, studentID).
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

			command := &commands.LiveLessonCommand{
				WrapperDBConnection: wrapperConnection,
				VideoTokenSuffix:    videoTokenSuffix,
				WhiteboardSvc:       whiteboardSvc,
				AgoraTokenSvc:       agoraTokenSvc,
				VirtualLessonRepo:   lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				StudentsRepo:        studentsRepo,
				CourseRepo:          courseRepo,
			}

			service := &controller.VirtualClassroomModifierService{
				WrapperDBConnection:        wrapperConnection,
				JSM:                        jsm,
				Cfg:                        config,
				LiveLessonCommand:          command,
				VirtualClassRoomLogService: virtualClassroomLogService,
			}

			response, err := service.JoinLiveLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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

			mock.AssertExpectationsForObjects(t, db, lessonRepo, studentsRepo, lessonMemberRepo, courseRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestLeaveLiveLesson(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	config := configurations.Config{
		Agora:      configurations.AgoraConfig{},
		Whiteboard: configs.WhiteboardConfig{AppID: "app-id"},
	}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	jsm := &mock_nats.JetStreamManagement{}

	lessonID := "lesson-id1"
	teacherID := "user-id1"
	studentID := "user-id2"

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.LeaveLiveLessonRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher leave live lesson",
			reqUserID: teacherID,
			req: &vpb.LeaveLiveLessonRequest{
				LessonId: lessonID,
				UserId:   teacherID,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, lessonID).Once().
					Return(&domain.VirtualLesson{
						LessonID: lessonID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)
			},
		},
		{
			name:      "student leave live lesson",
			reqUserID: studentID,
			req: &vpb.LeaveLiveLessonRequest{
				LessonId: lessonID,
				UserId:   studentID,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, lessonID).Once().
					Return(&domain.VirtualLesson{
						LessonID: lessonID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, studentID).Once().
					Return(true, nil)

				jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().
					Run(func(args mock.Arguments) {}).
					Return("", nil)
			},
		},
		{
			name:      "student leave live lesson with different user ID in request",
			reqUserID: studentID,
			req: &vpb.LeaveLiveLessonRequest{
				LessonId: lessonID,
				UserId:   teacherID,
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, lessonID).Once().
					Return(&domain.VirtualLesson{
						LessonID: lessonID,
					}, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
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

			command := &commands.LiveLessonCommand{
				WrapperDBConnection: wrapperConnection,
				VirtualLessonRepo:   lessonRepo,
				StudentsRepo:        studentsRepo,
			}

			service := &controller.VirtualClassroomModifierService{
				WrapperDBConnection: wrapperConnection,
				JSM:                 jsm,
				Cfg:                 config,
				LiveLessonCommand:   command,
			}

			_, err := service.LeaveLiveLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentsRepo, mockUnleashClient)
		})
	}
}

func TestEndLiveLesson(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	videoTokenSuffix := "samplevideosuffix"
	config := configurations.Config{
		Agora:      configurations.AgoraConfig{},
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
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	studentsRepo := &mock_repositories.MockStudentsRepo{}
	courseRepo := &mock_repositories.MockCourseRepo{}
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	virtualClassroomLogService := &logger_controller.VirtualClassRoomLogService{
		WrapperConnection: wrapperConnection,
		Repo:              logRepo,
	}
	jsm := &mock_nats.JetStreamManagement{}

	request := &vpb.EndLiveLessonRequest{
		LessonId: "lesson-id1",
	}
	teacherID := "user-id1"
	// studentID := "user-id2"
	courseID := "course-id1"
	logID := "log-id1"
	lessonGroupID := "lesson-group-1"
	mediaIDs := []string{"media-1", "media-2", "media-3"}
	learnerIDs := []string{"learner-1", "learner-2", "learner-3"}
	teacherIDs := []string{"teacher-1", "teacher-2", "teacher-3"}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.EndLiveLessonRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher end live lesson successfully",
			reqUserID: teacherID,
			req:       request,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(10)
				db.On("Begin", ctx).Return(tx, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
						CourseID: courseID,
					}, nil)

				courseRepo.On("GetValidCoursesByCourseIDsAndStatus", ctx, mock.Anything, []string{courseID}, domain.StatusActive).Once().
					Return([]*domain.Course{{ID: courseID}}, nil)

				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID: request.LessonId,
					}, nil)

				lessonRepo.On("GetLearnerIDsOfLesson", ctx, mock.Anything, request.LessonId).Once().
					Return(learnerIDs, nil)

				lessonRepo.On("GetTeacherIDsOfLesson", ctx, mock.Anything, request.LessonId).Once().
					Return(teacherIDs, nil)

				studentsRepo.On("IsUserIDAStudent", ctx, mock.Anything, teacherID).Once().
					Return(false, nil)

				lessonRepo.On("GetVirtualLessonByID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.VirtualLesson{
						LessonID:      request.LessonId,
						CourseID:      courseID,
						LessonGroupID: lessonGroupID,
						RoomState: domain.OldLessonRoomState{
							CurrentMaterial: &domain.CurrentMaterial{
								MediaID: mediaIDs[0],
							},
						},
					}, nil)

				lessonGroupRepo.On("GetByIDAndCourseID", ctx, mock.Anything, request.LessonId, mock.Anything).
					Return(&repo.LessonGroupDTO{
						LessonGroupID: database.Text(lessonGroupID),
						CourseID:      database.Text(courseID),
						MediaIDs:      database.TextArray(mediaIDs),
					}, nil).Once()

				lessonRoomStateRepo.On("UpsertCurrentMaterialState", ctx, mock.Anything, request.LessonId, mock.Anything).Once().
					Return(nil)

				lessonMemberRepo.On("UpsertAllLessonMemberStateByStateType", ctx, mock.Anything, request.LessonId, mock.Anything, mock.Anything).Once().
					Return(nil)

				lessonMemberRepo.On("UpsertAllLessonMemberStateByStateType", ctx, mock.Anything, request.LessonId, mock.Anything, mock.Anything).Once().
					Return(nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, mock.Anything, request.LessonId).Once().
					Return(&domain.LessonRoomState{
						CurrentPolling: &domain.CurrentPolling{
							IsShared: true,
						},
					}, nil)

				lessonRoomStateRepo.On("UpsertCurrentPollingState", ctx, mock.Anything, request.LessonId, mock.Anything).Once().
					Return(nil)

				lessonMemberRepo.On("UpsertAllLessonMemberStateByStateType", ctx, mock.Anything, request.LessonId, mock.Anything, mock.Anything).Once().
					Return(nil)

				lessonRoomStateRepo.On("UpsertWhiteboardZoomState", ctx, mock.Anything, request.LessonId, mock.Anything).Once().
					Return(nil)

				lessonRoomStateRepo.On("UnSpotlight", ctx, mock.Anything, request.LessonId).Once().
					Return(nil)

				lessonMemberRepo.On("UpsertAllLessonMemberStateByStateType", ctx, mock.Anything, request.LessonId, mock.Anything, mock.Anything).Once().
					Return(nil)

				lessonRoomStateRepo.On("UpsertRecordingState", ctx, mock.Anything, request.LessonId, mock.Anything).Once().
					Return(nil)

				lessonRepo.On("EndLiveLesson", ctx, mock.Anything, request.LessonId, mock.Anything).Once().
					Return(nil)

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(false),
					}, nil).Once()

				logRepo.On("CompleteLogByLessonID", ctx, db, request.LessonId).
					Return(nil).Once()

				logRepo.On("GetLatestByLessonID", ctx, db, request.LessonId).
					Return(&logger_repo.VirtualClassRoomLogDTO{
						LogID:       database.Text(logID),
						LessonID:    database.Text(request.LessonId),
						IsCompleted: database.Bool(true),
					}, nil).Once()

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

			command := &commands.LiveLessonCommand{
				WrapperDBConnection: wrapperConnection,
				VideoTokenSuffix:    videoTokenSuffix,
				WhiteboardSvc:       whiteboardSvc,
				AgoraTokenSvc:       agoraTokenSvc,
				VirtualLessonRepo:   lessonRepo,
				LessonMemberRepo:    lessonMemberRepo,
				StudentsRepo:        studentsRepo,
				CourseRepo:          courseRepo,
			}

			service := &controller.VirtualClassroomModifierService{
				WrapperDBConnection:        wrapperConnection,
				JSM:                        jsm,
				Cfg:                        config,
				LiveLessonCommand:          command,
				VirtualClassRoomLogService: virtualClassroomLogService,
				VirtualLessonRepo:          lessonRepo,
				StudentsRepo:               studentsRepo,
				LessonGroupRepo:            lessonGroupRepo,
				LessonMemberRepo:           lessonMemberRepo,
				LessonRoomStateRepo:        lessonRoomStateRepo,
			}

			_, err := service.EndLiveLesson(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, studentsRepo, lessonMemberRepo, courseRepo, logRepo, mockUnleashClient)
		})
	}
}

func TestPreparePublish(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	activityLogRepo := &mock_repositories.MockActivityLogRepo{}

	maximumLearnerStreamings := 5
	teacherID := "user-id1"
	studentID := "user-id2"
	studentID3 := "user-id3"
	lessonID := "lesson-id1"
	learnerIDs := []string{studentID3, "user-id4"}

	tcs := []struct {
		name           string
		reqUserID      string
		req            *vpb.PreparePublishRequest
		expectedStatus vpb.PrepareToPublishStatus
		setup          func(ctx context.Context)
		hasError       bool
	}{
		{
			name:      "teacher prepare publish successfully",
			reqUserID: teacherID,
			req: &vpb.PreparePublishRequest{
				LessonId:  lessonID,
				LearnerId: studentID,
			},
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, lessonID, true).Once().
					Return(learnerIDs, nil)

				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, lessonID, studentID, maximumLearnerStreamings).Once().
					Return(nil)

				payload := map[string]interface{}{
					"lesson_id": lessonID,
				}
				activityLogRepo.On("Create", ctx, mock.Anything, studentID, constant.LogActionTypePublish, payload).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher gets prepared before status",
			reqUserID: teacherID,
			req: &vpb.PreparePublishRequest{
				LessonId:  lessonID,
				LearnerId: studentID3,
			},
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, lessonID, true).Once().
					Return(learnerIDs, nil)
			},
		},
		{
			name:      "teacher gets reached maximum learner streamings limit status",
			reqUserID: teacherID,
			req: &vpb.PreparePublishRequest{
				LessonId:  lessonID,
				LearnerId: studentID,
			},
			expectedStatus: vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, lessonID, true).Once().
					Return(learnerIDs, nil)

				lessonRepo.On("IncreaseNumberOfStreaming", ctx, mock.Anything, lessonID, studentID, maximumLearnerStreamings).Once().
					Return(fmt.Errorf("no rows updated"))
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
				WrapperDBConnection:      wrapperConnection,
				MaximumLearnerStreamings: maximumLearnerStreamings,
				VirtualLessonRepo:        lessonRepo,
				ActivityLogRepo:          activityLogRepo,
			}

			service := &controller.VirtualClassroomModifierService{
				WrapperDBConnection: wrapperConnection,
				LiveLessonCommand:   command,
			}

			res, err := service.PreparePublish(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedStatus, res.Status)
			mock.AssertExpectationsForObjects(t, db, lessonRepo, activityLogRepo, mockUnleashClient)
		})
	}
}

func TestUnpublish(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockVirtualLessonRepo{}
	activityLogRepo := &mock_repositories.MockActivityLogRepo{}

	teacherID := "user-id1"
	studentID := "user-id2"
	studentID3 := "user-id3"
	lessonID := "lesson-id1"
	learnerIDs := []string{studentID3, "user-id4"}

	tcs := []struct {
		name           string
		reqUserID      string
		req            *vpb.UnpublishRequest
		expectedStatus vpb.UnpublishStatus
		setup          func(ctx context.Context)
		hasError       bool
	}{
		{
			name:      "teacher unpublish successfully",
			reqUserID: teacherID,
			req: &vpb.UnpublishRequest{
				LessonId:  lessonID,
				LearnerId: studentID,
			},
			expectedStatus: vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_NONE,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()

				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, lessonID, false).Once().
					Return(learnerIDs, nil)

				lessonRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, lessonID, studentID).Once().
					Return(nil)

				payload := map[string]interface{}{
					"lesson_id": lessonID,
				}
				activityLogRepo.On("Create", ctx, mock.Anything, studentID, constant.LogActionTypeUnpublish, payload).Once().
					Return(nil)
			},
		},
		{
			name:      "teacher gets prepared before status",
			reqUserID: teacherID,
			req: &vpb.UnpublishRequest{
				LessonId:  lessonID,
				LearnerId: studentID,
			},
			expectedStatus: vpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()

				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, lessonID, false).Once().
					Return(learnerIDs, nil)

				lessonRepo.On("DecreaseNumberOfStreaming", ctx, mock.Anything, lessonID, studentID).Once().
					Return(fmt.Errorf("no rows updated"))
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
				VirtualLessonRepo:   lessonRepo,
				ActivityLogRepo:     activityLogRepo,
			}

			service := &controller.VirtualClassroomModifierService{
				WrapperDBConnection: wrapperConnection,
				LiveLessonCommand:   command,
			}

			res, err := service.Unpublish(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedStatus, res.Status)
			mock.AssertExpectationsForObjects(t, db, lessonRepo, activityLogRepo, mockUnleashClient)
		})
	}
}
