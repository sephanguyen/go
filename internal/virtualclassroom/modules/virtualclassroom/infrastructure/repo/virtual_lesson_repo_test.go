package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
	vl_payloads "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/application/queries/payloads"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func VirtualLessonRepoWithSqlMock() (*VirtualLessonRepo, *testutil.MockDB) {
	r := &VirtualLessonRepo{}
	return r, testutil.NewMockDB()
}

func TestVirtualLessonRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualLessonRepoWithSqlMock()
	// virtual lesson
	e := VirtualLesson{}
	fields, values := e.FieldMap()
	// mock lesson-teacher
	lessonTeacher := &LessonTeacher{}
	lessonTeacherFields, lessonTeacherValues := lessonTeacher.FieldMap()

	// mock lesson-member
	lessonMemberDTO := &LessonMemberDTO{}
	lessonMemberFields, lessonMemberValues := lessonMemberDTO.FieldMap()
	id := "test-lesson-id-1"

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		lessons, err := r.GetVirtualLessonByID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&id,
		)

		//lesson-member
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonMemberFields, lessonMemberValues)
		//lesson-teacher
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, lessonTeacherFields, lessonTeacherValues)

		e.LessonID = database.Text(id)
		mockDB.MockRowScanFields(nil, fields, values)
		lessons, err := r.GetVirtualLessonByID(ctx, mockDB.DB, id)
		assert.NoError(t, err)
		assert.NotNil(t, lessons)
	})
}

func TestVirtualLessonRepo_GetVirtualLessonOnlyByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := VirtualLessonRepoWithSqlMock()

	e := VirtualLesson{}
	fields, values := e.FieldMap()
	id := "test-lesson-id-1"

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &id)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		lessons, err := r.GetVirtualLessonOnlyByID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, lessons)
	})

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), &id)
		mockDB.MockRowScanFields(nil, fields, values)

		lessons, err := r.GetVirtualLessonOnlyByID(ctx, mockDB.DB, id)
		assert.NoError(t, err)
		assert.NotNil(t, lessons)
	})
}

func TestVirtualLessonRepo_GetVirtualLessonByLessonIDsAndCourseIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
	mockLesson := &VirtualLesson{}
	fields, values := mockLesson.FieldMap()

	lessonIDs := []string{"lesson-id1", "lesson-id2"}
	courseIDs := []string{"course-id1", "course-id2"}

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonIDs, &courseIDs)
		mockDB.MockScanFields(nil, fields, values)

		lessons, err := virtualLessonRepo.GetVirtualLessonByLessonIDsAndCourseIDs(ctx, mockDB.DB, lessonIDs, courseIDs)
		assert.NoError(t, err)
		assert.NotNil(t, lessons)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonIDs, &courseIDs)

		lessons, err := virtualLessonRepo.GetVirtualLessonByLessonIDsAndCourseIDs(ctx, mockDB.DB, lessonIDs, courseIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})
}

func TestVirtualLessonRepo_GetVirtualLessonsByLessonIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
	mockLesson := &VirtualLesson{}
	fields, values := mockLesson.FieldMap()

	lessonIDs := []string{"lesson-id1", "lesson-id2"}

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &lessonIDs)
		mockDB.MockScanFields(nil, fields, values)

		lessons, err := virtualLessonRepo.GetVirtualLessonsByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.NoError(t, err)
		assert.NotNil(t, lessons)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &lessonIDs)

		lessons, err := virtualLessonRepo.GetVirtualLessonsByLessonIDs(ctx, mockDB.DB, lessonIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})
}

func TestVirtualLessonRepo_UpdateRoomID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	roomID := "room-id1"
	lessonID := "lesson-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&roomID,
		&lessonID)

	t.Run("upsert failed", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := virtualLessonRepo.UpdateRoomID(ctx, mockDB.DB, lessonID, roomID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.UpdateRoomID(ctx, mockDB.DB, lessonID, roomID)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.UpdateRoomID(ctx, mockDB.DB, lessonID, roomID)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestVirtualLessonRepo_EndLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id1"
	endTime := time.Now()
	var endAt pgtype.Timestamptz
	endAt.Set(endTime)

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&endAt,
		&lessonID)

	t.Run("upsert failed", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := virtualLessonRepo.EndLiveLesson(ctx, mockDB.DB, lessonID, endTime)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.EndLiveLesson(ctx, mockDB.DB, lessonID, endTime)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.EndLiveLesson(ctx, mockDB.DB, lessonID, endTime)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestVirtualLessonRepo_GetStreamingLearners(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var learnerIDs pgtype.TextArray
	fields, values := []string{"learner_ids"}, []interface{}{&learnerIDs}
	lessonID := "lesson-id1"

	t.Run("successful", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := virtualLessonRepo.GetStreamingLearners(ctx, mockDB.DB, lessonID, true)
		assert.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &lessonID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		_, err := virtualLessonRepo.GetStreamingLearners(ctx, mockDB.DB, lessonID, true)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})
}

func TestVirtualLessonRepo_IncreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	learnerID := "student-id1"
	lessonID := "lesson-id1"
	maximumLearnerStreamings := 20

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&learnerID,
		&lessonID,
		&maximumLearnerStreamings)

	t.Run("update failed", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := virtualLessonRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID, maximumLearnerStreamings)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("no rows affected after update", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID, maximumLearnerStreamings)
		assert.EqualError(t, err, "no rows updated")
	})

	t.Run("update successful", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.IncreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID, maximumLearnerStreamings)
		assert.Nil(t, err)
	})
}

func TestVirtualLessonRepo_DecreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	learnerID := "student-id1"
	lessonID := "lesson-id1"

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		&learnerID,
		&lessonID)

	t.Run("update failed", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := virtualLessonRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("no rows affected after update", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID)
		assert.EqualError(t, err, "no rows updated")
	})

	t.Run("update successful", func(t *testing.T) {
		virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := virtualLessonRepo.DecreaseNumberOfStreaming(ctx, mockDB.DB, lessonID, learnerID)
		assert.Nil(t, err)
	})
}

func TestVirtualLessonRepo_GetVirtualLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	virtualLessonRepo, mockDB := VirtualLessonRepoWithSqlMock()
	lesson := &VirtualLesson{}
	fields, values := lesson.FieldMap()

	var total pgtype.Int8
	totalFields := []string{"total"}
	totalValues := []interface{}{&total}

	now := time.Now()
	studentIDs := []string{"student_id1", "student_id2"}
	courseIDs := []string{"course_id1", "course_id2"}
	locationIDs := []string{"location_id1", "location_id2"}
	statuses := []domain.LessonSchedulingStatus{domain.LessonSchedulingStatusPublished}
	startDate, endDate := now, now.Add(24*time.Hour)
	page, limit := int32(1), int32(10)

	t.Run("successful with all parameters", func(t *testing.T) {
		payload := &vl_payloads.GetVirtualLessonsArgs{
			StudentIDs:               studentIDs,
			CourseIDs:                courseIDs,
			LocationIDs:              locationIDs,
			StartDate:                startDate,
			EndDate:                  endDate,
			LessonSchedulingStatuses: statuses,
			Limit:                    limit,
			Page:                     page,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // location
			mock.Anything,                 // student
			mock.Anything,                 // course
			mock.Anything,                 // statuses
			mock.Anything,                 // end date
			mock.Anything,                 // start date
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // limit
			mock.Anything, // page
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, values)

		lessons, total, err := virtualLessonRepo.GetVirtualLessons(ctx, mockDB.DB, payload)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int32(0))
		assert.NotNil(t, lessons)
	})

	t.Run("successful with only locations", func(t *testing.T) {
		payload := &vl_payloads.GetVirtualLessonsArgs{
			LocationIDs: locationIDs,
			Limit:       limit,
			Page:        page,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // location
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // limit
			mock.Anything, // page
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, values)

		lessons, total, err := virtualLessonRepo.GetVirtualLessons(ctx, mockDB.DB, payload)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int32(0))
		assert.NotNil(t, lessons)
	})

	t.Run("failed", func(t *testing.T) {
		payload := &vl_payloads.GetVirtualLessonsArgs{
			StudentIDs:               studentIDs,
			CourseIDs:                courseIDs,
			LocationIDs:              locationIDs,
			StartDate:                startDate,
			EndDate:                  endDate,
			LessonSchedulingStatuses: statuses,
			Limit:                    limit,
			Page:                     page,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // location
			mock.Anything,                 // student
			mock.Anything,                 // course
			mock.Anything,                 // statuses
			mock.Anything,                 // end date
			mock.Anything,                 // start date
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // limit
			mock.Anything, // page
		)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, queryArgs...)

		users, total, err := virtualLessonRepo.GetVirtualLessons(ctx, mockDB.DB, payload)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Equal(t, int32(0), total)
		assert.Nil(t, users)
	})
}

func TestVirtualLessonRepo_GetLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().UTC()

	var total pgtype.Int8
	totalFields := []string{"total"}
	totalValues := []interface{}{&total}

	var preTotalLessons pgtype.Int8
	var preOffset pgtype.Text
	offsetFields := []string{"offset_lesson_id", "pre_total"}
	offsetValues := []interface{}{&preOffset, &preTotalLessons}

	lesson := &VirtualLesson{}
	fields := []string{"lesson_id", "name", "start_time", "end_time",
		"teaching_method", "teaching_medium", "center_id", "course_id",
		"class_id", "scheduling_status", "lesson_capacity", "end_at", "zoom_link",
	}
	scanFields := database.GetScanFields(lesson, fields)

	t.Run("failed to get total", func(t *testing.T) {
		mockRepo, mockDB := VirtualLessonRepoWithSqlMock()

		payload := payloads.GetLessonsArgs{
			CurrentTime:       now,
			TimeLookup:        payloads.TimeLookupEndTime,
			LessonTimeCompare: payloads.LessonTimeCompareFuture,
			SortAscending:     false,
			SchoolID:          "1",
			Limit:             2,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // school ID
			mock.Anything,                 // current time
		}

		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, totalFields, totalValues)

		lessons, total, offsetID, preTotal, err := mockRepo.GetLessons(ctx, mockDB.DB, payload)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
		assert.Empty(t, total)
		assert.Empty(t, offsetID)
		assert.Empty(t, preTotal)
	})

	t.Run("failed to get lessons", func(t *testing.T) {
		mockRepo, mockDB := VirtualLessonRepoWithSqlMock()

		payload := payloads.GetLessonsArgs{
			CurrentTime:       now,
			TimeLookup:        payloads.TimeLookupStartTime,
			LessonTimeCompare: payloads.LessonTimeComparePast,
			SortAscending:     false,
			SchoolID:          "1",
			Limit:             2,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // school ID
			mock.Anything,                 // current time
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // offset ID
			mock.Anything, // limit
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(pgx.ErrNoRows, fields, scanFields)

		lessons, _, offsetID, preTotal, err := mockRepo.GetLessons(ctx, mockDB.DB, payload)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, lessons)
		assert.Empty(t, offsetID)
		assert.Empty(t, preTotal)
	})

	t.Run("failed to get offset lessons", func(t *testing.T) {
		mockRepo, mockDB := VirtualLessonRepoWithSqlMock()

		payload := payloads.GetLessonsArgs{
			CurrentTime:       now,
			TimeLookup:        payloads.TimeLookupStartTime,
			LessonTimeCompare: payloads.LessonTimeComparePastAndEqual,
			SortAscending:     false,
			SchoolID:          "1",
			OffsetLessonID:    "lesson-id1",
			Limit:             2,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // school ID
			mock.Anything,                 // current time
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // offset ID
			mock.Anything, // limit
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, scanFields)

		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, offsetFields, offsetValues)

		lessons, _, offsetID, preTotal, err := mockRepo.GetLessons(ctx, mockDB.DB, payload)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.NotNil(t, lessons)
		assert.Empty(t, offsetID)
		assert.Empty(t, preTotal)
	})

	t.Run("successfully get lessons without offset", func(t *testing.T) {
		mockRepo, mockDB := VirtualLessonRepoWithSqlMock()

		payload := payloads.GetLessonsArgs{
			CurrentTime:       now,
			TimeLookup:        payloads.TimeLookupStartTime,
			LessonTimeCompare: payloads.LessonTimeComparePastAndEqual,
			SortAscending:     false,
			SchoolID:          "1",
			Limit:             2,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // school ID
			mock.Anything,                 // current time
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // offset ID
			mock.Anything, // limit
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, scanFields)

		lessons, _, offsetID, preTotal, err := mockRepo.GetLessons(ctx, mockDB.DB, payload)
		assert.Empty(t, err)
		assert.NotNil(t, lessons)
		assert.Empty(t, offsetID)
		assert.Empty(t, preTotal)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})

	t.Run("successfully get lessons with all filters", func(t *testing.T) {
		mockRepo, mockDB := VirtualLessonRepoWithSqlMock()

		payload := payloads.GetLessonsArgs{
			CurrentTime:              now,
			TimeLookup:               payloads.TimeLookupEndTimeIncludeWithoutEndAt,
			LessonTimeCompare:        payloads.LessonTimeComparePastAndEqual,
			SortAscending:            false,
			SchoolID:                 "1",
			OffsetLessonID:           "lesson-id1",
			Limit:                    2,
			LocationIDs:              []string{"loc-id1", "loc-id2"},
			TeacherIDs:               []string{"teach-id1", "teach-id2"},
			StudentIDs:               []string{"student-id1", "student-id2"},
			CourseIDs:                []string{"course-id1", "course-id2"},
			LessonSchedulingStatuses: []domain.LessonSchedulingStatus{domain.LessonSchedulingStatusDraft, domain.LessonSchedulingStatusCompleted},
			LiveLessonStatus:         payloads.LiveLessonStatusEnded,
			FromDate:                 now,
			ToDate:                   now,
		}

		queryArgs := []interface{}{
			mock.Anything,                 // context
			mock.AnythingOfType("string"), // query string
			mock.Anything,                 // school ID
			mock.Anything,                 // current time
			mock.Anything,                 // location IDs
			mock.Anything,                 // teacher IDs
			mock.Anything,                 // student IDs
			mock.Anything,                 // course IDs
			mock.Anything,                 // lesson status
			mock.Anything,                 // end date
			mock.Anything,                 // start date
		}
		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, totalFields, totalValues)

		queryArgs = append(queryArgs,
			mock.Anything, // offset ID
			mock.Anything, // limit
		)
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, scanFields)

		mockDB.MockQueryRowArgs(t, queryArgs...)
		mockDB.MockRowScanFields(nil, offsetFields, offsetValues)

		lessons, _, _, _, err := mockRepo.GetLessons(ctx, mockDB.DB, payload)
		assert.Empty(t, err)
		assert.NotNil(t, lessons)
	})
}
