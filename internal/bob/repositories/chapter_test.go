package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ChapterRepoWithSqlMock() (*ChapterRepo, *testutil.MockDB) {
	r := &ChapterRepo{}
	return r, testutil.NewMockDB()
}

func TestChapterRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		chapters, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, chapters)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &entities.Chapter{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.Name.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		chapters, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string]*entities.Chapter{
			e.ID.String: e,
		}, chapters)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"chapter_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestChapterRepo_FindSchoolIDsOnChapters(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		schoolIDs, err := r.FindSchoolIDsOnChapters(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		schoolIDs, err := r.FindSchoolIDsOnChapters(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []int32{e.SchoolID}, schoolIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "chapters", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"chapter_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestChapterRepo_UpsertWithoutDisplayOrderWhenUpdate(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	chapterRepo := &ChapterRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Chapter{
				{
					ID:       database.Text("mock-chapter-id-1"),
					Name:     database.Text("mock-chapter-name-1"),
					SchoolID: database.Int4(1),
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
			req: []*entities.Chapter{
				{
					ID: database.Text("mock-chapter-id-1"),
				},
				{
					ID: database.Text("mock-chapter-id-2"),
				},
				{
					ID: database.Text("mock-chapter-id-3"),
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
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := chapterRepo.UpsertWithoutDisplayOrderWhenUpdate(ctx, db, testCase.req.([]*entities.Chapter))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestChapterRepo_FindByID(t *testing.T) {
	// t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()
	chapterID := database.Text("mock-chapter-id")
	e := &entities.Chapter{}
	_, values := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &chapterID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(puddle.ErrClosedPool)
		chapter, err := r.FindByID(ctx, mockDB.DB, chapterID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, chapter)
	})

	t.Run("success query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &chapterID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(nil)
		chapter, err := r.FindByID(ctx, mockDB.DB, chapterID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, chapter)
		mockDB.RawStmt.AssertSelectedFields(t, "chapter_id", "name", "country", "subject", "grade", "display_order", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_topic_display_order")
		mockDB.RawStmt.AssertSelectedTable(t, "chapters", "")
	})
}

func TestChapterRepo_UpdateCurrentTopicDisplayOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()

	chapterID := database.Text("mock-chapter-id")
	totalGeneratedTopicDisplayOrder := database.Int4(4)
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedTopicDisplayOrder, chapterID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.UpdateCurrentTopicDisplayOrder(ctx, mockDB.DB, totalGeneratedTopicDisplayOrder, chapterID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedTopicDisplayOrder, chapterID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateCurrentTopicDisplayOrder(ctx, mockDB.DB, totalGeneratedTopicDisplayOrder, chapterID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "chapters")
		mockDB.RawStmt.AssertUpdatedFields(t, "current_topic_display_order", "updated_at")
	})
}
