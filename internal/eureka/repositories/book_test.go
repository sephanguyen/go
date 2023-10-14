package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
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

func BookRepoWithSqlMock() (*BookRepo, *testutil.MockDB) {
	r := &BookRepo{}
	return r, testutil.NewMockDB()
}

func TestBookRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities.Book{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.Name.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string]*entities.Book{
			e.ID.String: e,
		}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"book_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestBookRepo_FindWithFilter(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
	r, mockDB := BookRepoWithSqlMock()
	type Params struct {
		courseID string
		limit    uint32
		offset   uint32
	}
	type Result struct {
		books []*entities.Book
		count int
	}
	courseID := "course-id"
	book := &entities.Book{
		ID: database.Text("book-id-1"),
	}
	fields, values := book.FieldMap()
	count := 1
	testCases := []TestCase{
		{
			name: "error select books",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &courseID)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&count})
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &courseID)
				mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{values})
			},
			req: Params{
				courseID: "course-id",
			},
			expectedErr: fmt.Errorf("database.Select: rows.Scan: %w", pgx.ErrNoRows),
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &courseID)
				mockDB.MockRowScanFields(nil, []string{"count"}, []interface{}{&count})
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &courseID)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			req: Params{
				courseID: "course-id",
			},
			expectedResp: Result{
				books: []*entities.Book{book},
				count: 1,
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			params := testCase.req.(Params)
			_, _, err := r.FindWithFilter(ctx, mockDB.DB, params.courseID, params.limit, params.offset)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestBookRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	bookID := database.Text("mock-book-id")
	e := &entities.Book{}
	_, values := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &bookID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(pgx.ErrNoRows)
		book, err := r.FindByID(ctx, mockDB.DB, bookID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, book)
	})

	t.Run("success query", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &bookID)
		mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", values...).Once().Return(nil)
		book, err := r.FindByID(ctx, mockDB.DB, bookID)
		assert.True(t, errors.Is(err, nil))
		assert.NotNil(t, book)
		mockDB.RawStmt.AssertSelectedFields(t, "book_id", "name", "country", "subject", "grade", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_chapter_display_order", "book_type", "is_v2")
		mockDB.RawStmt.AssertSelectedTable(t, "books", "")
	})
}

func TestBookRepo_UpdateCurrentChapterDisplayOrder(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()

	bookID := database.Text("mock-book-id")
	totalGeneratedChapterDisplayOrder := database.Int4(4)
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedChapterDisplayOrder, bookID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.UpdateCurrentChapterDisplayOrder(ctx, mockDB.DB, totalGeneratedChapterDisplayOrder, bookID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, totalGeneratedChapterDisplayOrder, bookID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateCurrentChapterDisplayOrder(ctx, mockDB.DB, totalGeneratedChapterDisplayOrder, bookID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "books")
		mockDB.RawStmt.AssertUpdatedFields(t, "current_chapter_display_order")
	})
}

func TestBookRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()

	bookIDs := database.TextArray([]string{"book-id-1", "book-id-2"})

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &bookIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.SoftDelete(ctx, mockDB.DB, bookIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &bookIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.SoftDelete(ctx, mockDB.DB, bookIDs)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "books")
	})
}

func TestBookRepo_RetrieveBookTreeByBookID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	bookID := database.Text("book-id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &bookID)

		books, err := r.RetrieveBookTreeByBookID(ctx, mockDB.DB, bookID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &bookID)

		e := &entities.Book{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		_ = e.Name.Set(ksuid.New().String())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.RetrieveBookTreeByBookID(ctx, mockDB.DB, bookID)
		assert.Nil(t, err)
	})
}

func TestBookRepo_RetrieveBookTreeByTopicIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	topicIDs := database.TextArray([]string{"topic-id-1", "topic-id-2"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &topicIDs)

		books, err := r.RetrieveBookTreeByTopicIDs(ctx, mockDB.DB, topicIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &topicIDs)

		ce := &entities.Chapter{}
		te := &entities.Topic{}
		loe := &entities.LearningObjective{}

		mockDB.MockScanArray(nil, []string{"lo.lo_id", "tp.topic_id", "ct.chapter_id", "lo.display_order", "tp.display_order", "ct.display_order"}, [][]interface{}{
			{
				&loe.ID, &te.ID, &ce.ID, &loe.DisplayOrder, &te.DisplayOrder, &ce.DisplayOrder,
			},
		})

		_, err := r.RetrieveBookTreeByTopicIDs(ctx, mockDB.DB, topicIDs)
		assert.Nil(t, err)
	})
}

func TestBookRepo_ListBooks(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	listBooksArgs := &ListBooksArgs{}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		books, err := r.ListBooks(ctx, mockDB.DB, listBooksArgs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		b := &entities.Book{}
		fields, values := b.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.ListBooks(ctx, mockDB.DB, listBooksArgs)
		assert.Nil(t, err)
	})
}

func TestBookRepo_DuplicateBook(t *testing.T) {
	t.Parallel()
	_, mockDB := BookRepoWithSqlMock()
	bookRepo := &BookRepo{}
	type duplicateBookReq struct {
		Name string
		ID   string
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &duplicateBookReq{
				Name: "book-name",
				ID:   "book-id",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				expectBookId := "book_id_return"
				fields := []string{"book_id"}
				values := []interface{}{&expectBookId}
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			req := testCase.req.(*duplicateBookReq)
			_, err := bookRepo.DuplicateBook(ctx, mockDB.DB, database.Text(req.Name), database.Text(req.ID))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestBookRepo_RetrieveAdHocBookByCourseIDAndStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookRepoWithSqlMock()
	courseID := database.Text("course-id")
	studentID := database.Text("student-id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		books, err := r.RetrieveAdHocBookByCourseIDAndStudentID(ctx, mockDB.DB, courseID, studentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		b := &entities.Book{}
		fields, values := b.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.RetrieveAdHocBookByCourseIDAndStudentID(ctx, mockDB.DB, courseID, studentID)
		assert.Nil(t, err)
	})
}

func TestBookRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	bookRepo := &BookRepo{}
	validBookReq := []*entities.Book{
		{
			ID:       database.Text("book-id-1"),
			Name:     database.Text("book-name-1"),
			SchoolID: database.Int4(1),
		},
		{
			ID:       database.Text("book-id-2"),
			Name:     database.Text("book-name-2"),
			SchoolID: database.Int4(1),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validBookReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookReq); i++ {
					_, field := validBookReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validBookReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertBook error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookReq); i++ {
					_, field := validBookReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validBookReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertBook error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookReq); i++ {
					_, field := validBookReq[i].FieldMap()
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
		err := bookRepo.Upsert(ctx, db, testCase.req.([]*entities.Book))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
