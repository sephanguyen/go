package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentProductService_RetrieveActiveStudentProductsOfStudentInLocation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetActiveStudentProductsByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get student product by student ID and location ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetActiveStudentProductsByStudentIDAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentProduct{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				DB:                 db,
				StudentProductRepo: studentProductRepo,
			}
			_, err := s.RetrieveActiveStudentProductsOfStudentInLocation(testCase.Ctx, db, mock.Anything, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestStudentProductService_TestRetrieveStudentProductByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get student product by ID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				DB:                 db,
				StudentProductRepo: studentProductRepo,
			}
			_, err := s.RetrieveStudentProductByID(testCase.Ctx, db, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestStudentProductService_TestRetrieveStudentsProductByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get student product by IDs",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				DB:                 db,
				StudentProductRepo: studentProductRepo,
			}
			_, err := s.RetrieveStudentProductsByIDs(testCase.Ctx, db, []string{mock.Anything})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestStudentProductService_TestRetrieveStudentsProductByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		studentProductRepo *mockRepositories.MockStudentProductRepo
		orderItemRepo      *mockRepositories.MockOrderItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetStudentProductIDsByOrderID", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get student product ids by order IDs",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetStudentProductIDsByOrderID", ctx, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get student product by IDs",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetStudentProductIDsByOrderID", ctx, mock.Anything, mock.Anything).Return([]string{}, nil)
				studentProductRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.StudentProduct{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductRepo = new(mockRepositories.MockStudentProductRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				DB:                 db,
				StudentProductRepo: studentProductRepo,
				OrderItemRepo:      orderItemRepo,
			}
			_, err := s.RetrieveStudentProductsByOrderID(testCase.Ctx, db, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductRepo)
		})
	}
}

func TestRetrieveDiscountOfStudentProduct(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		billItemRepo *mockRepositories.MockBillItemRepo
		discountRepo *mockRepositories.MockDiscountRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error on discount GetLastBillItemOfStudentProduct",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLastBillItemOfStudentProduct", ctx, mock.Anything, mock.Anything).Return(entities.BillItem{
					DiscountID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error on discount GetByID",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLastBillItemOfStudentProduct", ctx, mock.Anything, mock.Anything).Return(entities.BillItem{
					DiscountID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, nil)
				discountRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: With discount attached",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLastBillItemOfStudentProduct", ctx, mock.Anything, mock.Anything).Return(entities.BillItem{
					DiscountID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
				}, nil)
				discountRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Discount{}, nil)
			},
		},
		{
			Name:        "Happy case: Without discount attached",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemRepo.On("GetLastBillItemOfStudentProduct", ctx, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			billItemRepo = new(mockRepositories.MockBillItemRepo)
			discountRepo = new(mockRepositories.MockDiscountRepo)
			testCase.Setup(testCase.Ctx)
			s := &StudentProductService{
				DB:           db,
				BillItemRepo: billItemRepo,
				DiscountRepo: discountRepo,
			}
			_, err := s.RetrieveDiscountOfStudentProduct(testCase.Ctx, mock.Anything)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, billItemRepo, discountRepo)
		})
	}
}
