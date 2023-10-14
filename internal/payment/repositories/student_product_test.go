package repositories

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentProductRepoWithSqlMock() (*StudentProductRepo, *testutil.MockDB) {
	repo := &StudentProductRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentProductRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentProduct{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Create(ctx, mockDB.DB, *mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert student product fail", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, *mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentProduct: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert student product", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := repo.Create(ctx, mockDB.DB, *mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert StudentProduct: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentProductRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.StudentProduct{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Update(ctx, mockDB.DB, *mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student product fail", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := repo.Update(ctx, mockDB.DB, *mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Student Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update student product", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := repo.Update(ctx, mockDB.DB, *mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Student Product: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentProductRepo_GetLatestEndDateStudentProductWithProductIDAndStudentID(t *testing.T) {
	t.Parallel()

	r, mockDB := StudentProductRepoWithSqlMock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	var mocks []interface{}
	for range studentProductFieldNames {
		mocks = append(mocks, mock.Anything)
	}

	t.Run(constant.HappyCase, func(t *testing.T) {
		expectedStudentProductIDs := []string{"student_product_id_1", "student_product_id_2"}
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		for range expectedStudentProductIDs {
			rows.On("Next").Once().Return(true)
			rows.On("Scan", mocks...).Once().Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)
	})
	t.Run("Empty case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentProducts, err := r.GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)
		assert.Equal(t, len(studentProducts), 0)
	})
	t.Run(constant.FailCaseErrorQuery, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, constant.ErrDefault)
		_, err := r.GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx, mockDB.DB, "1", "1")
		assert.Equal(t, constant.ErrDefault.Error(), err.Error())
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mocks...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetLatestEndDateStudentProductWithProductIDAndStudentID(ctx, mockDB.DB, "1", "1")
		assert.Equal(t, constant.ErrDefault.Error(), err.Error())
	})

	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
}

func TestStudentProductRepo_GetStudentProductForUpdateByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepoWithSqlMock, mockDB := StudentProductRepoWithSqlMock()

	studentProductID := "student_product_id_1"
	mockEntity := &entities.StudentProduct{}
	fields, values := mockEntity.FieldMap()
	testCases := []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Ctx:          nil,
			Req:          studentProductID,
			ExpectedResp: &entities.StudentProduct{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
		{
			Name:        "Fail case: Error when scan row",
			Ctx:         nil,
			Req:         studentProductID,
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			req := (testCase.Req).(string)
			_, err := studentProductRepoWithSqlMock.GetStudentProductForUpdateByStudentProductID(ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestStudentProductRepo_UpdateStatusStudentProductAndResetStudentProductLabel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.UpdateStatusStudentProductAndResetStudentProductLabel(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update student product fail", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := repo.UpdateStatusStudentProductAndResetStudentProductLabel(ctx, mockDB.DB, "1", "1")
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Student Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update student product", func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := repo.UpdateStatusStudentProductAndResetStudentProductLabel(ctx, mockDB.DB, "1", "1")
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Student Product: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentProductRepo_GetStudentProductsByStudentProductLabelForUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductsByStudentProductLabelForUpdate(ctx, mockDB.DB, []string{"1"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductsByStudentProductLabelForUpdate(ctx, mockDB.DB, []string{"1"})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestGetUniqueProductsByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetUniqueProductsByStudentID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetUniqueProductsByStudentID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestGetUniqueProductsByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetUniqueProductsByStudentIDs(ctx, mockDB.DB, []string{"1"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetUniqueProductsByStudentIDs(ctx, mockDB.DB, []string{"1"})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentProductRepo_GetStudentProductByStudentProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepoWithSqlMock, mockDB := StudentProductRepoWithSqlMock()

	studentProductID := "student_product_id_1"
	mockEntity := &entities.StudentProduct{}
	fields, values := mockEntity.FieldMap()
	testCases := []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Ctx:          nil,
			Req:          studentProductID,
			ExpectedResp: &entities.StudentProduct{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.MockRowScanFields(nil, fields, values)
			},
		},
		{
			Name:        "Fail case: Error when scan row",
			Ctx:         nil,
			Req:         studentProductID,
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Once().Return(mockDB.Row)
				mockDB.MockRowScanFields(constant.ErrDefault, fields, values)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx)
			req := (testCase.Req).(string)
			_, err := studentProductRepoWithSqlMock.GetStudentProductByStudentProductID(ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.ExpectedErr, err)
		})
	}
}

func TestCountStudentProductIDsByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.CountStudentProductIDsByStudentIDAndLocationIDs(ctx, mockDB.DB, "1", []string{"location_1", "location_2"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.CountStudentProductIDsByStudentIDAndLocationIDs(ctx, mockDB.DB, "1", []string{})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestGetStudentProductIDsByRootStudentProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductIDsByRootStudentProductID(ctx, mockDB.DB, "1")
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductIDsByRootStudentProductID(ctx, mockDB.DB, "1")
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestGetByStudentIDAndLocationIDsWithPaging(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetByStudentIDAndLocationIDsWithPaging(ctx, mockDB.DB, "1", []string{"location_1", "location_2"}, int64(1), int64(1))
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetByStudentIDAndLocationIDsWithPaging(ctx, mockDB.DB, "1", []string{}, int64(1), int64(1))
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetByStudentIDAndLocationIDsWithPaging(ctx, mockDB.DB, "1", []string{"location_1", "location_2"}, int64(1), int64(1))
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestStudentProductRepo_GetByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockStudentProductRepo, mockDB := StudentProductRepoWithSqlMock()
	testcases := buildStudentProductGetByIDsTestcases(mockDB)

	for _, testcase := range testcases {
		refTestcase := testcase
		t.Run(refTestcase.Name, func(t *testing.T) {
			refTestcase.Setup(ctx)
			expectedStudentProducts := refTestcase.ExpectedResp.([]*entities.StudentProduct)
			studentProducts, err := mockStudentProductRepo.GetByIDs(ctx, mockDB.DB, refTestcase.Req.([]string))
			if refTestcase.ExpectedErr != nil {
				assert.Equal(t, refTestcase.ExpectedErr.Error(), err.Error())
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, len(expectedStudentProducts), len(studentProducts))
			for idx, expectedStudentProduct := range expectedStudentProducts {
				assert.Equal(t, expectedStudentProduct.StudentProductID, studentProducts[idx].StudentProductID)
				assert.Equal(t, expectedStudentProduct.StudentID, studentProducts[idx].StudentID)
				assert.Equal(t, expectedStudentProduct.ProductID, studentProducts[idx].ProductID)
				assert.Equal(t, expectedStudentProduct.UpcomingBillingDate, studentProducts[idx].UpcomingBillingDate)
				assert.Equal(t, expectedStudentProduct.StartDate, studentProducts[idx].StartDate)
				assert.Equal(t, expectedStudentProduct.EndDate, studentProducts[idx].EndDate)
				assert.Equal(t, expectedStudentProduct.ProductStatus, studentProducts[idx].ProductStatus)
				assert.Equal(t, expectedStudentProduct.ApprovalStatus, studentProducts[idx].ApprovalStatus)
				assert.Equal(t, expectedStudentProduct.UpdatedAt, studentProducts[idx].UpdatedAt)
				assert.Equal(t, expectedStudentProduct.CreatedAt, studentProducts[idx].CreatedAt)
				assert.Equal(t, expectedStudentProduct.DeletedAt, studentProducts[idx].DeletedAt)
				assert.Equal(t, expectedStudentProduct.LocationID, studentProducts[idx].LocationID)
				assert.Equal(t, expectedStudentProduct.StudentProductLabel, studentProducts[idx].StudentProductLabel)
				assert.Equal(t, expectedStudentProduct.UpdatedFromStudentProductID, studentProducts[idx].UpdatedFromStudentProductID)
				assert.Equal(t, expectedStudentProduct.UpdatedToStudentProductID, studentProducts[idx].UpdatedToStudentProductID)
				assert.Equal(t, expectedStudentProduct.IsUnique, studentProducts[idx].IsUnique)
				assert.Equal(t, expectedStudentProduct.ResourcePath, studentProducts[idx].ResourcePath)
			}
		})
	}
	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
}

func buildStudentProductGetByIDsTestcases(mockDB *testutil.MockDB) []utils.TestCase {
	mockEntity := &entities.StudentProduct{}
	fieldNames, _ := mockEntity.FieldMap()

	studentProductIDs := []string{"1", "2"}
	stmt := fmt.Sprintf(
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = ANY($1)
		`,
		strings.Join(fieldNames, ","),
		mockEntity.TableName(),
	)
	args := []interface{}{
		mock.Anything,
		stmt,
		studentProductIDs,
	}

	expectedStudentProducts := []*entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: "1",
			},
			StudentID: pgtype.Text{
				String: "1",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			UpcomingBillingDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			StartDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			EndDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ProductStatus: pgtype.Text{
				String: "",
			},
			ApprovalStatus: pgtype.Text{
				String: "",
			},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			DeletedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			LocationID: pgtype.Text{
				String: "",
			},
			UpdatedFromStudentProductID: pgtype.Text{
				String: "",
			},
			UpdatedToStudentProductID: pgtype.Text{
				String: "",
			},
			StudentProductLabel: pgtype.Text{
				String: "WITHDRAWAL_SCHEDULED",
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "2",
			},
			StudentID: pgtype.Text{
				String: "2",
			},
			ProductID: pgtype.Text{
				String: "2",
			},
			UpcomingBillingDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			StartDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			EndDate: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ProductStatus: pgtype.Text{
				String: "",
			},
			ApprovalStatus: pgtype.Text{
				String: "",
			},
			UpdatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			CreatedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			DeletedAt: pgtype.Timestamptz{
				Time: time.Now(),
			},
			ResourcePath: pgtype.Text{
				String: "",
			},
			LocationID: pgtype.Text{
				String: "",
			},
			UpdatedFromStudentProductID: pgtype.Text{
				String: "",
			},
			UpdatedToStudentProductID: pgtype.Text{
				String: "",
			},
			StudentProductLabel: pgtype.Text{
				String: "",
			},
		},
	}

	return []utils.TestCase{
		{
			Name:         constant.HappyCase,
			Req:          studentProductIDs,
			ExpectedResp: expectedStudentProducts,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Times(len(expectedStudentProducts)).Return(true)

				studentProduct := &entities.StudentProduct{}
				fields, _ := studentProduct.FieldMap()
				scanFields := database.GetScanFields(studentProduct, fields)

				for _, expectedStudentProduct := range expectedStudentProducts {
					refExpectedStudentProduct := expectedStudentProduct
					rows.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
						args[0].(*pgtype.Text).String = refExpectedStudentProduct.StudentProductID.String
						args[1].(*pgtype.Text).String = refExpectedStudentProduct.StudentID.String
						args[2].(*pgtype.Text).String = refExpectedStudentProduct.ProductID.String
						args[3].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.UpcomingBillingDate.Time
						args[4].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.StartDate.Time
						args[5].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.EndDate.Time
						args[6].(*pgtype.Text).String = refExpectedStudentProduct.ProductStatus.String
						args[7].(*pgtype.Text).String = refExpectedStudentProduct.ApprovalStatus.String
						args[8].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.UpdatedAt.Time
						args[9].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.CreatedAt.Time
						args[10].(*pgtype.Timestamptz).Time = refExpectedStudentProduct.DeletedAt.Time
						args[11].(*pgtype.Text).String = refExpectedStudentProduct.LocationID.String
						args[12].(*pgtype.Text).String = refExpectedStudentProduct.StudentProductLabel.String
						args[13].(*pgtype.Text).String = refExpectedStudentProduct.UpdatedFromStudentProductID.String
						args[14].(*pgtype.Text).String = refExpectedStudentProduct.UpdatedToStudentProductID.String
						args[15].(*pgtype.Bool).Bool = refExpectedStudentProduct.IsUnique.Bool
						args[16].(*pgtype.Text).String = refExpectedStudentProduct.ResourcePath.String
					}).Return(nil)
				}
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         "empty case",
			Req:          studentProductIDs,
			ExpectedResp: []*entities.StudentProduct{},
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:         constant.FailCaseErrorQuery,
			Req:          studentProductIDs,
			ExpectedResp: []*entities.StudentProduct{},
			ExpectedErr:  constant.ErrDefault,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, constant.ErrDefault)
			},
		},
		{
			Name:         constant.FailCaseErrorRow,
			Req:          studentProductIDs,
			ExpectedResp: []*entities.StudentProduct{},
			ExpectedErr:  fmt.Errorf("row.Scan: %w", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", args...).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				order := &entities.StudentProduct{}
				fields, _ := order.FieldMap()
				scanFields := database.GetScanFields(order, fields)
				rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
}

func TestGetStudentProductAssociatedByStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run(constant.HappyCase, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductAssociatedByStudentProductID(ctx, mockDB.DB, []string{"student_product_1", "student_product_2"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		repo, mockDB := StudentProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)
		studentProduct := &entities.StudentProduct{}
		fields, _ := studentProduct.FieldMap()
		scanFields := database.GetScanFields(studentProduct, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)

		//rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := repo.GetStudentProductAssociatedByStudentProductID(ctx, mockDB.DB, []string{"student_product_1", "student_product_1"})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func Test_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentProductRepoWithSqlMock, mockDB := StudentProductRepoWithSqlMock()

	const studentProductID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			studentProductID,
		)
		entity := &entities.StudentProduct{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		product, err := studentProductRepoWithSqlMock.GetByID(ctx, mockDB.DB, studentProductID)
		assert.Nil(t, err)
		assert.NotNil(t, product)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			studentProductID,
		)
		e := &entities.StudentProduct{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		product, err := studentProductRepoWithSqlMock.GetByID(ctx, mockDB.DB, studentProductID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, product)

	})
}
func TestStudentProductRepo_GetActiveRecurringProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	r, mockDB := StudentProductRepoWithSqlMock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	var mocks []interface{}
	for range studentProductFieldNames {
		mocks = append(mocks, mock.Anything)
	}

	t.Run(constant.HappyCase, func(t *testing.T) {
		expectedStudentProductIDs := []string{"student_product_id_1", "student_product_id_2"}
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		for range expectedStudentProductIDs {
			rows.On("Next").Once().Return(true)
			rows.On("Scan", mocks...).Once().Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetActiveRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1", []string{})
		assert.Nil(t, err)
	})
	t.Run(constant.HappyCase, func(t *testing.T) {
		expectedStudentProductIDs := []string{"student_product_id_1", "student_product_id_2"}
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		for range expectedStudentProductIDs {
			rows.On("Next").Once().Return(true)
			rows.On("Scan", mocks...).Once().Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetActiveRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1", []string{"student_product_id_4"})
		assert.Nil(t, err)
	})
	t.Run("Empty case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentProducts, err := r.GetActiveRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1", []string{})
		assert.Nil(t, err)
		assert.Equal(t, len(studentProducts), 0)
	})
	t.Run(constant.FailCaseErrorQuery, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, constant.ErrDefault)
		_, err := r.GetActiveRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1", []string{})
		assert.Equal(t, constant.ErrDefault.Error(), err.Error())
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mocks...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetActiveRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1", []string{})
		assert.Contains(t, err.Error(), constant.ErrDefault.Error())
	})

	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
}

func TestStudentProductRepo_GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	r, mockDB := StudentProductRepoWithSqlMock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentProductEntity := &entities.StudentProduct{}
	studentProductFieldNames, _ := studentProductEntity.FieldMap()
	var mocks []interface{}
	for range studentProductFieldNames {
		mocks = append(mocks, mock.Anything)
	}

	t.Run(constant.HappyCase, func(t *testing.T) {
		expectedStudentProductIDs := []string{"student_product_id_1", "student_product_id_2"}
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		for range expectedStudentProductIDs {
			rows.On("Next").Once().Return(true)
			rows.On("Scan", mocks...).Once().Return(nil)
		}
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)
	})
	t.Run("Empty case", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		studentProducts, err := r.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1")
		assert.Nil(t, err)
		assert.Equal(t, len(studentProducts), 0)
	})
	t.Run(constant.FailCaseErrorQuery, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, constant.ErrDefault)
		_, err := r.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1")
		assert.Equal(t, constant.ErrDefault.Error(), err.Error())
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", mocks...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := r.GetIgnoreStudentProductIDOfRecurringProductsOfStudentInLocation(ctx, mockDB.DB, "1", "1")
		assert.Contains(t, err.Error(), constant.ErrDefault.Error())
	})

	mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
}
