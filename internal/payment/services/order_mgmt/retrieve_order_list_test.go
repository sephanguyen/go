package ordermgmt

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRetrieveOrdersListService_RetrieveListOfOrdersWithFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)
	now := time.Now()
	expectedOrders := []entities.Order{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			StudentID:           pgtype.Text{String: "student_id_1"},
			LocationID:          pgtype.Text{String: "location_id_1"},
			OrderSequenceNumber: pgtype.Int4{Int: 1},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			StudentID:           pgtype.Text{String: "student_id_2"},
			LocationID:          pgtype.Text{String: "location_id_2"},
			OrderSequenceNumber: pgtype.Int4{Int: 2},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			StudentID:           pgtype.Text{String: "student_id_3"},
			LocationID:          pgtype.Text{String: "location_id_3"},
			OrderSequenceNumber: pgtype.Int4{Int: 3},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			StudentID:           pgtype.Text{String: "student_id_4"},
			LocationID:          pgtype.Text{String: "location_id_4"},
			OrderSequenceNumber: pgtype.Int4{Int: 4},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_LOA.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
	}

	expectedOrderProducts := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
	}

	expectedProducts := []entities.Product{
		{
			ProductID: pgtype.Text{
				String: "1",
			},
			Name: pgtype.Text{
				String: "product_1",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
	}
	expectedLocations := []entities.Location{
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_2",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_3",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_4",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
	}
	expectedCreators := []entities.OrderCreator{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
				Status: 2,
			},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
				Status: 2,
			},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
				Status: 2,
			},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
				Status: 2,
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "Fail case: Error when get order stats by filter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, constant.ErrDefault)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get list of orders by filter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, constant.ErrDefault)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get order items by order ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, constant.ErrDefault)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get products by product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get creator info of orders",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedCreators, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.OrderId, item.OrderId)
					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
					assert.Equal(t, expectedItem.StudentId, item.StudentId)
					assert.Equal(t, expectedItem.StudentName, item.StudentName)
					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
					assert.Equal(t, expectedItem.OrderType, item.OrderType)
					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
			}
		})
	}
}

func TestRetrieveOrdersListService_RetrieveListOfOrdersWithoutFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)
	now := time.Now()
	expectedOrders := []entities.Order{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			StudentID:           pgtype.Text{String: "student_id_1"},
			LocationID:          pgtype.Text{String: "location_id_1"},
			OrderSequenceNumber: pgtype.Int4{Int: 1},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			StudentID:           pgtype.Text{String: "student_id_2"},
			LocationID:          pgtype.Text{String: "location_id_2"},
			OrderSequenceNumber: pgtype.Int4{Int: 2},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			StudentID:           pgtype.Text{String: "student_id_3"},
			LocationID:          pgtype.Text{String: "location_id_3"},
			OrderSequenceNumber: pgtype.Int4{Int: 3},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			StudentID:           pgtype.Text{String: "student_id_4"},
			LocationID:          pgtype.Text{String: "location_id_4"},
			OrderSequenceNumber: pgtype.Int4{Int: 4},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_LOA.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
	}
	expectedOrderProducts := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
	}
	expectedProducts := []entities.Product{
		{
			ProductID: pgtype.Text{
				String: "1",
			},
			Name: pgtype.Text{
				String: "product_1",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
	}
	expectedLocations := []entities.Location{
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_2",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_3",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_4",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "Fail case: Error when get order stats by filter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, constant.ErrDefault)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get list of orders by filter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, constant.ErrDefault)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get order items by order ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, constant.ErrDefault)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
			},
		},
		{
			Name: "Fail case: Error when get products by product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case (without filter)",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						UserID: pgtype.Text{
							String: constant.UserID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.OrderId, item.OrderId)
					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
					assert.Equal(t, expectedItem.StudentId, item.StudentId)
					assert.Equal(t, expectedItem.StudentName, item.StudentName)
					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
					assert.Equal(t, expectedItem.OrderType, item.OrderType)
					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
			}
		})
	}
}

func TestRetrieveOrdersListService_RetrieveListOfOrdersWithEmptyProducts(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)
	now := time.Now()
	expectedOrders := []entities.Order{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			StudentID:           pgtype.Text{String: "student_id_1"},
			LocationID:          pgtype.Text{String: "location_id_1"},
			OrderSequenceNumber: pgtype.Int4{Int: 1},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			StudentID:           pgtype.Text{String: "student_id_2"},
			LocationID:          pgtype.Text{String: "location_id_2"},
			OrderSequenceNumber: pgtype.Int4{Int: 2},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			StudentID:           pgtype.Text{String: "student_id_3"},
			LocationID:          pgtype.Text{String: "location_id_3"},
			OrderSequenceNumber: pgtype.Int4{Int: 3},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			StudentID:           pgtype.Text{String: "student_id_4"},
			LocationID:          pgtype.Text{String: "location_id_4"},
			OrderSequenceNumber: pgtype.Int4{Int: 4},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_LOA.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
	}
	expectedOrderProducts := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
	}
	expectedProducts := []entities.Product{}
	expectedLocations := []entities.Location{
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_2",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_3",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_4",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "happy case (with empty products)",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						IsReviewed:          true,
					},
				},
				PreviousPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						UserID: pgtype.Text{
							String: constant.UserID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.OrderId, item.OrderId)
					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
					assert.Equal(t, expectedItem.StudentId, item.StudentId)
					assert.Equal(t, expectedItem.StudentName, item.StudentName)
					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
					assert.Equal(t, expectedItem.OrderType, item.OrderType)
					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
			}
		})
	}
}

func TestRetrieveOrdersListService_RetrieveListOfOrdersWithCaseEmptyOrdersCauseFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)
	now := time.Now()
	expectedOrders := []entities.Order{}
	expectedOrderProducts := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
	}
	expectedProducts := []entities.Product{
		{
			ProductID: pgtype.Text{
				String: "1",
			},
			Name: pgtype.Text{
				String: "product_1",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
		{
			ProductID: pgtype.Text{
				String: "3",
			},
			Name: pgtype.Text{
				String: "product_3",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
	}
	expectedLocations := []entities.Location{
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_2",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_3",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_4",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "happy case (with empty orders cause filter)",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items:                    []*pb.RetrieveListOfOrdersResponse_Order{},
				PreviousPage:             nil,
				NextPage:                 nil,
				TotalItems:               58,
				TotalOfSubmitted:         20,
				TotalOfPending:           15,
				TotalOfRejected:          10,
				TotalOfVoided:            8,
				TotalOfInvoiced:          5,
				TotalOfOrderNeedToReview: 5,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 58,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 20,
					},
					TotalOfPending: pgtype.Int8{
						Int: 15,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 10,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 8,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 5,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						UserID: pgtype.Text{
							String: constant.UserID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.OrderId, item.OrderId)
					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
					assert.Equal(t, expectedItem.StudentId, item.StudentId)
					assert.Equal(t, expectedItem.StudentName, item.StudentName)
					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
					assert.Equal(t, expectedItem.OrderType, item.OrderType)
					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
			}
		})
	}
}

func TestRetrieveOrdersListService_RetrieveListOfOrdersWithFilterStudentName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)
	now := time.Now()
	expectedOrders := []entities.Order{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			StudentID:           pgtype.Text{String: "student_id_1"},
			StudentFullName:     pgtype.Text{String: "manabie"},
			LocationID:          pgtype.Text{String: "location_id_1"},
			OrderSequenceNumber: pgtype.Int4{Int: 1},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			StudentID:           pgtype.Text{String: "student_id_2"},
			StudentFullName:     pgtype.Text{String: "manabian"},
			LocationID:          pgtype.Text{String: "location_id_2"},
			OrderSequenceNumber: pgtype.Int4{Int: 2},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			StudentID:           pgtype.Text{String: "student_id_3"},
			StudentFullName:     pgtype.Text{String: "new manabian"},
			LocationID:          pgtype.Text{String: "location_id_3"},
			OrderSequenceNumber: pgtype.Int4{Int: 3},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			StudentID:           pgtype.Text{String: "student_id_4"},
			StudentFullName:     pgtype.Text{String: "old manabian"},
			LocationID:          pgtype.Text{String: "location_id_4"},
			OrderSequenceNumber: pgtype.Int4{Int: 4},
			OrderStatus:         pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
			OrderType:           pgtype.Text{String: pb.OrderType_ORDER_TYPE_LOA.String()},
			UpdatedAt:           pgtype.Timestamptz{Time: now},
			CreatedAt:           pgtype.Timestamptz{Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time},
			IsReviewed:          pgtype.Bool{Bool: true},
		},
	}
	expectedOrderProducts := []entities.OrderItem{
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRA",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRB",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRC",
			},
			ProductID: pgtype.Text{
				String: "1",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
		{
			OrderID: pgtype.Text{
				String: "01FWZC3BGP8J7D3Z4XV1B5CRRD",
			},
			ProductID: pgtype.Text{
				String: "3",
			},
			DiscountID: pgtype.Text{
				String: "0",
			},
			StartDate: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
		},
	}
	expectedProducts := []entities.Product{
		{
			ProductID: pgtype.Text{
				String: "1",
			},
			Name: pgtype.Text{
				String: "product_1",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
		{
			ProductID: pgtype.Text{
				String: "3",
			},
			Name: pgtype.Text{
				String: "product_3",
			},
			ProductType: pgtype.Text{
				String: "",
			},
			TaxID: pgtype.Text{
				String: "0",
			},
			AvailableFrom:       pgtype.Timestamptz{Time: now},
			AvailableUntil:      pgtype.Timestamptz{Time: now},
			CustomBillingPeriod: pgtype.Timestamptz{Time: now},
			BillingScheduleID: pgtype.Text{
				String: "0",
			},
			DisableProRatingFlag: pgtype.Bool{Bool: false},
			Remarks: pgtype.Text{
				String: "",
			},
			IsArchived: pgtype.Bool{Bool: true},
			UpdatedAt:  pgtype.Timestamptz{Time: now},
			CreatedAt:  pgtype.Timestamptz{Time: now},
		},
	}
	expectedLocations := []entities.Location{
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_2",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_3",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_4",
			},
			Name: pgtype.Text{
				String: constant.LocationName,
			},
		},
	}
	testCases := []utils.TestCase{
		{
			Name: "happy case (with filter student name)",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "mana",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter: &pb.RetrieveListOfOrdersFilter{
					CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
					CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
					OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
					ProductIds:  []string{"1", "2", "3", "4"},
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			ExpectedResp: &pb.RetrieveListOfOrdersResponse{
				Items: []*pb.RetrieveListOfOrdersResponse_Order{
					{
						OrderSequenceNumber: 1,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
						StudentId:           "student_id_1",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_NEW,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						StudentName:         "manabie",
					},
					{
						OrderSequenceNumber: 2,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
						StudentId:           "student_id_2",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						StudentName:         "manabian",
					},
					{
						OrderSequenceNumber: 3,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
						StudentId:           "student_id_3",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
						ProductDetails:      "product_1",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						StudentName:         "new manabian",
					},
					{
						OrderSequenceNumber: 4,
						OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
						StudentId:           "student_id_4",
						OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
						OrderType:           pb.OrderType_ORDER_TYPE_LOA,
						ProductDetails:      "product_3",
						CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
						StudentName:         "old manabian",
					},
				},
				TotalItems:               4,
				TotalOfSubmitted:         0,
				TotalOfPending:           0,
				TotalOfRejected:          0,
				TotalOfVoided:            0,
				TotalOfInvoiced:          4,
				TotalOfOrderNeedToReview: 4,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderStatByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderStats{
					TotalItems: pgtype.Int8{
						Int: 4,
					},
					TotalOfSubmitted: pgtype.Int8{
						Int: 0,
					},
					TotalOfPending: pgtype.Int8{
						Int: 0,
					},
					TotalOfRejected: pgtype.Int8{
						Int: 0,
					},
					TotalOfVoided: pgtype.Int8{
						Int: 0,
					},
					TotalOfInvoiced: pgtype.Int8{
						Int: 4,
					},
					TotalOfNeedToReview: pgtype.Int8{
						Int: 4,
					},
				}, nil)
				orderService.On("GetListOfOrdersByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedOrders, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedLocations, nil)
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedOrderProducts, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(expectedProducts, nil)
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						UserID: pgtype.Text{
							String: constant.UserID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.OrderId, item.OrderId)
					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
					assert.Equal(t, expectedItem.StudentId, item.StudentId)
					assert.Equal(t, expectedItem.StudentName, item.StudentName)
					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
					assert.Equal(t, expectedItem.OrderType, item.OrderType)
					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
			}
		})
	}
}

func TestRetrieveOrdersListService_buildMapProductsWithOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
	)

	testCases := []utils.TestCase{
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
				},
				{
					OrderID: pgtype.Text{
						String: "Order-124",
					},
				},
			},
			ExpectedResp: map[string][]entities.Product{
				"Order-123": {
					{
						ProductID: pgtype.Text{
							String: "Product-1",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-2",
						},
					},
				},
				"Order-124": {
					{
						ProductID: pgtype.Text{
							String: "Product-2",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-3",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-4",
						},
					},
				},
			},
			Setup: func(ctx context.Context) {
				orderItemService.On("GetOrderItemsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: constant.ProductID,
						},
						OrderItemID: pgtype.Text{
							String: "OrderItemID-1",
						},
					},
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						ProductID: pgtype.Text{
							String: "Product-2",
						},
						OrderItemID: pgtype.Text{
							String: "OrderItemID-2",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "Order-124",
						},
						ProductID: pgtype.Text{
							String: "Product-2",
						},
						OrderItemID: pgtype.Text{
							String: "OrderItemID-3",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "Order-124",
						},
						ProductID: pgtype.Text{
							String: "Product-3",
						},
						OrderItemID: pgtype.Text{
							String: "OrderItemID-4",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "Order-124",
						},
						ProductID: pgtype.Text{
							String: "Product-4",
						},
						OrderItemID: pgtype.Text{
							String: "OrderItemID-4",
						},
					},
				}, nil)
				productService.On("GetProductsByIDs", mock.Anything, mock.Anything, mock.Anything).Twice().Return([]entities.Product{
					{
						ProductID: pgtype.Text{
							String: "Product-1",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-2",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-3",
						},
					},
					{
						ProductID: pgtype.Text{
							String: "Product-4",
						},
					},
				}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
			}
			req := testCase.Req.([]entities.Order)
			mapProductsWithOrderID, err := s.buildMapProductsWithOrderID(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedMapProductsWithOrderID := testCase.ExpectedResp.(map[string][]entities.Product)
				for expectedOrderID, expectedProducts := range expectedMapProductsWithOrderID {
					for orderID, products := range mapProductsWithOrderID {
						if expectedOrderID == orderID {
							assert.Equal(t, len(expectedProducts), len(products))
						}
					}
				}
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, mapProductsWithOrderID)
			}
		})
	}
}

func TestRetrieveOrdersListService_getLocationsOfOrdersReturningMapLocationWithOrderID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)

	testCases := []utils.TestCase{
		{
			Name: "Fail case: Error when get locations by ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
				{
					OrderID: pgtype.Text{
						String: "Order-124",
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
			},
			ExpectedErr:  constant.ErrDefault,
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.Location{}, constant.ErrDefault)

			},
		},
		{
			Name: "Happy case (location id of order is invalid)",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
				{
					OrderID: pgtype.Text{
						String: "Order-124",
					},
					LocationID: pgtype.Text{String: "invalid_location_id"},
				},
			},
			ExpectedResp: map[string]entities.Location{
				constant.OrderID: {
					LocationID: pgtype.Text{
						String: constant.LocationID,
					},
					Name: pgtype.Text{String: constant.LocationName},
				},
				"Order-124": {},
			},
			Setup: func(ctx context.Context) {
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.Location{
					{
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						Name: pgtype.Text{String: constant.LocationName},
					},
				}, nil)

			},
		},

		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
				{
					OrderID: pgtype.Text{
						String: "Order-124",
					},
					LocationID: pgtype.Text{String: "constant.LocationID-12"},
				},
			},
			ExpectedResp: map[string]entities.Location{
				constant.OrderID: {
					LocationID: pgtype.Text{
						String: constant.LocationID,
					},
					Name: pgtype.Text{
						String: constant.LocationName,
					},
				},
				"Order-124": {
					LocationID: pgtype.Text{
						String: "constant.LocationID-12",
					},
					Name: pgtype.Text{
						String: constant.LocationName,
					},
				},
			},
			Setup: func(ctx context.Context) {
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.Location{
					{
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
					{
						LocationID: pgtype.Text{
							String: "constant.LocationID-12",
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)

			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.([]entities.Order)
			mapProductsWithOrderID, err := s.getLocationsOfOrdersReturningMapLocationWithOrderID(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedMapLocationWithOrderID := testCase.ExpectedResp.(map[string]entities.Location)
				for expectedOrderID, expectedLocation := range expectedMapLocationWithOrderID {
					location, _ := mapProductsWithOrderID[expectedOrderID]
					assert.Equal(t, expectedLocation.Name.String, location.Name.String)
					assert.Equal(t, expectedLocation.LocationID.String, location.LocationID.String)
				}
				assert.Equal(t, len(expectedMapLocationWithOrderID), len(mapProductsWithOrderID))
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, mapProductsWithOrderID)
			}
		})
	}
}

func TestRetrieveOrdersListService_getMapOrderIDWithCreatorInfo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db               *mockDb.Ext
		orderService     *mockServices.IOrderServiceForOrderList
		orderItemService *mockServices.IOrderItemServiceForOrderList
		productService   *mockServices.IProductServiceForOrderList
		locationService  *mockServices.ILocationServiceForOrderList
	)

	testCases := []utils.TestCase{
		{
			Name: "Fail case: Error when get order creators by order_ids",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
				{
					OrderID: pgtype.Text{
						String: "Order-124",
					},
					LocationID: pgtype.Text{String: constant.LocationID},
				},
			},
			ExpectedErr:  constant.ErrDefault,
			ExpectedResp: nil,
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{}, constant.ErrDefault)
			},
		},

		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: []entities.Order{
				{
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
				},
			},
			ExpectedResp: map[string]entities.OrderCreator{
				constant.OrderID: {
					OrderID: pgtype.Text{
						String: constant.OrderID,
					},
					UserID: pgtype.Text{
						String: constant.UserID,
					},
					Name: pgtype.Text{
						String: constant.LocationName,
					},
				},
			},
			Setup: func(ctx context.Context) {
				orderService.On("GetOrderCreatorsByOrderIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderCreator{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						UserID: pgtype.Text{
							String: constant.UserID,
						},
						Name: pgtype.Text{
							String: constant.LocationName,
						},
					},
				}, nil)

			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderService = new(mockServices.IOrderServiceForOrderList)
			orderItemService = new(mockServices.IOrderItemServiceForOrderList)
			productService = new(mockServices.IProductServiceForOrderList)
			locationService = new(mockServices.ILocationServiceForOrderList)

			testCase.Setup(testCase.Ctx)
			s := &OrderList{
				DB:               db,
				OrderService:     orderService,
				OrderItemService: orderItemService,
				ProductService:   productService,
				LocationService:  locationService,
			}
			req := testCase.Req.([]entities.Order)
			mapOrderIDWithOrderCreator, err := s.getMapOrderIDWithCreatorInfo(testCase.Ctx, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedMapOrderIDWithOrderCreator := testCase.ExpectedResp.(map[string]entities.OrderCreator)
				for expectedOrderID, expectedLocation := range expectedMapOrderIDWithOrderCreator {
					orderCreator, _ := mapOrderIDWithOrderCreator[expectedOrderID]
					assert.Equal(t, expectedLocation.Name.String, orderCreator.Name.String)
					assert.Equal(t, expectedLocation.UserID.String, orderCreator.UserID.String)
				}
				assert.Equal(t, len(expectedMapOrderIDWithOrderCreator), len(mapOrderIDWithOrderCreator))
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, mapOrderIDWithOrderCreator)
			}
		})
	}
}

//func TestRetrieveOrdersListService_RetrieveOrdersUsingFilter(t *testing.T) {
//	t.Parallel()
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//
//	mockSearchEngine := &mockServices.Engine{}
//	s := &OrderList{
//		SearchEngine: mockSearchEngine,
//	}
//
//	testCases := []utils.TestCase{
//		caseRetrieveOrdersListWithFilter(ctx, mockSearchEngine),
//		caseRetrieveOrdersListWithoutFilter(ctx, mockSearchEngine),
//		caseRetrieveOrdersListWithEmptyProducts(ctx, mockSearchEngine),
//		caseRetrieveEmptyOrdersListCauseFilterProduct(ctx, mockSearchEngine),
//		caseRetrieveOrdersListWithFilterStudentName(ctx, mockSearchEngine),
//	}
//	for _, testCase := range testCases {
//		testCase := testCase
//		t.Run(testCase.Name, func(t *testing.T) {
//			testCase.Setup(testCase.Ctx)
//			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)
//			resp, err := s.RetrieveListOfOrders(testCase.Ctx, req)
//			if testCase.ExpectedErr != nil {
//				assert.Error(t, err)
//				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
//			} else {
//				assert.NoError(t, err)
//				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfOrdersResponse)
//				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
//				for idx, expectedItem := range expectedResp.Items {
//					item := resp.Items[idx]
//					assert.Equal(t, expectedItem.OrderId, item.OrderId)
//					assert.Equal(t, expectedItem.OrderSequenceNumber, item.OrderSequenceNumber)
//					assert.Equal(t, expectedItem.StudentId, item.StudentId)
//					assert.Equal(t, expectedItem.StudentName, item.StudentName)
//					assert.Equal(t, expectedItem.OrderStatus, item.OrderStatus)
//					assert.Equal(t, expectedItem.OrderType, item.OrderType)
//					assert.Equal(t, expectedItem.ProductDetails, item.ProductDetails)
//					assert.Equal(t, expectedItem.CreateDate, item.CreateDate)
//				}
//
//				if expectedResp.PreviousPage == nil {
//					assert.Nil(t, resp.PreviousPage)
//				} else {
//					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
//					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
//				}
//
//				if expectedResp.NextPage == nil {
//					assert.Nil(t, resp.NextPage)
//				} else {
//					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
//					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
//				}
//				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
//				assert.Equal(t, expectedResp.TotalOfSubmitted, resp.TotalOfSubmitted)
//				assert.Equal(t, expectedResp.TotalOfPending, resp.TotalOfPending)
//				assert.Equal(t, expectedResp.TotalOfRejected, resp.TotalOfRejected)
//				assert.Equal(t, expectedResp.TotalOfVoided, resp.TotalOfVoided)
//				assert.Equal(t, expectedResp.TotalOfInvoiced, resp.TotalOfInvoiced)
//				assert.Equal(t, expectedResp.TotalOfOrderNeedToReview, resp.TotalOfOrderNeedToReview)
//			}
//		})
//	}
//}
//
//func caseRetrieveOrdersListWithFilter(ctx context.Context, mockSearchEngine *mockServices.Engine) utils.TestCase {
//	now := time.Now().UTC()
//
//	expectedOrders := make([]interface{}, 0)
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		StudentID:           "student_id_1",
//		LocationID:          "location_id_1",
//		OrderSequenceNumber: 1,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_NEW.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		StudentID:           "student_id_2",
//		LocationID:          "location_id_2",
//		OrderSequenceNumber: 2,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		StudentID:           "student_id_3",
//		LocationID:          "location_id_3",
//		OrderSequenceNumber: 3,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		StudentID:           "student_id_4",
//		LocationID:          "location_id_4",
//		OrderSequenceNumber: 4,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_LOA.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//
//	expectedOrderProducts := make([]interface{}, 0)
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//
//	expectedProducts := make([]interface{}, 0)
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "1",
//		Name:                 "product_1",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "3",
//		Name:                 "product_3",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//
//	return utils.TestCase{
//		Name: "happy case",
//		Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
//		Req: &pb.RetrieveListOfOrdersRequest{
//			CurrentTime: timestamppb.New(now),
//			Keyword:     "",
//			OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
//			Filter: &pb.RetrieveListOfOrdersFilter{
//				CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
//				CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
//				OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
//				ProductIds:  []string{"1", "2", "3", "4"},
//			},
//			Paging: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 10,
//				},
//			},
//		},
//		ExpectedResp: &pb.RetrieveListOfOrdersResponse{
//			Items: []*pb.RetrieveListOfOrdersResponse_Order{
//				{
//					OrderSequenceNumber: 1,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//					StudentId:           "student_id_1",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_NEW,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 2,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//					StudentId:           "student_id_2",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 3,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//					StudentId:           "student_id_3",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 4,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//					StudentId:           "student_id_4",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_LOA,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//			},
//			PreviousPage: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 0,
//				},
//			},
//			NextPage:                 nil,
//			TotalItems:               58,
//			TotalOfSubmitted:         20,
//			TotalOfPending:           15,
//			TotalOfRejected:          10,
//			TotalOfVoided:            8,
//			TotalOfInvoiced:          5,
//			TotalOfOrderNeedToReview: 5,
//		},
//		ExpectedErr: nil,
//		Setup: func(ctx context.Context) {
//			mockSearchEngine.On(
//				"Search",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Once().Return(expectedOrders, nil)
//
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticOrderItemTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Twice().Return(expectedOrderProducts, nil)
//
//			mapProductIDs := make(map[string]bool)
//			for _, item := range expectedOrderProducts {
//				orderProduct := item.(*entities.ElasticOrderItem)
//				mapProductIDs[orderProduct.ProductID] = true
//			}
//			products := make([]interface{}, 0, len(mapProductIDs))
//			for _, item := range expectedProducts {
//				product := item.(*entities.ElasticProduct)
//				if mapProductIDs[product.ProductID] {
//					products = append(products, product)
//				}
//			}
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticProductTableName,
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					productIDConditions := make([]op.Condition, 0, len(products))
//					for _, item := range products {
//						product := item.(*entities.ElasticProduct)
//						productIDConditions = append(productIDConditions, op.Equal("product_id", product.ProductID))
//					}
//					expectedCondition := op.Or(productIDConditions...)
//					return condition.String() == expectedCondition.String()
//				}),
//				mock.Anything,
//				mock.Anything,
//			).Times(len(expectedOrders)*2).Return(products, nil)
//
//			// Mock stats
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					return condition != nil
//				}),
//			).Once().Return(uint32(58), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(20), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_PENDING.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(15), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_REJECTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(10), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_VOIDED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(8), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_INVOICED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.And(
//						op.Equal("is_reviewed", false),
//						op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//					)
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//
//		},
//	}
//}
//
//func caseRetrieveOrdersListWithoutFilter(ctx context.Context, mockSearchEngine *mockServices.Engine) utils.TestCase {
//	now := time.Now().UTC()
//
//	expectedOrders := make([]interface{}, 0)
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		StudentID:           "student_id_1",
//		LocationID:          "location_id_1",
//		OrderSequenceNumber: 1,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_NEW.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		StudentID:           "student_id_2",
//		LocationID:          "location_id_2",
//		OrderSequenceNumber: 2,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		StudentID:           "student_id_3",
//		LocationID:          "location_id_3",
//		OrderSequenceNumber: 3,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		StudentID:           "student_id_4",
//		LocationID:          "location_id_4",
//		OrderSequenceNumber: 4,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_LOA.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//
//	expectedOrderProducts := make([]interface{}, 0)
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//
//	expectedProducts := make([]interface{}, 0)
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "1",
//		Name:                 "product_1",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "3",
//		Name:                 "product_3",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//
//	return utils.TestCase{
//		Name: "case without filter",
//		Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
//		Req: &pb.RetrieveListOfOrdersRequest{
//			CurrentTime: timestamppb.New(now),
//			Keyword:     "",
//			OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
//			Filter:      nil,
//			Paging: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 10,
//				},
//			},
//		},
//		ExpectedResp: &pb.RetrieveListOfOrdersResponse{
//			Items: []*pb.RetrieveListOfOrdersResponse_Order{
//				{
//					OrderSequenceNumber: 1,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//					StudentId:           "student_id_1",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_NEW,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 2,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//					StudentId:           "student_id_2",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 3,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//					StudentId:           "student_id_3",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 4,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//					StudentId:           "student_id_4",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_LOA,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//			},
//			PreviousPage: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 0,
//				},
//			},
//			NextPage:                 nil,
//			TotalItems:               58,
//			TotalOfSubmitted:         20,
//			TotalOfPending:           15,
//			TotalOfRejected:          10,
//			TotalOfVoided:            8,
//			TotalOfInvoiced:          5,
//			TotalOfOrderNeedToReview: 5,
//		},
//		ExpectedErr: nil,
//		Setup: func(ctx context.Context) {
//			mockSearchEngine.On(
//				"Search",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Once().Return(expectedOrders, nil)
//
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticOrderItemTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Twice().Return(expectedOrderProducts, nil)
//
//			mapProductIDs := make(map[string]bool)
//			for _, item := range expectedOrderProducts {
//				orderProduct := item.(*entities.ElasticOrderItem)
//				mapProductIDs[orderProduct.ProductID] = true
//			}
//			products := make([]interface{}, 0, len(mapProductIDs))
//			for _, item := range expectedProducts {
//				product := item.(*entities.ElasticProduct)
//				if mapProductIDs[product.ProductID] {
//					products = append(products, product)
//				}
//			}
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticProductTableName,
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					productIDConditions := make([]op.Condition, 0, len(products))
//					for _, item := range products {
//						product := item.(*entities.ElasticProduct)
//						productIDConditions = append(productIDConditions, op.Equal("product_id", product.ProductID))
//					}
//					expectedCondition := op.Or(productIDConditions...)
//					return condition.String() == expectedCondition.String()
//				}),
//				mock.Anything,
//				mock.Anything,
//			).Times(len(expectedOrders)*2).Return(products, nil)
//
//			// Mock stats
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					return condition != nil
//				}),
//			).Once().Return(uint32(58), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(20), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_PENDING.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(15), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_REJECTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(10), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_VOIDED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(8), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_INVOICED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.And(
//						op.Equal("is_reviewed", false),
//						op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//					)
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//		},
//	}
//}
//
//func caseRetrieveOrdersListWithEmptyProducts(ctx context.Context, mockSearchEngine *mockServices.Engine) utils.TestCase {
//	now := time.Now().UTC()
//
//	expectedOrders := make([]interface{}, 0)
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		StudentID:           "student_id_1",
//		LocationID:          "location_id_1",
//		OrderSequenceNumber: 1,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_NEW.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		StudentID:           "student_id_2",
//		LocationID:          "location_id_2",
//		OrderSequenceNumber: 2,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		StudentID:           "student_id_3",
//		LocationID:          "location_id_3",
//		OrderSequenceNumber: 3,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		StudentID:           "student_id_4",
//		LocationID:          "location_id_4",
//		OrderSequenceNumber: 4,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_LOA.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//
//	expectedOrderProducts := make([]interface{}, 0)
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//
//	expectedProducts := make([]interface{}, 0)
//
//	return utils.TestCase{
//		Name: "case empty products",
//		Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
//		Req: &pb.RetrieveListOfOrdersRequest{
//			CurrentTime: timestamppb.New(now),
//			Keyword:     "",
//			OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
//			Filter: &pb.RetrieveListOfOrdersFilter{
//				CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
//				CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
//				OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
//				ProductIds:  []string{"1", "2", "3", "4"},
//			},
//			Paging: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 10,
//				},
//			},
//		},
//		ExpectedResp: &pb.RetrieveListOfOrdersResponse{
//			Items: []*pb.RetrieveListOfOrdersResponse_Order{
//				{
//					OrderSequenceNumber: 1,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//					StudentId:           "student_id_1",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_NEW,
//					ProductDetails:      "",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 2,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//					StudentId:           "student_id_2",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 3,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//					StudentId:           "student_id_3",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//				{
//					OrderSequenceNumber: 4,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//					StudentId:           "student_id_4",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_LOA,
//					ProductDetails:      "",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					IsReviewed:          true,
//				},
//			},
//			PreviousPage: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 0,
//				},
//			},
//			NextPage:                 nil,
//			TotalItems:               58,
//			TotalOfSubmitted:         20,
//			TotalOfPending:           15,
//			TotalOfRejected:          10,
//			TotalOfVoided:            8,
//			TotalOfInvoiced:          5,
//			TotalOfOrderNeedToReview: 5,
//		},
//		ExpectedErr: nil,
//		Setup: func(ctx context.Context) {
//			mockSearchEngine.On(
//				"Search",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Once().Return(expectedOrders, nil)
//
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticOrderItemTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Twice().Return(expectedOrderProducts, nil)
//
//			mapProductIDs := make(map[string]bool)
//			for _, item := range expectedOrderProducts {
//				orderProduct := item.(*entities.ElasticOrderItem)
//				mapProductIDs[orderProduct.ProductID] = true
//			}
//			products := make([]interface{}, 0, len(mapProductIDs))
//			for _, item := range expectedProducts {
//				product := item.(*entities.ElasticProduct)
//				if mapProductIDs[product.ProductID] {
//					products = append(products, product)
//				}
//			}
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticProductTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Times(len(expectedOrders)*2).Return(products, nil)
//
//			// Mock stats
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					return condition != nil
//				}),
//			).Once().Return(uint32(58), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(20), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_PENDING.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(15), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_REJECTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(10), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_VOIDED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(8), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_INVOICED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.And(
//						op.Equal("is_reviewed", false),
//						op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//					)
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//		},
//	}
//}
//
//func caseRetrieveEmptyOrdersListCauseFilterProduct(ctx context.Context, mockSearchEngine *mockServices.Engine) utils.TestCase {
//	now := time.Now().UTC()
//
//	expectedOrders := make([]interface{}, 0)
//
//	expectedOrderProducts := make([]interface{}, 0)
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//
//	expectedProducts := make([]interface{}, 0)
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "1",
//		Name:                 "product_1",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "3",
//		Name:                 "product_3",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//
//	return utils.TestCase{
//		Name: "case empty orders list cause filter products",
//		Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
//		Req: &pb.RetrieveListOfOrdersRequest{
//			CurrentTime: timestamppb.New(now),
//			Keyword:     "",
//			OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
//			Filter: &pb.RetrieveListOfOrdersFilter{
//				CreatedFrom: timestamppb.New(now.Add(-30 * time.Hour)),
//				CreatedTo:   timestamppb.New(now.Add(30 * time.Hour)),
//				OrderTypes:  []pb.OrderType{pb.OrderType_ORDER_TYPE_NEW, pb.OrderType_ORDER_TYPE_CUSTOM_BILLING, pb.OrderType_ORDER_TYPE_LOA},
//				ProductIds:  []string{"2", "3", "4"},
//			},
//			Paging: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 10,
//				},
//			},
//		},
//		ExpectedResp: &pb.RetrieveListOfOrdersResponse{
//			Items:                    []*pb.RetrieveListOfOrdersResponse_Order{},
//			PreviousPage:             nil,
//			NextPage:                 nil,
//			TotalItems:               53,
//			TotalOfSubmitted:         20,
//			TotalOfPending:           15,
//			TotalOfRejected:          10,
//			TotalOfVoided:            8,
//			TotalOfInvoiced:          0,
//			TotalOfOrderNeedToReview: 5,
//		},
//		ExpectedErr: nil,
//		Setup: func(ctx context.Context) {
//			mockSearchEngine.On(
//				"Search",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Once().Return(expectedOrders, nil)
//
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticOrderItemTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Twice().Return(expectedOrderProducts, nil)
//
//			mapProductIDs := make(map[string]bool)
//			for _, item := range expectedOrderProducts {
//				orderProduct := item.(*entities.ElasticOrderItem)
//				mapProductIDs[orderProduct.ProductID] = true
//			}
//			products := make([]interface{}, 0, len(mapProductIDs))
//			for _, item := range expectedProducts {
//				product := item.(*entities.ElasticProduct)
//				if mapProductIDs[product.ProductID] {
//					products = append(products, product)
//				}
//			}
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticProductTableName,
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					productIDConditions := make([]op.Condition, 0, len(products))
//					for _, item := range products {
//						product := item.(*entities.ElasticProduct)
//						productIDConditions = append(productIDConditions, op.Equal("product_id", product.ProductID))
//					}
//					expectedCondition := op.Or(productIDConditions...)
//					return condition.String() == expectedCondition.String()
//				}),
//				mock.Anything,
//				mock.Anything,
//			).Times(len(expectedOrders)*2).Return(products, nil)
//
//			// Mock stats
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					return condition != nil
//				}),
//			).Once().Return(uint32(53), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(20), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_PENDING.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(15), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_REJECTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(10), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_VOIDED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(8), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_INVOICED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(0), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.And(
//						op.Equal("is_reviewed", false),
//						op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//					)
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(5), nil)
//		},
//	}
//}
//
//func caseRetrieveOrdersListWithFilterStudentName(ctx context.Context, mockSearchEngine *mockServices.Engine) utils.TestCase {
//	now := time.Now().UTC()
//
//	expectedOrders := make([]interface{}, 0)
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		StudentID:           "student_id_1",
//		StudentName:         "manabie",
//		LocationID:          "location_id_1",
//		OrderSequenceNumber: 1,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_NEW.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		StudentID:           "student_id_2",
//		StudentName:         "manabian",
//		LocationID:          "location_id_2",
//		OrderSequenceNumber: 2,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		StudentID:           "student_id_3",
//		StudentName:         "new manabian",
//		LocationID:          "location_id_3",
//		OrderSequenceNumber: 3,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//	expectedOrders = append(expectedOrders, &entities.ElasticOrder{
//		OrderID:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		StudentID:           "student_id_4",
//		StudentName:         "old manabian",
//		LocationID:          "location_id_4",
//		OrderSequenceNumber: 4,
//		OrderComment:        "",
//		OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
//		OrderType:           pb.OrderType_ORDER_TYPE_LOA.String(),
//		UpdatedAt:           now,
//		CreatedAt:           database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
//		IsReviewed:          true,
//	})
//
//	expectedOrderProducts := make([]interface{}, 0)
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//	expectedOrderProducts = append(expectedOrderProducts, &entities.ElasticOrderItem{
//		OrderID:    "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//		ProductID:  "1",
//		DiscountID: "0",
//		StartDate:  now,
//		CreatedAt:  now,
//	})
//
//	expectedProducts := make([]interface{}, 0)
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "1",
//		Name:                 "product_1",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//	expectedProducts = append(expectedProducts, &entities.ElasticProduct{
//		ProductID:            "3",
//		Name:                 "product_3",
//		ProductType:          "",
//		TaxID:                "0",
//		AvailableFrom:        now,
//		AvailableUntil:       now,
//		CustomBillingPeriod:  now,
//		BillingScheduleID:    "0",
//		DisableProRatingFlag: false,
//		Remarks:              "",
//		IsArchived:           true,
//		UpdatedAt:            now,
//		CreatedAt:            now,
//	})
//
//	expectedStudents := make([]entities.User, 0)
//	expectedStudents = append(expectedStudents, entities.User{
//		UserID: pgtype.Text{
//			String: "student_id_1",
//		},
//		Name: pgtype.Text{
//			String: "manabie",
//		},
//		Group: pgtype.Text{
//			String: "UserGroup_USER_GROUP_STUDENT",
//		},
//	})
//	expectedStudents = append(expectedStudents, entities.User{
//		UserID: pgtype.Text{
//			String: "student_id_2",
//		},
//		Name: pgtype.Text{
//			String: "manabian",
//		},
//		Group: pgtype.Text{
//			String: "UserGroup_USER_GROUP_STUDENT",
//		},
//	})
//	expectedStudents = append(expectedStudents, entities.User{
//		UserID: pgtype.Text{
//			String: "student_id_3",
//		},
//		Name: pgtype.Text{
//			String: "new manabian",
//		},
//		Group: pgtype.Text{
//			String: "UserGroup_USER_GROUP_STUDENT",
//		},
//	})
//	expectedStudents = append(expectedStudents, entities.User{
//		UserID: pgtype.Text{
//			String: "student_id_4",
//		},
//		Name: pgtype.Text{
//			String: "old manabian",
//		},
//		Group: pgtype.Text{
//			String: "UserGroup_USER_GROUP_STUDENT",
//		},
//	})
//
//	return utils.TestCase{
//		Name: "case with filter student-name",
//		Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
//		Req: &pb.RetrieveListOfOrdersRequest{
//			CurrentTime: timestamppb.New(now),
//			Keyword:     "mana",
//			OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
//			Filter:      nil,
//			Paging: &cpb.Paging{
//				Limit: 10,
//				Offset: &cpb.Paging_OffsetInteger{
//					OffsetInteger: 0,
//				},
//			},
//		},
//		ExpectedResp: &pb.RetrieveListOfOrdersResponse{
//			Items: []*pb.RetrieveListOfOrdersResponse_Order{
//				{
//					OrderSequenceNumber: 1,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRA",
//					StudentId:           "student_id_1",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_NEW,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					StudentName:         "manabie",
//				},
//				{
//					OrderSequenceNumber: 2,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRB",
//					StudentId:           "student_id_2",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					StudentName:         "manabian",
//				},
//				{
//					OrderSequenceNumber: 3,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRC",
//					StudentId:           "student_id_3",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					StudentName:         "new manabian",
//				},
//				{
//					OrderSequenceNumber: 4,
//					OrderId:             "01FWZC3BGP8J7D3Z4XV1B5CRRD",
//					StudentId:           "student_id_4",
//					OrderStatus:         pb.OrderStatus_ORDER_STATUS_INVOICED,
//					OrderType:           pb.OrderType_ORDER_TYPE_LOA,
//					ProductDetails:      "product_1",
//					CreateDate:          timestamppb.New(database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time),
//					StudentName:         "old manabian",
//				},
//			},
//			TotalItems:               4,
//			TotalOfSubmitted:         0,
//			TotalOfPending:           0,
//			TotalOfRejected:          0,
//			TotalOfVoided:            0,
//			TotalOfInvoiced:          4,
//			TotalOfOrderNeedToReview: 4,
//		},
//		ExpectedErr: nil,
//		Setup: func(ctx context.Context) {
//			mockSearchEngine.On(
//				"Search",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Once().Return(expectedOrders, nil)
//
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticOrderItemTableName,
//				mock.Anything,
//				mock.Anything,
//				mock.Anything,
//			).Twice().Return(expectedOrderProducts, nil)
//
//			mapProductIDs := make(map[string]bool)
//			for _, item := range expectedOrderProducts {
//				orderProduct := item.(*entities.ElasticOrderItem)
//				mapProductIDs[orderProduct.ProductID] = true
//			}
//			products := make([]interface{}, 0, len(mapProductIDs))
//			for _, item := range expectedProducts {
//				product := item.(*entities.ElasticProduct)
//				if mapProductIDs[product.ProductID] {
//					products = append(products, product)
//				}
//			}
//			mockSearchEngine.On(
//				"SearchWithoutPaging",
//				mock.Anything,
//				constant.ElasticProductTableName,
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					productIDConditions := make([]op.Condition, 0, len(products))
//					for _, item := range products {
//						product := item.(*entities.ElasticProduct)
//						productIDConditions = append(productIDConditions, op.Equal("product_id", product.ProductID))
//					}
//					expectedCondition := op.Or(productIDConditions...)
//					return condition.String() == expectedCondition.String()
//				}),
//				mock.Anything,
//				mock.Anything,
//			).Times(len(expectedOrders)*2).Return(products, nil)
//
//			// Mock stats
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					return condition != nil
//				}),
//			).Once().Return(uint32(4), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(0), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_PENDING.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(0), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_REJECTED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(0), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_VOIDED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(0), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_INVOICED.String())
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(4), nil)
//
//			mockSearchEngine.On(
//				"CountValue",
//				mock.Anything,
//				constant.ElasticOrderTableName,
//				"order_sequence_number",
//				mock.MatchedBy(func(condition op.Condition) bool {
//					if condition == nil {
//						return false
//					}
//					expectedCondition := op.And(
//						op.Equal("is_reviewed", false),
//						op.Equal("order_status", pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()),
//					)
//					return strings.Contains(condition.String(), expectedCondition.String())
//				}),
//			).Once().Return(uint32(4), nil)
//		},
//	}
//}
