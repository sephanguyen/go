package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsertMediaBatch(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	mediaRepo := &MediaRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Media{
				{
					MediaID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			req: []*entities.Media{
				{
					MediaID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					MediaID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("UpsertMediaBatch batchResults.Exec: %w", pgx.ErrTxClosed),
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
		err := mediaRepo.UpsertMediaBatch(ctx, db, testCase.req.([]*entities.Media))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestRetrieveByIDs(t *testing.T) {
	t.Parallel()

	r := &MediaRepo{}
	mockDB := testutil.NewMockDB()

	ids := pgtype.TextArray{}
	_ = ids.Set([]string{"id"})

	e := &entities.Media{}
	fields, values := e.FieldMap()

	testCases := []TestCase{
		{
			name:        "error query",
			req:         ids,
			expectedErr: fmt.Errorf("database.Select: err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
					mock.AnythingOfType("string"),
					ids,
				)
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "happy case",
			req:         ids,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					ids,
				)
				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.RetrieveByIDs(ctx, mockDB.DB, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"media_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	}

	return
}
