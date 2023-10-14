package staff

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStaffService_UpdateStaffSetting(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	staffId := "staffId"

	staffRepo := new(mock_repositories.MockStaffRepo)
	jsm := new(mock_nats.JetStreamManagement)
	staffService := &StaffService{
		StaffRepo: staffRepo,
		JSM:       jsm,
	}

	existingStaff := &entity.Staff{}

	testCases := []TestCase{
		{
			name: "error staff id is empty",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             "",
				AutoCreateTimesheet: true,
			},
			expectedErr: status.Error(codes.InvalidArgument, "staff id cannot be null or empty"),
		},
		{
			name: "error cannot find staff",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             staffId,
				AutoCreateTimesheet: true,
			},
			setup: func(ctx context.Context) {
				staffRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("error find staff"))
			},
			expectedErr: status.Error(codes.Unknown, "error find staff"),
		},
		{
			name: "error cannot update staff",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             "jfsdfjsfsdf",
				AutoCreateTimesheet: true,
			},
			setup: func(ctx context.Context) {
				staffRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entity.Staff{}, nil)
				staffRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("error update staff"))
			},
			expectedErr: status.Error(codes.Unknown, "error update staff"),
		},
		{
			name: "error publish event",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             staffId,
				AutoCreateTimesheet: true,
			},
			setup: func(ctx context.Context) {
				staffRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				staffRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(nil, errors.New("error publish event"))
			},
			expectedErr: status.Error(codes.Unknown, "publishStaffSettingEvent error: error publish event"),
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             staffId,
				AutoCreateTimesheet: true,
			},
			setup: func(ctx context.Context) {
				staffRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				staffRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "happy case update auto create config same value with existing staff",
			ctx:  ctx,
			req: &pb.UpdateStaffSettingRequest{
				StaffId:             staffId,
				AutoCreateTimesheet: true,
			},
			setup: func(ctx context.Context) {
				existingStaff.AutoCreateTimesheet = pgtype.Bool{
					Bool:   true,
					Status: pgtype.Present,
				}
				staffRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				staffRepo.On("Update", mock.Anything, mock.Anything, mock.Anything).Once().Return(existingStaff, nil)
				jsm.On("PublishContext", mock.Anything, constants.SubjectStaffUpsertTimesheetConfig, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}

			_, err := staffService.UpdateStaffSetting(testCase.ctx, testCase.req.(*pb.UpdateStaffSettingRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
