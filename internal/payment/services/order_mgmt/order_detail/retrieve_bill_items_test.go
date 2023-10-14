package order_detail

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOrderDetail_RetrieveBillItemsOfOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductForOrderDetail
		billItemService       *mockServices.IBillItemServiceForOrderDetail
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: with nil paging",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveBillingOfOrderDetailsRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: fmt.Errorf("invalid paging data with error"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: when get billing description by order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveBillingOfOrderDetailsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billItemService.On("GetBillItemDescriptionsByOrderIDWithPaging",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{{}}, 0, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: when verified total",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveBillingOfOrderDetailsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: status.Error(codes.Internal, "Error offset"),
			Setup: func(ctx context.Context) {
				billItemService.On("GetBillItemDescriptionsByOrderIDWithPaging",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: pgtype.Numeric{
								Int:    big.NewInt(123),
								Status: pgtype.Present,
							},
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: pgtype.Numeric{
								Int:    big.NewInt(123),
								Status: pgtype.Present,
							},
						},
					},
				}, 0, nil)
			},
		},
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveBillingOfOrderDetailsRequest{
				OrderId: constant.OrderID,
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billItemService.On("GetBillItemDescriptionsByOrderIDWithPaging",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]utils.BillItemForRetrieveApi{
					{
						BillItemEntity: entities.BillItem{
							AdjustmentPrice: pgtype.Numeric{
								Int:    big.NewInt(123),
								Status: pgtype.Present,
							},
						},
					},
					{
						BillItemEntity: entities.BillItem{
							FinalPrice: pgtype.Numeric{
								Int:    big.NewInt(123),
								Status: pgtype.Present,
							},
						},
					},
				}, 20, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductForOrderDetail)
			billItemService = new(mockServices.IBillItemServiceForOrderDetail)
			testCase.Setup(testCase.Ctx)
			s := &OrderDetail{
				DB:                    db,
				StudentProductService: studentProductService,
				BillItemService:       billItemService,
			}

			resp, err := s.RetrieveBillItemsOfOrder(testCase.Ctx, testCase.Req.(*pb.RetrieveBillingOfOrderDetailsRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				billItemService,
				studentProductService,
			)
		})
	}
}
