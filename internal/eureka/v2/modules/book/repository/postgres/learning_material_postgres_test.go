package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestLearningMaterialRepo_UpdatePublishStatusLearningMaterials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mockDB := &mock_database.Ext{}

	lmRepo := &LearningMaterialRepo{
		DB: mockDB,
	}

	testCases := []TestCase{
		{
			name: "Happy case",
			req: []domain.LearningMaterial{
				{
					ID:        "LM_1",
					Published: true,
				},
				{
					ID:        "LM_2",
					Published: false,
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))

				mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
			expectedResp: nil,
		},
		{
			name: "Batch Exec error",
			req: []domain.LearningMaterial{
				{
					ID:        "LM_1",
					Published: true,
				},
				{
					ID:        "LM_2",
					Published: false,
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))

				mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)

				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
			expectedErr: fmt.Errorf("updatePublishStatusLearningMaterials batchResults.Exec: %w", puddle.ErrClosedPool),
		},
		{
			name: "Batch Exec with no RowsAffected",
			req: []domain.LearningMaterial{
				{
					ID:        "LM_1",
					Published: true,
				},
				{
					ID:        "LM_2",
					Published: false,
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))

				mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)

				batchResults.On("Close").Once().Return(nil)
			},
			expectedErr: fmt.Errorf("updatePublishStatusLearningMaterials no item updated"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)

			err := lmRepo.UpdatePublishStatusLearningMaterials(ctx, testCase.req.([]domain.LearningMaterial))

			if testCase.expectedErr != nil {
				assert.Equal(t, err, testCase.expectedErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestLearningMaterialRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	repo := &LearningMaterialRepo{DB: mockDB.DB}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name:    "happy case",
			Ctx:     ctx,
			Request: "lm_id",
			Setup: func(ctx context.Context) {
				lmDto := dto.LearningMaterialDto{}
				fields, values := lmDto.FieldMap()

				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("lm_id"))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.LearningMaterial{},
			ExpectedError:    nil,
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: "lm_id",
			Setup: func(ctx context.Context) {
				lmDto := dto.LearningMaterialDto{}
				fields, values := lmDto.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text("lm_id"))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.LearningMaterial{},
			ExpectedError:    fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", pgx.ErrNoRows)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			res, err := repo.GetByID(tc.Ctx, tc.Request.(string))
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestLearningMaterialRepo_GetManyByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	keys := []string{"key", "key-1"}
	textArr := database.TextArray(keys)

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &LearningMaterialRepo{mockDB.DB}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), textArr)
		expectedErr := errors.NewDBError("LearningMaterialRepo.GetManyByIDs", puddle.ErrClosedPool)

		// act
		actual, err := repo.GetManyByIDs(ctx, keys)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed return conversion error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &LearningMaterialRepo{DB: mockDB.DB}
		asm := &dto.LearningMaterialDto{}
		_, val := asm.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"), textArr).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("LearningMaterialRepo.scanLearningMaterials", scanErr)

		// act
		c, err := repo.GetManyByIDs(ctx, keys)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &LearningMaterialRepo{DB: mockDB.DB}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), textArr)
		e := &dto.LearningMaterialDto{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByIDs(ctx, keys)

		// assert
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
