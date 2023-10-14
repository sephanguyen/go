package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTaxService_validateTaxWithBillItem(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db      *mockDb.Ext
		taxRepo *mockRepositories.MockTaxRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when inconsistency tax between tax in product and tax in bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InconsistentTax, nil),
			Req: []interface{}{
				utils.BillingItemData{BillingItem: &pb.BillingItem{
					TaxItem: nil,
				}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when tax category is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with ID %v change tax category from %v to %v",
				constant.ProductID,
				pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String(),
				pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String()),
			Req: []interface{}{
				utils.BillingItemData{BillingItem: &pb.BillingItem{
					TaxItem: &pb.TaxBillItem{
						TaxId:       constant.TaxID,
						TaxCategory: pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
					},
					ProductId: constant.ProductID,
				}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when tax category is exclusive",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This tax category is not supported in this version"),
			Req: []interface{}{
				utils.BillingItemData{BillingItem: &pb.BillingItem{
					TaxItem: &pb.TaxBillItem{
						TaxId:       constant.TaxID,
						TaxCategory: pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
					},
					ProductId: constant.ProductID,
				}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String(),
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: Error when tax percentages is different",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with ID %v change tax percentage from %v to %v",
				constant.ProductID,
				1,
				2),
			Req: []interface{}{
				utils.BillingItemData{BillingItem: &pb.BillingItem{
					TaxItem: &pb.TaxBillItem{
						TaxId:         constant.TaxID,
						TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
						TaxPercentage: 1.0,
					},
					ProductId: constant.ProductID,
				}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 2,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when tax amount is incorrect",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Incorrect tax amount actual = %v vs expected = %v", 2, 0.8910891),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						TaxItem: &pb.TaxBillItem{
							TaxId:         constant.TaxID,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxPercentage: 1.0,
							TaxAmount:     2,
						},
						ProductId: constant.ProductID,
						Price:     100.0,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountType:        0,
							DiscountAmountType:  0,
							DiscountAmountValue: 0,
							DiscountAmount:      10,
						},
					}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.BillingItemData{
					BillingItem: &pb.BillingItem{
						TaxItem: &pb.TaxBillItem{
							TaxId:         constant.TaxID,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxPercentage: 1.0,
							TaxAmount:     0.8910891,
						},
						ProductId: constant.ProductID,
						Price:     100.0,
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          constant.DiscountID,
							DiscountType:        0,
							DiscountAmountType:  0,
							DiscountAmountValue: 0,
							DiscountAmount:      10,
						},
					}},
				entities.Tax{
					TaxID: pgtype.Text{
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			taxRepo = &mockRepositories.MockTaxRepo{}
			s := &TaxService{
				taxRepo: taxRepo,
			}
			testCase.Setup(testCase.Ctx)

			billingDataReq := testCase.Req.([]interface{})[0].(utils.BillingItemData)
			taxReq := testCase.Req.([]interface{})[1].(entities.Tax)
			err := s.validateTaxWithBillItem(billingDataReq, taxReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, taxRepo)
		})
	}
}

func TestTaxService_getTax(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db      *mockDb.Ext
		taxRepo *mockRepositories.MockTaxRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get tax by id for update",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"getting tax of product %v have error %s",
				pgtype.Text{
					String: constant.ProductID,
				},
				constant.ErrDefault),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: When tax status is not present",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Null,
					},
				},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			taxRepo = &mockRepositories.MockTaxRepo{}
			s := &TaxService{
				taxRepo: taxRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)
			_, err := s.getTax(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, taxRepo)
		})
	}
}

func TestTaxService_IsValidTaxForCustomOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db      *mockDb.Ext
		taxRepo *mockRepositories.MockTaxRepo
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get tax by id for update",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal,
				"Error when retrieving tax id %v with error: %s",
				constant.TaxID,
				constant.ErrDefault),
			Req: &pb.CustomBillingItem{
				Name: constant.CustomBillingItemName,
				TaxItem: &pb.TaxBillItem{
					TaxId:         constant.TaxID,
					TaxPercentage: 1,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     0,
				},
				Price: 100,
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when tax category name is incorrect",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition,
				"Product with name %v changed tax category from %v to %v",
				constant.CustomBillingItemName,
				pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
				pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String()),
			Req: &pb.CustomBillingItem{
				Name: constant.CustomBillingItemName,
				TaxItem: &pb.TaxBillItem{
					TaxId:         constant.TaxID,
					TaxPercentage: 1,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     0,
				},
				Price: 100,
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String(),
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when tax category is TAX_CATEGORY_EXCLUSIVE (note supported in this version)",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This tax category %v is not supported in this version", pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String()),
			Req: &pb.CustomBillingItem{
				Name: constant.CustomBillingItemName,
				TaxItem: &pb.TaxBillItem{
					TaxId:         constant.TaxID,
					TaxPercentage: 1,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE,
					TaxAmount:     0,
				},
				Price: 100,
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_EXCLUSIVE.String(),
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when tax amount is incorrect",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Incorrect tax amount actual = %v vs expected = %v", 1, 0.990099),
			Req: &pb.CustomBillingItem{
				Name: constant.CustomBillingItemName,
				TaxItem: &pb.TaxBillItem{
					TaxId:         constant.TaxID,
					TaxPercentage: 1,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     1,
				},
				Price: 100,
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CustomBillingItem{
				Name: constant.CustomBillingItemName,
				TaxItem: &pb.TaxBillItem{
					TaxId:         constant.TaxID,
					TaxPercentage: 1,
					TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
					TaxAmount:     0.990099,
				},
				Price: 100,
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				}, nil)
			},
		},
		{
			Name: "Happy case: When tax of custom order is nil",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:  &pb.CustomBillingItem{TaxItem: nil},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			taxRepo = &mockRepositories.MockTaxRepo{}
			s := &TaxService{
				taxRepo: taxRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(*pb.CustomBillingItem)
			err := s.IsValidTaxForCustomOrder(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, taxRepo)
		})
	}
}

func TestTaxService_IsValidTaxForRecurringBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db      *mockDb.Ext
		taxRepo *mockRepositories.MockTaxRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when validate tax with bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InconsistentTax, nil),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							TaxItem: nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							TaxItem: &pb.TaxBillItem{
								TaxId:         constant.TaxID,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxPercentage: 1.0,
								TaxAmount:     0.8910891,
							},
							ProductId: constant.ProductID,
							Price:     100.0,
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountType:        0,
								DiscountAmountType:  0,
								DiscountAmountValue: 0,
								DiscountAmount:      10,
							},
						}},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			taxRepo = &mockRepositories.MockTaxRepo{}
			s := &TaxService{
				taxRepo: taxRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)
			err := s.IsValidTaxForRecurringBilling(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				fmt.Println(err)
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, taxRepo)
		})
	}
}

func TestTaxService_IsValidTaxForOneTimeBilling(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db      *mockDb.Ext
		taxRepo *mockRepositories.MockTaxRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when validate tax with bill item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: utils.StatusErrWithDetail(codes.FailedPrecondition, constant.InconsistentTax, nil),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							TaxItem: nil,
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				}, nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: utils.OrderItemData{
				ProductInfo: entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
				},
				BillItems: []utils.BillingItemData{
					{
						BillingItem: &pb.BillingItem{
							TaxItem: &pb.TaxBillItem{
								TaxId:         constant.TaxID,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxPercentage: 1.0,
								TaxAmount:     0.8910891,
							},
							ProductId: constant.ProductID,
							Price:     100.0,
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          constant.DiscountID,
								DiscountType:        0,
								DiscountAmountType:  0,
								DiscountAmountValue: 0,
								DiscountAmount:      10,
							},
						}},
				},
			},
			Setup: func(ctx context.Context) {
				taxRepo.On("GetByIDForUpdate", ctx, db, constant.TaxID).Return(entities.Tax{
					TaxID: pgtype.Text{
						String: constant.TaxID,
						Status: pgtype.Present,
					},
					TaxCategory: pgtype.Text{
						String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(),
					},
					TaxPercentage: pgtype.Int4{
						Int: 1,
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			taxRepo = &mockRepositories.MockTaxRepo{}
			s := &TaxService{
				taxRepo: taxRepo,
			}
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(utils.OrderItemData)
			err := s.IsValidTaxForOneTimeBilling(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				fmt.Println(err)
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, taxRepo)
		})
	}
}
