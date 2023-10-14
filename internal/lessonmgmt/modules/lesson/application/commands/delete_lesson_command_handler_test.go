package commands

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	report_mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	mock_service "github.com/manabie-com/backend/mock/lessonmgmt/zoom/service"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteLessonCommandHandler_DeleteLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockLessonRepo)
	reallocationRepo := new(mock_repositories.MockReallocationRepo)
	lessonReportRepo := new(report_mock_repositories.MockLessonReportRepo)
	mockZoomService := &mock_service.MockZoomService{}

	testCases := []struct {
		name               string
		setup              func(ctx context.Context)
		isRecurringDeleted bool
		lessonIDs          []string
		hasError           bool
	}{
		{
			name:               "happy case - delete lesson one time",
			isRecurringDeleted: false,
			lessonIDs:          []string{"lesson-id-1"},
			setup: func(ctx context.Context) {
				lessonIDs := []string{"lesson-id-1"}
				lessonRepo.On("GetLessonWithSchedulerInfoByLessonID", ctx, db, "lesson-id-1").Return(&domain.Lesson{
					LessonID: "lesson-id-1",
					ZoomID:   "zoom-id-1",
					SchedulerInfo: &domain.SchedulerInfo{
						Freq: "once",
					},
				}, nil).Once()
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockZoomService.On("RetryDeleteZoomLink", ctx, "zoom-id-1").Return(true, nil).Once()
				lessonReportRepo.On("DeleteReportsBelongToLesson", ctx, tx, lessonIDs).Return(nil).Once()
				lessonRepo.On("RemoveZoomLinkByLessonID", ctx, tx, "lesson-id-1").Return(nil).Once()
				lessonRepo.On("Delete", ctx, tx, lessonIDs).Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				reallocationRepo.On("DeleteByOriginalLessonID", ctx, tx, lessonIDs).Return(nil).Once()
				reallocationRepo.On("CancelReallocationByLessonID", ctx, tx, lessonIDs).Return(nil).Once()
			},
		},
		{
			name:               "happy case - delete lesson recurring",
			isRecurringDeleted: true,
			lessonIDs:          []string{"lesson-id-1", "lesson-id-2", "lesson-id-3"},
			setup: func(ctx context.Context) {
				lessonIDs := []string{"lesson-id-1", "lesson-id-2", "lesson-id-3"}
				db.On("Begin", ctx).Return(tx, nil).Once()
				tx.On("Commit", ctx).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetFutureRecurringLessonIDs", ctx, db, "lesson-id-1").Return(lessonIDs, nil).Once()
				lessonReportRepo.On("DeleteReportsBelongToLesson", ctx, tx, lessonIDs).Return(nil).Once()
				lessonRepo.On("Delete", ctx, tx, lessonIDs).Return(nil).Once()
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(true, nil).Once()
				reallocationRepo.On("DeleteByOriginalLessonID", ctx, tx, lessonIDs).Return(nil).Once()
				reallocationRepo.On("CancelReallocationByLessonID", ctx, tx, lessonIDs).Return(nil).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			handler := LessonCommandHandler{
				WrapperConnection: wrapperConnection,
				LessonRepo:        lessonRepo,
				UnleashClientIns:  mockUnleashClient,
				Env:               "local",
				ReallocationRepo:  reallocationRepo,
				LessonReportRepo:  lessonReportRepo,
				ZoomService:       mockZoomService,
			}
			res, err := handler.DeleteLesson(ctx, DeleteLessonRequest{
				LessonID:           "lesson-id-1",
				IsDeletedRecurring: tc.isRecurringDeleted,
			})
			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, res, tc.lessonIDs)
				mock.AssertExpectationsForObjects(t, db, tx, lessonRepo, lessonReportRepo, mockUnleashClient)
			}
		})
	}
}
