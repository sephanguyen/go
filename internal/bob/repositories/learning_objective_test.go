package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
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

func TestLearningObjectiveRepo_Create(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	learningObjectiveRepo := &LearningObjectiveRepo{}
	learningObjective := &entities.LearningObjective{
		ID:            database.Text("id"),
		Name:          database.Text("name"),
		Country:       database.Text("country"),
		Grade:         database.Int2(1),
		Subject:       database.Text("subject"),
		TopicID:       database.Text("topic-id"),
		MasterLoID:    database.Text("mater-lo-id"),
		DisplayOrder:  database.Int2(1),
		VideoScript:   database.Text("video-script"),
		Prerequisites: database.TextArray([]string{"prerequisites-1"}),
		Video:         database.Text("video"),
		StudyGuide:    database.Text("study-guide"),
		SchoolID:      database.Int4(1),
		CreatedAt:     database.Timestamptz(time.Now()),
		UpdatedAt:     database.Timestamptz(time.Now()),
		DeletedAt:     database.Timestamptz(time.Now()),
		Type:          database.Text("type"),
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         learningObjective,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "error TxClosed",
			req:         learningObjective,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error no rows affected",
			req:         learningObjective,
			expectedErr: errors.New("cannot insert new learning_objectives"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := learningObjectiveRepo.Create(ctx, db, testCase.req.(*entities.LearningObjective))
			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
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
	scanFields := database.GetScanFields(e, fields)
	var (
		bookID pgtype.Text
	)
	scanFields = append(scanFields, &bookID)

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
		})
	}
}
