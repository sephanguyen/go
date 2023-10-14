package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductLocationRepoWithSqlMock() (*ProductLocationRepo, *testutil.MockDB) {
	productLocationRepo := &ProductLocationRepo{}
	return productLocationRepo, testutil.NewMockDB()
}

func TestProductLocationRepo_GetByLocationIDAndProductIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	productLocationRepoWithSqlMock, mockDB := ProductLocationRepoWithSqlMock()

	const locationID string = "1"
	const productID string = "1"

	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			locationID,
			productID,
		)
		entity := &entities.ProductLocation{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		productLocation, err := productLocationRepoWithSqlMock.GetByLocationIDAndProductIDForUpdate(ctx, mockDB.DB, locationID, productID)
		assert.Nil(t, err)
		assert.NotNil(t, productLocation)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			locationID,
			productID,
		)
		entity := &entities.ProductLocation{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		productLocation, err := productLocationRepoWithSqlMock.GetByLocationIDAndProductIDForUpdate(ctx, mockDB.DB, locationID, productID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, productLocation)
	})
}

func TestProductLocationRepoReplace(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	productLocationRepo := &ProductLocationRepo{}
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Req: []*entities.ProductLocation{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					LocationID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "happy case: replace multiple parents",
			Req: []*entities.ProductLocation{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					LocationID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					LocationID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(constant.SuccessCommandTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			Name: "error send batch",
			Req: []*entities.ProductLocation{
				{
					ProductID:  pgtype.Text{String: "1", Status: pgtype.Present},
					LocationID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
			},
			ExpectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			Setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		err := productLocationRepo.Replace(ctx, db, pgtype.Text{String: "1", Status: pgtype.Present}, testCase.Req.([]*entities.ProductLocation))
		if testCase.ExpectedErr != nil {
			assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			continue
		}
		assert.Equal(t, testCase.ExpectedErr, err)
	}
}
