package dto

import (
	"fmt"
	"testing"
	"time"

	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimesheetConfirmationPeriod_ValidateGetPeriodInfo(t *testing.T) {
	t.Parallel()

	var (
		requestWithEmptyDate = &tpb.GetTimesheetConfirmationPeriodByDateRequest{}
		requestWithSmallDate = &tpb.GetTimesheetConfirmationPeriodByDateRequest{
			Date: timestamppb.New(time.Now().AddDate(-100, 0, 0)),
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "fail with empty date",
			request:      requestWithEmptyDate,
			expectedResp: fmt.Errorf("date must not be empty"),
		},
		{
			name:         "fail with small date",
			request:      requestWithSmallDate,
			expectedResp: fmt.Errorf("date must be larger than 2022-01-01"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := ValidateGetPeriodInfo(testcase.request.(*tpb.GetTimesheetConfirmationPeriodByDateRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewTimesheetConfirmationPeriodToRPCResponse(t *testing.T) {
	t.Parallel()

	var (
		startDate                   = time.Now()
		endDate                     = startDate.Add(time.Hour * 1)
		timesheetConfirmationPeriod = &TimesheetConfirmationPeriod{
			ID:        "1",
			StartDate: startDate,
			EndDate:   endDate,
		}
		timesheetConfirmationPeriodExpect = &tpb.GetTimesheetConfirmationPeriodByDateResponse{
			TimesheetConfirmationPeriod: &tpb.TimesheetConfirmationPeriod{
				Id:        "1",
				StartDate: timestamppb.New(timesheetConfirmationPeriod.StartDate),
				EndDate:   timestamppb.New(timesheetConfirmationPeriod.EndDate),
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new timesheet confirmation period to RPC response success",
			request:      timesheetConfirmationPeriod,
			expectedResp: timesheetConfirmationPeriodExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewTimesheetConfirmationPeriodToRPCResponse(testcase.request.(*TimesheetConfirmationPeriod))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
