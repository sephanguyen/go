package repositories

import (
	"context"
	"reflect"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIndividualStudyPlanRepo_BulkUpdateTime(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	individualStudyPlanRepo := &IndividualStudyPlan{}
	testCases := []TestCase{
		{
			name: "happy Case",
			req: []*entities.IndividualStudyPlan{
				&entities.IndividualStudyPlan{},
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
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := individualStudyPlanRepo.BulkUpdateTime(ctx, db, testCase.req.([]*entities.IndividualStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, nil, err)
		}
	}
}

func TestBulkSyncIndividualStudyPlan(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	individualStudyPlanRepo := &IndividualStudyPlan{}

	testCases := []TestCase{
		{
			name: "all individual study plan are new",
			req: []*entities.IndividualStudyPlan{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new3", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			expectedResp: []*entities.IndividualStudyPlan{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new3", Status: pgtype.Present},
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				batchResults.On("Close").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

				row1 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row1)
				row1.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new1", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row2 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row2)
				row2.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row3 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row3)
				row3.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new3", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)
			},
		},
		{
			name: "some study plan items exist in DB",
			req: []*entities.IndividualStudyPlan{
				{
					ID: pgtype.Text{String: "exist1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			expectedResp: []*entities.IndividualStudyPlan{
				{
					ID: pgtype.Text{String: "new1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "new2", Status: pgtype.Present},
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				batchResults.On("Close").Once().Return(nil)

				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)

				row1 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row1)
				row1.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "exist2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row2 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row2)
				row2.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new1", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)

				row3 := &mock_database.Row{}
				batchResults.On("QueryRow").Once().Return(row3)
				row3.On("Scan", mock.Anything).Once().Run(func(args mock.Arguments) {
					id := pgtype.Text{String: "new2", Status: pgtype.Present}
					reflect.ValueOf(args[0]).Elem().Set(reflect.ValueOf(id))
				}).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		resp, err := individualStudyPlanRepo.BulkSync(ctx, db, testCase.req.([]*entities.IndividualStudyPlan))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedResp, resp)
		}
	}
}
