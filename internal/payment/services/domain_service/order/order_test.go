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
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestOrderService_CreateOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                     *mockDb.Ext
		orderRepo              *mockRepositories.MockOrderRepo
		orderActionLogRepo     *mockRepositories.MockOrderActionLogRepo
		orderLeavingReasonRepo *mockRepositories.MockOrderLeavingReasonRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when order type is withdrawal/graduate without effective date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing effective date for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_GRADUATE,
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when order type is LOA without start date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing start date or end date for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					EndDate:              timestamppb.Now(),
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when order type is LOA without end date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing start date or end date for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					StartDate:            timestamppb.Now(),
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "Fail case: Error when create order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order have error %v", constant.ErrDefault),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_GRADUATE,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", ctx, db, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when missing start_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing start date or end date for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Happy case: Error when start_date before now",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "start_date must not be before current time for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, -1).Unix()},
					EndDate:              timestamppb.Now(),
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Happy case: Error when end_date before start_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "start_date must not be before current time for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, -1).Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, -10).Unix()},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Happy case: Error when mising leaving reasons",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing leaving reasons for update student status"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Happy case: Error when create order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order have error"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
					LeavingReasonIds:     []string{constant.LeavingReasonID},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when create leaving reason",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "error when create order leaving reason"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
					LeavingReasonIds:     []string{constant.LeavingReasonID},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderLeavingReasonRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: Error when create order action log",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order action log have error"),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
					LeavingReasonIds:     []string{constant.LeavingReasonID},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderLeavingReasonRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", ctx, db, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				&pb.CreateOrderRequest{
					StudentId:            constant.StudentID,
					LocationId:           constant.LocationID,
					OrderComment:         constant.OrderComment,
					OrderType:            pb.OrderType_ORDER_TYPE_LOA,
					EffectiveDate:        timestamppb.Now(),
					BillingItems:         []*pb.BillingItem{},
					UpcomingBillingItems: nil,
					Timezone:             7,
					StartDate:            &timestamppb.Timestamp{Seconds: time.Now().Unix()},
					EndDate:              &timestamppb.Timestamp{Seconds: time.Now().AddDate(0, 0, 10).Unix()},
					LeavingReasonIds:     []string{constant.LeavingReasonID},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderLeavingReasonRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", ctx, db, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			orderLeavingReasonRepo = new(mockRepositories.MockOrderLeavingReasonRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:              orderRepo,
				orderActionLogRepo:     orderActionLogRepo,
				orderLeavingReasonRepo: orderLeavingReasonRepo,
			}
			createOrderReq := testCase.Req.([]interface{})[0].(*pb.CreateOrderRequest)
			studentName := testCase.Req.([]interface{})[1].(string)
			orderStatus := testCase.Req.([]interface{})[2].(pb.OrderStatus)
			_, err := s.CreateOrder(testCase.Ctx, db, createOrderReq, studentName, orderStatus)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db,
				orderRepo,
				orderActionLogRepo,
				orderLeavingReasonRepo,
			)
		})
	}
}

func TestOrderService_CreateCustomOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when create order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order have error %v", constant.ErrDefault),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:          constant.StudentID,
					LocationId:         constant.LocationID,
					OrderComment:       constant.OrderComment,
					OrderType:          pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
					CustomBillingItems: []*pb.CustomBillingItem{},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", ctx, db, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.FailCaseErrorCreateActionLog,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order action log have error %v", constant.ErrDefault),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:          constant.StudentID,
					LocationId:         constant.LocationID,
					OrderComment:       constant.OrderComment,
					OrderType:          pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
					CustomBillingItems: []*pb.CustomBillingItem{},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", ctx, db, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", ctx, db, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.CreateCustomBillingRequest{
					StudentId:          constant.StudentID,
					LocationId:         constant.LocationID,
					OrderComment:       constant.OrderComment,
					OrderType:          pb.OrderType_ORDER_TYPE_CUSTOM_BILLING,
					CustomBillingItems: []*pb.CustomBillingItem{},
				},
				constant.StudentName,
				pb.OrderStatus_ORDER_STATUS_SUBMITTED,
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("Create", ctx, db, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", ctx, db, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
			}
			createOrderReq := testCase.Req.([]interface{})[0].(*pb.CreateCustomBillingRequest)
			studentName := testCase.Req.([]interface{})[1].(string)
			orderStatus := testCase.Req.([]interface{})[2].(pb.OrderStatus)
			_, err := s.CreateCustomOrder(testCase.Ctx, db, createOrderReq, studentName, orderStatus)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_UpdateOrderReview(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get order by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting order by order id %v have error : %v", constant.OrderID, constant.ErrDefault),
			Req: []interface{}{
				constant.OrderID,
				true,
				int32(0),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", ctx, db, mock.Anything).Return(entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when update is_review flag by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "updating is review flag by order id %v have error: %v", constant.OrderID, constant.ErrDefault),
			Req: []interface{}{
				constant.OrderID,
				true,
				int32(0),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", ctx, db, mock.Anything).Return(entities.Order{}, nil)
				orderRepo.On("UpdateIsReviewFlagByOrderID", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        constant.FailCaseErrorCreateActionLog,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "creating order action log for update review flag have error : %v", constant.ErrDefault),
			Req: []interface{}{
				constant.OrderID,
				true,
				int32(0),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", ctx, db, mock.Anything).Return(entities.Order{}, nil)
				orderRepo.On("UpdateIsReviewFlagByOrderID", ctx, db, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", ctx, db, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
			}
			orderId := testCase.Req.([]interface{})[0].(string)
			isReview := testCase.Req.([]interface{})[1].(bool)
			versionNumber := testCase.Req.([]interface{})[2].(int32)
			err := s.UpdateOrderReview(testCase.Ctx, db, orderId, isReview, versionNumber)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_GetListOfOrdersByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)
	now := time.Now()
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get orders by filter",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when getting orders by filter with error: %v", constant.ErrDefault),
			Req: []interface{}{
				&pb.RetrieveListOfOrdersRequest{
					CurrentTime: timestamppb.New(now),
					Keyword:     "mana",
					OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
					Filter:      nil,
					Paging: &cpb.Paging{
						Limit: 10,
						Offset: &cpb.Paging_OffsetInteger{
							OffsetInteger: 0,
						},
					},
				},
				int64(0), int64(10),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrdersByFilter", ctx, db, mock.Anything).Return([]entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.RetrieveListOfOrdersRequest{
					CurrentTime: timestamppb.New(now),
					Keyword:     "mana",
					OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
					Filter:      nil,
					Paging: &cpb.Paging{
						Limit: 10,
						Offset: &cpb.Paging_OffsetInteger{
							OffsetInteger: 0,
						},
					},
				},
				int64(0),
				int64(10),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrdersByFilter", ctx, db, mock.Anything).Return([]entities.Order{
					{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						OrderStatus: pgtype.Text{
							String: pb.OrderStatus_ORDER_STATUS_INVOICED.String(),
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_NEW.String(),
						},
						IsReviewed: pgtype.Bool{
							Bool: true,
						},
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
			}
			req := testCase.Req.([]interface{})[0].(*pb.RetrieveListOfOrdersRequest)
			from := testCase.Req.([]interface{})[1].(int64)
			limit := testCase.Req.([]interface{})[2].(int64)
			_, err := s.GetListOfOrdersByFilter(testCase.Ctx, db, req, from, limit)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_GetOrderStatByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderItemRepo      *mockRepositories.MockOrderItemRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)
	now := time.Now()
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get order stats by filter",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when getting order stat by filter with error: %v", constant.ErrDefault),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "mana",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_INVOICED,
				Filter:      nil,
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderStatsByFilter", ctx, db, mock.Anything).Return(entities.OrderStats{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrdersRequest{
				CurrentTime: timestamppb.New(now),
				Keyword:     "",
				OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
				Filter: &pb.RetrieveListOfOrdersFilter{
					OrderTypes: []pb.OrderType{
						pb.OrderType_ORDER_TYPE_NEW,
						pb.OrderType_ORDER_TYPE_ENROLLMENT,
					},
					ProductIds: []string{
						"01G9KBZ2NVVF0694TNQ77ZQG88",
					},
					OnlyNotReviewed: true,
				},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemsByProductIDs", mock.Anything, mock.Anything, mock.Anything).Return([]entities.OrderItem{
					{
						OrderID: pgtype.Text{
							String: "01GHZT1DZ8YRYCK5GG14VJFP2Q",
						},
						ProductID: pgtype.Text{
							String: "01G9KBZ2NVVF0694TNQ77ZQG88",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "01GJC45148M9FM3P1B48TRS8RH",
						},
						ProductID: pgtype.Text{
							String: "01G9KBZ2NVVF0694TNQ77ZQG88",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "01GJC46B03W75VEXZYT6XFB9S8",
						},
						ProductID: pgtype.Text{
							String: "01G9KBZ2NVVF0694TNQ77ZQG88",
						},
					},
					{
						OrderID: pgtype.Text{
							String: "01GK3NBQ8WH4CA4ANEE5HZTK6Z",
						},
						ProductID: pgtype.Text{
							String: "01G9KBZ2NVVF0694TNQ77ZQG88",
						},
					},
				}, nil)
				orderRepo.On("GetOrderStatsByFilter", ctx, db, mock.Anything).Return(entities.OrderStats{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
				orderItemRepo:      orderItemRepo,
			}
			req := testCase.Req.(*pb.RetrieveListOfOrdersRequest)

			_, err := s.GetOrderStatByFilter(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_GetOrdersByStudentIDAndLocationIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)
	// now := time.Now()
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: counting order by student id and location ids with error:",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "counting order by student id and location ids with error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderRepo.On("CountOrderByStudentIDAndLocationIDs", ctx, db, mock.Anything, mock.Anything).Return(int(2), constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: getting order by student id and location ids and pagination",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting order by student id and location ids and pagination with error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderRepo.On("CountOrderByStudentIDAndLocationIDs", ctx, db, mock.Anything, mock.Anything).Return(int(2), nil)
				orderRepo.On("GetOrderByStudentIDAndLocationIDsPaging", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				orderRepo.On("CountOrderByStudentIDAndLocationIDs", ctx, db, mock.Anything, mock.Anything).Return(int(2), nil)
				orderRepo.On("GetOrderByStudentIDAndLocationIDsPaging", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Order{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
			}

			_, _, err := s.GetOrdersByStudentIDAndLocationIDs(testCase.Ctx, db, "student_1", []string{"location_1", "location_2"}, int64(2), int64(2))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_setFieldsForStudentStatusUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	testcases := []utils.TestCase{
		// {
		// 	Name:        "Fail case: Error missing reason in student subscription",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing reason in update student subscription"),
		// 	Req: []interface{}{
		// 		&pb.CreateOrderRequest{
		// 			Background:     &wrapperspb.StringValue{Value: "Sample background"},
		// 			FutureMeasures: &wrapperspb.StringValue{Value: "Sample future measures"},
		// 		},
		// 		&entities.Order{},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		// Do nothing
		// 	},
		// },
		// {
		// 	Name:        "Fail case: Error missing background in student subscription",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing background in update student subscription"),
		// 	Req: []interface{}{
		// 		&pb.CreateOrderRequest{
		// 			Reason:         &wrapperspb.StringValue{Value: "Sample reason"},
		// 			FutureMeasures: &wrapperspb.StringValue{Value: "Sample future measures"},
		// 		},
		// 		&entities.Order{},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		// Do nothing
		// 	},
		// },
		// {
		// 	Name:        "Fail case: Error missing future measures in student subscription",
		// 	Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	ExpectedErr: status.Errorf(codes.FailedPrecondition, "missing future measures in update student subscription"),
		// 	Req: []interface{}{
		// 		&pb.CreateOrderRequest{
		// 			Reason:         &wrapperspb.StringValue{Value: "Sample reason"},
		// 			Background:     &wrapperspb.StringValue{Value: "Sample background"},
		// 		},
		// 		&entities.Order{},
		// 	},
		// 	Setup: func(ctx context.Context) {
		// 		// Do nothing
		// 	},
		// },
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				&pb.CreateOrderRequest{
					LeavingReasonIds: []string{"Sample reasons"},
					Background:       &wrapperspb.StringValue{Value: "Sample background"},
					FutureMeasures:   &wrapperspb.StringValue{Value: "Sample future measures"},
				},
				&entities.Order{},
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			createOrderReq := testCase.Req.([]interface{})[0].(*pb.CreateOrderRequest)
			order := testCase.Req.([]interface{})[1].(*entities.Order)
			err := setFieldsForStudentStatusUpdate(createOrderReq, order)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestOrderService_GetOrderByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: missing order ID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Missing order ID when getting order by ID"),
			Req: []interface{}{
				"",
			},
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Fail case: error getting order by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "Error when getting order by ID: %v", constant.ErrDefault),
			Req: []interface{}{
				"orderid",
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", ctx, db, mock.Anything).Return(entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				"orderid",
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", ctx, db, mock.Anything).Return(entities.Order{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
			}

			_, err := s.GetOrderByID(testCase.Ctx, db, testCase.Req.([]interface{})[0].(string))

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_VoidOrderReturnOrderAndStudentProductIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                 *mockDb.Ext
		orderRepo          *mockRepositories.MockOrderRepo
		orderItemRepo      *mockRepositories.MockOrderItemRepo
		orderActionLogRepo *mockRepositories.MockOrderActionLogRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when getting order by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when voiding invoiced/voided order",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "error when void an invoiced/voided order"),
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_INVOICED.String()},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when getting all order items by order id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()},
					OrderType:   pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
					OrderID:     pgtype.Text{String: constant.OrderID},
				}, nil)
				orderItemRepo.On("GetAllByOrderID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when voiding an update order when effective date have passed",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "cannot void an update order when effective_date have passed"),
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()},
					OrderType:   pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
					OrderID:     pgtype.Text{String: constant.OrderID},
				}, nil)
				orderItemRepo.On("GetAllByOrderID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.OrderItem{
					{
						EffectiveDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, -1),
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when updating order status by order id and version",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()},
					OrderType:   pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
					OrderID:     pgtype.Text{String: constant.OrderID},
				}, nil)
				orderItemRepo.On("GetAllByOrderID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.OrderItem{
					{
						EffectiveDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 1),
							Status: pgtype.Present,
						},
					},
				}, nil)
				orderRepo.On("UpdateOrderStatusByOrderIDAndVersion", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when creating order action log",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()},
					OrderType:   pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
					OrderID:     pgtype.Text{String: constant.OrderID},
				}, nil)
				orderItemRepo.On("GetAllByOrderID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.OrderItem{
					{
						EffectiveDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 1),
							Status: pgtype.Present,
						},
					},
				}, nil)
				orderRepo.On("UpdateOrderStatusByOrderIDAndVersion", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: []interface{}{
				constant.OrderID,
				int32(constant.OrderVersionNumber),
			},
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByIDForUpdate", mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
					VersionNumber: pgtype.Int4{
						Int:    constant.OrderVersionNumber,
						Status: pgtype.Present,
					},
					OrderStatus: pgtype.Text{String: pb.OrderStatus_ORDER_STATUS_SUBMITTED.String()},
					OrderType:   pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
					OrderID:     pgtype.Text{String: constant.OrderID},
				}, nil)
				orderItemRepo.On("GetAllByOrderID", mock.Anything, mock.Anything, mock.Anything).Return([]*entities.OrderItem{
					{
						EffectiveDate: pgtype.Timestamptz{
							Time:   time.Now().AddDate(0, 0, 1),
							Status: pgtype.Present,
						},
					},
				}, nil)
				orderRepo.On("UpdateOrderStatusByOrderIDAndVersion", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				orderActionLogRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			orderActionLogRepo = new(mockRepositories.MockOrderActionLogRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:          orderRepo,
				orderActionLogRepo: orderActionLogRepo,
				orderItemRepo:      orderItemRepo,
			}
			orderIDReq := testCase.Req.([]interface{})[0].(string)
			orderVersionNumberReq := testCase.Req.([]interface{})[1].(int32)
			_, _, err := s.VoidOrderReturnOrderAndStudentProductIDs(testCase.Ctx, db, orderIDReq, orderVersionNumberReq)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderItemRepo, orderActionLogRepo)
		})
	}
}

func TestOrderService_GetStudentProductIDsForResume(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		orderRepo     *mockRepositories.MockOrderRepo
		orderItemRepo *mockRepositories.MockOrderItemRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: getting orderID by studentID and locationID have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting orderID by studentID and locationID have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByStudentIDAndLocationIDForResume", ctx, db, mock.Anything, mock.Anything).Return("", constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: getting order item list by order_id have error",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting order item list by order_id have error: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByStudentIDAndLocationIDForResume", ctx, db, mock.Anything, mock.Anything).Return("order_id", nil)
				orderItemRepo.On("GetAllByOrderID", ctx, db, mock.Anything).Return([]*entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Setup: func(ctx context.Context) {
				orderRepo.On("GetOrderByStudentIDAndLocationIDForResume", ctx, db, mock.Anything, mock.Anything).Return("order_id", nil)
				orderItemRepo.On("GetAllByOrderID", ctx, db, mock.Anything).Return([]*entities.OrderItem{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			orderRepo = new(mockRepositories.MockOrderRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			testCase.Setup(testCase.Ctx)
			s := &OrderService{
				orderRepo:     orderRepo,
				orderItemRepo: orderItemRepo,
			}
			_, err := s.GetStudentProductIDsForResume(testCase.Ctx, db, "1", "2")

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, orderRepo, orderItemRepo)
		})
	}
}
