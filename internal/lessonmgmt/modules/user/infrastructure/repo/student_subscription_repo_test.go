package repo

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonStudentSubscriptionRepoSqlMock() (*StudentSubscriptionRepo, *testutil.MockDB) {
	r := &StudentSubscriptionRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonStudentSubscriptionRepo_Retrieves(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonStudentSubscriptionRepoSqlMock()
	args := &payloads.ListStudentSubScriptionsArgs{
		Limit:      uint32(2),
		SchoolID:   "5",
		Grades:     []int32{4},
		GradesV2:   []string{"Grade-1"},
		LessonDate: time.Now(),
	}

	t.Run("err select", func(t *testing.T) {

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"),
			mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		studentsSubs, _, _, _, err := r.RetrieveStudentSubscription(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentsSubs)
	})
}

func TestLessonStudentSubscriptionRepo_BulkUpsertStudentSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	sampleData := domain.StudentSubscriptions{
		{
			StudentSubscriptionID: "student-sub-id-1",
			SubscriptionID:        "id-1",
			StudentID:             "student-1",
			StudentFirstName:      "name-1",
			StudentLastName:       "last name-1",
			CourseID:              "course-1",
			PackageType:           "sample-type",
			CourseSlot:            2,
			StartAt:               now,
			EndAt:                 now.Add(24 * time.Hour),
		},
		{
			StudentSubscriptionID: "student-sub-id-2",
			SubscriptionID:        "id-2",
			StudentID:             "student-2",
			StudentFirstName:      "name-2",
			StudentLastName:       "last name-2",
			CourseID:              "course-2",
			PackageType:           "sample-type",
			CourseSlotPerWeek:     2,
			StartAt:               now,
			EndAt:                 now.Add(24 * time.Hour),
		},
		{
			StudentSubscriptionID: "student-sub-id-3",
			SubscriptionID:        "id-3",
			StudentID:             "student-3",
			StudentFirstName:      "name-3",
			StudentLastName:       "last name-3",
			CourseID:              "course-3",
			PackageType:           "sample-type",
			CourseSlot:            2,
			StartAt:               now,
			EndAt:                 now.Add(24 * time.Hour),
		},
	}

	t.Run("err bulk upsert", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscription(ctx, mockDB.DB, sampleData)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, nil).Once()
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscription(ctx, mockDB.DB, sampleData)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("bulk upsert successful", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(sampleData); i++ {
			batchResults.On("Exec").Return(cmdTag, nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscription(ctx, mockDB.DB, sampleData)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}

func TestLessonStudentSubscriptionRepo_GetStudentSubscriptionIDByUniqueIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "student-id-1"
	courseID := "course-id-1"
	subscriptionID := "subscription-id-1"
	var sampleResultString pgtype.Text

	t.Run("error", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &subscriptionID, &studentID, &courseID)
		mockDB.MockRowScanFields(errors.New("error"), []string{"student_subscription_id"}, []interface{}{&sampleResultString})
		studentSubID, err := r.GetStudentSubscriptionIDByUniqueIDs(ctx, mockDB.DB, subscriptionID, studentID, courseID)
		assert.NotNil(t, err)
		assert.Equal(t, "", studentSubID)
	})
	t.Run("success with no rows fetched", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &subscriptionID, &studentID, &courseID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, []string{"student_subscription_id"}, []interface{}{&sampleResultString})
		studentSubID, err := r.GetStudentSubscriptionIDByUniqueIDs(ctx, mockDB.DB, subscriptionID, studentID, courseID)

		assert.Nil(t, err)
		assert.Equal(t, "", studentSubID)
	})
	t.Run("success", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &subscriptionID, &studentID, &courseID)
		mockDB.MockRowScanFields(nil, []string{"student_subscription_id"}, []interface{}{&sampleResultString})
		studentSubID, err := r.GetStudentSubscriptionIDByUniqueIDs(ctx, mockDB.DB, subscriptionID, studentID, courseID)

		assert.Nil(t, err)
		assert.NotNil(t, studentSubID)
	})
}

func TestLessonStudentSubscriptionRepo_UpdateMultiStudentNameByStudents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LessonStudentSubscriptionRepoSqlMock()
	students := user_domain.Users{{ID: "student-1", LastName: "last name", FirstName: "first name"}}

	t.Run("err update", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, errors.New("batchResults.Exec: closed pool"))
		batchResults.On("Close").Once().Return(nil)
		err := l.UpdateMultiStudentNameByStudents(ctx, mockDB.DB, students)
		assert.Equal(t, "batchResults.Exec: batchResults.Exec: closed pool", err.Error())
	})

	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := l.UpdateMultiStudentNameByStudents(ctx, mockDB.DB, students)
		assert.Equal(t, nil, err)
	})
}

func TestLessonStudentSubscriptionRepo_RetrieveStudentPendingReallocate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonStudentSubscriptionRepoSqlMock()
	params := domain.RetrieveStudentPendingReallocateDto{
		Limit:  5,
		Offset: 0,
		// SchoolID: "5",
		// Grades:   []int32{4},
		// GradesV2: []string{"Grade-1"},
		// LessonDate: time.Now(),
	}
	var (
		total            pgtype.Int8
		studentId        pgtype.Text
		originalLessonId pgtype.Text
		courseId         pgtype.Text
		startAt          pgtype.Timestamptz
		endAt            pgtype.Timestamptz
		gradeId          pgtype.Text
		classId          pgtype.Text
		locationId       pgtype.Text
	)

	t.Run("success", func(t *testing.T) {
		fields := []string{"total", "student_id", "original_lesson_id", "course_id", "start_at", "end_at", "grade_id", "class_id", "location_id"}
		values := []interface{}{&total, &studentId, &originalLessonId, &courseId, &startAt, &endAt, &gradeId, &classId, &locationId}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, values)

		var count uint32

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, []string{"total"}, []interface{}{&count})

		rs, _, err := r.RetrieveStudentPendingReallocate(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, rs)
	})
}

func TestLessonStudentSubscriptionRepo_GetStudentCoursesAndClasses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "student_id"
	t.Run("err select", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		var (
			classID    pgtype.Text
			className  pgtype.Text
			courseID   pgtype.Text
			courseName pgtype.Text
		)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), studentID)
		values := []interface{}{&classID, &className, &courseID, &courseName}
		mockDB.MockScanFields(errors.New("error"), []string{"class_id", "class_name", "course_id", "course_name"}, values)

		_, err := r.GetStudentCoursesAndClasses(ctx, mockDB.DB, studentID)
		assert.EqualValues(t, err, fmt.Errorf("rows.Scan: %w", errors.New("error")))
	})

	t.Run("select successfully", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), studentID)
		classID := database.Text("class_id")
		className := database.Text("class_name")
		courseID := database.Text("course_id")
		courseName := database.Text("course_name")
		values := []interface{}{&classID, &className, &courseID, &courseName}
		mockDB.MockScanFields(nil, []string{"class_id", "class_name", "course_id", "course_name"}, values)

		res, err := r.GetStudentCoursesAndClasses(ctx, mockDB.DB, studentID)
		require.NoError(t, err)
		expected := &domain.StudentCoursesAndClasses{
			StudentID: studentID,
			Courses: []*domain.StudentCoursesAndClassesCourses{
				{
					CourseID: courseID.String,
					Name:     courseName.String,
				},
			},
			Classes: []*domain.StudentCoursesAndClassesClasses{
				{
					ClassID:  classID.String,
					Name:     className.String,
					CourseID: courseID.String,
				},
			},
		}
		assert.Equal(t, *expected, *res)
	})

	t.Run("there are no record", func(t *testing.T) {
		r, mockDB := LessonStudentSubscriptionRepoSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), studentID)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(true)

		res, err := r.GetStudentCoursesAndClasses(ctx, mockDB.DB, studentID)
		require.NoError(t, err)
		expected := &domain.StudentCoursesAndClasses{
			StudentID: studentID,
		}
		assert.Equal(t, *expected, *res)
	})
}
