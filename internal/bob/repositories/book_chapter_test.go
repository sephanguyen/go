package repositories

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BookChapterRepoWithSqlMock() (*BookChapterRepo, *testutil.MockDB) {
	r := &BookChapterRepo{}
	return r, testutil.NewMockDB()
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

		_ = e.BookID.Set(ksuid.New().String())
		_ = e.ChapterID.Set(ksuid.New().String())

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
			expectedResp: map[string][]ContentStructure{
				"lo1": {
					{
						BookID:    "book1",
						TopicID:   "topic1",
						ChapterID: "chapter1",
					},
				},
				"lo2": {
					{

						BookID:    "book2",
						TopicID:   "topic2",
						ChapterID: "chapter2",
					},
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

				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					reflect.ValueOf(args[0]).Elem().SetString("book2")
					reflect.ValueOf(args[1]).Elem().SetString("chapter2")
					reflect.ValueOf(args[2]).Elem().SetString("topic2")
					reflect.ValueOf(args[3]).Elem().SetString("lo2")
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
			expectedResp: map[string][]ContentStructure{
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

func TestBookChapterRepo_FindByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := BookChapterRepoWithSqlMock()
	chapterIDs := database.TextArray([]string{"mock-chapter-id-1", "mock-chapter-id-2"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, mock.Anything, mock.Anything, &chapterIDs)
		bookChapters, err := r.FindByIDs(ctx, mockDB.DB, chapterIDs)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, bookChapters)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &chapterIDs)

		e := &entities.BookChapter{}
		fields, values := e.FieldMap()
		_ = e.BookID.Set(idutil.ULIDNow())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		_, err := r.FindByIDs(ctx, mockDB.DB, chapterIDs)
		assert.Nil(t, err)
	})
}
