package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestAutoCreateFlagActivityLog_ToEntity(t *testing.T) {
	changeTime := time.Now()
	var (
		autoCreateTimesheetFlag = &AutoCreateFlagActivityLog{
			ID:         "id_1",
			StaffID:    "staff1_1",
			ChangeTime: changeTime,
			FlagOn:     true,
		}

		autoCreateTimesheetFlagExpect = &entity.AutoCreateFlagActivityLog{
			ID:         database.Text("id_1"),
			StaffID:    database.Text("staff1_1"),
			FlagOn:     database.Bool(true),
			ChangeTime: pgtype.Timestamptz{Time: changeTime, Status: 0x2},
			CreatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:  pgtype.Timestamptz{Status: pgtype.Null},
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
			resp := testcase.request.(*AutoCreateFlagActivityLog).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewAutoCreateFlagActivityLogFromEntity(t *testing.T) {
	t.Parallel()

	changeTime := time.Now()
	var (
		autoCreateTimesheetFlagExpect = &AutoCreateFlagActivityLog{
			ID:         "id_1",
			StaffID:    "staff1_1",
			ChangeTime: changeTime,
			FlagOn:     true,
		}

		autoCreateTimesheetFlag = &entity.AutoCreateFlagActivityLog{
			ID:         database.Text("id_1"),
			StaffID:    database.Text("staff1_1"),
			FlagOn:     database.Bool(true),
			ChangeTime: pgtype.Timestamptz{Time: changeTime, Status: 0x2},
			CreatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:  pgtype.Timestamptz{Status: pgtype.Null},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new auto create flag from entity success",
			request:      autoCreateTimesheetFlag,
			expectedResp: autoCreateTimesheetFlagExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewAutoCreateFlagActivityLogFromEntity(testcase.request.(*entity.AutoCreateFlagActivityLog))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
