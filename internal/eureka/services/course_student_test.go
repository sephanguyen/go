package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_repositories "github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCourseStudentService_SyncStudentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := new(mock_nats.JetStreamManagement)

	t.Run("[ActionKind UPSERT] should create CourseStudent success", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		c := &CourseStudentService{
			DB:                   db,
			JSM:                  jsm,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
			CourseStudyPlanRepo:  courseStudyPlanRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId").Once().
			Return(nil)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, "studentId")
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[0]),
				StudentID:  database.Text("studentId"),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[1]),
				StudentID:  database.Text("studentId"),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			expected[1].CreatedAt = s[1].CreatedAt
			expected[1].UpdatedAt = s[1].UpdatedAt
			expected[1].DeletedAt.Set(nil)
			expected[1].ID = s[1].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()

		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseIDs[0]),
				StudyPlanID: database.Text("study-plan-id"),
			},
			{
				CourseID:    database.Text(courseIDs[1]),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}
		studyPlans := []*entities.StudyPlan{
			{
				ID:              database.Text("study-plan-id-1"),
				MasterStudyPlan: database.Text("study-plan-id"),
			},
		}
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)
		studentStudyPlanRepo.On("FindAllStudentStudyPlan", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(studyPlans, nil)
		studentStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		// Action
		err := c.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: courseIDs,
							StartDate: startAt,
							EndDate:   endAt,
						},
						{
							CourseIds: courseIDs,
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
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}

		c := &CourseStudentService{
			DB:                   db,
			JSM:                  jsm,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Second()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Second()),
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId").Once().
			Return(nil)

		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, "studentId")
			}).
			Return(courseStudentMap, errors.New("error insert"))

		// Action
		err := c.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: courseIDs,
							StartDate: startAt,
							EndDate:   endAt,
						},
						{
							CourseIds: courseIDs,
							StartDate: startAt,
							EndDate:   endAt,
						},
					},
				},
			},
		})

		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err upsert student course of studentID studentId: err s.CourseStudentRepo.BulkUpsert: error insert")
	})

	t.Run("[ActionKind DELETE] should soft delete CourseStudent", func(t *testing.T) {
		// Arrange
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}

		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		c := &CourseStudentService{
			DB:                   db,
			JSM:                  jsm,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
		}
		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		studentStudyPlanRepo.On("FindStudentStudyPlanWithCourseIDs", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, courseIDs, actualCourseIds)
			}).
			Return(nil)

		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		// Action
		err := c.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: courseIDs,
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

	t.Run("[ActionKind DELETE] should throw error", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		c := &CourseStudentService{
			DB:                db,
			JSM:               jsm,
			CourseStudentRepo: courseStudentRepo,
		}
		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, courseIDs, actualCourseIds)
			}).
			Return(errors.New("error softdelete"))

		// Action
		err := c.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
				{
					StudentId:  "studentId",
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					Packages: []*npb.EventSyncStudentPackage_Package{
						{
							CourseIds: courseIDs,
							StartDate: startAt,
							EndDate:   endAt,
						},
					},
				},
			},
		})

		// Assert
		assert.NotNil(t, err)
		assert.EqualError(t, err, "err soft delete student course of studentID studentId: err s.CourseStudentRepo.SoftDelete: error softdelete")
	})

	t.Run("[ActionKind UPSERT] should continue even if any student doesn't have any courses", func(t *testing.T) {
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		c := &CourseStudentService{
			DB:                   db,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
			CourseStudyPlanRepo:  courseStudyPlanRepo,
			JSM:                  jsm,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		now := timestamppb.Now()
		courseIDs := []string{"courseID1", "courseID2"}
		studentPackages := []*npb.EventSyncStudentPackage_StudentPackage{
			{
				StudentId:  "studentId1",
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: courseIDs,
						StartDate: now,
						EndDate:   now,
					},
					{
						CourseIds: courseIDs,
						StartDate: now,
						EndDate:   now,
					},
				},
			},
			{
				StudentId:  "studentId2",
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			},
			{
				StudentId:  "studentId3",
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
				Packages: []*npb.EventSyncStudentPackage_Package{
					{
						CourseIds: courseIDs,
						StartDate: now,
						EndDate:   now,
					},
				},
			},
			{
				StudentId:  "studentId4",
				ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
			},
		}
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)
		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId1").Once().Return(nil)
		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId2").Once().Return(nil)
		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId3").Once().Return(nil)
		courseStudentRepo.On("SoftDeleteByStudentID", mock.Anything, tx, "studentId4").Once().Return(nil)
		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Twice().Return(courseStudentMap, nil)
		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[0]),
				StudentID:  database.Text("studentId1"),
				StartAt:    database.Timestamptz(now.AsTime()),
				EndAt:      database.Timestamptz(now.AsTime()),
			},
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[1]),
				StudentID:  database.Text("studentId1"),
				StartAt:    database.Timestamptz(now.AsTime()),
				EndAt:      database.Timestamptz(now.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			expected[1].CreatedAt = s[1].CreatedAt
			expected[1].UpdatedAt = s[1].UpdatedAt
			expected[1].DeletedAt.Set(nil)
			expected[1].ID = s[1].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			expected[0].StudentID = database.Text("studentId3")
			expected[0].CourseID = database.Text(courseIDs[0])
			expected[1].StudentID = database.Text("studentId3")
			expected[1].CourseID = database.Text(courseIDs[1])

			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			expected[1].CreatedAt = s[1].CreatedAt
			expected[1].UpdatedAt = s[1].UpdatedAt
			expected[1].DeletedAt.Set(nil)
			expected[1].ID = s[1].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Twice().Return("", nil)
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Twice().Return(nil, nil)

		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Times(4).Return(nil, nil)
		studentStudyPlanRepo.On("FindAllStudentStudyPlan", mock.Anything, tx, mock.Anything, mock.Anything).Twice().Return(nil, nil)
		studentStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Twice().Return(nil)

		studyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Twice().Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		err := c.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: studentPackages,
		})

		assert.Nil(t, err)

		courseStudentRepo.AssertNumberOfCalls(t, "SoftDeleteByStudentID", 4)
		courseStudentRepo.AssertNumberOfCalls(t, "BulkUpsert", 2)

		studentStudyPlanRepo.AssertNumberOfCalls(t, "FindByStudentIDs", 4)

		courseStudyPlanRepo.AssertNumberOfCalls(t, "FindByCourseIDs", 2)

		studentStudyPlanRepo.AssertNumberOfCalls(t, "FindAllStudentStudyPlan", 2)
		studentStudyPlanRepo.AssertNumberOfCalls(t, "BulkUpsert", 2)

		studyPlanRepo.AssertNumberOfCalls(t, "BulkUpsert", 2)
	})
}

func TestCourseStudentService_HandleStudentPackageEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	t.Run("status active should create CourseStudent success", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}
		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:                      zap.NewNop(),
			JSM:                         jsm,
			DB:                          db,
			CourseStudentRepo:           courseStudentRepo,
			StudentStudyPlanRepo:        studentStudyPlanRepo,
			StudyPlanRepo:               studyPlanRepo,
			CourseStudyPlanRepo:         courseStudyPlanRepo,
			CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now().UTC()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Contains(t, courseIDs, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, "studentId")
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[0]),
				StudentID:  database.Text("studentId"),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseIDs[1]),
				StudentID:  database.Text("studentId"),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			expected[1].CreatedAt = s[1].CreatedAt
			expected[1].UpdatedAt = s[1].UpdatedAt
			expected[1].DeletedAt.Set(nil)
			expected[1].ID = s[1].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageEventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Once().Return("", nil)

		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseIDs[0]),
				StudyPlanID: database.Text("study-plan-id"),
			},
			{
				CourseID:    database.Text(courseIDs[1]),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}
		studyPlans := []*entities.StudyPlan{
			{
				ID:              database.Text("study-plan-id-1"),
				MasterStudyPlan: database.Text("study-plan-id"),
			},
		}
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)
		studentStudyPlanRepo.On("FindAllStudentStudyPlan", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(studyPlans, nil)
		studentStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		courseStudentAccessPathRepo.On("DeleteLatestCourseStudentAccessPathsByCourseStudentIDs", mock.Anything, tx, mock.Anything).Return(nil)
		courseStudentAccessPathRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Return(nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		// Action
		err := c.HandleStudentPackageEvent(ctx, &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: "studentId",
				Package: &npb.EventStudentPackage_Package{
					LocationIds: []string{"location-1", "location-2"},
					CourseIds:   courseIDs,
					StartDate:   startAt,
					EndDate:     endAt,
				},
				IsActive: true,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("status inactive should delete CourseStudent success", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:               zap.NewNop(),
			DB:                   db,
			JSM:                  jsm,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
			CourseStudyPlanRepo:  courseStudyPlanRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseIDs[0]),
				StudyPlanID: database.Text("study-plan-id"),
			},
			{
				CourseID:    database.Text(courseIDs[1]),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)

		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, courseIDs, actualCourseIds)
			}).
			Return(nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		// Action
		err := c.HandleStudentPackageEvent(ctx, &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: "studentId",
				Package: &npb.EventStudentPackage_Package{
					LocationIds: []string{"location-1", "location-2"},
					CourseIds:   courseIDs,
					StartDate:   startAt,
					EndDate:     endAt,
				},
				IsActive: false,
			},
		})
		assert.NoError(t, err)
	})
	t.Run("status inactive should delete CourseStudent throw error", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:               zap.NewNop(),
			DB:                   db,
			JSM:                  jsm,
			CourseStudentRepo:    courseStudentRepo,
			StudentStudyPlanRepo: studentStudyPlanRepo,
			StudyPlanRepo:        studyPlanRepo,
			CourseStudyPlanRepo:  courseStudyPlanRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		timeNow := time.Now()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseIDs := []string{"courseID1", "courseID2"}
		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseIDs[0]),
				StudyPlanID: database.Text("study-plan-id"),
			},
			{
				CourseID:    database.Text(courseIDs[1]),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)

		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, courseIDs, actualCourseIds)
			}).
			Return(nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(errors.New("error delete"))
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		// Action
		err := c.HandleStudentPackageEvent(ctx, &npb.EventStudentPackage{
			StudentPackage: &npb.EventStudentPackage_StudentPackage{
				StudentId: "studentId",
				Package: &npb.EventStudentPackage_Package{
					CourseIds: courseIDs,
					StartDate: startAt,
					EndDate:   endAt,
				},
				IsActive: false,
			},
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "database.ExecInTxWithRetry: s.softDeleteStudentStudyPlanByCourseStudent: StudyPlanRepo.SoftDelete: error delete")
	})
}

func TestCourseStudentService_HandleStudentPackageEventV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	t.Run("status active should create CourseStudent, ClassStudent success", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}

		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:                      zap.NewNop(),
			JSM:                         jsm,
			DB:                          db,
			CourseStudentRepo:           courseStudentRepo,
			StudentStudyPlanRepo:        studentStudyPlanRepo,
			StudyPlanRepo:               studyPlanRepo,
			CourseStudyPlanRepo:         courseStudyPlanRepo,
			CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now().UTC()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseID := "courseID1"
		classID := "classID"
		studentID := "studentId"
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Equal(t, courseID, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, studentID)
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseID),
				StudentID:  database.Text(studentID),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Once().Return("", nil)

		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseID),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}

		studyPlans := []*entities.StudyPlan{
			{
				ID:              database.Text("study-plan-id-1"),
				MasterStudyPlan: database.Text("study-plan-id"),
			},
		}
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)
		studentStudyPlanRepo.On("FindAllStudentStudyPlan", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(studyPlans, nil)
		studentStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		courseStudentAccessPathRepo.On("DeleteLatestCourseStudentAccessPathsByCourseStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		courseStudentAccessPathRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		// Action
		err := c.HandleStudentPackageEventV2(ctx, &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "studentId",
				Package: &npb.EventStudentPackageV2_PackageV2{
					ClassId:    classID,
					LocationId: "locationId",
					CourseId:   courseID,
					StartDate:  startAt,
					EndDate:    endAt,
				},
				IsActive: true,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("status active should create CourseStudent, delete ClassStudent when class_id empty", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}

		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:                      zap.NewNop(),
			JSM:                         jsm,
			DB:                          db,
			CourseStudentRepo:           courseStudentRepo,
			StudentStudyPlanRepo:        studentStudyPlanRepo,
			StudyPlanRepo:               studyPlanRepo,
			CourseStudyPlanRepo:         courseStudyPlanRepo,
			CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now().UTC()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseID := "courseID1"
		studentID := "studentId"
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Equal(t, courseID, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, studentID)
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseID),
				StudentID:  database.Text(studentID),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Once().Return("", nil)

		courseStudyPlan := []*entities.CourseStudyPlan{
			{
				CourseID:    database.Text(courseID),
				StudyPlanID: database.Text("study-plan-id"),
			},
		}

		studyPlans := []*entities.StudyPlan{
			{
				ID:              database.Text("study-plan-id-1"),
				MasterStudyPlan: database.Text("study-plan-id"),
			},
		}
		courseStudyPlanRepo.On("FindByCourseIDs", mock.Anything, tx, mock.Anything).Once().Return(courseStudyPlan, nil)
		studentStudyPlanRepo.On("FindAllStudentStudyPlan", mock.Anything, tx, mock.Anything, mock.Anything).Once().Return(studyPlans, nil)
		studentStudyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		jsm.On("PublishAsyncContext", mock.Anything, mock.Anything, mock.Anything).Once().Return("", nil)
		courseStudentAccessPathRepo.On("DeleteLatestCourseStudentAccessPathsByCourseStudentIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		courseStudentAccessPathRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		// Action
		err := c.HandleStudentPackageEventV2(ctx, &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "studentId",
				Package: &npb.EventStudentPackageV2_PackageV2{
					//ClassId:    classID,
					LocationId: "locationId",
					CourseId:   courseID,
					StartDate:  startAt,
					EndDate:    endAt,
				},
				IsActive: true,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("status inactive should delete CourseStudent, ClassStudent success", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		classStudentRepo := &mock_repositories.MockClassStudentRepo{}
		courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}

		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:                      zap.NewNop(),
			JSM:                         jsm,
			DB:                          db,
			CourseStudentRepo:           courseStudentRepo,
			StudentStudyPlanRepo:        studentStudyPlanRepo,
			StudyPlanRepo:               studyPlanRepo,
			CourseStudyPlanRepo:         courseStudyPlanRepo,
			CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Commit", mock.Anything).Once().Return(nil)

		timeNow := time.Now().UTC()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseID := "courseID1"
		classID := "classID"
		studentID := "studentId"
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Equal(t, courseID, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, studentID)
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseID),
				StudentID:  database.Text(studentID),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Once().Return("", nil)

		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, []string{courseID}, actualCourseIds)
			}).
			Return(nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)
		classStudentRepo.On("BulkSoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)

		// Action
		err := c.HandleStudentPackageEventV2(ctx, &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "studentId",
				Package: &npb.EventStudentPackageV2_PackageV2{
					ClassId:    classID,
					LocationId: "locationId",
					CourseId:   courseID,
					StartDate:  startAt,
					EndDate:    endAt,
				},
				IsActive: false,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("status inactive should delete CourseStudent ClassStudent throw error", func(t *testing.T) {
		// Arrange
		courseStudentRepo := &mock_repositories.MockCourseStudentRepo{}
		studentStudyPlanRepo := &mock_repositories.MockStudentStudyPlanRepo{}
		studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
		courseStudyPlanRepo := &mock_repositories.MockCourseStudyPlanRepo{}
		courseStudentAccessPathRepo := &mock_repositories.MockCourseStudentAccessPathRepo{}

		jsm := new(mock_nats.JetStreamManagement)
		c := &CourseStudentService{
			Logger:                      zap.NewNop(),
			JSM:                         jsm,
			DB:                          db,
			CourseStudentRepo:           courseStudentRepo,
			StudentStudyPlanRepo:        studentStudyPlanRepo,
			StudyPlanRepo:               studyPlanRepo,
			CourseStudyPlanRepo:         courseStudyPlanRepo,
			CourseStudentAccessPathRepo: courseStudentAccessPathRepo,
		}

		db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
		tx.On("Rollback", mock.Anything).Once().Return(nil)

		timeNow := time.Now().UTC()
		startAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, -1).Unix()),
		}
		endAt := &timestamppb.Timestamp{
			Seconds: int64(timeNow.AddDate(0, 0, 1).Unix()),
		}
		courseID := "courseID1"
		classID := "classID"
		studentID := "studentId"
		courseStudentMap := make(map[eureka_repositories.CourseStudentKey]string)

		courseStudentRepo.On("BulkUpsert", mock.Anything, tx, mock.AnythingOfType("[]*entities.CourseStudent")).
			Run(func(args mock.Arguments) {
				s := args[2].([]*entities.CourseStudent)
				assert.Equal(t, courseID, s[0].CourseID.String)
				assert.Equal(t, s[0].StudentID.String, studentID)
			}).
			Return(courseStudentMap, nil).Once()

		expected := entities.CourseStudents{
			{
				BaseEntity: entities.BaseEntity{},
				ID:         pgtype.Text{},
				CourseID:   database.Text(courseID),
				StudentID:  database.Text(studentID),
				StartAt:    database.Timestamptz(startAt.AsTime()),
				EndAt:      database.Timestamptz(endAt.AsTime()),
			},
		}
		courseStudentRepo.On("GetByCourseStudents", mock.Anything, tx, mock.Anything).Run(func(args mock.Arguments) {
			s := args[2].(entities.CourseStudents)
			assert.NotEmpty(t, s[0].ID)
			assert.NotEmpty(t, s[1].ID)
			expected[0].CreatedAt = s[0].CreatedAt
			expected[0].UpdatedAt = s[0].UpdatedAt
			expected[0].DeletedAt.Set(nil)
			expected[0].ID = s[0].ID
			assert.Equal(t, expected, s)
		}).
			Return(expected, nil).Once()

		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectStudentPackageV2EventNats, mock.Anything).Run(func(args mock.Arguments) {
			data := args[2].([]byte)
			mess := &npb.SyncStudentSubscriptionJobData{}
			err := proto.Unmarshal(data, mess)
			require.NoError(t, err)
			assert.Lenf(t, mess.CourseStudents, len(expected), "missing some student subscription item when PublishAsyncContext")
			for i, actual := range mess.CourseStudents {
				assert.Equal(t, expected[i].ID.String, actual.CourseStudentId)
				assert.Equal(t, expected[i].CourseID.String, actual.CourseId)
				assert.Equal(t, expected[i].StudentID.String, actual.StudentId)
				assert.Equal(t, expected[i].StartAt.Time, actual.StartAt.AsTime())
				assert.Equal(t, expected[i].EndAt.Time, actual.EndAt.AsTime())
			}
		}).Once().Return("", nil)

		courseStudentRepo.On("SoftDelete", mock.Anything, tx, mock.AnythingOfType("[]string"), mock.AnythingOfType("[]string")).
			Run(func(args mock.Arguments) {
				actualCourseIds := args[3].([]string)
				assert.Equal(t, []string{courseID}, actualCourseIds)
			}).
			Return(nil)
		studentStudyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(nil)
		studyPlanRepo.On("SoftDelete", mock.Anything, tx, mock.Anything).Once().Return(errors.New("error delete"))
		studentStudyPlanRepo.On("FindByStudentIDs", mock.Anything, tx, mock.Anything).Once().Return([]string{"student-study-plan-id"}, nil)

		// Action
		err := c.HandleStudentPackageEventV2(ctx, &npb.EventStudentPackageV2{
			StudentPackage: &npb.EventStudentPackageV2_StudentPackageV2{
				StudentId: "studentId",
				Package: &npb.EventStudentPackageV2_PackageV2{
					ClassId:    classID,
					LocationId: "locationId",
					CourseId:   courseID,
					StartDate:  startAt,
					EndDate:    endAt,
				},
				IsActive: false,
			},
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "database.ExecInTx: s.softDeleteStudentStudyPlanByCourseStudent: StudyPlanRepo.SoftDelete: error delete")
	})
}

func TestCourseStudentService_convStudentPackageToCourseStudents(t *testing.T) {
	var nilEventStudentPackage *npb.EventStudentPackage
	testCases := []TestCase{
		{
			name:        "nil request",
			req:         nilEventStudentPackage,
			expectedErr: fmt.Errorf("empty request"),
		},
		{
			name:        "nil student package",
			req:         &npb.EventStudentPackage{},
			expectedErr: fmt.Errorf("empty request"),
		},
		{
			name: "nil package",
			req: &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: "mock-student-id",
					IsActive:  true,
				},
			},
			expectedErr: fmt.Errorf("empty request"),
		},
		{
			name: "nil courses",
			req: &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: "mock-student-id",
					IsActive:  true,
					Package: &npb.EventStudentPackage_Package{
						StartDate: timestamppb.Now(),
					},
				},
			},
			expectedErr: fmt.Errorf("empty request"),
		},
		{
			name: "happy case",
			req: &npb.EventStudentPackage{
				StudentPackage: &npb.EventStudentPackage_StudentPackage{
					StudentId: "mock-student-id",
					IsActive:  true,
					Package: &npb.EventStudentPackage_Package{
						CourseIds: []string{"course-id"},
						StartDate: timestamppb.Now(),
						EndDate:   timestamppb.Now(),
					},
				},
			},
			expectedErr: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.req.(*npb.EventStudentPackage)
			_, _, err := convStudentPackageToCourseStudents(req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_upsertCourseStudent(t *testing.T) {
	t.Parallel()
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.CourseStudent{
				{
					StudentID: database.Text("mock-id-1"),
				},
				{
					StudentID: database.Text("mock-id-2"),
				},
				{
					StudentID: database.Text("mock-id-2"),
				},
			},
			expectedErr:  nil,
			expectedResp: []string{"mock-id-1", "mock-id-2"},
		},
		{
			name:         "empty request",
			req:          []*entities.CourseStudent{},
			expectedErr:  nil,
			expectedResp: []string{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase := testCase
			req := testCase.req.([]*entities.CourseStudent)
			expectedResp := testCase.expectedResp.([]string)
			resp := retrieveStudentIDsFromCourseStudents(req)
			assert.ElementsMatch(t, expectedResp, resp)
		})
	}
}

func Test_retrieveStudentIDsFromEventSyncStudentPackage(t *testing.T) {
	t.Parallel()
	testCases := []TestCase{
		{
			name: "empty request",
			req:  &npb.EventSyncStudentPackage{},

			expectedErr:  nil,
			expectedResp: []string{},
		},
		{
			name: "empty request",
			req: &npb.EventSyncStudentPackage{
				StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
					{
						StudentId: "mock-student-id-1",
					},
					{
						StudentId: "mock-student-id-2",
					},
					{
						StudentId: "mock-student-id-2",
					},
				},
			},
			expectedErr:  nil,
			expectedResp: []string{"mock-student-id-1", "mock-student-id-2"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase := testCase
			req := testCase.req.(*npb.EventSyncStudentPackage)
			expectedResp := testCase.expectedResp.([]string)
			resp := retrieveStudentIDsFromEventSyncStudentPackage(req)
			assert.ElementsMatch(t, expectedResp, resp)
		})
	}
}

func Test_constructMapSyncCourseStudent(t *testing.T) {
	t.Parallel()
	testCases := []TestCase{
		{
			name: "empty request",
			req:  &npb.EventSyncStudentPackage{},

			expectedErr:  nil,
			expectedResp: courseStudentSyncInfo{},
		},
		{
			name: "empty course in upsert action",
			req: &npb.EventSyncStudentPackage{
				StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
					{
						StudentId:  "mock-student-id-1",
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					},
				},
			},
			expectedErr:  fmt.Errorf("constructMapSyncCourseStudent: Upsert Action Kind expect exist courseStudents"),
			expectedResp: courseStudentSyncInfo{},
		},
		{
			name: "empty course in delete action",
			req: &npb.EventSyncStudentPackage{
				StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
					{
						StudentId:  "mock-student-id-1",
						ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					},
				},
			},
			expectedErr:  fmt.Errorf("constructMapSyncCourseStudent: Delete Action Kind expect exist courseStudents"),
			expectedResp: courseStudentSyncInfo{},
		},
		{
			name: "upsert action with course",
			req: &npb.EventSyncStudentPackage{
				StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
					{
						StudentId:  "mock-student-id-1",
						ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
						Packages:   []*npb.EventSyncStudentPackage_Package{{CourseIds: []string{"course-1", "course-2"}}},
					},
				},
			},
			expectedErr: nil,
			expectedResp: courseStudentSyncInfo{
				courseStudent: []*entities.CourseStudent{
					{
						CourseID:  database.Text("course-1"),
						StudentID: database.Text("mock-student-id-1"),
					},
					{
						CourseID:  database.Text("course-2"),
						StudentID: database.Text("mock-student-id-1"),
					},
				},
				studentIDs: []string{"mock-student-id-1"},
				courseIDs:  nil,
			},
		},
		{
			name: "delete action with course",
			req: &npb.EventSyncStudentPackage{
				StudentPackages: []*npb.EventSyncStudentPackage_StudentPackage{
					{
						StudentId:  "mock-student-id-1",
						ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
						Packages:   []*npb.EventSyncStudentPackage_Package{{CourseIds: []string{"course-1", "course-2"}}},
					},
				},
			},
			expectedErr: nil,
			expectedResp: courseStudentSyncInfo{
				courseStudent: nil,
				studentIDs:    []string{"mock-student-id-1", "mock-student-id-1"},
				courseIDs:     []string{"course-1", "course-2"},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase := testCase
			req := testCase.req.(*npb.EventSyncStudentPackage)
			expectedResp := testCase.expectedResp.(courseStudentSyncInfo)
			resp, err := constructMapSyncCourseStudent(req)
			if err != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if val, ok := resp[npb.ActionKind_ACTION_KIND_UPSERTED]; ok {
				assert.Equal(t, expectedResp.courseIDs, val.courseIDs)
				assert.Equal(t, expectedResp.studentIDs, val.studentIDs)
			}
			if val, ok := resp[npb.ActionKind_ACTION_KIND_DELETED]; ok {
				assert.Equal(t, expectedResp.courseIDs, val.courseIDs)
				assert.Equal(t, expectedResp.studentIDs, val.studentIDs)
			}

		})
	}

}
