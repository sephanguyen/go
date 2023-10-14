package repositories

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BookChapterRepoWithSqlMock() (*BookChapterRepo, *testutil.MockDB) {
	r := &BookChapterRepo{}
	return r, testutil.NewMockDB()
}

func TestBookChapterRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	bookChapterRepo := &BookChapterRepo{}
	validBookChapterReq := []*entities.BookChapter{
		{
			BookID:    database.Text("book-id-1"),
			ChapterID: database.Text("book-id-1"),
		},
		{
			BookID:    database.Text("book-id-2"),
			ChapterID: database.Text("book-id-2"),
		},
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         validBookChapterReq,
			expectedErr: nil,
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookChapterReq); i++ {
					_, field := validBookChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "exec error",
			req:         validBookChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertBookChapter error: exec error"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookChapterReq); i++ {
					_, field := validBookChapterReq[i].FieldMap()
					fields = append(fields, field...)
				}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, fields...)
				db.On("Exec", args...).Once().Return(cmdTag, fmt.Errorf("exec error"))
			},
		},
		{
			name:        "no row affected",
			req:         validBookChapterReq,
			expectedErr: fmt.Errorf("eureka_db.BulkUpsertBookChapter error: no row affected"),
			setup: func(context.Context) {
				var fields []interface{}
				for i := 0; i < len(validBookChapterReq); i++ {
					_, field := validBookChapterReq[i].FieldMap()
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
		err := bookChapterRepo.Upsert(ctx, db, testCase.req.([]*entities.BookChapter))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestBookChapterRepo_FindByBookIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookChapterRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, mock.Anything, mock.Anything, &ids)

		books, err := r.FindByBookIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, books)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &entities.BookChapter{}
		fields, values := e.FieldMap()

		_ = e.BookID.Set(idutil.ULIDNow())
		_ = e.ChapterID.Set(idutil.ULIDNow())

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		books, err := r.FindByBookIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string][]*entities.BookChapter{
			e.BookID.String: {e},
		}, books)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"book_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestBookChapterRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	r, mockDB := BookChapterRepoWithSqlMock()

	type Req struct {
		ChapterIds pgtype.TextArray
		BookIds    pgtype.TextArray
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &Req{
				ChapterIds: database.TextArray([]string{"1"}),
				BookIds:    database.TextArray([]string{"1"}),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "happy case",
			req: &Req{
				ChapterIds: database.TextArray([]string{"1"}),
				BookIds:    database.TextArray([]string{"1"}),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.SoftDelete(ctx, mockDB.DB, testCase.req.(*Req).ChapterIds, testCase.req.(*Req).BookIds)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestRetrieveContentStructuresByLOs(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	r := &BookChapterRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []interface{}{"lo1", "lo2"},
			expectedErr: nil,
			expectedResp: map[string]entities.ContentStructure{
				"lo1": {
					BookID:    "book1",
					TopicID:   "topic1",
					ChapterID: "chapter1",
				},
			},
			setup: func(ctx context.Context) {
				loIDs := database.TextArray([]string{"lo1", "lo2"})
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), &loIDs).
					Once().
					Return(rows, nil)

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("book1")
					reflect.ValueOf(args[1]).Elem().SetString("chapter1")
					reflect.ValueOf(args[2]).Elem().SetString("topic1")
					reflect.ValueOf(args[3]).Elem().SetString("lo1")
				}).Return(nil)

				rows.On("Next").Once().Return(false)

				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})

		resp, err := r.RetrieveContentStructuresByLOs(
			ctx,
			db,
			database.TextArray([]string{req[0].(string), req[1].(string)}),
		)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestRetrieveContentStructuresByTopics(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	rows := &mock_database.Rows{}
	r := &BookChapterRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         []interface{}{"topic1", "topic2"},
			expectedErr: nil,
			expectedResp: map[string][]entities.ContentStructure{
				"topic1": {
					{
						BookID:    "book1",
						TopicID:   "topic1",
						ChapterID: "chapter1",
					},
				},
				"topic2": {
					{

						BookID:    "book2",
						TopicID:   "topic2",
						ChapterID: "chapter2",
					},
				},
			},
			setup: func(ctx context.Context) {
				topicIDs := database.TextArray([]string{"topic1", "topic2"})
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), &topicIDs).
					Once().
					Return(rows, nil)

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("book1")
					reflect.ValueOf(args[1]).Elem().SetString("chapter1")
					reflect.ValueOf(args[2]).Elem().SetString("topic1")
				}).Return(nil)

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("book2")
					reflect.ValueOf(args[1]).Elem().SetString("chapter2")
					reflect.ValueOf(args[2]).Elem().SetString("topic2")
				}).Return(nil)

				rows.On("Next").Once().Return(false)

				rows.On("Close").Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "not found",
			req:         []interface{}{"topic1", "topic2"},
			expectedErr: fmt.Errorf("db.Query: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				topicIDs := database.TextArray([]string{"topic1", "topic2"})
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), &topicIDs).Once().Return(rows, pgx.ErrNoRows)
			},
		},
		{
			name:        "error scan row",
			req:         []interface{}{"topic1", "topic2"},
			expectedErr: fmt.Errorf("rows.Scan: %v", errors.New("scan failed")),
			setup: func(ctx context.Context) {
				topicIDs := database.TextArray([]string{"topic1", "topic2"})
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), &topicIDs).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(errors.New("scan failed"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]interface{})

		resp, err := r.RetrieveContentStructuresByTopics(
			ctx,
			db,
			database.TextArray([]string{req[0].(string), req[1].(string)}),
		)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}

func TestBookChapterRepo_SoftDeleteByChapterIDs(t *testing.T) {
	r, db := BookChapterRepoWithSqlMock()
	t.Parallel()

	ids := []string{"1", "2", "3"}
	testCases := []TestCase{
		{
			name: "happy case",
			req:  ids,
			setup: func(ctx context.Context) {
				db.MockExecArgs(t, pgconn.CommandTag("3"), nil, mock.Anything, mock.Anything, database.TextArray(ids))
			},
			expectedErr: nil,
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

		err := r.SoftDeleteByChapterIDs(ctx, db.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		}
	}
}
