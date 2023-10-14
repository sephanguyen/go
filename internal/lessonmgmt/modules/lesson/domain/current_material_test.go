package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	media_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_media_module_adapter "github.com/manabie-com/backend/mock/lessonmgmt/lesson/media_module_adapter"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCurrentMaterial_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	mediaID := "media-id"

	db := &mock_database.Ext{}
	mediaModulePort := new(mock_media_module_adapter.MockMediaModuleAdapter)
	lessonRepo := new(mock_repositories.MockLessonRepo)

	tcs := []struct {
		name            string
		currentMaterial *domain.CurrentMaterial
		setup           func(ctx context.Context)
		isValid         bool
	}{
		{
			name: "full fields with media is the video",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "full fields with media is the video and is pausing",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "full fields with media is the video and is playing",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "full fields with media is the video and ended",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(-2 * time.Minute),
					PlayerState: domain.PlayerStateEnded,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "full fields with media is not the video",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypePDF,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing media id and video state",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				UpdatedAt: now,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
			},
			isValid: true,
		},

		// failed cases
		{
			name: "missing lesson id",
			currentMaterial: &domain.CurrentMaterial{
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "missing media id but has video state",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "media is the video without video state",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypeVideo,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "media has video state without player state",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(-2 * time.Minute),
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "media has video state and is pausing but has invalid current time",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(-2 * time.Minute),
					PlayerState: domain.PlayerStatePause,
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "media has video state and is playing but has invalid current time",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(-2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup:   func(ctx context.Context) {},
			isValid: false,
		},
		{
			name: "media is not the video but has video state",
			currentMaterial: &domain.CurrentMaterial{
				LessonID: "lesson-id",
				MediaID:  &mediaID,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
				UpdatedAt: now,
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{
						{
							ID:        mediaID,
							Name:      "media",
							Resource:  "video-id",
							Type:      media_domain.MediaTypePDF,
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "media id is not belong to lesson",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{"media-id-100"}},
					}, nil).Once()
			},
			isValid: false,
		},
		{
			name: "media id is not exist",
			currentMaterial: &domain.CurrentMaterial{
				LessonID:  "lesson-id",
				MediaID:   &mediaID,
				UpdatedAt: now,
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(2 * time.Minute),
					PlayerState: domain.PlayerStatePlaying,
				},
			},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-id").
					Return(&domain.Lesson{
						LessonID: "lesson-id",
						Material: &domain.LessonMaterial{MediaIDs: []string{mediaID}},
					}, nil).Once()
				mediaModulePort.On("RetrieveMediasByIDs", ctx, []string{mediaID}).
					Return(media_domain.Medias{}, nil).Once()
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)
			tc.currentMaterial.LessonRepo = lessonRepo
			tc.currentMaterial.MediaModulePort = mediaModulePort
			err := tc.currentMaterial.IsValid(ctx, db)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, lessonRepo, mediaModulePort)
		})
	}
}
