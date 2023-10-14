package dto

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewOtherWorkingHoursFromRPCRequest(t *testing.T) {
	t.Parallel()
	var (
		startTime   = time.Now().UTC()
		endTime     = startTime.Add(time.Hour * 1)
		timesheetId = "timesheet-1"
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "new other working hours from rpc success",
			request: &pb.OtherWorkingHoursRequest{
				OtherWorkingHoursId: "owh-1",
				TimesheetConfigId:   "timesheet-config-1",
				Remarks:             "",
				StartTime:           timestamppb.New(startTime),
				EndTime:             timestamppb.New(endTime),
			},
			expectedResp: &OtherWorkingHours{
				ID:                "owh-1",
				TimesheetID:       timesheetId,
				TimesheetConfigID: "timesheet-config-1",
				StartTime:         startTime,
				EndTime:           endTime,
				Remarks:           "",
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewOtherWorkingHoursFromRPCRequest(timesheetId, testcase.request.(*pb.OtherWorkingHoursRequest))
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestNewOtherWorkingHoursFromEntity(t *testing.T) {
	t.Parallel()
	var (
		startTime          = time.Now().UTC()
		endTime            = startTime.Add(time.Hour * 1)
		otherWorkingHoursE = &entity.OtherWorkingHours{
			ID:                database.Text("other_working_hours_1"),
			TimesheetID:       database.Text("timesheetID_1"),
			TimesheetConfigID: database.Text("timesheet_configID_1"),
			StartTime:         database.Timestamptz(startTime),
			EndTime:           database.Timestamptz(endTime),
			TotalHour:         database.Int2(1),
			Remarks:           database.Text(""),
			CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		}
		deletedOtherWorkingHoursE = &entity.OtherWorkingHours{
			ID:                database.Text("other_working_hours_1"),
			TimesheetID:       database.Text("timesheetID_1"),
			TimesheetConfigID: database.Text("timesheet_configID_1"),
			StartTime:         database.Timestamptz(startTime),
			EndTime:           database.Timestamptz(endTime),
			TotalHour:         database.Int2(1),
			Remarks:           database.Text(""),
			CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:         database.Timestamptz(time.Now()),
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:    "new other working hours from entity success",
			request: otherWorkingHoursE,
			expectedResp: &OtherWorkingHours{
				ID:                "other_working_hours_1",
				TimesheetID:       "timesheetID_1",
				TimesheetConfigID: "timesheet_configID_1",
				StartTime:         startTime,
				EndTime:           endTime,
				TotalHour:         1,
				Remarks:           "",
			},
		},
		{
			name:    "new other working hours from entity set deleted",
			request: deletedOtherWorkingHoursE,
			expectedResp: &OtherWorkingHours{
				ID:                "other_working_hours_1",
				TimesheetID:       "timesheetID_1",
				TimesheetConfigID: "timesheet_configID_1",
				StartTime:         startTime,
				EndTime:           endTime,
				TotalHour:         1,
				Remarks:           "",
				IsDeleted:         true,
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewOtherWorkingHoursFromEntity(testcase.request.(*entity.OtherWorkingHours))
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestOtherWorkingHours_IsEqual(t *testing.T) {
	t.Parallel()
	otherWorkingHours1 := &OtherWorkingHours{
		ID:                "other_working_hours_1",
		TimesheetID:       "timesheetID_1",
		TimesheetConfigID: "timesheet_configID_1",
	}
	otherWorkingHours2 := &OtherWorkingHours{
		ID:                "other_working_hours_1",
		TimesheetID:       "timesheetID_1",
		TimesheetConfigID: "timesheet_configID_1",
	}
	otherWorkingHours3 := &OtherWorkingHours{
		ID:                "other_working_hours_3",
		TimesheetID:       "timesheetID_3",
		TimesheetConfigID: "timesheet_configID_3",
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "compare success with true",
			request:      []*OtherWorkingHours{otherWorkingHours1, otherWorkingHours2},
			expectedResp: true,
		},
		{
			name:         "compare success with false",
			request:      []*OtherWorkingHours{otherWorkingHours1, otherWorkingHours3},
			expectedResp: false,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.([]*OtherWorkingHours)[0].IsEqual(testcase.request.([]*OtherWorkingHours)[1])
			assert.Equal(t, testcase.expectedResp.(bool), resp)
		})
	}
}

func TestOtherWorkingHours_ToEntity(t *testing.T) {
	t.Parallel()
	startTime := time.Now()
	endTime := startTime.Add(time.Hour * 1)
	var (
		otherWorkingHoursDto = &OtherWorkingHours{
			ID:                "other_working_hours_1",
			TimesheetID:       "timesheetID_1",
			TimesheetConfigID: "timesheet_configID_1",
			Remarks:           "",
			TotalHour:         1,
			StartTime:         startTime,
			EndTime:           endTime,
		}
		otherWorkingHoursE = &entity.OtherWorkingHours{
			ID:                database.Text("other_working_hours_1"),
			TimesheetID:       database.Text("timesheetID_1"),
			TimesheetConfigID: database.Text("timesheet_configID_1"),
			StartTime:         database.Timestamptz(startTime),
			EndTime:           database.Timestamptz(endTime),
			TotalHour:         database.Int2(1),
			Remarks:           database.Text(""),
			CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entity success",
			request:      otherWorkingHoursDto,
			expectedResp: otherWorkingHoursE,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*OtherWorkingHours).ToEntity()
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestOtherWorkingHours_Validate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name             string
		request          interface{}
		expectedResponse interface{}
	}{
		{
			name: "empty timesheet config id",
			request: &OtherWorkingHours{
				TimesheetConfigID: "",
			},
			expectedResponse: fmt.Errorf("other working type must not be empty"),
		},
		{
			name: "empty start time",
			request: &OtherWorkingHours{
				TimesheetConfigID: "1",
				StartTime:         time.Time{},
			},
			expectedResponse: fmt.Errorf("other working hours start time must not be nil"),
		},
		{
			name: "empty end time",
			request: &OtherWorkingHours{
				TimesheetConfigID: "1",
				StartTime:         time.Now(),
				EndTime:           time.Time{},
			},
			expectedResponse: fmt.Errorf("other working hours end time must not be nil"),
		},
		{
			name: "end time must be after start time",
			request: &OtherWorkingHours{
				TimesheetConfigID: "1",
				StartTime:         time.Now(),
				EndTime:           time.Now().Add(-2 * 24 * time.Hour),
			},
			expectedResponse: fmt.Errorf("other working hours end time must after start time"),
		},
		{
			name: fmt.Sprintf("remarks length over %d limit", constant.KOtherWorkingHoursRemarksLimit),
			request: &OtherWorkingHours{
				TimesheetConfigID: "1",
				StartTime:         time.Now(),
				EndTime:           time.Now(),
				Remarks:           strings.Repeat("a", constant.KOtherWorkingHoursRemarksLimit+1),
			},
			expectedResponse: fmt.Errorf("other working hours remarks must limit to 100 characters"),
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.request.(*OtherWorkingHours).Validate(), testcase.expectedResponse)
		})
	}
}

func TestListOtherWorkingHours_IsEqual(t *testing.T) {
	t.Parallel()
	startTime := time.Now()
	endTime := startTime.Add(time.Hour * 1)
	testCases := []struct {
		name             string
		request          interface{}
		toCompare        interface{}
		expectedResponse interface{}
	}{
		{
			name: "happy case",
			request: &ListOtherWorkingHours{
				{
					ID:                "other_working_hours_1",
					TimesheetID:       "timesheetID_1",
					TimesheetConfigID: "timesheet_configID_1",
					Remarks:           "",
					TotalHour:         1,
					StartTime:         startTime,
					EndTime:           endTime,
				},
			},
			toCompare: ListOtherWorkingHours{
				{
					ID:                "other_working_hours_1",
					TimesheetID:       "timesheetID_1",
					TimesheetConfigID: "timesheet_configID_1",
					Remarks:           "",
					TotalHour:         1,
					StartTime:         startTime,
					EndTime:           endTime,
				},
			},
			expectedResponse: true,
		},
		{
			name: "error case other working hours list array length mismatch",
			request: &ListOtherWorkingHours{
				{
					ID: "other_working_hours_1",
				},
			},
			toCompare: ListOtherWorkingHours{
				{
					ID: "other_working_hours_1",
				},
				{
					ID: "other_working_hours_2",
				},
			},
			expectedResponse: false,
		},
		{
			name:             "case both other working hours lists are empty",
			request:          &ListOtherWorkingHours{},
			toCompare:        ListOtherWorkingHours{},
			expectedResponse: true,
		},
		{
			name: "case lists are not equal",
			request: &ListOtherWorkingHours{
				{
					ID:                "other_working_hours_1",
					TimesheetID:       "timesheetID_1",
					TimesheetConfigID: "timesheet_configID_1",
					Remarks:           "",
					TotalHour:         1,
					StartTime:         startTime,
					EndTime:           endTime,
				},
			},
			toCompare: ListOtherWorkingHours{
				{
					ID:                "other_working_hours_2",
					TimesheetID:       "timesheetID_2",
					TimesheetConfigID: "timesheet_configID_2",
					Remarks:           "1",
					TotalHour:         1,
					StartTime:         startTime,
					EndTime:           endTime,
				},
			},
			expectedResponse: false,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.request.(*ListOtherWorkingHours).IsEqual(testcase.toCompare.(ListOtherWorkingHours)), testcase.expectedResponse)
		})
	}
}

func TestNewListOtherWorkingHoursFromRPCRequest(t *testing.T) {
	var (
		startTime   = time.Now().UTC()
		endTime     = startTime.Add(time.Hour * 1)
		timesheetId = "timesheet-1"
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "new list other working hours from rpc success",
			request: []*pb.OtherWorkingHoursRequest{
				{
					OtherWorkingHoursId: "owh-1",
					TimesheetConfigId:   "timesheet-config-1",
					Remarks:             "",
					StartTime:           timestamppb.New(startTime),
					EndTime:             timestamppb.New(endTime),
				},
			},
			expectedResp: ListOtherWorkingHours{
				{
					ID:                "owh-1",
					TimesheetID:       timesheetId,
					TimesheetConfigID: "timesheet-config-1",
					StartTime:         startTime,
					EndTime:           endTime,
					Remarks:           "",
				},
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewListOtherWorkingHoursFromRPCRequest(timesheetId, testcase.request.([]*pb.OtherWorkingHoursRequest))
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestListOtherWorkingHours_ToEntities(t *testing.T) {
	t.Parallel()
	startTime := time.Now()
	endTime := startTime.Add(time.Hour * 1)
	var (
		listOtherWorkingHours = ListOtherWorkingHours{
			{
				ID:                "other_working_hours_1",
				TimesheetID:       "timesheetID_1",
				TimesheetConfigID: "timesheet_configID_1",
				Remarks:           "",
				TotalHour:         1,
				StartTime:         startTime,
				EndTime:           endTime,
			},
		}
		otherWorkingHoursE = []*entity.OtherWorkingHours{
			{
				ID:                database.Text("other_working_hours_1"),
				TimesheetID:       database.Text("timesheetID_1"),
				TimesheetConfigID: database.Text("timesheet_configID_1"),
				StartTime:         database.Timestamptz(startTime),
				EndTime:           database.Timestamptz(endTime),
				TotalHour:         database.Int2(1),
				Remarks:           database.Text(""),
				CreatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:         pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
			},
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entities success",
			request:      listOtherWorkingHours,
			expectedResp: otherWorkingHoursE,
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(ListOtherWorkingHours).ToEntities()
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestListOtherWorkingHours_Validate(t *testing.T) {
	t.Parallel()
	startTime := time.Now()
	endTime := startTime.Add(time.Hour * 1)
	listOtherWorkingHours := &ListOtherWorkingHours{
		{
			ID:          "1",
			TimesheetID: "timesheetID_1",
		},
		{
			ID:          "2",
			TimesheetID: "timesheetID_2",
		},
		{
			ID:          "3",
			TimesheetID: "timesheetID_3",
		},
		{
			ID:          "4",
			TimesheetID: "timesheetID_4",
		},
		{
			ID:          "5",
			TimesheetID: "timesheetID_5",
		},
		{
			ID:          "6",
			TimesheetID: "timesheetID_6",
		},
	}

	testCases := []struct {
		name             string
		request          interface{}
		expectedResponse interface{}
	}{
		{
			name: "happy case",
			request: &ListOtherWorkingHours{
				{
					ID:                "other_working_hours_1",
					TimesheetID:       "timesheetID_1",
					TimesheetConfigID: "timesheet_configID_1",
					Remarks:           "",
					TotalHour:         1,
					StartTime:         startTime,
					EndTime:           endTime,
				},
			},
			expectedResponse: nil,
		},
		{
			name:             "validate other working hours list length over max limit",
			request:          listOtherWorkingHours,
			expectedResponse: fmt.Errorf("list other working hours must be limit to 5 rows"),
		},
		{
			name: "error case list other working hours has validation errors",
			request: &ListOtherWorkingHours{
				{
					ID:                "other_working_hours_1",
					TimesheetID:       "1",
					TimesheetConfigID: "timesheet_configID_1",
					Remarks:           "",
					TotalHour:         1,
					StartTime:         time.Time{},
					EndTime:           time.Time{},
				},
			},
			expectedResponse: fmt.Errorf("other working hours start time must not be nil"),
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			assert.Equal(t, testcase.request.(*ListOtherWorkingHours).Validate(), testcase.expectedResponse)
		})
	}
}
