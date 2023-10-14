package dto

import (
	"fmt"
	"testing"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewTimesheetActionLogDTOFromNATSRPCRequest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	var (
		timesheetActionLogReq = &pb.TimesheetActionLogRequest{
			TimesheetId: "timesheet_1",
			ExecutedBy:  "user_1",
			IsSystem:    false,
			Action:      pb.TimesheetAction_EDITED,
			ExecutedAt:  timestamppb.New(now),
		}
		timesheetActionLogExpect = &TimesheetActionLogReq{
			TimesheetID: "timesheet_1",
			UserID:      "user_1",
			IsSystem:    false,
			Action:      "EDITED",
			ExecutedAt:  timestamppb.New(now).AsTime(),
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new action log request from NATS rpc request",
			request:      timesheetActionLogReq,
			expectedResp: timesheetActionLogExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewTimesheetActionLogDTOFromNATSRPCRequest(testcase.request.(*pb.TimesheetActionLogRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetActionLogReq_ValidateCreateInfo(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "happy case",
			request: &TimesheetActionLogReq{
				TimesheetID: "timesheet_1",
				UserID:      "user_1",
				IsSystem:    false,
				Action:      "EDITED",
				ExecutedAt:  timestamppb.New(now).AsTime(),
			},
			expectedResp: nil,
		},
		{
			name: "timesheet id empty",
			request: &TimesheetActionLogReq{
				TimesheetID: "",
				UserID:      "user_1",
				IsSystem:    false,
				Action:      "EDITED",
				ExecutedAt:  timestamppb.New(now).AsTime(),
			},
			expectedResp: fmt.Errorf("timesheet id must not be empty"),
		},
		{
			name: "user id empty",
			request: &TimesheetActionLogReq{
				TimesheetID: "timesheet_1",
				UserID:      "",
				IsSystem:    false,
				Action:      "EDITED",
				ExecutedAt:  timestamppb.New(now).AsTime(),
			},
			expectedResp: fmt.Errorf("user id must not be empty"),
		},
		{
			name: "is system set",
			request: &TimesheetActionLogReq{
				TimesheetID: "timesheet_1",
				UserID:      "",
				IsSystem:    true,
				Action:      "EDITED",
				ExecutedAt:  timestamppb.New(now).AsTime(),
			},
			expectedResp: nil,
		},
		{
			name: "action empty",
			request: &TimesheetActionLogReq{
				TimesheetID: "timesheet_1",
				UserID:      "user_1",
				IsSystem:    false,
				Action:      "",
				ExecutedAt:  timestamppb.New(now).AsTime(),
			},
			expectedResp: fmt.Errorf("action must not be empty"),
		},
		{
			name: "executed at empty",
			request: &TimesheetActionLogReq{
				TimesheetID: "timesheet_1",
				UserID:      "user_1",
				IsSystem:    false,
				Action:      "EDITED",
				ExecutedAt:  time.Time{},
			},
			expectedResp: fmt.Errorf("executed at must not be empty"),
		},
	}

	for _, testcase := range testCases {
		resp := testcase.request.(*TimesheetActionLogReq).ValidateCreateInfo()
		assert.Equal(t, testcase.expectedResp, resp)
	}
}
