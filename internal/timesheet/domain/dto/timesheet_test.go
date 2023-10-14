package dto

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

var (
	now = time.Now()

	_timesheet1 = &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		IsCreated:       true,
	}
	_timesheet2 = &Timesheet{
		ID:                       "timesheet2_1",
		StaffID:                  "staff2_1",
		LocationID:               "location2_1",
		TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:            now,
		ListTimesheetLessonHours: nil,
		IsCreated:                true,
	}
	_timesheet3 = &Timesheet{
		ID:              "timesheet3_1",
		StaffID:         "staff3_1",
		LocationID:      "location3_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,

		IsCreated: true,
	}
)

func TestNewTimesheetFromRPCCreateRequest(t *testing.T) {
	var (
		createTimesheetReq = &tpb.CreateTimesheetRequest{
			StaffId:    "staff-1",
			LocationId: "loc-1",
			Remark:     "",
			ListOtherWorkingHours: []*tpb.OtherWorkingHoursRequest{
				{
					OtherWorkingHoursId: "1",
				},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:    "new timesheet from rpc create request",
			request: createTimesheetReq,
			expectedResp: &Timesheet{
				StaffID:         "staff-1",
				LocationID:      "loc-1",
				TimesheetStatus: "",

				IsCreated: false,
				ListOtherWorkingHours: ListOtherWorkingHours{
					{
						ID: "1",
					},
				},
			},
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewTimesheetFromRPCCreateRequest(testcase.request.(*tpb.CreateTimesheetRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewTimesheetFromRPCUpdateRequest(t *testing.T) {
	var (
		createTimesheetReq = &tpb.UpdateTimesheetRequest{
			TimesheetId: "1",
			Remark:      "test",
			ListOtherWorkingHours: []*tpb.OtherWorkingHoursRequest{
				{
					OtherWorkingHoursId: "1",
				},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:    "new timesheet from rpc update request",
			request: createTimesheetReq,
			expectedResp: &Timesheet{
				ID:     "1",
				Remark: "test",

				IsCreated: false,
				ListOtherWorkingHours: ListOtherWorkingHours{
					{
						ID:          "1",
						TimesheetID: "1",
					},
				},
			},
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewTimesheetFromRPCUpdateRequest(testcase.request.(*tpb.UpdateTimesheetRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_IsEqual(t *testing.T) {
	t.Parallel()
	compareTimesheet1 := &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,

		IsCreated: true,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_1",
				IsCreated:   true,
			},
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_2",
				IsCreated:   true,
			},
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_3",
				IsCreated:   true,
			},
		},
		ListOtherWorkingHours: ListOtherWorkingHours{
			{
				ID:                "other_working_hours_2_1",
				TimesheetID:       "timesheet2_1",
				TimesheetConfigID: "timesheet_config_2_1",
			},
			{
				ID:                "other_working_hours_2_2",
				TimesheetID:       "timesheet2_1",
				TimesheetConfigID: "timesheet_config_2_2",
			},
		},
	}
	compareTimesheet2 := &Timesheet{
		ID:              "timesheet2_1",
		StaffID:         "staff2_1",
		LocationID:      "location2_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListOtherWorkingHours: ListOtherWorkingHours{
			{
				ID:                "other_working_hours_2_1",
				TimesheetID:       "timesheet2_1",
				TimesheetConfigID: "timesheet_config_2_1",
			},
		},
		IsCreated: true,
	}
	compareTimesheet3 := &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		IsCreated:       true,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_3",
				IsCreated:   true,
			},
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_1",
				IsCreated:   true,
			},
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID1_2",
				IsCreated:   true,
			},
		},
		ListOtherWorkingHours: ListOtherWorkingHours{
			{
				ID:                "other_working_hours_2_2",
				TimesheetID:       "timesheet2_1",
				TimesheetConfigID: "timesheet_config_2_2",
			},
			{
				ID:                "other_working_hours_2_1",
				TimesheetID:       "timesheet2_1",
				TimesheetConfigID: "timesheet_config_2_1",
			},
		},
	}
	compareTimesheet4 := &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		IsCreated:       true,
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "IsEqual success with true",
			request:      []*Timesheet{_timesheet1, compareTimesheet4},
			expectedResp: true,
		},
		{
			name:         "IsEqual success with true when different order timesheet lesson hours and other working hours",
			request:      []*Timesheet{compareTimesheet1, compareTimesheet3},
			expectedResp: true,
		},
		{
			name:         "IsEqual success with false",
			request:      []*Timesheet{_timesheet1, _timesheet2},
			expectedResp: false,
		},
		{
			name:         "IsEqual success with false when different listTimesheetLessonHours",
			request:      []*Timesheet{_timesheet1, compareTimesheet1},
			expectedResp: false,
		},
		{
			name:         "IsEqual success with false when different listOtherWorkingHours",
			request:      []*Timesheet{_timesheet2, compareTimesheet2},
			expectedResp: false,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.([]*Timesheet)[0].IsEqual(testcase.request.([]*Timesheet)[1])
			assert.Equal(t, testcase.expectedResp.(bool), resp)
		})
	}
}

func TestTimesheet_Merge(t *testing.T) {
	t.Parallel()
	mergeTimesheet1 := &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "",
				LessonID:    "lessonID_1",
				IsCreated:   true,
			},
		},
		IsCreated: true,
	}
	mergeTimesheet2 := &Timesheet{
		ID:              "timesheet2_2",
		StaffID:         "staff2_1",
		LocationID:      "location2_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "",
				LessonID:    "lessonID_1",
				IsCreated:   true,
			},
		},
		IsCreated: true,
	}
	mergeTimesheet3 := &Timesheet{
		ID:              "timesheet3_3_1",
		StaffID:         "staff3_3_1",
		LocationID:      "location3_3_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "",
				LessonID:    "lessonID_1",
				IsCreated:   true,
			},
		},
		IsCreated: true,
	}
	mergeTimesheet4 := &Timesheet{
		ID:              "",
		StaffID:         "staff3_1",
		LocationID:      "location3_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "",
				LessonID:    "lessonID_1",
			},
			{
				TimesheetID: "",
				LessonID:    "lessonID_2",
			},
		},
	}
	expectMergedTimesheet1 := &Timesheet{
		ID:              "timesheet1_1",
		StaffID:         "staff1_1",
		LocationID:      "location1_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "timesheet1_1",
				LessonID:    "lessonID_1",
				IsCreated:   true,
			},
		},
		IsCreated: true,
	}
	expectMergedTimesheet2 := &Timesheet{
		ID:              "timesheet3_1",
		StaffID:         "staff3_1",
		LocationID:      "location3_1",
		TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:   now,
		ListTimesheetLessonHours: ListTimesheetLessonHours{
			{
				TimesheetID: "timesheet3_1",
				LessonID:    "lessonID_1",
				IsCreated:   false,
			},
			{
				TimesheetID: "timesheet3_1",
				LessonID:    "lessonID_2",
				IsCreated:   false,
			},
		},
		IsCreated: true,
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
		expectedErr  interface{}
	}{
		{
			name:         "merge success",
			request:      []*Timesheet{_timesheet1, mergeTimesheet1},
			expectedResp: expectMergedTimesheet1,
			expectedErr:  nil,
		},
		{
			name:         "merge success with one timesheet empty timesheet ID",
			request:      []*Timesheet{_timesheet3, mergeTimesheet4},
			expectedResp: expectMergedTimesheet2,
			expectedErr:  nil,
		},
		{
			name:         "merge failed with different timesheet ID",
			request:      []*Timesheet{_timesheet2, mergeTimesheet2},
			expectedResp: (*Timesheet)(nil),
			expectedErr:  fmt.Errorf("validateMerge failed"),
		},
		{
			name:         "merge failed with different timesheet general info",
			request:      []*Timesheet{_timesheet3, mergeTimesheet3},
			expectedResp: (*Timesheet)(nil),
			expectedErr:  fmt.Errorf("validateMerge failed"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp, success := testcase.request.([]*Timesheet)[0].Merge(testcase.request.([]*Timesheet)[1])
			assert.Equal(t, testcase.expectedErr, success)
			assert.True(t, testcase.expectedResp.(*Timesheet).IsEqual(resp))
		})
	}
}

func TestTimesheet_DeleteTimesheetLessonHours(t *testing.T) {
	t.Parallel()
	var (
		lessonID  = "lessonID1_1"
		timesheet = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,

			IsCreated: true,
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_1",
					IsCreated:   true,
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_2",
					IsCreated:   true,
				},
			},
		}
		timesheetSuccess = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,

			IsCreated: true,
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_1",
					IsCreated:   true,
					IsDeleted:   true,
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_2",
					IsCreated:   true,
				},
			},
		}
		timesheetDeletedFail = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			IsCreated:       true,
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_1",
					IsCreated:   true,
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_2",
					IsCreated:   true,
				},
			},
		}
	)

	testCases := []struct {
		name              string
		request           interface{}
		input             interface{}
		expectedResp      interface{}
		expectedTimesheet interface{}
	}{
		{
			name:              "delete timesheet lesson hours success",
			request:           timesheet,
			input:             lessonID,
			expectedResp:      true,
			expectedTimesheet: timesheetSuccess,
		},
		{
			name:              "delete timesheet lesson hours failed not contains lessonID",
			request:           timesheet,
			input:             "not_contains_lessonID",
			expectedResp:      false,
			expectedTimesheet: timesheetDeletedFail,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			res := testcase.request.(*Timesheet).DeleteTimesheetLessonHours(testcase.input.(string))
			assert.Equal(t, testcase.expectedResp, res)
			assert.True(t, testcase.request.(*Timesheet).IsEqual(testcase.expectedTimesheet.(*Timesheet)))
		})
	}
}

func TestTimesheet_NormalizedData(t *testing.T) {
	var (
		timesheet = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),

			IsCreated: true,
		}
		timesheetExpect = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 0, 0, 0, 0, timeutil.Timezone(pbc.COUNTRY_JP)),

			IsCreated: true,
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "normalized data success",
			request:      timesheet,
			expectedResp: timesheetExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.request.(*Timesheet).NormalizedData()
			assert.Equal(t, testcase.expectedResp, testcase.request.(*Timesheet))
		})
	}
}

func TestTimesheet_ToEntity(t *testing.T) {
	var (
		timesheet = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
		}

		timesheetExpect = &entity.Timesheet{
			TimesheetID:     database.Text("timesheet1_1"),
			StaffID:         database.Text("staff1_1"),
			LocationID:      database.Text("location1_1"),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
			TimesheetDate:   database.Timestamptz(time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP))),
			Remark:          database.Text("test remark"),
			CreatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
		}

		timesheetEmptyRemark = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "",
		}

		timesheetEmptyRemarkExpect = &entity.Timesheet{
			TimesheetID:     database.Text("timesheet1_1"),
			StaffID:         database.Text("staff1_1"),
			LocationID:      database.Text("location1_1"),
			TimesheetStatus: database.Text(tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()),
			TimesheetDate:   database.Timestamptz(time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP))),
			Remark:          pgtype.Text{Status: pgtype.Null},
			CreatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:       pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:       pgtype.Timestamptz{Status: pgtype.Null},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entity success",
			request:      timesheet,
			expectedResp: timesheetExpect,
		},

		{
			name:         "convert to timesheet entity remark success",
			request:      timesheetEmptyRemark,
			expectedResp: timesheetEmptyRemarkExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*Timesheet).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_MakeNewID(t *testing.T) {
	var (
		timesheet = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					LessonID: "1",
				},
			},
			ListOtherWorkingHours: ListOtherWorkingHours{
				{
					ID: "1",
				},
			},
		}
	)

	testCases := []struct {
		name     string
		request  interface{}
		validate func(timesheet *Timesheet) bool
	}{
		{
			name:    "make new id success",
			request: timesheet,
			validate: func(timesheet *Timesheet) bool {
				return timesheet.ID != "timesheetId_1"
			},
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.request.(*Timesheet).MakeNewID()
			assert.Equal(t, testcase.validate(testcase.request.(*Timesheet)), true)
		})
	}
}

func TestTimesheet_GetTimesheetLessonHoursNew(t *testing.T) {
	var (
		timesheet = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_1",
					IsCreated:   false,
				},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:    "generate timesheet lesson hours new success",
			request: timesheet,
			expectedResp: []*TimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lessonID1_1",
					IsCreated:   false,
				},
			},
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*Timesheet).GetTimesheetLessonHoursNew()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_ValidateCreateInfo(t *testing.T) {
	t.Parallel()
	longThan500CharacterRemark := "宇利波米婆 胡藤母意母保由 久利波米婆 麻斯提斯農波由 伊豆久欲曲 【キョク】 composition, piece of music, song, track (on a record), tune, melody, air, enjoyment, fun, interest, pleasure 曲折 【キョクセツ】 bending, winding, meandering, zigzagging, ups and downs, twists and turns, complications, difficulties, vicissitudes 歌謡曲 【カヨウキョク】 kayōkyoku, form of Japanese popular music that developed during the Show宇利波米婆 胡藤母意母保由 久利波米婆 麻斯提斯農波由 伊豆久欲曲 【キョク】 composition, piece of music, song, track (on a record), tune, melody, air, enjoyment, fun, intere2022"

	var (
		otherWorkingHours1 = OtherWorkingHours{
			ID:                "other_working_hour_1",
			TimesheetID:       "timesheet_1",
			TimesheetConfigID: "timesheet_config_1",
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour * 1), //add 1 hour
			TotalHour:         1,
			Remarks:           "other working hour remark",
		}

		transportationExpense = TransportationExpenses{
			TransportExpenseID: "transport_expense_1",
			TimesheetID:        "timesheet_1",
			TransportationType: tpb.TransportationType_TYPE_BUS.String(),
			Remarks:            "transportation remark",
		}

		timesheetEmptyStaffID = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "",
			LocationID:      "location_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
		}

		timesheetEmptyLocationID = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
		}

		timesheetEmptyTimesheetDate = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
		}

		timesheetWithDateSmaller = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2021, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
		}

		timesheetWithRemarkIsTooLong = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          longThan500CharacterRemark,
		}

		timesheetWithEmptyOWHs = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
		}

		timesheetWithMoreThan5OWHs = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
			},
		}

		timesheetWithMoreThan10TransportExpense = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				&otherWorkingHours1,
			},
			ListTransportationExpenses: ListTransportationExpenses{
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "fail with empty staff id",
			request:      timesheetEmptyStaffID,
			expectedResp: fmt.Errorf("staff id must not be empty"),
		},
		{
			name:         "fail with empty location id",
			request:      timesheetEmptyLocationID,
			expectedResp: fmt.Errorf("location id must not be empty"),
		},
		{
			name:         "fail with empty timesheet date",
			request:      timesheetEmptyTimesheetDate,
			expectedResp: fmt.Errorf("date must not be nil"),
		},
		{
			name:         "fail with timesheet date invalid",
			request:      timesheetWithDateSmaller,
			expectedResp: fmt.Errorf("date must be greater than 1st Jan 2022"),
		},
		{
			name:         "fail with timesheet remark is too long",
			request:      timesheetWithRemarkIsTooLong,
			expectedResp: fmt.Errorf("remark must be limit to 500 characters"),
		},
		{
			name:         "fail with timesheet not has other working hours",
			request:      timesheetWithEmptyOWHs,
			expectedResp: fmt.Errorf("other working hours must be not empty"),
		},
		{
			name:         "fail with timesheet has more than 5 other working hours",
			request:      timesheetWithMoreThan5OWHs,
			expectedResp: fmt.Errorf("list other working hours must be limit to 5 rows"),
		},

		{
			name:         "fail with timesheet has more than 10 transportation expense record",
			request:      timesheetWithMoreThan10TransportExpense,
			expectedResp: fmt.Errorf("list transportation expenses must be limit to 10 rows"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*Timesheet).ValidateCreateInfo()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_ValidateUpdateInfo(t *testing.T) {
	t.Parallel()
	longThan500CharacterRemark := "宇利波米婆 胡藤母意母保由 久利波米婆 麻斯提斯農波由 伊豆久欲曲 【キョク】 composition, piece of music, song, track (on a record), tune, melody, air, enjoyment, fun, interest, pleasure 曲折 【キョクセツ】 bending, winding, meandering, zigzagging, ups and downs, twists and turns, complications, difficulties, vicissitudes 歌謡曲 【カヨウキョク】 kayōkyoku, form of Japanese popular music that developed during the Show宇利波米婆 胡藤母意母保由 久利波米婆 麻斯提斯農波由 伊豆久欲曲 【キョク】 composition, piece of music, song, track (on a record), tune, melody, air, enjoyment, fun, intere2022"

	var (
		otherWorkingHours1 = OtherWorkingHours{
			ID:                "other_working_hour_1",
			TimesheetID:       "timesheet_1",
			TimesheetConfigID: "timesheet_config_1",
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(time.Hour * 1), //add 1 hour
			TotalHour:         1,
			Remarks:           "other working hour remark",
		}

		otherWorkingHoursWithTimeInvalid = OtherWorkingHours{
			ID:                "other_working_hour_1",
			TimesheetID:       "timesheet_1",
			TimesheetConfigID: "timesheet_config_1",
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(-24 * time.Hour), //yesterday
			TotalHour:         1,
			Remarks:           "other working hour remark",
		}

		transportationExpense = TransportationExpenses{
			TransportExpenseID: "transport_expense_1",
			TimesheetID:        "timesheet_1",
			TransportationType: tpb.TransportationType_TYPE_BUS.String(),
			Remarks:            "transportation remark",
		}

		timesheetEmptyID = &Timesheet{
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
		}

		timesheetWithRemarkIsTooLong = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          longThan500CharacterRemark,
		}

		timesheetWithOWHInvalid = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				&otherWorkingHoursWithTimeInvalid,
			},
		}

		timesheetWithMoreThan5OWHs = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
				&otherWorkingHours1,
			},
		}

		timesheetWithMoreThan10TransportExpense = &Timesheet{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location_1",
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				&otherWorkingHours1,
			},
			ListTransportationExpenses: ListTransportationExpenses{
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
				&transportationExpense,
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "fail with empty id",
			request:      timesheetEmptyID,
			expectedResp: fmt.Errorf("timesheet id must not be empty"),
		},
		{
			name:         "fail with empty invalid other working hour",
			request:      timesheetWithOWHInvalid,
			expectedResp: fmt.Errorf("other working hours end time must after start time"),
		},
		{
			name:         "fail with timesheet remark is too long",
			request:      timesheetWithRemarkIsTooLong,
			expectedResp: fmt.Errorf("remark must be limit to 500 characters"),
		},
		{
			name:         "fail with timesheet has more than 5 other working hours",
			request:      timesheetWithMoreThan5OWHs,
			expectedResp: fmt.Errorf("list other working hours must be limit to 5 rows"),
		},

		{
			name:         "fail with timesheet has more than 10 transportation expense record",
			request:      timesheetWithMoreThan10TransportExpense,
			expectedResp: fmt.Errorf("list transportation expenses must be limit to 10 rows"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*Timesheet).ValidateUpdateInfo()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheet_IsTimesheetEmpty(t *testing.T) {
	var (
		timesheet = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
		}
		timesheetHasOtherWorkingHours = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
			ListOtherWorkingHours: ListOtherWorkingHours{
				{
					IsDeleted: false,
				},
			},
		}
		timesheetHasTimesheetLessonHours = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
			ListTimesheetLessonHours: ListTimesheetLessonHours{
				{
					LessonID:  "1",
					IsDeleted: false,
				},
			},
		}
		timesheetHasTransportationExpenses = &Timesheet{
			ID:              "timesheetId_1",
			StaffID:         "staffId_1",
			LocationID:      "locationId_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   time.Date(2022, 01, 01, 15, 30, 30, 30, timeutil.Timezone(pbc.COUNTRY_JP)),
			Remark:          "test remark",
			ListTransportationExpenses: ListTransportationExpenses{
				{
					IsDeleted: false,
				},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "not empty timesheet",
			request:      timesheet,
			expectedResp: true,
		},
		{
			name:         "timesheet has other working hours",
			request:      timesheetHasOtherWorkingHours,
			expectedResp: false,
		},
		{
			name:         "timesheet has lesson hours",
			request:      timesheetHasTimesheetLessonHours,
			expectedResp: false,
		},
		{
			name:         "timesheet has transportation expenses",
			request:      timesheetHasTransportationExpenses,
			expectedResp: false,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*Timesheet).IsTimesheetEmpty()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
