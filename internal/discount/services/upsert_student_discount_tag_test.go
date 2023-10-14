package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mockRepo "github.com/manabie-com/backend/mock/discount/repositories"
	mockServices "github.com/manabie-com/backend/mock/discount/services/domain_service"
	mockDatabase "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDiscountService_UpsertStudentDiscountTag(t *testing.T) {
	t.Parallel()

	const (
		ctxUserID = "user-id"
	)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Mock objects
	mockTx := &mockDatabase.Tx{}
	mockDB := new(mockDatabase.Ext)
	mockDiscountRepo := new(mockRepo.MockDiscountRepo)
	mockDiscountTagService := new(mockServices.MockDiscountTagService)

	s := &DiscountService{
		DB:                 mockDB,
		DiscountRepo:       mockDiscountRepo,
		DiscountTagService: mockDiscountTagService,
	}

	userDiscountTagRecords := []*entities.UserDiscountTag{
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()),
			DiscountTagID: database.Text("test-1"),
		},
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()),
			DiscountTagID: database.Text("test-2"),
		},
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_FAMILY.String()),
			DiscountTagID: database.Text("test-3"),
		},
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_SINGLE_PARENT.String()),
			DiscountTagID: database.Text("test-4"),
		},
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_SIBLING.String()),
			DiscountTagID: database.Text("test-5"),
		},
		{
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_COMBO.String()),
			DiscountTagID: database.Text("test-6"),
		},
	}

	employeeFullTimeDiscounts := []*entities.Discount{
		{
			DiscountID:    database.Text("1"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()),
			DiscountTagID: database.Text("test-1"),
		},
		{
			DiscountID:    database.Text("2"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()),
			DiscountTagID: database.Text("test-1"),
		},
	}

	employeePartTimeDiscounts := []*entities.Discount{
		{
			DiscountID:    database.Text("3"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()),
			DiscountTagID: database.Text("test-2"),
		},
		{
			DiscountID:    database.Text("4"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()),
			DiscountTagID: database.Text("test-2"),
		},
	}

	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase + " no existing record to delete and discount tag ids on request are empty",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " delete user discount tag records when request discount tag ids are empty",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(userDiscountTagRecords, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " no changes happen, all discount tag ids in payload are in existing records",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3", "test-4", "test-5", "test-6"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(userDiscountTagRecords, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " single record on create user discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " multiple record on create user discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeePartTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " remain existing record and create new record",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.UserDiscountTag{userDiscountTagRecords[2]}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeePartTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " remain existing records and delete existing record not in discount tag ids payload",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(userDiscountTagRecords, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: constant.HappyCase + " remain existing records, create new records and delete existing record not in discount tag ids payload",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3"},
			},
			ExpectedResp: &pb.UpsertStudentDiscountTagResponse{
				Successful: true,
			},
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.UserDiscountTag{userDiscountTagRecords[2], userDiscountTagRecords[3]}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeePartTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				mockTx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: empty student id",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId: "",
			},
			ExpectedErr: status.Error(codes.FailedPrecondition, "student id should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name:        "Fail case: not existing student id on payload",
			Ctx:         interceptors.ContextWithUserID(ctx, mock.Anything),
			Req:         &pb.UpsertStudentDiscountTagRequest{},
			ExpectedErr: status.Error(codes.FailedPrecondition, "student id should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: cannot retrieve user discount tags of student",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: cannot delete user discount tag records when request discount tag ids are empty",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(userDiscountTagRecords, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: cannot get discount records when creating new user discount tag records",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3", "test-4"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil, constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: create user discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3", "test-4"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: multiple record on create user discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeePartTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: remain existing record and create new record",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.UserDiscountTag{userDiscountTagRecords[2]}, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeeFullTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				mockDiscountRepo.On("GetByDiscountTagIDs", ctx, mock.Anything, mock.Anything).Once().Return(employeePartTimeDiscounts, nil)
				mockDiscountTagService.On("CreateUserDiscountTag", ctx, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "Fail case: remain existing records and delete existing records not in discount tag ids payload",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pb.UpsertStudentDiscountTagRequest{
				StudentId:      "test-student-1",
				DiscountTagIds: []string{"test-1", "test-2", "test-3"},
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				mockDiscountTagService.On("RetrieveEligibleDiscountTagsOfStudent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(userDiscountTagRecords, nil)
				mockDB.On("Begin", ctx).Once().Return(mockTx, nil)
				mockDiscountTagService.On("SoftDeleteUserDiscountTagsByTypesAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.ErrDefault)
				mockTx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			response, err := s.UpsertStudentDiscountTag(testCase.Ctx, testCase.Req.(*pb.UpsertStudentDiscountTagRequest))

			if testCase.ExpectedErr == nil {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, response)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB, mockDiscountTagService, mockDiscountRepo)
		})
	}
}
