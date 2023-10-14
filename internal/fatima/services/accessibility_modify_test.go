package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/fatima/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCourseService_SyncStudentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	db.On("Begin", ctx).Return(tx, nil)
	jsm := new(mock_nats.JetStreamManagement)
	t.Run("[ActionKind UPSERT] should create StudentPackage success", func(t *testing.T) {
		// Arrange
		studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
		studentPackageAccessPathRepo := &mock_repositories.MockStudentPackageAccessPathRepo{}
		c := &AccessibilityModifyService{
			DB:                           db,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			JSM:                          jsm,
		}
		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		defaultLocation := database.Text(constants.JPREPOrgLocation)

		tx.On("Commit", mock.Anything).Once().Return(nil)

		studentPackageRepo.On("SoftDelete", mock.Anything, tx, database.Text("studentId")).
			Once().Return(nil)

		studentPackageRepo.On("BulkInsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.StudentPackage")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.StudentPackage)
				assert.Equal(t, s[0].StartAt.Time, startAt.AsTime())
				assert.Equal(t, s[0].EndAt.Time, endAt.AsTime())
				assert.Equal(t, s[0].StudentID.String, "studentId")
			}).
			Return(nil)

		studentPackageAccessPathRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.StudentPackageAccessPath")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.StudentPackageAccessPath)
				assert.Equal(t, s[0].LocationID, defaultLocation)
				assert.Equal(t, s[0].CourseID.String, "courseID1")
			}).
			Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectSyncJprepStudentPackageEventNats, mock.Anything).Once().Return("", nil)

		// Action
		err := c.SyncStudentPackage(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: []string{"courseID1", "courseID2"},
							StartDate: startAt,
							EndDate:   endAt,
						},
						{
							CourseIds: []string{"courseID1", "courseID2"},
							StartDate: startAt,
							EndDate:   endAt,
						},
					},
				},
			},
		})

		// Assert
		assert.Nil(t, err)
	})

	t.Run("[ActionKind UPSERT] should throw err", func(t *testing.T) {
		// Arrange
		studentPackageRepo := &mock_repositories.MockStudentPackageRepo{}
		studentPackageAccessPathRepo := &mock_repositories.MockStudentPackageAccessPathRepo{}
		c := &AccessibilityModifyService{
			DB:                           db,
			StudentPackageRepo:           studentPackageRepo,
			StudentPackageAccessPathRepo: studentPackageAccessPathRepo,
			JSM:                          jsm,
		}
		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Second()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Second()),
		}

		tx.On("Rollback", mock.Anything).Once().Return(nil)

		studentPackageRepo.On("SoftDelete", mock.Anything, tx, database.Text("studentId")).
			Once().Return(nil)

		studentPackageRepo.On("BulkInsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.StudentPackage")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.StudentPackage)
				assert.Equal(t, s[0].StartAt.Time, startAt.AsTime())
				assert.Equal(t, s[0].EndAt.Time, endAt.AsTime())
				assert.Equal(t, s[0].StudentID.String, "studentId")
			}).
			Return(errors.New("error insert"))

		studentPackageAccessPathRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.StudentPackageAccessPath")).
			Once().Return(nil)

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectSyncJprepStudentPackageEventNats, mock.Anything).Once().Return("", nil)

		// Action
		err := c.SyncStudentPackage(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: []string{"courseID1", "courseID2"},
							StartDate: startAt,
							EndDate:   endAt,
						},
						{
							CourseIds: []string{"courseID1", "courseID2"},
							StartDate: startAt,
							EndDate:   endAt,
						},
					},
				},
			},
		})

		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err s.StudentPackageRepo.BulkInsert, studentID studentId: error insert")
	})
}
