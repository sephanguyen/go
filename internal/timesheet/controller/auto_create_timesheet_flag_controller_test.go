package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_auto_flag_services "github.com/manabie-com/backend/mock/timesheet/service/autocreatetimesheetflag"
	mock_mastermgmt_configuration_services "github.com/manabie-com/backend/mock/timesheet/service/mastermgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	AutoCreateStaffID = "user_id"
	AutoCreateFlag    = "test flag"
)

var (
	UpdateStaffId = idutil.ULIDNow()
	UpdateFlagOn  = false
)

func TestAutoCreateTimesheetFlagController_UpdateAutoCreateTimesheetFlag(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	autoCreateFlagSV := new(mock_auto_flag_services.MockAutoCreateTimesheetFlagServiceImpl)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)
	mockJsm := new(mock_nats.JetStreamManagement)

	ctl := &AutoCreateTimesheetFlagController{
		JSM:                            mockJsm,
		AutoCreateFlagService:          autoCreateFlagSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: UpdateStaffId,
				FlagOn:  UpdateFlagOn},
			expectedErr:  nil,
			expectedResp: &pb.UpdateAutoCreateTimesheetFlagResponse{Successful: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				autoCreateFlagSV.On("UpsertFlag", ctx, mock.Anything).Return(nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, constants.SubjectTimesheetAutoCreateFlag, mock.Anything, mock.Anything).Once().Return("", nil)
			},
		},
		{
			name: "error case upsert auto create timesheet flag failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: UpdateStaffId,
				FlagOn:  UpdateFlagOn},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find auto create timesheet flag error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.UpdateAutoCreateTimesheetFlagResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				autoCreateFlagSV.On("UpsertFlag", ctx, mock.Anything).Return(status.Error(codes.Internal, fmt.Sprintf("find auto create timesheet flag error: %s", pgx.ErrNoRows.Error()))).Once()
			},
		},
		{
			name: "error case publish update auto create timesheet flag failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: UpdateStaffId,
				FlagOn:  UpdateFlagOn},
			expectedErr:  status.Error(codes.Internal, "jsm error: PublishUpdateAutoCreateFlagEvent JSM.PublishAsyncContext failed, msgID: MsgID, Error"),
			expectedResp: (*pb.UpdateAutoCreateTimesheetFlagResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				autoCreateFlagSV.On("UpsertFlag", ctx, mock.Anything).Return(nil).Once()
				mockJsm.On("PublishAsyncContext", mock.Anything, constants.SubjectTimesheetAutoCreateFlag, mock.Anything, mock.Anything).Once().Return("MsgID", fmt.Errorf("Error"))
			},
		},
		{
			name: "error case invalid request",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: "",
				FlagOn:  UpdateFlagOn},
			expectedErr:  status.Error(codes.InvalidArgument, "staff id must not be empty"),
			expectedResp: (*pb.UpdateAutoCreateTimesheetFlagResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
			},
		},
		{
			name: "error case when get value from Master Service",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: "",
				FlagOn:  UpdateFlagOn},
			expectedErr:  status.Error(codes.Internal, "unknown service"),
			expectedResp: (*pb.UpdateAutoCreateTimesheetFlagResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, errors.New("unknown service")).Once()
			},
		},
		{
			name: "error case when get value from Master Service",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateAutoCreateTimesheetFlagRequest{
				StaffId: "",
				FlagOn:  UpdateFlagOn},
			expectedErr:  status.Error(codes.FailedPrecondition, "don't have permission to modify timesheet"),
			expectedResp: (*pb.UpdateAutoCreateTimesheetFlagResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateAutoCreateTimesheetFlagRequest)
			resp, err := ctl.UpdateAutoCreateTimesheetFlag(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
