package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_mastermgmt_configuration_services "github.com/manabie-com/backend/mock/timesheet/service/mastermgmt"
	mock_services "github.com/manabie-com/backend/mock/timesheet/service/timesheet_state_machine"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTimesheetController_CancelSubmissionTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetStateMachineSV := new(mock_services.MockTimesheetStateMachineService)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)

	ctl := &TimesheetStateMachineController{
		TimesheetStateMachineService:   timesheetStateMachineSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CancelSubmissionTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  nil,
			expectedResp: &pb.CancelSubmissionTimesheetResponse{Success: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("CancelSubmissionTimesheet", ctx, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CancelSubmissionTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.PermissionDenied, "don't have permission to modify timesheet"),
			expectedResp: (*pb.CancelSubmissionTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CancelSubmissionTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.CancelSubmissionTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()

			},
		},
		{
			name: "error cancel submission timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CancelSubmissionTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.CancelSubmissionTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("CancelSubmissionTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
		{
			name: "error case invalid request",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CancelSubmissionTimesheetRequest{
				TimesheetId: "",
			},
			expectedErr:  status.Error(codes.InvalidArgument, "timesheet id cannot be empty"),
			expectedResp: (*pb.CancelSubmissionTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.CancelSubmissionTimesheetRequest)
			resp, err := ctl.CancelSubmissionTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetController_DeleteTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetStateMachineSV := new(mock_services.MockTimesheetStateMachineService)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)

	ctl := &TimesheetStateMachineController{
		TimesheetStateMachineService:   timesheetStateMachineSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.DeleteTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  nil,
			expectedResp: &pb.DeleteTimesheetResponse{Success: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("DeleteTimesheet", ctx, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.DeleteTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.PermissionDenied, "don't have permission to modify timesheet"),
			expectedResp: (*pb.DeleteTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.DeleteTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.DeleteTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()

			},
		},
		{
			name: "error cancel delete timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.DeleteTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.DeleteTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("DeleteTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
		{
			name: "error case invalid request",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.DeleteTimesheetRequest{
				TimesheetId: "",
			},
			expectedErr:  status.Error(codes.InvalidArgument, "timesheet id cannot be empty"),
			expectedResp: (*pb.DeleteTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.DeleteTimesheetRequest)
			resp, err := ctl.DeleteTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetController_ApproveTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetStateMachineSV := new(mock_services.MockTimesheetStateMachineService)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)

	ctl := &TimesheetStateMachineController{
		TimesheetStateMachineService:   timesheetStateMachineSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
	}
	timesheetIds := []string{"ts-1", "ts-2"}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.ApproveTimesheetRequest{
				TimesheetIds: timesheetIds,
			},
			expectedErr:  nil,
			expectedResp: &pb.ApproveTimesheetResponse{Success: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("ApproveTimesheet", ctx, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.ApproveTimesheetRequest{
				TimesheetIds: timesheetIds,
			},
			expectedErr:  status.Error(codes.PermissionDenied, "don't have permission to modify timesheet"),
			expectedResp: (*pb.ApproveTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.ApproveTimesheetRequest{
				TimesheetIds: timesheetIds,
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.ApproveTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()

			},
		},
		{
			name: "error cancel delete timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.ApproveTimesheetRequest{
				TimesheetIds: timesheetIds,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.ApproveTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("ApproveTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
		{
			name:         "error case invalid request",
			ctx:          interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req:          &pb.ApproveTimesheetRequest{},
			expectedErr:  status.Error(codes.InvalidArgument, "timesheet ids cannot be empty"),
			expectedResp: (*pb.ApproveTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.ApproveTimesheetRequest)
			resp, err := ctl.ApproveTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetStateMachineController_SubmitTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	timesheetStateMachineSV := new(mock_services.MockTimesheetStateMachineService)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)

	ctl := &TimesheetStateMachineController{
		TimesheetStateMachineService:   timesheetStateMachineSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
	}
	timesheetId := "ts-1"
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.SubmitTimesheetRequest{
				TimesheetId: timesheetId,
			},
			expectedErr:  nil,
			expectedResp: &pb.SubmitTimesheetResponse{Success: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("SubmitTimesheet", ctx, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.SubmitTimesheetRequest{
				TimesheetId: timesheetId,
			},
			expectedErr:  status.Error(codes.PermissionDenied, "don't have permission to modify timesheet"),
			expectedResp: (*pb.SubmitTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.SubmitTimesheetRequest{
				TimesheetId: timesheetId,
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.SubmitTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()

			},
		},
		{
			name: "error case when timesheetId is empty",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.SubmitTimesheetRequest{
				TimesheetId: "",
			},
			expectedErr:  status.Error(codes.InvalidArgument, "timesheet id cannot be empty"),
			expectedResp: (*pb.SubmitTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("SubmitTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
		{
			name: "error case when submit failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.SubmitTimesheetRequest{
				TimesheetId: CreateTimesheetID,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.SubmitTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetStateMachineSV.
					On("SubmitTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.SubmitTimesheetRequest)
			resp, err := ctl.SubmitTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
