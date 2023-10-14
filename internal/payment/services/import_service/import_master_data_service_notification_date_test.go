package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestImportService_ImportNotificationDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	tx := new(mockDb.Tx)
	db := new(mockDb.Ext)
	mockNotificationDateRepo := new(mockRepositories.MockNotificationDateRepo)

	s := &ImportMasterDataService{
		DB:                   db,
		NotificationDateRepo: mockNotificationDateRepo,
	}

	testcases := []utils.TestCase{
		{
			Name:        constant.NoDataInCsvFile,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req:         &pb.ImportNotificationDateRequest{},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "only headers in csv file",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, constant.NoDataInCsvFile),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - number of columns != 4",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - number of column should be 4"),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date
				1,ORDER_TYPE_NEW,1
				2,ORDER_TYPE_NEW,2
				3,ORDER_TYPE_NEW,3`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - first column name (toLowerCase) != notification_date_id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - first column (toLowerCase) should be 'notification_date_id'"),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`invalid_notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0
				2,ORDER_TYPE_NEW,2,0
				3,ORDER_TYPE_NEW,3,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - second column name (toLowerCase) != order_type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - second column (toLowerCase) should be 'order_type'"),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,invalid_order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0
				2,ORDER_TYPE_NEW,2,0
				3,ORDER_TYPE_NEW,3,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - third column name (toLowerCase) != notification_date",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - third column (toLowerCase) should be 'notification_date'"),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,invalid_notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0
				2,ORDER_TYPE_NEW,2,0
				3,ORDER_TYPE_NEW,3,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name:        "invalid file - fourth column name (toLowerCase) != is_archived",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Error(codes.InvalidArgument, "csv file invalid format - fourth column (toLowerCase) should be 'is_archived'"),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,invalid_is_archived
				1,ORDER_TYPE_NEW,1,0
				2,ORDER_TYPE_NEW,2,0
				3,ORDER_TYPE_NEW,3,1`),
			},
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing valid file (capitalized header still valid) with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "upsert notification date fields with no error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
		{
			Name: "missing mandatory data (except ID), error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,,1,0
				2,ORDER_TYPE_NEW,,0
				3,ORDER_TYPE_NEW,3,`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseNotificationDate, fmt.Errorf("missing mandatory data: order_type")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf(constant.UnableToParseNotificationDate, fmt.Errorf("missing mandatory data: notification_date")),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf(constant.UnableToParseNotificationDate, fmt.Errorf("missing mandatory data: is_archived")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "wrong number of data in a record",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,0
				2,ORDER_TYPE_NEW,2,0
				3,ORDER_TYPE_NEW,`),
			},
			ExpectedErr: status.Error(codes.InvalidArgument, "record on line 4: wrong number of fields"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "parsing is_archived with error lines in response",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,1,3`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf(constant.UnableToParseNotificationDate, fmt.Errorf("error parsing is_archived: strconv.ParseBool: parsing \"3\": invalid syntax")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when notification_date is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,33,0
				1,ORDER_TYPE_NEW,0,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("invalid notification date: %s", fmt.Errorf("notification_date should be greater than 30 and smaller than 0")),
					},
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("invalid notification date: %s", fmt.Errorf("notification_date should be greater than 30 and smaller than 0")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when order_type is invalid",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW_INVALID,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("invalid notification date: %s", fmt.Errorf("order_type is invalid")),
					},
				},
			},
			Setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when duplicate order type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,10,0
				2,ORDER_TYPE_NEW,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 3,
						Error:     fmt.Sprintf("invalid notification date: %s", fmt.Errorf("duplicate order type: %s", pb.OrderType_ORDER_TYPE_NEW.String())),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when update notification date",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to update notification date item: %s", constant.ErrDefault),
					},
				},
			},

			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(constant.ErrDefault)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when create notification date",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				,ORDER_TYPE_NEW,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to create new notification date item: %s", constant.ErrDefault),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(constant.ErrDefault)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: "Fail case: Error when create and update notification dates",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,10,0
				2,ORDER_TYPE_ENROLLMENT,10,0
				,ORDER_TYPE_UPDATE,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{
					{
						RowNumber: 2,
						Error:     fmt.Sprintf("unable to update notification date item: %s", constant.ErrDefault),
					},
					{
						RowNumber: 4,
						Error:     fmt.Sprintf("unable to create new notification date item: %s", constant.ErrDefault),
					},
				},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(constant.ErrDefault)
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(constant.ErrDefault)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.ImportNotificationDateRequest{
				Payload: []byte(`notification_date_id,order_type,notification_date,is_archived
				1,ORDER_TYPE_NEW,10,0
				2,ORDER_TYPE_ENROLLMENT,10,0
				,ORDER_TYPE_UPDATE,10,0`),
			},
			ExpectedResp: &pb.ImportNotificationDateResponse{
				Errors: []*pb.ImportNotificationDateResponse_ImportNotificationDateError{},
			},
			Setup: func(ctx context.Context) {
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				mockNotificationDateRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)

			resp, err := s.ImportNotificationDate(testCase.Ctx, testCase.Req.(*pb.ImportNotificationDateRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)

				for i, expectedErr := range testCase.ExpectedResp.(*pb.ImportNotificationDateResponse).Errors {
					assert.Equal(t, expectedErr.RowNumber, resp.Errors[i].RowNumber)
					assert.Contains(t, expectedErr.Error, resp.Errors[i].Error)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockNotificationDateRepo)
		})
	}
}
