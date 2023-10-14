package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LearningObjectiveRepoWithSqlMock() (*LearningObjectiveRepo, *testutil.MockDB) {
	r := &LearningObjectiveRepo{}
	return r, testutil.NewMockDB()
}

func TestQuizRepo_RetrieveByIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := LearningObjectiveRepoWithSqlMock()
	ids := database.TextArray([]string{"ids"})

	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool,
					mock.Anything,
					mock.Anything,
					&ids,
				)
			},
			expectedErr: fmt.Errorf("rows.Err: err db.Query: %w", puddle.ErrClosedPool),
		},
		{
			name: "success with select all fields",
			setup: func(ctx context.Context) {
				lo := &entities.LearningObjective{}
				fields, values := lo.FieldMap()
				lo.ID.Set("id")
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					&ids,
				)
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.RetrieveByIDs(ctx, mockDB.DB, ids)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestQuizRepo_RetrieveByTopicIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := LearningObjectiveRepoWithSqlMock()
	ids := database.TextArray([]string{"ids"})

	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool,
					mock.Anything,
					mock.Anything,
					&ids,
				)
			},
			expectedErr: fmt.Errorf("rows.Err: err db.Query: %w", puddle.ErrClosedPool),
		},
		{
			name: "success with retrieve by topic ids",
			setup: func(ctx context.Context) {
				lo := &entities.LearningObjective{}
				fields, values := lo.FieldMap()
				lo.ID.Set("id")
				mockDB.MockQueryArgs(t, nil,
					mock.Anything,
					mock.Anything,
					&ids,
				)
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.RetrieveByTopicIDs(ctx, mockDB.DB, ids)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestBulkImport_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	learningObjectiveRepo := &LearningObjectiveRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.LearningObjective{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			req: []*entities.LearningObjective{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				}, {
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := learningObjectiveRepo.BulkImport(ctx, db, testCase.req.([]*entities.LearningObjective))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func Test_RetrieveBookLoByIntervalTime(t *testing.T) {
	t.Parallel()
	type TestCase struct {
		name         string
		input1       pgtype.Text
		expectedResp interface{}
		expectedErr  error
		setup        func(ctx context.Context)
	}
	mockDB := testutil.NewMockDB()
	rows := mockDB.Rows
	r := &LearningObjectiveRepo{}
	e := &entities.LearningObjective{}
	fields, _ := e.FieldMap()
	totalFields := append(fields, "book_id", "chapter_id", "topic_id")
	scanFields := database.GetScanFields(e, fields)
	var (
		bookID    pgtype.Text
		chapterID pgtype.Text
		topicID   pgtype.Text
	)
	scanFields = append(scanFields, &bookID, &chapterID, &topicID)

	intervalTime := database.Text("15 mins")
	testCases := []TestCase{
		{
			name:        "happy case",
			input1:      intervalTime,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, intervalTime)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
			},
		},
		{
			name:        "err query",
			input1:      intervalTime,
			expectedErr: fmt.Errorf("row.Err: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, intervalTime)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(pgx.ErrNoRows)
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.setup(ctx)
			_, err := r.RetrieveBookLoByIntervalTime(ctx, mockDB.DB, tc.input1)
			assert.Equal(t, tc.expectedErr, err)
			mockDB.RawStmt.AssertSelectedFields(t, totalFields...)
		})
	}
}

func TestLearningObjectiveRepo_UpdateDisplayOrders(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	learningObjectiveRepo := &LearningObjectiveRepo{}
	mDisplayOrder := map[pgtype.Text]pgtype.Int2{
		database.Text("lo-1"): database.Int2(1),
	}
	testCases := []TestCase{
		{
			name:        "happy case",
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
			name:        "error send batch",
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := learningObjectiveRepo.UpdateDisplayOrders(ctx, db, mDisplayOrder)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestLearningObjectiveRepo_SoftDeleteWithLoIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	LoRepo := &LearningObjectiveRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"lo_id-1", "lo_id-2"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "unexpected error",
			req:         database.TextArray([]string{"lo_id-1", "lo_id-2"}),
			expectedErr: fmt.Errorf("err db.Exec: %w", fmt.Errorf("unexpected error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("unexpected error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		_, err := LoRepo.SoftDeleteWithLoIDs(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestLearningObjectiveRepo_UpdateName(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	r := &LearningObjectiveRepo{}
	id := database.Text("LO_ID_1")
	name := database.Text("LO name edited")
	query := "UPDATE learning_objectives SET name = $1, updated_at = now() WHERE lo_id = $2::TEXT AND deleted_at IS NULL"

	testCases := []TestCase{
		{
			name: "Happy case",
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, query, name, id).Once().Return(pgconn.CommandTag("1"), nil)
			},
			expectedResp: 1,
		},
		{
			name: "db.Exec error",
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, query, name, id).Once().Return(pgconn.CommandTag("0"), puddle.ErrClosedPool)
			},
			expectedErr: fmt.Errorf("db.Exec: %w", puddle.ErrClosedPool),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			rowAff, err := r.UpdateName(ctx, db, id, name)

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedResp, int(rowAff))
				assert.NoError(t, err)
			}
		})

	}
}
