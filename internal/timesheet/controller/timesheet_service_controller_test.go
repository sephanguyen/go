package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_mastermgmt_configuration_services "github.com/manabie-com/backend/mock/timesheet/service/mastermgmt"
	mock_timesheet_confirmation_services "github.com/manabie-com/backend/mock/timesheet/service/timesheet_confirmation"
	mock_services "github.com/manabie-com/backend/mock/timesheet/services/timesheet"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CreateTimesheetStaffID       = "user_id"
	CreateTimeSheetRemark        = "test remark"
	CreateTimesheetUserIDSuccess = CreateTimesheetStaffID
	CreateTimesheetUserIDFail    = "failed_user_id"
	CreateTimesheetLocationID    = "location_id"
)

var (
	CreateTimesheetID            = idutil.ULIDNow()
	UpdateTimesheetID            = idutil.ULIDNow()
	CreateTimesheetConfigID      = idutil.ULIDNow()
	CreateTimesheetTimesheetDate = timestamppb.Now()
)

type TestCase struct {
	name            string
	ctx             context.Context
	req             interface{}
	expectedResp    interface{}
	expectedErr     error
	setup           func(ctx context.Context)
	reqString       string
	reqTimesheetIDs []string
}

func TestTimesheetController_CreateTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetSV := new(mock_services.MockServiceImpl)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)
	timesheetConfirmationSV := new(mock_timesheet_confirmation_services.MockConfirmationWindowServiceImpl)

	ctl := &TimesheetServiceController{
		TimesheetService:               timesheetSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
		ConfirmationWindowService:      timesheetConfirmationSV,
	}

	owhs := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: CreateTimesheetConfigID,
		StartTime:         timestamppb.Now(),
		EndTime:           timestamppb.Now(),
	}
	testCases := []TestCase{

		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
				Remark: CreateTimeSheetRemark,
			},
			expectedErr:  nil,
			expectedResp: &pb.CreateTimesheetResponse{TimesheetId: CreateTimesheetID},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.On("CreateTimesheet", ctx, mock.Anything).
					Return(CreateTimesheetID, nil).Once()

			},
		},

		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
			},
			expectedErr:  status.Error(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()
				timesheetSV.
					On("CreateTimesheet", ctx, mock.Anything).
					Return("", status.Error(codes.Internal, "Already exists timesheet with same info")).
					Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()
				timesheetSV.
					On("CreateTimesheet", ctx, mock.Anything).
					Return("", status.Error(codes.Internal, "Already exists timesheet with same info")).
					Once()

			},
		},

		{
			name: "error case create timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
			},
			expectedErr:  status.Error(codes.Internal, "Already exists timesheet with same info"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.
					On("CreateTimesheet", ctx, mock.Anything).
					Return("", status.Error(codes.Internal, "Already exists timesheet with same info")).
					Once()

			},
		},
		{
			name: "error case invalid request",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       "",
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				Remark:        CreateTimeSheetRemark},
			expectedErr:  status.Error(codes.InvalidArgument, "staff id must not be empty"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).
					Return(true, nil).Once()
			},
		},

		{
			name: "error case when timesheet in a period is confirmed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
				Remark: CreateTimeSheetRemark,
			},
			expectedErr:  status.Error(codes.FailedPrecondition, "all data in this period have been confirm"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
		},

		{
			name: "error case check confirmation info error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.CreateTimesheetRequest{
				StaffId:       CreateTimesheetStaffID,
				LocationId:    CreateTimesheetLocationID,
				TimesheetDate: CreateTimesheetTimesheetDate,
				ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
					owhs,
				},
				Remark: CreateTimeSheetRemark,
			},
			expectedErr:  status.Error(codes.Internal, "error when get confirmation info"),
			expectedResp: (*pb.CreateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetDateAndLocation", ctx, mock.Anything, mock.Anything).
					Return(false, fmt.Errorf("error when get confirmation info")).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.CreateTimesheetRequest)
			resp, err := ctl.CreateTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetController_UpdateTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	timesheetSV := new(mock_services.MockServiceImpl)
	mastermgmtConfigurationSV := new(mock_mastermgmt_configuration_services.MockMasterConfigurationServiceImpl)
	timesheetConfirmationSV := new(mock_timesheet_confirmation_services.MockConfirmationWindowServiceImpl)

	ctl := &TimesheetServiceController{
		TimesheetService:               timesheetSV,
		MastermgmtConfigurationService: mastermgmtConfigurationSV,
		ConfirmationWindowService:      timesheetConfirmationSV,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
				Remark:      CreateTimeSheetRemark},
			expectedErr:  nil,
			expectedResp: &pb.UpdateTimesheetResponse{Success: true},
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetID", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.
					On("UpdateTimesheet", ctx, mock.Anything).
					Return(nil).
					Once()
			},
		},
		{
			name: "error case when timesheet service is off",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
				Remark:      CreateTimeSheetRemark,
			},
			expectedErr:  status.Error(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet"),
			expectedResp: (*pb.UpdateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetID", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.
					On("CreateTimesheet", ctx, mock.Anything).
					Return("", status.Error(codes.Internal, "Already exists timesheet with same info")).
					Once()

			},
		},
		{
			name: "error case when get configuration from Mastermgmt service have internal error",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
				Remark:      CreateTimeSheetRemark,
			},
			expectedErr:  status.Error(codes.Internal, "s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey"),
			expectedResp: (*pb.UpdateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey")).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetID", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.
					On("CreateTimesheet", ctx, mock.Anything).
					Return("", status.Error(codes.Internal, "Already exists timesheet with same info")).
					Once()

			},
		},
		{
			name: "error case create timesheet failed",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateTimesheetRequest{
				TimesheetId: UpdateTimesheetID,
				Remark:      CreateTimeSheetRemark},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error())),
			expectedResp: (*pb.UpdateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetID", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetSV.
					On("UpdateTimesheet", ctx, mock.Anything).
					Return(status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", pgx.ErrNoRows.Error()))).
					Once()
			},
		},
		{
			name: "error case invalid request",
			ctx:  interceptors.ContextWithUserID(ctx, CreateTimesheetUserIDSuccess),
			req: &pb.UpdateTimesheetRequest{
				TimesheetId: "",
				Remark:      CreateTimeSheetRemark},
			expectedErr:  status.Error(codes.InvalidArgument, "timesheet id must not be empty"),
			expectedResp: (*pb.UpdateTimesheetResponse)(nil),
			setup: func(ctx context.Context) {
				mastermgmtConfigurationSV.On("CheckPartnerTimesheetServiceIsOn", ctx, mock.Anything).
					Return(true, nil).Once()
				timesheetConfirmationSV.On("CheckModifyConditionByTimesheetID", ctx, mock.Anything).
					Return(true, nil).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpdateTimesheetRequest)
			resp, err := ctl.UpdateTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
