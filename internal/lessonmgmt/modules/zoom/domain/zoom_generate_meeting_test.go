package domain

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	"github.com/nleeper/goment"
	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGenerateMeeting_ConverterZoomGenerateMeetingRequest(t *testing.T) {
	t.Parallel()
	startTime := time.Now()
	endTime := time.Now()
	timeZone := "Asia/Tokyo"
	duration := endTime.Sub(startTime)
	durationAddedOneHour := int(duration.Minutes())
	startTimestamp := &timestamppb.Timestamp{Seconds: startTime.Unix()}
	endTimestamp := &timestamppb.Timestamp{Seconds: endTime.Unix()}
	endDateTimestamp := &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	endDate, _ := ConvertDatetoCountryTZ(endDateTimestamp.AsTime(), timeZone)

	gEndDate, _ := goment.New(endDate)
	eDate := gEndDate.EndOf("d")
	tEndDate := eDate.ToTime()
	type TestCase struct {
		name         string
		req          *lpb.GenerateZoomLinkRequest
		expectedResp *ZoomGenerateMeetingRequest
		expectedErr  error
		setup        func(ctx context.Context)
	}
	dayOfWeek, _ := ConvertWeekDayGoToWeekDayZoom(startTimestamp.AsTime(), "")
	tc := []TestCase{
		{
			name: "should convert zoom link success for one time",
			req: &lpb.GenerateZoomLinkRequest{
				AccountOwner: "abc@gmail.com",
				Method:       lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
				StartTime:    startTimestamp,
				EndTime:      endTimestamp,
				TimeZone:     timeZone,
			},
			expectedErr: nil,
			expectedResp: &ZoomGenerateMeetingRequest{
				TypeZoom:  InstantMeeting,
				StartTime: support.ConvertToStringRFCFormat(startTimestamp.AsTime()),
				Duration:  durationAddedOneHour,
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "should convert zoom link success for recurrence time",
			req: &lpb.GenerateZoomLinkRequest{
				AccountOwner: "abc@gmail.com",
				Method:       lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
				StartTime:    startTimestamp,
				EndTime:      endTimestamp,
				TimeZone:     timeZone,
				EndDate:      endDateTimestamp,
			},
			expectedErr: nil,
			expectedResp: &ZoomGenerateMeetingRequest{
				TypeZoom:  RecurringFixedMeeting,
				StartTime: support.ConvertToStringRFCFormat(startTimestamp.AsTime()),
				Duration:  durationAddedOneHour,
				RecurrenceSetting: &RecurrenceSetting{
					EndDateTime: support.ConvertToStringRFCFormat(tEndDate),
					Type:        Weekly,
					WeeklyDays:  dayOfWeek,
				},
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {

			resp, err := ConverterZoomGenerateMeetingRequest(testCase.req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestGenerateMeeting_ConverterMultiZoomGenerateMeetingRequest(t *testing.T) {
	t.Parallel()
	startTime := time.Date(2022, time.Month(2), 11, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2022, time.Month(2), 11, 1, 0, 0, 0, time.UTC)

	endDateTime := time.Date(2022, time.Month(2), 11, 0, 0, 0, 0, time.UTC).AddDate(0, 4, 0)
	endDateTime2 := time.Date(2022, time.Month(2), 11, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	endDateTime3 := time.Date(2022, time.Month(2), 11, 0, 0, 0, 0, time.UTC).AddDate(0, 16, 0)

	type TestCase struct {
		name         string
		req          *MultiZoomGenerateMeetingRequest
		expectedResp []*ZoomGenerateMeetingRequest
		expectedErr  error
		setup        func(ctx context.Context)
	}

	tc := []TestCase{
		{
			name: "should convert multi zoom link success - return one",
			req: &MultiZoomGenerateMeetingRequest{
				StartTime:   startTime,
				EndTime:     endTime,
				EndDateTime: endDateTime,
				TimeZone:    "VN",
			},
			expectedErr: nil,
			expectedResp: []*ZoomGenerateMeetingRequest{
				{
					Topic:     "",
					TypeZoom:  8,
					StartTime: support.ConvertToStringRFCFormat(startTime.Round(0)),
					Duration:  60,
					Settings:  nil,
					RecurrenceSetting: &RecurrenceSetting{
						EndDateTime: "2022-06-11T16:59:59Z",
						Type:        2,
						WeeklyDays:  6,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "should convert multi zoom link success",
			req: &MultiZoomGenerateMeetingRequest{
				StartTime:   startTime,
				EndTime:     endTime,
				EndDateTime: endDateTime2,
				TimeZone:    "VN",
			},
			expectedErr: nil,
			expectedResp: []*ZoomGenerateMeetingRequest{
				{
					Topic:     "",
					TypeZoom:  8,
					StartTime: support.ConvertToStringRFCFormat(startTime.Round(0)),
					Duration:  60,
					Settings:  nil,
					RecurrenceSetting: &RecurrenceSetting{
						EndDateTime: "2022-03-11T16:59:59Z",
						Type:        2,
						WeeklyDays:  6,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "should convert multi zoom link success - return multi",
			req: &MultiZoomGenerateMeetingRequest{
				StartTime:   startTime,
				EndTime:     endTime,
				EndDateTime: endDateTime3,
				TimeZone:    "VN",
			},
			expectedErr: nil,
			expectedResp: []*ZoomGenerateMeetingRequest{
				{
					Topic:     "",
					TypeZoom:  8,
					StartTime: support.ConvertToStringRFCFormat(startTime.Round(0)),
					Duration:  60,
					Settings:  nil,
					RecurrenceSetting: &RecurrenceSetting{
						EndDateTime: "2023-06-11T16:59:59Z",
						Type:        2,
						WeeklyDays:  6,
					},
				},
				{
					Topic:     "",
					TypeZoom:  8,
					StartTime: support.ConvertToStringRFCFormat(startTime.AddDate(0, 0, 7*60).Round(0)),
					Duration:  60,
					Settings:  nil,
					RecurrenceSetting: &RecurrenceSetting{
						EndDateTime: "2023-06-11T16:59:59Z",
						Type:        2,
						WeeklyDays:  6,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {

			resp, err := ConverterMultiZoomGenerateMeetingRequest(testCase.req)
			if testCase.expectedErr != nil {

				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
