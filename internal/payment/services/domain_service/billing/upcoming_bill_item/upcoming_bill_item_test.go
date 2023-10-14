package service

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillingForCustomOrderService_CreateBillItemForCustomOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		mockUpcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when voiding upcoming bill items by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Order{
				OrderID: pgtype.Text{
					String: constant.OrderID,
					Status: pgtype.Present,
				},
			},
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("VoidUpcomingBillItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: entities.Order{
				OrderID: pgtype.Text{
					String: constant.OrderID,
					Status: pgtype.Present,
				},
			},
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("VoidUpcomingBillItemsByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUpcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &UpcomingBillItemService{
				UpcomingBillItemRepo: mockUpcomingBillItemRepo,
			}

			testCase.Setup(testCase.Ctx)

			orderReq := testCase.Req.(entities.Order)

			err := s.VoidUpcomingBillItemsByOrder(testCase.Ctx, db, orderReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockUpcomingBillItemRepo)
		})
	}
}

func TestUpcomingBillItemService_CreateUpcomingBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		mockUpcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testCases := []utils.TestCase{
		{
			Name:         "happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          entities.BillItem{},
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUpcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &UpcomingBillItemService{
				UpcomingBillItemRepo: mockUpcomingBillItemRepo,
			}

			testCase.Setup(testCase.Ctx)

			billItem := entities.BillItem{
				BillItemSequenceNumber: pgtype.Int4{
					Int:    1,
					Status: pgtype.Present,
				},
				StudentID: pgtype.Text{
					String: "student-id",
					Status: pgtype.Present,
				},
				StudentProductID: pgtype.Text{
					String: "student-product",
					Status: pgtype.Present,
				},
				OrderID: pgtype.Text{
					String: "order-id",
					Status: pgtype.Present,
				},
				BillDate: pgtype.Timestamptz{
					Time:   time.Now(),
					Status: pgtype.Present,
				},
				BillSchedulePeriodID: pgtype.Text{
					String: "bp-ID",
					Status: pgtype.Present,
				},
				ProductID: pgtype.Text{
					String: "product-id",
					Status: pgtype.Present,
				},
				ProductDescription: pgtype.Text{
					String: "product-1",
					Status: pgtype.Present,
				},
				TaxID: pgtype.Text{
					String: "tax-id",
					Status: pgtype.Present,
				},
				DiscountID: pgtype.Text{
					String: "discount-id",
					Status: pgtype.Present,
				},
			}
			err := s.CreateUpcomingBillItem(testCase.Ctx, db, billItem)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockUpcomingBillItemRepo)
		})
	}

}

func TestUpcomingBillItemService_GetUpcomingBillItemsForGenerate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		mockUpcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  entities.BillItem{},
			ExpectedResp: []entities.UpcomingBillItem{
				{
					OrderID: pgtype.Text{
						String: "order-id",
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("GetUpcomingBillItemsForGenerate", ctx, mock.Anything).Return([]entities.UpcomingBillItem{
					{
						OrderID: pgtype.Text{
							String: "order-id",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUpcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &UpcomingBillItemService{
				UpcomingBillItemRepo: mockUpcomingBillItemRepo,
			}

			testCase.Setup(testCase.Ctx)

			_, err := s.GetUpcomingBillItemsForGenerate(testCase.Ctx, db)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockUpcomingBillItemRepo)
		})
	}
}

func TestUpcomingBillItemService_AddExecuteNoteForCurrentUpcomingBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		mockUpcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  entities.BillItem{},
			ExpectedResp: []entities.UpcomingBillItem{
				{
					OrderID: pgtype.Text{
						String: "order-id",
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("AddUpcomingExecuteNote", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUpcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &UpcomingBillItemService{
				UpcomingBillItemRepo: mockUpcomingBillItemRepo,
			}

			testCase.Setup(testCase.Ctx)
			entity := entities.UpcomingBillItem{}
			importErr := errors.Errorf("err")
			err := s.AddExecuteNoteForCurrentUpcomingBillItem(testCase.Ctx, db, entity, importErr)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockUpcomingBillItemRepo)
		})
	}
}

func TestUpcomingBillItemService_UpdateCurrentUpcomingBillItemStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                       *mockDb.Ext
		mockUpcomingBillItemRepo *mockRepositories.MockUpcomingBillItemRepo
	)

	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  entities.BillItem{},
			ExpectedResp: []entities.UpcomingBillItem{
				{
					OrderID: pgtype.Text{
						String: "order-id",
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockUpcomingBillItemRepo.On("UpdateCurrentUpcomingBillItemStatus", ctx, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUpcomingBillItemRepo = &mockRepositories.MockUpcomingBillItemRepo{}
			s := &UpcomingBillItemService{
				UpcomingBillItemRepo: mockUpcomingBillItemRepo,
			}

			testCase.Setup(testCase.Ctx)
			entity := entities.UpcomingBillItem{}
			err := s.UpdateCurrentUpcomingBillItemStatus(testCase.Ctx, db, entity)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockUpcomingBillItemRepo)
		})
	}
}
