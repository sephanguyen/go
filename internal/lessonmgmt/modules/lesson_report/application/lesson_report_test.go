package application

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLessonReportCommand_ValidateOptimisticLockingLessonReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	lessonReportDetailRepo := new(mock_repositories.MockLessonReportDetailRepo)
	const (
		StudentID1            = "StudentI_1"
		StudentID2            = "StudentID_2"
		LessonReportID        = "LessonReportID"
		LessonReportDetailID1 = "LessonReportDetailID_1"
		LessonReportDetailID2 = "LessonReportDetailID_2"
		LessonId              = "LessonId"
	)
	tcs := []struct {
		name     string
		req      *domain.LessonReport
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "success to version is not locked",
			req: &domain.LessonReport{
				UnleashToggles: map[string]bool{
					"Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport": true,
				},
				Details: []*domain.LessonReportDetail{
					{
						StudentID:     StudentID1,
						ReportVersion: 1,
					},
					{
						StudentID:     StudentID2,
						ReportVersion: 2,
					},
				},
				LessonReportID: LessonReportID,
				LessonID:       LessonId,
			},
			hasError: false,
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					{ReportVersion: 1, LessonReportDetailID: LessonReportDetailID1, StudentID: StudentID1, LessonReportID: LessonReportID},
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID2, StudentID: StudentID2, LessonReportID: LessonReportID},
				}
				lessonReportDetailRepo.On("GetReportVersionByLessonID", ctx, db, LessonId).Return(
					lessonReportDetails, nil).Once()
			},
		},
		{
			name: "success to version is not locked",
			req: &domain.LessonReport{
				UnleashToggles: map[string]bool{
					"Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport": true,
				},
				Details: []*domain.LessonReportDetail{
					{
						StudentID:     StudentID1,
						ReportVersion: 1,
					},
				},
				LessonReportID:   LessonReportID,
				IsSavePerStudent: true,
				LessonID:         LessonId,
			},
			hasError: false,
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					{ReportVersion: 1, LessonReportDetailID: LessonReportDetailID1, StudentID: StudentID1, LessonReportID: LessonReportID},
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID2, StudentID: StudentID2, LessonReportID: LessonReportID},
				}
				lessonReportDetailRepo.On("GetReportVersionByLessonID", ctx, db, LessonId).Return(
					lessonReportDetails, nil).Once()
			},
		},
		{
			name: "failed to lesson version out date",
			req: &domain.LessonReport{
				UnleashToggles: map[string]bool{
					"Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport": true,
				},
				Details: []*domain.LessonReportDetail{
					{
						StudentID:            StudentID1,
						ReportVersion:        1,
						LessonReportDetailID: LessonReportDetailID1,
					},
					{
						StudentID:            StudentID2,
						ReportVersion:        2,
						LessonReportDetailID: LessonReportDetailID2,
					},
				},
				LessonReportID: LessonReportID,
				LessonID:       LessonId,
			},
			hasError: true,
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID1, StudentID: StudentID1, LessonReportID: LessonReportID},
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID2, StudentID: StudentID2, LessonReportID: LessonReportID},
				}
				lessonReportDetailRepo.On("GetReportVersionByLessonID", ctx, db, LessonId).Return(
					lessonReportDetails, nil).Once()
			},
		},
		{
			name: "failed to lesson report version invalid",
			req: &domain.LessonReport{
				UnleashToggles: map[string]bool{
					"Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport": true,
				},
				Details: []*domain.LessonReportDetail{
					{
						StudentID:            StudentID1,
						ReportVersion:        2,
						LessonReportDetailID: LessonReportDetailID1,
					},
				},
				LessonReportID: LessonReportID,
				LessonID:       LessonId,
			},
			hasError: true,
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					{ReportVersion: 1, LessonReportDetailID: LessonReportDetailID1, StudentID: StudentID1, LessonReportID: LessonReportID},
				}
				lessonReportDetailRepo.On("GetReportVersionByLessonID", ctx, db, LessonId).Return(
					lessonReportDetails, nil).Once()
			},
		},
		{
			name: "failed to version is not locked when save per student",
			req: &domain.LessonReport{
				UnleashToggles: map[string]bool{
					"Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport": true,
				},
				Details: []*domain.LessonReportDetail{
					{
						StudentID:     StudentID1,
						ReportVersion: 1,
					},
				},
				LessonReportID:   LessonReportID,
				IsSavePerStudent: true,
				LessonID:         LessonId,
			},
			hasError: true,
			setup: func(ctx context.Context) {
				lessonReportDetails := domain.LessonReportDetails{
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID1, StudentID: StudentID1, LessonReportID: LessonReportID},
					{ReportVersion: 2, LessonReportDetailID: LessonReportDetailID2, StudentID: StudentID2, LessonReportID: LessonReportID},
				}
				lessonReportDetailRepo.On("GetReportVersionByLessonID", ctx, db, LessonId).Return(
					lessonReportDetails, nil).Once()
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			tc.setup(ctx)
			lessonReport := LessonReportCommand{
				LessonReportDetailRepo: lessonReportDetailRepo,
				Logger:                 zap.NewNop(),
			}

			err := lessonReport.validateOptimisticLockingLessonReport(ctx, db, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
