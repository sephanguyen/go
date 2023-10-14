package dto

import (
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestAutoCreateTimesheetFlag_ToEntity(t *testing.T) {
	var (
		autoCreateTimesheetFlag = &AutoCreateTimesheetFlag{
			StaffID: "staff1_1",
			FlagOn:  true,
		}

		autoCreateTimesheetFlagExpect = &entity.AutoCreateTimesheetFlag{
			StaffID:   database.Text("staff1_1"),
			FlagOn:    database.Bool(true),
			CreatedAt: pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt: pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entity success",
			request:      autoCreateTimesheetFlag,
			expectedResp: autoCreateTimesheetFlagExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*AutoCreateTimesheetFlag).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestAutoCreateTimesheetFlag_ValidateUpsertInfo(t *testing.T) {
	t.Parallel()

	var (
		autoCreateTimesheetFlagWithEmptyStaffId = &AutoCreateTimesheetFlag{
			StaffID: "",
			FlagOn:  true,
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "fail with empty staff id",
			request:      autoCreateTimesheetFlagWithEmptyStaffId,
			expectedResp: fmt.Errorf("staff id must not be empty"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*AutoCreateTimesheetFlag).ValidateUpsertInfo()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewAutoCreateTimeSheetFlagFromRPCUpdateRequest(t *testing.T) {
	t.Parallel()

	var (
		staffId                         = "staff-1"
		autoCreateTimesheetFlagExpected = &AutoCreateTimesheetFlag{
			StaffID: staffId,
			FlagOn:  true,
		}
		autoCreateTimesheetFlagRequest = &pb.UpdateAutoCreateTimesheetFlagRequest{
			StaffId: staffId,
			FlagOn:  true,
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new auto create timesheet flag from rpc success",
			request:      autoCreateTimesheetFlagRequest,
			expectedResp: autoCreateTimesheetFlagExpected,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewAutoCreateTimeSheetFlagFromRPCUpdateRequest(testcase.request.(*pb.UpdateAutoCreateTimesheetFlagRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
