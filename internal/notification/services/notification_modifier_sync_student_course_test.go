package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_UpsertStudentCourse(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	notificationStudentCourseRepo := &mock_repositories.MockNotificationStudentCourseRepo{}

	svc := &NotificationModifierService{
		DB:                            mockDB,
		NotificationStudentCourseRepo: notificationStudentCourseRepo,
	}

	ctx := context.Background()

	t.Run("happy case create", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: true,
			},
		}

		filter := repositories.NewFindNotificationStudentCourseFilter()
		_ = multierr.Combine(
			filter.StudentID.Set(data.StudentPackage.StudentId),
			filter.CourseID.Set(data.StudentPackage.Package.CourseId),
		)

		notificationStudentCourseRepo.On("Find", ctx, mockTx, filter).Once().Return(entities.NotificationStudentCourses{}, nil)
		notificationStudentCourseRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

		err := svc.upsertStudentCourse(ctx, data, mockTx)
		assert.NoError(t, err)
	})

	t.Run("happy case update", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: true,
			},
		}

		filter := repositories.NewFindNotificationStudentCourseFilter()
		_ = multierr.Combine(
			filter.StudentID.Set(data.StudentPackage.StudentId),
			filter.CourseID.Set(data.StudentPackage.Package.CourseId),
		)

		findReturneds := entities.NotificationStudentCourses{
			{
				StudentCourseID: database.Text("student-course-id"),
				CourseID:        database.Text("course-id"),
				StudentID:       database.Text("student-id"),
				LocationID:      database.Text("location-id"),
				StartAt:         database.Timestamptz(time.Now()),
				EndAt:           database.Timestamptz(time.Now()),
			},
		}

		notificationStudentCourseRepo.On("Find", ctx, mockTx, filter).Once().Return(findReturneds, nil)

		studentCourseEnt, err := mappers.EventStudentPackageV2PbToNotificationStudentCourseEnt(data)
		studentCourseEnt.StudentCourseID = database.Text("student-course-id")
		studentCourseEnt.DeletedAt.Set(nil)
		notificationStudentCourseRepo.On("Upsert", ctx, mockTx, studentCourseEnt).Once().Return(nil)

		err = svc.upsertStudentCourse(ctx, data, mockTx)
		assert.NoError(t, err)
	})

	t.Run("case delete all, create 1", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: true,
			},
		}

		filter := repositories.NewFindNotificationStudentCourseFilter()
		_ = multierr.Combine(
			filter.StudentID.Set(data.StudentPackage.StudentId),
			filter.CourseID.Set(data.StudentPackage.Package.CourseId),
		)

		findReturneds := entities.NotificationStudentCourses{
			{
				StudentCourseID: database.Text("student-course-id-1"),
				CourseID:        database.Text("course-id"),
				StudentID:       database.Text("student-id"),
				LocationID:      database.Text("location-id"),
				StartAt:         database.Timestamptz(time.Now()),
				EndAt:           database.Timestamptz(time.Now()),
			},
			{
				StudentCourseID: database.Text("student-course-id-2"),
				CourseID:        database.Text("course-id"),
				StudentID:       database.Text("student-id"),
				LocationID:      database.Text("location-id"),
				StartAt:         database.Timestamptz(time.Now()),
				EndAt:           database.Timestamptz(time.Now()),
			},
		}

		notificationStudentCourseRepo.On("Find", ctx, mockTx, filter).Once().Return(findReturneds, nil)

		softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
		err := multierr.Combine(
			softDeleteFilter.StudentIDs.Set([]string{data.StudentPackage.Package.CourseId}),
			softDeleteFilter.CourseIDs.Set([]string{data.StudentPackage.StudentId}),
		)
		notificationStudentCourseRepo.On("SoftDelete", ctx, mockTx, softDeleteFilter).Once().Return(nil)

		notificationStudentCourseRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

		err = svc.upsertStudentCourse(ctx, data, mockTx)
		assert.NoError(t, err)
	})

	t.Run("case delete", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: false,
			},
		}

		filter := repositories.NewFindNotificationStudentCourseFilter()
		_ = multierr.Combine(
			filter.StudentID.Set(data.StudentPackage.StudentId),
			filter.CourseID.Set(data.StudentPackage.Package.CourseId),
		)

		findReturneds := entities.NotificationStudentCourses{
			{
				StudentCourseID: database.Text("student-course-id-1"),
				CourseID:        database.Text("course-id"),
				StudentID:       database.Text("student-id"),
				LocationID:      database.Text("location-id"),
				StartAt:         database.Timestamptz(time.Now()),
				EndAt:           database.Timestamptz(time.Now()),
			},
			{
				StudentCourseID: database.Text("student-course-id-2"),
				CourseID:        database.Text("course-id"),
				StudentID:       database.Text("student-id"),
				LocationID:      database.Text("location-id"),
				StartAt:         database.Timestamptz(time.Now()),
				EndAt:           database.Timestamptz(time.Now()),
			},
		}

		notificationStudentCourseRepo.On("Find", ctx, mockTx, filter).Once().Return(findReturneds, nil)

		softDeleteFilter := repositories.NewSoftDeleteNotificationStudentCourseFilter()
		err := multierr.Combine(
			softDeleteFilter.StudentIDs.Set([]string{data.StudentPackage.Package.CourseId}),
			softDeleteFilter.CourseIDs.Set([]string{data.StudentPackage.StudentId}),
		)
		notificationStudentCourseRepo.On("SoftDelete", ctx, mockTx, softDeleteFilter).Once().Return(nil)

		notificationStudentCourseRepo.On("Upsert", ctx, mockTx, mock.Anything).Once().Return(nil)

		err = svc.upsertStudentCourse(ctx, data, mockTx)
		assert.NoError(t, err)
	})
}

func Test_UpsertClassMember(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	notificationClassMemberRepo := &mock_repositories.MockNotificationClassMemberRepo{}

	svc := &NotificationModifierService{
		DB:                          mockDB,
		NotificationClassMemberRepo: notificationClassMemberRepo,
	}

	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: true,
			},
		}

		filter := repositories.NewNotificationClassMemberFilter()
		err := multierr.Combine(
			filter.StudentIDs.Set([]string{data.StudentPackage.StudentId}),
			filter.CourseIDs.Set([]string{data.StudentPackage.Package.CourseId}),
		)
		notificationClassMemberRepo.On("SoftDeleteByFilter", ctx, mockTx, filter).Once().Return(nil)

		notiClassMember, _ := mappers.EventStudentPackageV2PbToNotificationClassMemberEnt(data)

		notificationClassMemberRepo.On("Upsert", ctx, mockTx, notiClassMember).Once().Return(nil)

		err = svc.upsertClassMember(ctx, data, mockTx)
		assert.NoError(t, err)
	})

	t.Run("happy case delete (is active == false)", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					ClassId:    "class-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: false,
			},
		}

		filter := repositories.NewNotificationClassMemberFilter()
		err := multierr.Combine(
			filter.StudentIDs.Set([]string{data.StudentPackage.StudentId}),
			filter.CourseIDs.Set([]string{data.StudentPackage.Package.CourseId}),
		)
		notificationClassMemberRepo.On("SoftDeleteByFilter", ctx, mockTx, filter).Once().Return(nil)

		notiClassMember, _ := mappers.EventStudentPackageV2PbToNotificationClassMemberEnt(data)
		notificationClassMemberRepo.On("Upsert", ctx, mockTx, notiClassMember).Once().Return(nil)

		err = svc.upsertClassMember(ctx, data, mockTx)
		assert.NoError(t, err)
	})

	t.Run("happy case delete (missing class_id)", func(t *testing.T) {
		mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		data := &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "student-id",
				Package: &npb.EventStudentPackageV2_PackageV2{
					CourseId:   "course-id",
					LocationId: "location-id",
					StartDate:  timestamppb.Now(),
					EndDate:    timestamppb.Now(),
				},
				IsActive: false,
			},
		}

		filter := repositories.NewNotificationClassMemberFilter()
		err := multierr.Combine(
			filter.StudentIDs.Set([]string{data.StudentPackage.StudentId}),
			filter.CourseIDs.Set([]string{data.StudentPackage.Package.CourseId}),
		)
		notificationClassMemberRepo.On("SoftDeleteByFilter", ctx, mockTx, filter).Once().Return(nil)

		err = svc.upsertClassMember(ctx, data, mockTx)
		assert.NoError(t, err)
	})
}
