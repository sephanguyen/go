package controller

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_user_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/usermodadapter"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestLessonManagementGRPCService_ModifyLiveLessonState(t *testing.T) {
	t.Parallel()
	now := time.Now()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	userModuleAdapter := new(mock_user_module_adapter.MockUserModuleAdapter)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	roomStateRepo := new(mock_repositories.MockLessonRoomStateRepo)

	tcs := []struct {
		name      string
		reqUserID string
		req       *bpb.ModifyLiveLessonStateRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher command to share a material (video) in live lesson room",
			reqUserID: "teacher-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
						State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
							VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
								CurrentTime: durationpb.New(12 * time.Second),
								PlayerState: bpb.PlayerState_PLAYER_STATE_PAUSE,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userModuleAdapter.On("GetUserGroup", ctx, "teacher-2").
					Return(entities_bob.UserGroupTeacher, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID:       "lesson-1",
						Material:       &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
						TeachingMethod: domain.LessonTeachingMethodIndividual,
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
				roomStateRepo.
					On("UpsertCurrentMaterial", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actual := args.Get(2).(*domain.CurrentMaterial)
						assert.Equal(t, "lesson-1", actual.LessonID)
						assert.Equal(t, "media-2", *actual.MediaID)
						assert.False(t, actual.UpdatedAt.IsZero())
						assert.False(t, now.Equal(actual.UpdatedAt))
						assert.Equal(t, 12*time.Second, actual.VideoState.CurrentTime.Duration())
						assert.Equal(t, domain.PlayerStatePause, actual.VideoState.PlayerState)
					}).
					Return(nil, nil).Once()
			},
		},
		{
			name:      "teacher command to share a material (pdf) in live lesson room",
			reqUserID: "teacher-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userModuleAdapter.On("GetUserGroup", ctx, "teacher-2").
					Return(entities_bob.UserGroupTeacher, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID:       "lesson-1",
						Material:       &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
						TeachingMethod: domain.LessonTeachingMethodIndividual,
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "pdf-id",
							Type:      media_domain.MediaTypePDF,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
				roomStateRepo.
					On("UpsertCurrentMaterial", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actual := args.Get(2).(*domain.CurrentMaterial)
						assert.Equal(t, "lesson-1", actual.LessonID)
						assert.Equal(t, "media-2", *actual.MediaID)
						assert.False(t, actual.UpdatedAt.IsZero())
						assert.False(t, now.Equal(actual.UpdatedAt))
						assert.Nil(t, actual.VideoState)
					}).
					Return(nil, nil).Once()
			},
		},
		{
			name:      "learner command to share a material (pdf) in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id: "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
					ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
						MediaId: "media-2",
					},
				},
			},
			setup: func(ctx context.Context) {
				userModuleAdapter.On("GetUserGroup", ctx, "learner-1").
					Return(entities_bob.UserGroupStudent, nil).Once()
			},
			hasError: true,
		},
		{
			name:      "teacher command to stop sharing current material in live lesson room",
			reqUserID: "teacher-2",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userModuleAdapter.On("GetUserGroup", ctx, "teacher-2").
					Return(entities_bob.UserGroupTeacher, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID:       "lesson-1",
						Material:       &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
						TeachingMethod: domain.LessonTeachingMethodIndividual,
					}, nil).Once()
				roomStateRepo.
					On("UpsertCurrentMaterial", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actual := args.Get(2).(*domain.CurrentMaterial)
						assert.Equal(t, "lesson-1", actual.LessonID)
						assert.Nil(t, actual.MediaID)
						assert.False(t, actual.UpdatedAt.IsZero())
						assert.Nil(t, actual.VideoState)
					}).
					Return(nil, nil).Once()
			},
		},
		{
			name:      "learner command to stop sharing current material in live lesson room",
			reqUserID: "learner-1",
			req: &bpb.ModifyLiveLessonStateRequest{
				Id:      "lesson-1",
				Command: &bpb.ModifyLiveLessonStateRequest_StopSharingMaterial{},
			},
			setup: func(ctx context.Context) {
				userModuleAdapter.On("GetUserGroup", ctx, "learner-1").
					Return(entities_bob.UserGroupStudent, nil).Once()
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

			srv := NewLessonManagementGRPCService(
				wrapperConnection,
				lessonRepo,
				userModuleAdapter,
				mediaModulePort,
				roomStateRepo,
			)
			_, err := srv.ModifyLiveLessonState(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, userModuleAdapter, mediaModulePort, roomStateRepo, mockUnleashClient)
		})
	}
}
