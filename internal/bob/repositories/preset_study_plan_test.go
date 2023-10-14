package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePresetStudyPlan_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	presetStudyPlanRepo := &PresetStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.PresetStudyPlan{
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
			req: []*entities.PresetStudyPlan{
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
		err := presetStudyPlanRepo.CreatePresetStudyPlan(ctx, db, testCase.req.([]*entities.PresetStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
	return
}

func TestCreatePresetStudyPlanWeekly_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	presetStudyPlanRepo := &PresetStudyPlanRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.PresetStudyPlanWeekly{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				rows := &mock_database.Rows{}

				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once()
				rows.On("Err").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.PresetStudyPlanWeekly{
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
				rows := &mock_database.Rows{}

				db.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once()
				rows.On("Err").Once().Return(nil)

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
		err := presetStudyPlanRepo.CreatePresetStudyPlanWeekly(ctx, db, testCase.req.([]*entities.PresetStudyPlanWeekly))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
	return
}
