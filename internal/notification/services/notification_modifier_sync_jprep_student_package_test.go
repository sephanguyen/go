package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_SyncJprepStudentPackage(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	notificationStudentCourseRepo := &mock_repositories.MockNotificationStudentCourseRepo{}

	svc := &NotificationModifierService{
		DB:                            mockDB,
		NotificationStudentCourseRepo: notificationStudentCourseRepo,
	}

	ctx := context.Background()

	t.Run("happy case upsert", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		now := time.Now()
		studentPackages := []*npb.EventSyncStudentPackage_StudentPackage{
			{
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				StudentId:  "student-id-1",
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: []string{"course-1"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
					{
						CourseIds: []string{"course-2", "course-3"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
				},
			},
			{
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				StudentId:  "student-id-2",
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: []string{"course-1"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
					{
						CourseIds: []string{"course-2", "course-3"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
				},
			},
		}

		for _, studentPackage := range studentPackages {
			softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
			_ = multierr.Combine(
				softDeleteFilter.StudentIDs.Set([]string{studentPackage.StudentId}),
			)

			notificationStudentCourseRepo.On("SoftDelete", ctx, mockTx, softDeleteFilter).Once().Return(nil)
			notificationStudentCourseRepo.On("BulkCreate", ctx, mockTx, mock.Anything).Once().Return(nil)
		}

		err := svc.SyncJprepStudentPackage(ctx, studentPackages)
		assert.NoError(t, err)
	})

	t.Run("happy case delete", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		now := time.Now()
		studentPackages := []*npb.EventSyncStudentPackage_StudentPackage{
			{
				ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				StudentId:  "student-id-1",
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: []string{"course-1"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
					{
						CourseIds: []string{"course-2", "course-3"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
				},
			},
			{
				ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
				StudentId:  "student-id-2",
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: []string{"course-1"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
					{
						CourseIds: []string{"course-2", "course-3"},
						StartDate: timestamppb.New(now),
						EndDate:   timestamppb.New(now),
					},
				},
			},
		}

		for _, studentPackage := range studentPackages {
			courseIDs := make([]string, 0)
			for _, pkg := range studentPackage.Packages {
				courseIDs = append(courseIDs, pkg.CourseIds...)
			}
			softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
			_ = multierr.Combine(
				softDeleteFilter.StudentIDs.Set([]string{studentPackage.StudentId}),
				softDeleteFilter.CourseIDs.Set(courseIDs),
			)
			notificationStudentCourseRepo.On("SoftDelete", ctx, mockTx, softDeleteFilter).Once().Return(nil)
		}

		err := svc.SyncJprepStudentPackage(ctx, studentPackages)
		assert.NoError(t, err)
	})

}
