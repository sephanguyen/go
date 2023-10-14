package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductGradeRepoWithSqlMock() (*ProductGradeRepo, *testutil.MockDB) {
	r := &ProductGradeRepo{}
	return r, testutil.NewMockDB()
}

func TestProductGradeRepo_GetByGradeAndProductIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	productGradeRepoWithSqlMock, mockDB := ProductGradeRepoWithSqlMock()

	const grade string = "1"
	const productID string = "1"

	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			grade,
			productID,
		)
		entity := &entities.ProductGrade{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		billingRatio, err := productGradeRepoWithSqlMock.GetByGradeAndProductIDForUpdate(ctx, mockDB.DB, grade, productID)
		assert.Nil(t, err)
		assert.NotNil(t, billingRatio)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			grade,
			productID,
		)
		entity := &entities.ProductGrade{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		billingRatio, err := productGradeRepoWithSqlMock.GetByGradeAndProductIDForUpdate(ctx, mockDB.DB, grade, productID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, billingRatio)
	})
}

func TestProductGradeRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mockDb.Ext{}
	productGradeRepo := &ProductGradeRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.ProductGrade{
				{
					ProductID: pgtype.Text{String: "1", Status: pgtype.Present},
					GradeID:   pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: upsert multiple parents",
			Req: []*entities.ProductGrade{
				{
					ProductID: pgtype.Text{String: "1", Status: pgtype.Present},
					GradeID:   pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					ProductID: pgtype.Text{String: "2", Status: pgtype.Present},
					GradeID:   pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*entities.ProductGrade{
				{
					ProductID: pgtype.Text{String: "1", Status: pgtype.Present},
					GradeID:   pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mockDb.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := productGradeRepo.Upsert(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.ProductGrade))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.ExpectedErr, err)
		}
	}
}
