package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ChapterRepoWithSqlMock() (*ChapterRepo, *testutil.MockDB) {
	r := &ChapterRepo{}
	return r, testutil.NewMockDB()
}

func TestChapterRepo_FindByID(t *testing.T) {
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
		mockDB.RawStmt.AssertSelectedFields(t, "chapter_id", "name", "country", "subject", "grade", "display_order", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_topic_display_order", "book_id")
		mockDB.RawStmt.AssertSelectedTable(t, "chapters", "")
	})
}

func TestListChapters(t *testing.T) {
	t.Parallel()
	r, mockDB := ChapterRepoWithSqlMock()
	listChaptersArgs := ListChaptersArgs{}

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.Chapter{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "error no rows",
			expectedErr: fmt.Errorf("database.Select: err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.Chapter{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		e := &entities.Chapter{}
		_, err := r.ListChapters(ctx, mockDB.DB, &listChaptersArgs)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestChapterRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	chapterRepo := &ChapterRepo{}
	validChapterReq := []*entities.Chapter{
		{
			ID:           database.Text("chapter-id-1"),
			Name:         database.Text("chapter-name-1"),
			SchoolID:     database.Int4(1),
			DisplayOrder: database.Int2(1),
		},
		{
			ID:           database.Text("chapter-id-2"),
			Name:         database.Text("chapter-name-2"),
			SchoolID:     database.Int4(1),
			DisplayOrder: database.Int2(1),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validChapterReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertChapter error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertChapter error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := chapterRepo.Upsert(ctx, db, testCase.req.([]*entities.Chapter))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestChapterRepo_UpsertWithoutDisplayOrderWhenUpdate(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	chapterRepo := &ChapterRepo{}
	validChapterReq := []*entities.Chapter{
		{
			ID:           database.Text("chapter-id-1"),
			Name:         database.Text("chapter-name-1"),
			SchoolID:     database.Int4(1),
			DisplayOrder: database.Int2(1),
		},
		{
			ID:           database.Text("chapter-id-2"),
			Name:         database.Text("chapter-name-2"),
			SchoolID:     database.Int4(1),
			DisplayOrder: database.Int2(1),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validChapterReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertChapter error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertChapter error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validChapterReq); i++ {
					_, field := validChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := chapterRepo.UpsertWithoutDisplayOrderWhenUpdate(ctx, db, testCase.req.([]*entities.Chapter))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
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

		_ = e.ID.Set(idutil.ULIDNow())
		_ = e.Name.Set(idutil.ULIDNow())

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

func TestChapterRepo_ListChapters(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs, mock.Anything, mock.Anything, mock.Anything)

		args := &ListChaptersArgs{
			ChapterIDs: database.TextArray(ids),
		}
		chapters, err := r.ListChapters(ctx, mockDB.DB, args)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, chapters)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.Chapter{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(idutil.ULIDNow())
		_ = e.Name.Set(idutil.ULIDNow())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		args := &ListChaptersArgs{
			ChapterIDs: database.TextArray(ids),
		}
		_, err := r.ListChapters(ctx, mockDB.DB, args)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
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

func TestChapterRepo_SoftDelete(t *testing.T) {
	r, db := ChapterRepoWithSqlMock()
	t.Parallel()

	ids := []string{"1", "2", "3"}
	testCases := []TestCase{
		{
			name: "happy case",
			req:  ids,
			setup: func(ctx context.Context) {
				db.MockExecArgs(t, pgconn.CommandTag("3"), nil, mock.Anything, mock.Anything, database.TextArray(ids))
			},
			expectedResp: 3,
		},
		{
			name: "update error",
			req:  ids,
			setup: func(ctx context.Context) {
				db.MockExecArgs(t, pgconn.CommandTag("3"), puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(ids))
			},
			expectedErr: puddle.ErrClosedPool,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]string)

		resp, err := r.SoftDelete(ctx, db.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestChapterRepo_DuplicateChapters(t *testing.T) {
	t.Parallel()
	_, mockDB := ChapterRepoWithSqlMock()
	chapterRepo := &ChapterRepo{}

	bookID := "book-id-1"
	chapterIDs := []string{"chapter-id-1", "chapter-id-2"}

	t.Run("Happy case", func(t *testing.T) {
		ctx := context.Background()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &bookID, &chapterIDs)

		e := &entities.CopiedChapter{}
		fields, values := e.FieldMap()
		_ = e.CopyFromID.Set("chapter-id-1")
		_ = e.ID.Set("random-id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		copiedChapters, err := chapterRepo.DuplicateChapters(ctx, mockDB.DB, bookID, chapterIDs)

		assert.Nil(t, err)
		assert.Equal(t, copiedChapters, []*entities.CopiedChapter{e})
	})

}

func TestChapterRepo_FindByBookIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ChapterRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		chapters, err := r.FindByBookIDs(ctx, mockDB.DB, database.TextArray(ids))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, chapters)
	})
}
