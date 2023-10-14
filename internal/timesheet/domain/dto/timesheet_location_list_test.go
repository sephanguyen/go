package dto

import (
	"testing"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewGetTimesheetLocationListRequest(t *testing.T) {
	t.Parallel()
	var (
		FromDate = time.Now().UTC()
		ToDate   = FromDate.Add(time.Hour * 1)
		Keyword  = "location-test"
		Limit    = int32(10)
		Offset   = int32(0)
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "new get timesheet location list from rpc success",
			request: &pb.GetTimesheetLocationListRequest{
				FromDate: timestamppb.New(FromDate),
				ToDate:   timestamppb.New(ToDate),
				Keyword:  "location-test",
				Limit:    10,
				Offset:   0,
			},
			expectedResp: &GetTimesheetLocationListReq{
				FromDate: FromDate,
				ToDate:   ToDate,
				Keyword:  Keyword,
				Limit:    Limit,
				Offset:   Offset,
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewGetTimesheetLocationListRequest(testcase.request.(*pb.GetTimesheetLocationListRequest))
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}

func TestNewGetNonConfirmedLocationCountRequest(t *testing.T) {
	t.Parallel()
	var (
		PeriodDate = time.Now().UTC()
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "new get non confirmed location count from rpc success",
			request: &pb.GetNonConfirmedLocationCountRequest{
				PeriodDate: timestamppb.New(PeriodDate),
			},
			expectedResp: &GetNonConfirmedLocationCountReq{
				PeriodDate: PeriodDate,
			},
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewGetNonConfirmedLocationCountRequest(testcase.request.(*pb.GetNonConfirmedLocationCountRequest))
			assert.Equal(t, resp, testcase.expectedResp)
		})
	}
}
