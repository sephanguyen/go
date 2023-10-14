package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func MaterialRepoWithSqlMock() (*MaterialRepo, *testutil.MockDB, *mock_database.Tx) {
	materialRepo := &MaterialRepo{}
	return materialRepo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestMaterialRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	materialRepoWithSqlMock, mockDB, _ := MaterialRepoWithSqlMock()
	var materialID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			materialID,
		)
		entities := &entities.Material{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		material, err := materialRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, materialID)
		assert.Nil(t, err)
		assert.NotNil(t, material)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			materialID,
		)
		entities := &entities.Material{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		material, err := materialRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, materialID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, material)

	})
}

func TestMaterialRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Material{}
	_, fieldProductMap := mockEntities.Product.FieldMap()
	_, fieldMaterialMap := mockEntities.FieldMap()

	argsMaterial := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMaterialMap))...)
	argsQueryRow := append([]interface{}{mock.Anything}, genSliceMock(len(fieldProductMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		tx.On("Exec", argsMaterial...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return repo.Create(ctx, tx, mockEntities)
		})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert product fail", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert material fail", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsMaterial...).Once().Return(constant.SuccessCommandTag, pgx.ErrTxClosed)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Product Material: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert material", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsMaterial...).Return(constant.FailCommandTag, nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Product Material: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestMaterialRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Material{}
	_, fieldProductMap := mockEntities.Product.FieldMap()
	_, fieldMaterialMap := mockEntities.FieldMap()

	argsProduct := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldProductMap))...)
	argsMaterial := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMaterialMap))...)
	// argsQueryRow := append([]interface{}{mock.Anything}, genSliceMock(len(fieldProductMap)-1)...)

	t.Run("happy case", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("Exec", argsMaterial...).Return(constant.SuccessCommandTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update product fail", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, pgx.ErrTxClosed)
		tx.On("Exec", argsMaterial...).Return(constant.SuccessCommandTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update material fail", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("Exec", argsMaterial...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product Material: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update product", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		cmdProductTag := constant.FailCommandTag
		cmdMaterialTag := constant.SuccessCommandTag
		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(cmdProductTag, nil)
		tx.On("Exec", argsMaterial...).Return(cmdMaterialTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product: %d RowsAffected", cmdProductTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update material", func(t *testing.T) {
		materialRepoWithSqlMock, mockDB, tx := MaterialRepoWithSqlMock()

		cmdProductTag := constant.SuccessCommandTag
		cmdMaterialTag := constant.FailCommandTag
		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(cmdProductTag, nil)
		tx.On("Exec", argsMaterial...).Return(cmdMaterialTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return materialRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product Material: %d RowsAffected", cmdMaterialTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestMaterialRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	materialRepoWithSqlMock, mockDB, _ := MaterialRepoWithSqlMock()
	var materialID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			materialID,
		)
		entities := &entities.Material{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		material, err := materialRepoWithSqlMock.GetByID(ctx, mockDB.DB, materialID)
		assert.Nil(t, err)
		assert.NotNil(t, material)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			materialID,
		)
		entities := &entities.Material{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		material, err := materialRepoWithSqlMock.GetByID(ctx, mockDB.DB, materialID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, material)

	})
}

func TestMaterialRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		materialRepoWithSqlMock *MaterialRepo
		mockDB                  *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				entity := &entities.Material{}
				fields, _ := entity.FieldMap()
				scanFields := database.GetScanFields(entity, fields)
				rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				entity := &entities.Material{}
				fields, _ := entity.FieldMap()
				scanFields := database.GetScanFields(entity, fields)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			materialRepoWithSqlMock, mockDB, _ = MaterialRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)
			billingItem, err := materialRepoWithSqlMock.GetAll(ctx, mockDB.DB)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NotNil(t, billingItem)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}
