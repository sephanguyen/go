package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type GetTimesheetLocationListReq struct {
	FromDate time.Time
	ToDate   time.Time
	Keyword  string
	Limit    int32
	Offset   int32
}

type TimesheetLocation struct {
	LocationID       string
	Name             string
	IsConfirmed      bool
	DraftCount       int32
	SubmittedCount   int32
	ApprovedCount    int32
	ConfirmedCount   int32
	UnconfirmedCount int32
}

type TimesheetLocationAggregate struct {
	Count int32
}

type GetTimesheetLocationListOut struct {
	Locations         []*TimesheetLocation
	LocationAggregate *TimesheetLocationAggregate
}

type GetNonConfirmedLocationCountReq struct {
	PeriodDate time.Time
}

type GetNonConfirmedLocationCountOut struct {
	NonconfirmedCount int32
}

func NewGetTimesheetLocationListRequest(req *pb.GetTimesheetLocationListRequest) *GetTimesheetLocationListReq {
	return &GetTimesheetLocationListReq{
		FromDate: req.FromDate.AsTime(),
		ToDate:   req.ToDate.AsTime(),
		Keyword:  req.GetKeyword(),
		Limit:    req.GetLimit(),
		Offset:   req.GetOffset(),
	}
}

func NewGetNonConfirmedLocationCountRequest(req *pb.GetNonConfirmedLocationCountRequest) *GetNonConfirmedLocationCountReq {
	return &GetNonConfirmedLocationCountReq{
		PeriodDate: req.PeriodDate.AsTime(),
	}
}

func (t *GetTimesheetLocationListReq) ConvertTimeToJPTimezone() {
	t.FromDate = t.FromDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
	t.ToDate = t.ToDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
}

func (t *GetNonConfirmedLocationCountReq) ConvertTimeToJPTimezone() {
	t.PeriodDate = t.PeriodDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
}

func ConvertTimesheetLocationListToRPC(timesheetLocations []*TimesheetLocation) []*pb.TimesheetLocation {
	result := []*pb.TimesheetLocation{}
	for _, loc := range timesheetLocations {
		timesheetLocation := pb.TimesheetLocation{
			LocationId:       loc.LocationID,
			Name:             loc.Name,
			IsConfirmed:      loc.IsConfirmed,
			DraftCount:       loc.DraftCount,
			SubmittedCount:   loc.SubmittedCount,
			ApprovedCount:    loc.ApprovedCount,
			ConfirmedCount:   loc.ConfirmedCount,
			UnconfirmedCount: loc.UnconfirmedCount,
		}
		result = append(result, &timesheetLocation)
	}
	return result
}

func ConvertTimesheetLocationAggregateToRPC(timesheetLocationAggregate *TimesheetLocationAggregate) *pb.TimesheetLocationAggregate {
	return &pb.TimesheetLocationAggregate{
		Count: timesheetLocationAggregate.Count,
	}
}

func ConvertGetNonConfirmedLocationCountOutToRPC(getNonConfirmedLocationCountOut *GetNonConfirmedLocationCountOut) *pb.GetNonConfirmedLocationCountResponse {
	return &pb.GetNonConfirmedLocationCountResponse{
		NonConfirmedLocationCount: getNonConfirmedLocationCountOut.NonconfirmedCount,
	}
}
