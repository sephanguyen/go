package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LessonGroupRepoWithSqlMock() (*LessonGroupRepo, *testutil.MockDB) {
	r := &LessonGroupRepo{}
	return r, testutil.NewMockDB()
}

func TestLessonGroupRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonGroupRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities.LessonGroup{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.LessonGroup{}
		_, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.EqualError(t, err, "cannot insert new LessonGroup")
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.LessonGroup{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestLessonGroupRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonGroupRepoWithSqlMock()

	pgID := database.Text(idutil.ULIDNow())
	pgCourseID := database.Text(idutil.ULIDNow())

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&pgID,
			&pgCourseID,
		)

		e := &entities.LessonGroup{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := r.Get(ctx, mockDB.DB, pgID, pgCourseID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&pgID,
			&pgCourseID,
		)

		e := &entities.LessonGroup{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		results, err := r.Get(ctx, mockDB.DB, pgID, pgCourseID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"lesson_group_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"course_id":       {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}

func TestLessonGroupRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	lessonGroupRepo := &LessonGroupRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.LessonGroup{
				{
					LessonGroupID: database.Text("1"),
					CourseID:      database.Text("2"),
					MediaIDs:      database.TextArray([]string{}),
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
			req: []*entities.LessonGroup{
				{
					LessonGroupID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					LessonGroupID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
				{
					LessonGroupID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := lessonGroupRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.LessonGroup))
		assert.Equal(t, testCase.expectedErr, err)
	}

	return
}

func TestLessonGroupRepo_GetMediaIds(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LessonGroupRepoWithSqlMock()

	lessonGroupID := database.Text("lessonGroupID1")
	courseID := database.Text("courseID")
	limit := database.Int4(2)
	offset := database.Text("")
	t.Run("err get", func(t *testing.T) {
		mockDB.MockQueryArgs(t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.AnythingOfType("string"),
			lessonGroupID,
			courseID,
			offset,
			limit.Get(),
		)

		_, err := r.GetMedias(ctx, mockDB.DB, lessonGroupID, courseID, limit, offset)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			lessonGroupID,
			courseID,
			offset,
			limit.Get(),
		)

		e := &entities.Media{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		result, err := r.GetMedias(ctx, mockDB.DB, lessonGroupID, courseID, limit, offset)
		assert.Nil(t, err)
		assert.Equal(t, e, result[0])
	})
}
