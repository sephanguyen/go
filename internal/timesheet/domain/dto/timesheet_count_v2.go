package dto

import (
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type TimesheetCountV2Req struct {
	StaffName   string
	LocationIds []string
	StaffID     string
	FromDate    time.Time
	ToDate      time.Time
}

type TimesheetCountV2Out struct {
	AllCount       int64
	DraftCount     int64
	SubmittedCount int64
	ApprovedCount  int64
	ConfirmedCount int64
}

func NewTimesheetCountReqFromRPCCreateV2Request(req *pb.CountTimesheetsV2Request) *TimesheetCountV2Req {
	return &TimesheetCountV2Req{
		StaffName:   req.GetStaffName(),
		LocationIds: req.GetLocationIds(),
		StaffID:     req.GetStaffId(),
		FromDate:    req.FromDate.AsTime(),
		ToDate:      req.ToDate.AsTime(),
	}
}

func (t *TimesheetCountV2Out) FieldMap() ([]string, []interface{}) {
	return []string{
			"all_count", "draft_count", "submitted_count", "approved_count", "confirmed_count",
		}, []interface{}{
			&t.AllCount, &t.DraftCount, &t.SubmittedCount, &t.ApprovedCount, &t.ConfirmedCount,
		}
}

func (t *TimesheetCountV2Out) SQLFunctionName() string {
	return "get_timesheet_count_v2"
}
