package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonRepoWithSqlMock() (*LessonRepo, *testutil.MockDB) {
	r := &LessonRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray(ids)
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		lessons, err := r.FindByIDs(ctx, mockDB.DB, pgIDs, false)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, lessons)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &entities_bob.Lesson{}
		fields, values := e.FieldMap()
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		lessons, err := r.FindByIDs(ctx, mockDB.DB, pgIDs, false)
		assert.Nil(t, err)
		assert.Equal(t, map[pgtype.Text]*entities_bob.Lesson{
			e.LessonID: e,
		}, lessons)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLessonRepo_UpdateRoomID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()

	t.Run("err update", func(t *testing.T) {
		l := &entities_bob.Lesson{}

		// move primaryField to the last
		args := append(
			[]interface{}{
				mock.Anything,
				mock.AnythingOfType("string"),
			},
			&l.TeacherID,
			&l.ControlSettings,
			&l.LessonGroupID,
			&l.CourseID,
			&l.LessonType,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.TeachingMedium,
			&l.SchedulingStatus,
			&l.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.Update(ctx, mockDB.DB, l)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		l := &entities_bob.Lesson{}

		// move primaryField to the last
		args := append(
			[]interface{}{
				mock.Anything,
				mock.AnythingOfType("string"),
			},
			&l.TeacherID,
			&l.ControlSettings,
			&l.LessonGroupID,
			&l.CourseID,
			&l.LessonType,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.TeachingMedium,
			&l.SchedulingStatus,
			&l.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.Update(ctx, mockDB.DB, l)
		assert.EqualError(t, err, puddle.ErrClosedPool.Error())
	})

	t.Run("success", func(t *testing.T) {
		l := &entities_bob.Lesson{}

		// move primaryField to the last
		args := append(
			[]interface{}{
				mock.Anything,
				mock.AnythingOfType("string"),
			},
			&l.TeacherID,
			&l.ControlSettings,
			&l.LessonGroupID,
			&l.CourseID,
			&l.LessonType,
			&l.Name,
			&l.StartTime,
			&l.EndTime,
			&l.TeachingMedium,
			&l.SchedulingStatus,
			&l.LessonID,
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Update(ctx, mockDB.DB, l)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, l.TableName())
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(
			t,
			"updated_at",
			"teacher_id",
			"control_settings",
			"lesson_group_id",
			"course_id",
			"lesson_type",
			"name",
			"start_time",
			"end_time",
			"teaching_medium",
			"scheduling_status",
		)
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 11}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestLessonRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	lessonRepo := &LessonRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities_bob.Lesson{
				{
					LessonID:             database.Text("lesson-id"),
					TeacherID:            database.Text("teacher-id"),
					CourseID:             database.Text("course-id"),
					ControlSettings:      database.JSONB([]byte("")),
					LessonGroupID:        database.Text("lesson-group-id"),
					RoomID:               database.Text("room-id"),
					LessonType:           database.Text(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
					Status:               database.Text(""),
					StreamLearnerCounter: database.Int4(0),
					LearnerIds:           database.TextArray([]string{"learner1-id"}),
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities_bob.Lesson{
				{
					LessonID:             database.Text("lesson-id"),
					TeacherID:            database.Text("teacher-id"),
					CourseID:             database.Text("course-id"),
					ControlSettings:      database.JSONB([]byte("")),
					LessonGroupID:        database.Text("lesson-group-id"),
					RoomID:               database.Text("room-id"),
					LessonType:           database.Text(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
					Status:               database.Text(""),
					StreamLearnerCounter: database.Int4(0),
					LearnerIds:           database.TextArray([]string{"learner1-id"}),
				},
				{
					LessonID:             database.Text("lesson-id2"),
					TeacherID:            database.Text("teacher-id2"),
					CourseID:             database.Text("course-id2"),
					ControlSettings:      database.JSONB([]byte("")),
					LessonGroupID:        database.Text("lesson-group-id2"),
					RoomID:               database.Text("room-id2"),
					LessonType:           database.Text(cpb.LessonType_LESSON_TYPE_ONLINE.String()),
					Status:               database.Text(""),
					StreamLearnerCounter: database.Int4(2),
					LearnerIds:           database.TextArray([]string{"learner2-id"}),
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := lessonRepo.BulkUpsert(ctx, db, testCase.req.([]*entities_bob.Lesson))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestLessonRepo_GetLiveLessons(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonRepoWithSqlMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		locations, err := r.GetLiveLessons(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities_bob.Lesson{}
		fields, values := e.FieldMap()
		_ = e.LessonID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLiveLessons(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}
