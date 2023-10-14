package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	mockRepositories "github.com/manabie-com/backend/mock/discount/repositories"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockKafka "github.com/manabie-com/backend/mock/golibs/kafka"
	mockNats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDiscountEventService_PublishNotificationForStudentProductWithScheduleTag(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		userRepo      *mockRepositories.MockUserRepo
		productRepo   *mockRepositories.MockProductRepo
		orderItemRepo *mockRepositories.MockOrderItemRepo
		kafka         *mockKafka.KafkaManagement
		jsm           *mockNats.JetStreamManagement
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get product by id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get student user by id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get latest order item by student product id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, nil)
				orderItemRepo.On("GetLatestByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when get users ids by role names and location id",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, nil)
				orderItemRepo.On("GetLatestByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when missing reference id while publishing notification",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when missing reference_id when publish notification"),
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, nil)
				orderItemRepo.On("GetLatestByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
		{
			Name: "Fail case: Error when publish notification context with kafka",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, nil)
				orderItemRepo.On("GetLatestByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  ctx,
			Req: []interface{}{
				entities.StudentProduct{},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Product{}, nil)
				userRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.User{}, nil)
				orderItemRepo.On("GetLatestByStudentProductID", mock.Anything, mock.Anything, mock.Anything).Return(entities.OrderItem{
					OrderID: pgtype.Text{
						String: constant.OrderID,
						Status: pgtype.Present,
					},
				}, nil)
				userRepo.On("GetUserIDsByRoleNamesAndLocationID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			kafka = &mockKafka.KafkaManagement{}
			jsm = &mockNats.JetStreamManagement{}

			testCase.Setup(testCase.Ctx)
			s := &DiscountEventService{
				JSM:             jsm,
				Kafka:           kafka,
				DiscountService: nil,
				UserRepo:        userRepo,
				OrderItemRepo:   orderItemRepo,
				ProductRepo:     productRepo,
			}
			studentProductReq := testCase.Req.([]interface{})[0].(entities.StudentProduct)
			err := s.PublishNotificationForStudentProductWithScheduleTag(testCase.Ctx, db, studentProductReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userRepo, orderItemRepo, productRepo)
		})
	}
}

func TestDiscountEventService_PublishNotification(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db            *mockDb.Ext
		userRepo      *mockRepositories.MockUserRepo
		productRepo   *mockRepositories.MockProductRepo
		orderItemRepo *mockRepositories.MockOrderItemRepo
		kafka         *mockKafka.KafkaManagement
		jsm           *mockNats.JetStreamManagement
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when missing reference id",
			Ctx:  ctx,
			Req: []interface{}{
				payload.UpsertSystemNotification{},
			},
			ExpectedErr: status.Errorf(codes.Internal, "error when missing reference_id when publish notification"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: Error when publishing kafka notification message",
			Ctx:  ctx,
			Req: []interface{}{
				payload.UpsertSystemNotification{
					ReferenceID: constant.OrderID,
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  ctx,
			Req: []interface{}{
				payload.UpsertSystemNotification{
					ReferenceID: constant.OrderID,
				},
			},
			Setup: func(ctx context.Context) {
				kafka.On("TracedPublishContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			userRepo = new(mockRepositories.MockUserRepo)
			orderItemRepo = new(mockRepositories.MockOrderItemRepo)
			productRepo = new(mockRepositories.MockProductRepo)
			kafka = &mockKafka.KafkaManagement{}
			jsm = &mockNats.JetStreamManagement{}

			testCase.Setup(testCase.Ctx)
			s := &DiscountEventService{
				JSM:             jsm,
				Kafka:           kafka,
				DiscountService: nil,
				UserRepo:        userRepo,
				OrderItemRepo:   orderItemRepo,
				ProductRepo:     productRepo,
			}
			payloadReq := testCase.Req.([]interface{})[0].(payload.UpsertSystemNotification)
			err := s.PublishNotification(testCase.Ctx, payloadReq)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, userRepo, orderItemRepo, productRepo)
		})
	}
}
