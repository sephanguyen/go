package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type TimesheetCountReq struct {
	StaffName  string
	LocationID string
	StaffID    string
	FromDate   time.Time
	ToDate     time.Time
}

type TimesheetCountOut struct {
	AllCount       int64
	DraftCount     int64
	SubmittedCount int64
	ApprovedCount  int64
	ConfirmedCount int64
}

func NewTimesheetCountReqFromRPCCreateRequest(req *pb.CountTimesheetsRequest) *TimesheetCountReq {
	return &TimesheetCountReq{
		StaffName:  req.GetStaffName(),
		LocationID: req.GetLocationId(),
		StaffID:    req.GetStaffId(),
		FromDate:   req.FromDate.AsTime(),
		ToDate:     req.ToDate.AsTime(),
	}
}

func (t *TimesheetCountReq) ConvertTimeToJPTimezone() {
	t.FromDate = t.FromDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
	t.ToDate = t.ToDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
}

func (t *TimesheetCountOut) FieldMap() ([]string, []interface{}) {
	return []string{
			"all_count", "draft_count", "submitted_count", "approved_count", "confirmed_count",
		}, []interface{}{
			&t.AllCount, &t.DraftCount, &t.SubmittedCount, &t.ApprovedCount, &t.ConfirmedCount,
		}
}

func (t *TimesheetCountOut) SQLFunctionName() string {
	return "get_timesheet_count"
}
