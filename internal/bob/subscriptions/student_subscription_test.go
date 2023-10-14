package subscriptions

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/support"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	cls_repositories "github.com/manabie-com/backend/mock/lessonmgmt/course_location_schedule/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var FatimaStudentPackageMockData = &npb.EventStudentPackage{
	StudentPackage: &npb.EventStudentPackage_StudentPackage{
		StudentId: "test-student-id-1",
		Package: &npb.EventStudentPackage_Package{
			CourseIds:   []string{"test-course-id-1", "test-course-id-2", "test-course-id-3"},
			StartDate:   timestamppb.New(time.Date(2021, 10, 30, 12, 0, 0, 0, time.UTC)),
			EndDate:     timestamppb.New(time.Date(2021, 12, 30, 12, 0, 0, 0, time.UTC)),
			LocationIds: []string{"test-location-id-1", "test-location-id-2", "test-location-id-3"},
		},
		IsActive: true,
	},
}

func TestStudentSubscriptionJobEvent(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		// init mock vars
		ctx := context.Background()
		db := &mock_database.Ext{}
		tx := &mock_database.Tx{}
		mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
		wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
		studentSubscriptionMockRepo := &mock_repositories.MockStudentSubscriptionRepo{}
		studentSubscriptionAccessPathMockRepo := &mock_repositories.MockStudentSubscriptionAccessPathRepo{}
		courseLocationScheduleRepo := &cls_repositories.MockCourseLocationScheduleRepo{}

		// set triggers
		db.On("Begin", ctx).Return(tx, nil)
		tx.On("Commit", ctx).Return(nil)
		studentSubscriptionMockRepo.On("RetrieveStudentSubscriptionID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("test-id", nil)
		mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Return(false, nil).Times(3)
		studentSubscriptionMockRepo.On("DeleteByCourseIDAndStudentID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentSubscriptionAccessPathMockRepo.On("DeleteByStudentSubscriptionID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		studentSubscriptionMockRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(studentSubscriptions []*entities.StudentSubscription) bool {
			// check data after bulk upsert
			if len(studentSubscriptions) < 1 {
				return false
			}
			for i := 0; i < len(studentSubscriptions); i++ {
				if !strings.Contains(studentSubscriptions[i].CourseID.String, "test-course-id-") &&
					!strings.Contains(studentSubscriptions[i].StudentID.String, "test-student-id-") {
					return false
				}
			}
			return true
		})).Return(nil)
		studentSubscriptionAccessPathMockRepo.On("Upsert", mock.Anything, tx, mock.MatchedBy(func(studentSubscriptionAccessPaths []*entities.StudentSubscriptionAccessPath) bool {
			// check data after bulk upsert
			if len(studentSubscriptionAccessPaths) < 1 {
				return false
			}
			for i := 0; i < len(studentSubscriptionAccessPaths); i++ {
				if !strings.Contains(studentSubscriptionAccessPaths[i].LocationID.String, "test-location-id-") {
					return false
				}
			}
			return true
		})).Return(nil)
		data := &npb.EventStudentPackage{
			StudentPackage: FatimaStudentPackageMockData.StudentPackage,
		}

		// init NATSJS Subscriber
		zapLogger := zap.NewNop()
		s := &StudentSubscription{
			WrapperConnection:                 wrapperConnection,
			StudentSubscriptionRepo:           studentSubscriptionMockRepo,
			StudentSubscriptionAccessPathRepo: studentSubscriptionAccessPathMockRepo,
			Logger:                            zapLogger,
			UnleashClientIns:                  mockUnleashClient,
			Env:                               "local",
			CourseLocationScheduleRepo:        courseLocationScheduleRepo,
			LessonAllocationRepo:              nil,
			StudentCourseRepo:                 nil,
		}
		// handle data
		err := s.handle(ctx, data, tx)
		assert.Nil(t, err)
	})
}
