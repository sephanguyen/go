package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"
	ppb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestStudentCourseSlotInfoHandler(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	userRepo := new(mock_repositories.MockUserRepo)
	studentSubRepo := new(mock_repositories.MockStudentSubscriptionRepo)
	studentSubAccessPathRepo := new(mock_repositories.MockStudentSubscriptionAccessPathRepo)
	now := time.Now()

	tcs := []struct {
		name     string
		data     []*ppb.EventSyncStudentPackageCourse
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "invalid student subscription - missing student id",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name: "invalid student subscription access path - missing location id",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)
			},
			hasError: true,
		},
		{
			name: "failed to get student subscription id",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "failed to get user info",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(nil, errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "failed upsert student subscription",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(&domain.User{FirstName: "student first name 1", LastName: "student last name 2"}, nil)

				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				studentSubRepo.On("BulkUpsertStudentSubscription", mock.Anything, tx, mock.Anything).
					Once().Return(errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "failed delete student subscription access path",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(&domain.User{FirstName: "student first name 1", LastName: "student last name 2"}, nil)

				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				studentSubRepo.On("BulkUpsertStudentSubscription", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("DeleteByStudentSubscriptionIDs", mock.Anything, tx, mock.Anything).
					Once().Return(errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "failed upsert student subscription access path",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(&domain.User{FirstName: "student first name 1", LastName: "student last name 2"}, nil)

				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				studentSubRepo.On("BulkUpsertStudentSubscription", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("DeleteByStudentSubscriptionIDs", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("BulkUpsertStudentSubscriptionAccessPath", mock.Anything, tx, mock.Anything).
					Once().Return(errors.New("error"))
			},
			hasError: true,
		},
		{
			name: "successful handle",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:        "student-id-1",
					LocationId:       "location-id-1",
					CourseId:         "course-id-1",
					StudentPackageId: "student-package-id-1",
					StudentStartDate: timestamppb.New(now),
					StudentEndDate:   timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:      ppb.PackageType_PACKAGE_TYPE_SLOT_BASED,
					CourseSlot:       wrapperspb.Int32(3),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("student-subscription-id-1", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(&domain.User{FirstName: "student first name 1", LastName: "student last name 2"}, nil)

				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()

				studentSubRepo.On("BulkUpsertStudentSubscription", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("DeleteByStudentSubscriptionIDs", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("BulkUpsertStudentSubscriptionAccessPath", mock.Anything, tx, mock.Anything).
					Once().Return(nil)
			},
			hasError: false,
		},
		{
			name: "successful handle without existing student subscription id",
			data: []*ppb.EventSyncStudentPackageCourse{
				{
					StudentId:         "student-id-1",
					LocationId:        "location-id-1",
					CourseId:          "course-id-1",
					StudentPackageId:  "student-package-id-1",
					StudentStartDate:  timestamppb.New(now),
					StudentEndDate:    timestamppb.New(now.Add(24 * 2 * time.Hour)),
					PackageType:       ppb.PackageType_PACKAGE_TYPE_FREQUENCY,
					CourseSlotPerWeek: wrapperspb.Int32(5),
				},
			},
			setup: func(ctx context.Context) {
				studentSubRepo.On("GetStudentSubscriptionIDByUniqueIDs", mock.Anything, db, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Once().Return("", nil)

				userRepo.On("GetUserByUserID", mock.Anything, db, mock.AnythingOfType("string")).
					Once().Return(&domain.User{FirstName: "student first name 1", LastName: "student last name 2"}, nil)

				db.On("Begin", mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()

				studentSubRepo.On("BulkUpsertStudentSubscription", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("DeleteByStudentSubscriptionIDs", mock.Anything, tx, mock.Anything).
					Once().Return(nil)

				studentSubAccessPathRepo.On("BulkUpsertStudentSubscriptionAccessPath", mock.Anything, tx, mock.Anything).
					Once().Return(nil)
			},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.setup(ctx)

			handler := StudentCourseSlotInfoHandler{
				Logger:                            ctxzap.Extract(ctx),
				DB:                                db,
				JSM:                               jsm,
				UserRepo:                          userRepo,
				StudentSubscriptionRepo:           studentSubRepo,
				StudentSubscriptionAccessPathRepo: studentSubAccessPathRepo,
			}

			msgEvnt, _ := json.Marshal(tc.data)
			res, err := handler.Handle(ctx, msgEvnt)
			if tc.hasError {
				require.Error(t, err)
				require.False(t, res)
			} else {
				require.NoError(t, err)
				require.True(t, res)
			}

			mock.AssertExpectationsForObjects(t, db, tx, userRepo, studentSubRepo, studentSubAccessPathRepo)
		})
	}
}
