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
	mockServices "github.com/manabie-com/backend/mock/discount/services/domain_service"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	pmpb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
	paymentPb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDiscountService_RetrieveActiveStudentDiscountTag(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		discountTagService *mockServices.MockDiscountTagService
	)

	singleDiscountTag := &entities.DiscountTag{
		DiscountTagID:   database.Text("1"),
		DiscountTagName: database.Text("sample-discount-1"),
	}

	multiUserDiscountTag := []*entities.UserDiscountTag{
		{
			UserID:        database.Text("1"),
			CreatedAt:     database.Timestamptz(time.Now()),
			DiscountTagID: database.Text("1"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_FULL_TIME.String()),
		},
		{
			UserID:        database.Text("1"),
			CreatedAt:     database.Timestamptz(time.Now()),
			DiscountTagID: database.Text("2"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_EMPLOYEE_PART_TIME.String()),
		},
		{
			UserID:        database.Text("1"),
			CreatedAt:     database.Timestamptz(time.Now()),
			DiscountTagID: database.Text("3"),
			DiscountType:  database.Text(paymentPb.DiscountType_DISCOUNT_TYPE_FAMILY.String()),
		},
	}

	multiDiscountTag := []*entities.DiscountTag{
		{
			DiscountTagID:   database.Text("1"),
			DiscountTagName: database.Text("sample-discount-1"),
		},
		{
			DiscountTagID:   database.Text("2"),
			DiscountTagName: database.Text("sample-discount-2"),
		},
	}

	testcases := []utils.TestCase{
		{
			Name: constant.HappyCase + " no active user tag records retrieved",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedResp: &pmpb.RetrieveActiveStudentDiscountTagResponse{
				StudentId:          "test-1",
				DiscountTagDetails: []*pmpb.RetrieveActiveStudentDiscountTagResponse_DiscountTagDetail{},
			},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveActiveDiscountTagIDsByDateAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
		{
			Name: constant.HappyCase + " retrieve single discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedResp: &pmpb.RetrieveActiveStudentDiscountTagResponse{
				StudentId:          "test-1",
				DiscountTagDetails: []*pmpb.RetrieveActiveStudentDiscountTagResponse_DiscountTagDetail{},
			},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveActiveDiscountTagIDsByDateAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{singleDiscountTag.DiscountTagID.String}, nil)
				discountTagService.On("RetrieveDiscountTagByDiscountTagID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(singleDiscountTag, nil)
			},
		},
		{
			Name: constant.HappyCase + " retrieve multiple discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedResp: &pmpb.RetrieveActiveStudentDiscountTagResponse{
				StudentId:          "test-1",
				DiscountTagDetails: []*pmpb.RetrieveActiveStudentDiscountTagResponse_DiscountTagDetail{},
			},
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveActiveDiscountTagIDsByDateAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					multiUserDiscountTag[0].DiscountTagID.String,
					multiUserDiscountTag[1].DiscountTagID.String,
				}, nil)
				discountTagService.On("RetrieveDiscountTagByDiscountTagID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(multiDiscountTag[0], nil)
				discountTagService.On("RetrieveDiscountTagByDiscountTagID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(multiDiscountTag[1], nil)
			},
		},
		{
			Name: "Fail case: missing payload student id",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedErr: status.Error(codes.FailedPrecondition, "student id should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: empty payload student id",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedErr: status.Error(codes.FailedPrecondition, "student id should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: missing payload discount date",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId: "test-1",
			},
			ExpectedErr: status.Error(codes.FailedPrecondition, "discount date request should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: empty payload discount date",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: nil,
			},
			ExpectedErr: status.Error(codes.FailedPrecondition, "discount date request should be required"),
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "Fail case: error retrieve active user discount tag",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveActiveDiscountTagIDsByDateAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error retrieve discount tag record",
			Ctx:  interceptors.ContextWithUserID(ctx, mock.Anything),
			Req: &pmpb.RetrieveActiveStudentDiscountTagRequest{
				StudentId:           "test-1",
				DiscountDateRequest: timestamppb.Now(),
			},
			ExpectedErr: status.Error(codes.Internal, constant.ErrDefault.Error()),
			Setup: func(ctx context.Context) {
				discountTagService.On("RetrieveActiveDiscountTagIDsByDateAndUserID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{
					multiUserDiscountTag[0].DiscountTagID.String,
					multiUserDiscountTag[1].DiscountTagID.String,
				}, nil)
				discountTagService.On("RetrieveDiscountTagByDiscountTagID", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			discountTagService = new(mockServices.MockDiscountTagService)

			testCase.Setup(testCase.Ctx)
			s := &DiscountService{
				DB:                 db,
				DiscountTagService: discountTagService,
			}
			_, err := s.RetrieveActiveStudentDiscountTag(testCase.Ctx, testCase.Req.(*pmpb.RetrieveActiveStudentDiscountTagRequest))
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, discountTagService)
		})
	}
}
