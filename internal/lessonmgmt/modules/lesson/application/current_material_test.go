package application

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestModifyCurrentMaterialCommandHandler_Execute(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockLessonRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	roomStateRepo := &mock_repositories.MockLessonRoomStateRepo{}

	stringPtr := func(s string) *string {
		return &s
	}

	tcs := []struct {
		name     string
		command  *ModifyCurrentMaterialCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute share video material command successfully",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  stringPtr("media-2"),
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
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
			name: "execute share video material command without video state",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  stringPtr("media-2"),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
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
			},
			hasError: true,
		},
		{
			name: "execute share pdf material command without video state successfully",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  stringPtr("media-2"),
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "video-id",
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
			name: "execute share pdf material command with video state",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  stringPtr("media-2"),
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypePDF,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute modify material command without media id and video state successfully",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
					}, nil).Once()
				roomStateRepo.
					On("UpsertCurrentMaterial", ctx, tx, mock.Anything).
					Run(func(args mock.Arguments) {
						actual := args.Get(2).(*domain.CurrentMaterial)
						assert.Equal(t, "lesson-1", actual.LessonID)
						assert.Nil(t, actual.MediaID)
						assert.False(t, actual.UpdatedAt.IsZero())
						assert.False(t, now.Equal(actual.UpdatedAt))
						assert.Nil(t, actual.VideoState)
					}).
					Return(nil, nil).Once()
			},
		},
		{
			name: "execute share material command with media id not belong to lesson",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  stringPtr("media-2"),
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-1"}},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute modify material command with non-existing lesson id",
			command: &ModifyCurrentMaterialCommand{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			handler := &ModifyCurrentMaterialCommandHandler{
				command:           tc.command,
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				MediaModulePort:   mediaModulePort,
				RoomStateRepo:     roomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, mediaModulePort, roomStateRepo, mockUnleashClient)
		})
	}
}

func TestShareMaterialCommandCommandHandler_Execute(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	now := time.Now().UTC()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockLessonRepo{}
	mediaModulePort := &mock_media_module_adapter.MockMediaModuleAdapter{}
	roomStateRepo := &mock_repositories.MockLessonRoomStateRepo{}

	tcs := []struct {
		name     string
		command  *ShareMaterialCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute share video material command successfully",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  "media-2",
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
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
			name: "execute share video material command without video state",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  "media-2",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
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
			},
			hasError: true,
		},
		{
			name: "execute share pdf material command without video state successfully",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  "media-2",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "video-id",
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
			name: "execute share pdf material command with video state",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  "media-2",
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{"media-2"}).
					Return(media_domain.Medias{
						{
							ID:        "media-2",
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypePDF,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute share material command without media id and video state",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
			hasError: true,
		},
		{
			name: "execute share material command with media id not belong to lesson",
			command: &ShareMaterialCommand{
				LessonID: "lesson-1",
				MediaID:  "media-2",
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(12 * time.Second),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-1"}},
					}, nil).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			handler := &ShareMaterialCommandHandler{
				command:           tc.command,
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				MediaModulePort:   mediaModulePort,
				RoomStateRepo:     roomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, mediaModulePort, roomStateRepo, mockUnleashClient)
		})
	}
}

func TestStopSharingMaterialCommandHandler_Execute(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockLessonRepo{}
	roomStateRepo := &mock_repositories.MockLessonRoomStateRepo{}

	tcs := []struct {
		name     string
		command  *StopSharingMaterialCommand
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "execute stop share material command successfully",
			command: &StopSharingMaterialCommand{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-2"}},
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
			name: "execute stop share material command with non-existing lesson id",
			command: &StopSharingMaterialCommand{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Rollback", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, tx, "lesson-1").
					Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)

			handler := &StopSharingMaterialCommandHandler{
				command:           tc.command,
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				RoomStateRepo:     roomStateRepo,
			}
			err := handler.Execute(ctx)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, roomStateRepo, mockUnleashClient)
		})
	}
}
