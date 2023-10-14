package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LessonRepoWithSqlMock() (*LessonRepo, *testutil.MockDB) {
	r := &LessonRepo{}
	return r, testutil.NewMockDB()
}

var selectFields = []string{
	"lesson_id", "teacher_id", "course_id", "control_settings", "created_at",
	"updated_at", "deleted_at", "end_at", "lesson_group_id", "room_id", "lesson_type",
	"status", "stream_learner_counter", "learner_ids", "name", "start_time", "end_time", "room_state", "teaching_model", "class_id",
	"center_id", "teaching_method", "teaching_medium", "scheduling_status", "is_locked", "zoom_link", "zoom_id",
}

func TestLessonRepo_GetStreamingLearners(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	t.Run("err select", func(t *testing.T) {
		e := &entities.Lesson{}
		mockDB.MockQueryArgs(t, ErrNoRows, mock.Anything, mock.AnythingOfType("string"), database.Text("0"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := r.GetStreamingLearners(ctx, mockDB.DB, database.Text("0"))
		assert.True(t, errors.Is(err, ErrNoRows))
	})

	t.Run("happy case", func(t *testing.T) {
		e := &entities.Lesson{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text("0"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := r.GetStreamingLearners(ctx, mockDB.DB, database.Text("0"))
		assert.True(t, errors.Is(err, nil))
	})
}

func TestLessonStreamRepo_IncreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	pgTextZero := database.Text("0")
	mockMaximumLearnerStreamings := 13
	t.Run("err update status", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		err := r.IncreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("no effected row", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		err := r.IncreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		assert.Error(t, err)
		assert.True(t, err == ErrUnAffected)
	})
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		err := r.IncreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero, mockMaximumLearnerStreamings)
		assert.NoError(t, err)
	})
}

func TestLessonStreamRepo_DecreaseNumberOfStreaming(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	pgTextZero := database.Text("0")

	t.Run("err update status", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero)
		err := r.DecreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("no effected row", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero)
		err := r.DecreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero)
		assert.Error(t, err)
		assert.True(t, err == ErrUnAffected)
	})
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.AnythingOfType("string"), pgTextZero, pgTextZero)
		err := r.DecreaseNumberOfStreaming(ctx, mockDB.DB, pgTextZero, pgTextZero)
		assert.NoError(t, err)
	})
}

var pgSchedulingStatus = pgtype.Text{String: string(entities.LessonSchedulingStatusPublished), Status: pgtype.Present}
var pgSchedulingStatusNull = pgtype.Text{String: "", Status: pgtype.Null}

func TestFindLessonWithTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	courseIDs := database.TextArray([]string{"course_id"})
	startDate := database.Timestamptz(time.Now())
	endDate := database.Timestamptz(time.Now())

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		lessons, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatusNull)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatusNull)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
	t.Run("success with select with Scheduling Status", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatus)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
}
func TestFindLessonWithTimeAndLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	courseIDs := database.TextArray([]string{"course_id"})
	startDate := database.Timestamptz(time.Now())
	endDate := database.Timestamptz(time.Now())
	locationIDs := database.TextArray([]string{"location_id"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		lessons, _, err := r.FindLessonWithTimeAndLocations(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, &locationIDs, 100, 1, pgSchedulingStatusNull)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTimeAndLocations(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, &locationIDs, 100, 1, pgSchedulingStatusNull)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
	t.Run("success with select with Scheduling Status", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTimeAndLocations(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, &locationIDs, 100, 1, pgSchedulingStatus)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
}
func TestFindLessonJoined(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	userID := database.Text("user_id")
	courseIDs := database.TextArray([]string{"course_id"})
	startDate := database.Timestamptz(time.Now())
	endDate := database.Timestamptz(time.Now())

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		lessons, _, err := r.FindLessonJoined(ctx, mockDB.DB, userID, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatusNull)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatusNull)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})

	t.Run("success with select with Scheduling Status", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, pgSchedulingStatus, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatus)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
}

func TestFindLessonJoinedV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	nowTimestamptz := database.Timestamptz(time.Now())
	lessonFilter := LessonJoinedV2Filter{
		UserID:               database.Text("test-user-id-1"),
		CourseIDs:            database.TextArray([]string{"test-course-id-1", "test-course-id-2"}),
		BlacklistedCourseIDs: database.TextArray([]string{"test-blacklisted-course-id-1", "test-blacklisted-course-id-2"}),
		StartDate:            &nowTimestamptz,
		EndDate:              &nowTimestamptz,
	}
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, _, err := r.FindLessonJoinedV2(ctx, mockDB.DB, &lessonFilter, 100, 1)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonJoinedV2(ctx, mockDB.DB, &lessonFilter, 100, 1)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
}

func TestFindLessonJoinedWithLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	userID := database.Text("user_id")
	courseIDs := database.TextArray([]string{"course_id"})
	startDate := database.Timestamptz(time.Now())
	endDate := database.Timestamptz(time.Now())
	locationIDs := database.TextArray([]string{"location_id"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything)

		lessons, _, err := r.FindLessonJoinedWithLocations(ctx, mockDB.DB, userID, &courseIDs, &startDate, &endDate, &locationIDs, 100, 1, pgSchedulingStatusNull)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, pgSchedulingStatusNull, mock.Anything, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatusNull)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})

	t.Run("success with select with Scheduling Status", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, pgSchedulingStatus, mock.Anything, mock.Anything)

		e := LessonWithTime{}
		var total pgtype.Int8

		value := append(database.GetScanFields(&e.Lesson, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.Lesson.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessonsWithTime, _, err := r.FindLessonWithTime(ctx, mockDB.DB, &courseIDs, &startDate, &endDate, 100, 1, pgSchedulingStatus)
		assert.Nil(t, err)
		assert.Equal(t, []*LessonWithTime{
			{Lesson: entities.Lesson{LessonID: database.Text("id")}},
		}, lessonsWithTime)
	})
}

func TestLessonRepo_Retrieves(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	args := &ListLessonArgs{
		Limit:            5,
		LessonID:         pgtype.Text{Status: pgtype.Null},
		SchoolID:         database.Int4(1),
		Courses:          pgtype.TextArray{},
		StartTime:        database.Timestamptz(time.Now()),
		EndTime:          database.Timestamptz(time.Now()),
		StatusNotStarted: database.Text(""),
		StatusInProcess:  database.Text(""),
		StatusCompleted:  database.Text(""),
		KeyWord:          database.Text(""),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		lessons, _, _, _, err := r.Retrieve(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.Lesson{}
		var total pgtype.Int8
		var prePage pgtype.Text
		var preTotal pgtype.Int8
		selectFields := []string{"lesson_id", "created_at", "name", "start_time", "end_time", "lesson_type", "class_id"}
		value := append(database.GetScanFields(e, selectFields), &prePage, &preTotal, &total)
		selectFields = append(selectFields, "lesson_id", "total", "total")
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		lessons, _, _, _, err := r.Retrieve(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.Lesson{
			{LessonID: database.Text("id")},
		}, lessons)

		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestLessonRepo_FindPreviousPageOffset(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	args := &ListLessonArgs{
		Limit:    5,
		LessonID: pgtype.Text{Status: pgtype.Null},
		SchoolID: database.Int4(1),
	}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		lessons, err := r.FindPreviousPageOffset(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Equal(t, "", lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.Lesson{}
		selectFields := []string{"lesson_id", "created_at"}
		var total pgtype.Int8
		value := append(database.GetScanFields(e, selectFields), &total)
		selectFields = append(selectFields, "total")
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value, value, value, value, value,
		})

		lessons, err := r.FindPreviousPageOffset(ctx, mockDB.DB, args)
		assert.Nil(t, err)
		assert.Equal(t, "id", lessons)

		mockDB.RawStmt.AssertSelectedTable(t, "filter_time", "ft")
	})
}

func TestLessonRepo_CountLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	args := &ListLessonArgs{
		Limit:    5,
		LessonID: pgtype.Text{Status: pgtype.Null},
		SchoolID: database.Int4(1),
	}

	t.Run("count lesson", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		total, err := r.CountLesson(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Equal(t, int64(0), total)

		mockDB.RawStmt.AssertSelectedTable(t, "ls", "")
	})
}

func TestLessonRepo_GrantRecordingPermission(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lessonID := database.Text("lesson-id-1")
	state := "state"
	t.Run("successfully", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.GrantRecordingPermission(ctx, mockDB.DB, lessonID, database.JSONB(state))
		require.NoError(t, err)
	})

	t.Run("has error", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), errors.New("error"), args...)
		err := r.GrantRecordingPermission(ctx, mockDB.DB, lessonID, database.JSONB(state))
		require.Error(t, err)
	})
}

func TestLessonRepo_StopRecording(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	lessonID := database.Text("lesson-id-1")
	creator := database.Text("user-id-1")
	state := "state"
	t.Run("successfully", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.StopRecording(ctx, mockDB.DB, lessonID, creator, database.JSONB(state))
		require.NoError(t, err)
	})

	t.Run("has error", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &lessonID})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), errors.New("error"), args...)
		err := r.StopRecording(ctx, mockDB.DB, lessonID, creator, database.JSONB(state))
		require.Error(t, err)
	})
}

func TestLessonRepo_GetTeachersOfLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	ids := database.TextArray([]string{"lesson-1", "lesson-2"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		lessonTeachers, err := r.GetTeachersOfLessons(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessonTeachers)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities.LessonsTeachers{}
		fields, values := e.FieldMap()
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetTeachersOfLessons(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
