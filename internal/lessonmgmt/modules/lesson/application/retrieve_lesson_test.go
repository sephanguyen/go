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
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRetrieveLessonCommand_GetLessonByID(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := &mock_repositories.MockLessonRepo{}

	tcs := []struct {
		name     string
		command  *RetrieveLessonCommand
		lessonID string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "retrieve lesson command successfully",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
			},
			lessonID: "lesson-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-1").
					Return(&domain.Lesson{
						LessonID: "lesson-1",
					}, nil).Once()
			},
		},
		{
			name: "retrieve lesson command fail",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
			},
			lessonID: "lesson-2",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonByID", ctx, db, "lesson-2").
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

			handler := &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
			}
			lesson, err := handler.GetLessonByID(ctx, tc.lessonID)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, lesson.LessonID, tc.lessonID)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, mockUnleashClient)
		})
	}
}

func TestRetrieveLessonCommand_RetrieveLessonMembersByLessonArgs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}

	tcs := []struct {
		name     string
		command  *RetrieveLessonCommand
		args     *domain.ListStudentsByLessonArgs
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "retrieve lesson member by args command successfully",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonMemberRepo:  lessonMemberRepo,
			},
			args: &domain.ListStudentsByLessonArgs{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("ListStudentsByLessonArgs", ctx, db, &domain.ListStudentsByLessonArgs{
					LessonID: "lesson-1",
				}).
					Return([]*domain.User{{
						ID: "id-1",
					}}, nil).Once()
			},
		},
		{
			name: "retrieve lesson member by args command fail",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonMemberRepo:  lessonMemberRepo,
			},
			args: &domain.ListStudentsByLessonArgs{
				LessonID: "lesson-1",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("ListStudentsByLessonArgs", ctx, db, &domain.ListStudentsByLessonArgs{
					LessonID: "lesson-1",
				}).Return(nil, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)
			_, err := tc.command.RetrieveLessonMembersByLessonArgs(ctx, tc.args)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonMemberRepo, mockUnleashClient)
		})
	}
}

func TestRetrieveLessonCommand_RetrieveMediasByLessonArgs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonGroupRepo := &mock_repositories.MockLessonGroupRepo{}

	tcs := []struct {
		name     string
		command  *RetrieveLessonCommand
		args     *domain.ListMediaByLessonArgs
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "retrieve media by lesson args command successfully",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonGroupRepo:   lessonGroupRepo,
			},
			args: &domain.ListMediaByLessonArgs{
				LessonID: "lesson-1",
				Limit:    10,
				Offset:   "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonGroupRepo.On("ListMediaByLessonArgs", ctx, db, &domain.ListMediaByLessonArgs{
					LessonID: "lesson-1",
					Limit:    10,
					Offset:   "",
				}).
					Return(media_domain.Medias{{
						ID: "id-1",
					}}, nil).Once()
			},
		},
		{
			name: "retrieve medias by args command fail",
			command: &RetrieveLessonCommand{
				WrapperConnection: wrapperConnection,
				LessonGroupRepo:   lessonGroupRepo,
			},
			args: &domain.ListMediaByLessonArgs{
				LessonID: "lesson-1",
				Limit:    10,
				Offset:   "",
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonGroupRepo.On("ListMediaByLessonArgs", ctx, db, &domain.ListMediaByLessonArgs{
					LessonID: "lesson-1",
					Limit:    10,
					Offset:   "",
				}).Return(media_domain.Medias{}, pgx.ErrNoRows).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			tc.setup(ctx)
			_, err := tc.command.RetrieveMediasByLessonArgs(ctx, tc.args)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, lessonGroupRepo, mockUnleashClient)
		})
	}
}
