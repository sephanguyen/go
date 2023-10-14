package domain

import (
	"fmt"
	"math"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/imdario/mergo"
	"github.com/nleeper/goment"
)

type TypeZoom = int

const (
	InstantMeeting        TypeZoom = iota + 2
	RecurringFixedMeeting          = 8
)

type TypeRecurrence = int

const (
	Daily TypeRecurrence = iota + 1
	Weekly
	Monthly
)

type ZoomUserStatus = string

const (
	ZoomUserStatusActive   ZoomUserStatus = "active"
	ZoomUserStatusInActive ZoomUserStatus = "inactive"
	ZoomUserStatusPending  ZoomUserStatus = "pending"
)

const oneWeek = float64(24 * 7)
const MaximumRepeatInterval = float64(60)
const DaysOfWeek = 7

type ZoomGenerateMeetingResponse struct {
	UUID        string                      `json:"uuid"`
	ID          int                         `json:"id"`
	URL         string                      `json:"join_url"`
	Code        int                         `json:"code"`
	Message     string                      `json:"message"`
	Occurrences []*OccurrenceOfZoomResponse `json:"occurrences"`
}
type OccurrenceOfZoomResponse struct {
	OccurrenceID string `json:"occurrence_id"`
	StartTime    string `json:"start_time"`
	Duration     int    `json:"duration"`
	Status       string `json:"status"`
}
type ZoomGenerateMeetingRequest struct {
	Topic     string   `json:"topic"`
	TypeZoom  TypeZoom `json:"type"`
	StartTime string   `json:"start_time"`
	Duration  int      `json:"duration"`
	//	agenda //what mean
	Settings          *ZoomMeetingSetting `json:"settings"`
	RecurrenceSetting *RecurrenceSetting  `json:"recurrence"`
}

type ZoomGetListUserRequest struct {
	PageNumber int
	PageSize   int
}

type GenerateZoomLinkResponse struct {
	ZoomID      int
	URL         string
	Occurrences []*OccurrenceOfZoomResponse
}

type ZoomUserInfo struct {
	ID        string         `json:"id"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Email     string         `json:"email"`
	Status    ZoomUserStatus `json:"status"`
}

type UserZoomResponse struct {
	PageCount   int            `json:"page_count"`
	PageNumber  int            `json:"page_number"`
	PageSize    int            `json:"page_size"`
	TotalRecord int            `json:"total_records"`
	Users       []ZoomUserInfo `json:"users"`
	Code        int            `json:"code"`
	Message     string         `json:"message"`
}

type DeleteZoomResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RecurrenceSetting struct {
	EndDateTime string         `json:"end_date_time"`
	Type        TypeRecurrence `json:"type"`
	WeeklyDays  int            `json:"weekly_days"`
}

type ZoomMeetingSetting struct {
	HostVideo        bool `json:"host_video"`
	ParticipantVideo bool `json:"participant_video"`
}

type ZoomParamConverter interface {
	validation(c *lpb.GenerateZoomLinkRequest) error
	getParameter(c *lpb.GenerateZoomLinkRequest) *ZoomGenerateMeetingRequest
}

type ZoomParameter struct {
	Converter ZoomParamConverter
	Req       *lpb.GenerateZoomLinkRequest
}

func (z *ZoomParameter) validation() error {
	startTime := z.Req.StartTime
	if startTime.AsTime().IsZero() {
		return fmt.Errorf("startTime should not be empty")
	}
	endTime := z.Req.EndTime
	if endTime.AsTime().IsZero() {
		return fmt.Errorf("endTime should not be empty")
	}
	timeZone := z.Req.TimeZone
	if timeZone == "" {
		return fmt.Errorf("timezone should not be empty")
	}
	return z.Converter.validation(z.Req)
}

func (z *ZoomParameter) getParameter() (*ZoomGenerateMeetingRequest, error) {
	startTime := z.Req.GetStartTime().AsTime()
	endTime := z.Req.GetEndTime().AsTime()
	duration := endTime.Sub(startTime)

	commonParam := &ZoomGenerateMeetingRequest{
		StartTime: support.ConvertToStringRFCFormat(startTime),
		Duration:  int(duration.Minutes()),
	}

	otherParam := z.Converter.getParameter(z.Req)

	if err := mergo.Merge(otherParam, *commonParam); err != nil {
		return nil, err
	}
	return otherParam, nil
}

type OneTimeConverter struct {
}

func (z *OneTimeConverter) validation(c *lpb.GenerateZoomLinkRequest) error {
	return nil
}

func (z *OneTimeConverter) getParameter(c *lpb.GenerateZoomLinkRequest) *ZoomGenerateMeetingRequest {
	return &ZoomGenerateMeetingRequest{
		TypeZoom: InstantMeeting,
	}
}

type RecurrenceConverter struct {
}

func GetMaxDuration(durationRecurrence float64) int {
	// duration in [1, 12]
	return int(math.Max(1, math.Min(durationRecurrence, MaximumRepeatInterval)))
}

func GetDurationByWeek(startTime time.Time, endTime time.Time) float64 {
	return endTime.Sub(startTime).Hours() / oneWeek
}

func (z *RecurrenceConverter) validation(req *lpb.GenerateZoomLinkRequest) error {
	endDate := req.GetEndDate().AsTime()
	if endDate.IsZero() {
		return fmt.Errorf("endDate should not be empty")
	}
	return nil
}

const (
	ContryJP  = "COUNTRY_JP"
	CountryVN = "COUNTRY_VN"
)

var countryTZMap = map[string]string{
	ContryJP:  "Asia/Tokyo",
	CountryVN: "Asia/Ho_Chi_Minh",
}

func ConvertDatetoCountryTZ(date time.Time, country string) (time.Time, error) {
	timezone, ok := countryTZMap[country]
	if !ok {
		timezone = countryTZMap[CountryVN]
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return date, err
	}

	return date.In(location), nil
}

func ConvertWeekDayGoToWeekDayZoom(t time.Time, timeZone string) (int, error) {
	date, err := ConvertDatetoCountryTZ(t, timeZone)
	if err != nil {
		return -1, err
	}
	return int(date.Weekday()) + 1, nil
}

func (z *RecurrenceConverter) getParameter(req *lpb.GenerateZoomLinkRequest) *ZoomGenerateMeetingRequest {
	dayOfWeek, _ := ConvertWeekDayGoToWeekDayZoom(req.GetStartTime().AsTime(), req.TimeZone)
	endDate, _ := ConvertDatetoCountryTZ(req.GetEndDate().AsTime(), req.TimeZone)
	gEndDate, _ := goment.New(endDate)
	eDate := gEndDate.EndOf("d")
	tEndDate := eDate.ToTime()
	return &ZoomGenerateMeetingRequest{
		TypeZoom: RecurringFixedMeeting,
		RecurrenceSetting: &RecurrenceSetting{
			EndDateTime: support.ConvertToStringRFCFormat(tEndDate),
			Type:        Weekly,
			WeeklyDays:  dayOfWeek,
		},
	}
}

func ConverterZoomGenerateMeetingRequest(zoomLinkReq *lpb.GenerateZoomLinkRequest) (*ZoomGenerateMeetingRequest, error) {
	method := zoomLinkReq.GetMethod()
	zoomParameter := &ZoomParameter{
		Req: zoomLinkReq,
	}
	switch method {
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME:
		zoomParameter.Converter = &OneTimeConverter{}
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE:
		zoomParameter.Converter = &RecurrenceConverter{}
	default:
		return nil, fmt.Errorf("not support method")
	}
	if err := zoomParameter.validation(); err != nil {
		return nil, err
	}
	params, err := zoomParameter.getParameter()
	if err != nil {
		return nil, err
	}
	return params, nil
}

type MultiZoomGenerateMeetingRequest struct {
	StartTime   time.Time
	EndTime     time.Time
	EndDateTime time.Time
	TimeZone    string
}

func ConverterMultiZoomGenerateMeetingRequest(zoomLinkReq *MultiZoomGenerateMeetingRequest) ([]*ZoomGenerateMeetingRequest, error) {
	startTime := zoomLinkReq.StartTime
	endTime := zoomLinkReq.EndTime
	duration := endTime.Sub(startTime)
	endDate, _ := ConvertDatetoCountryTZ(zoomLinkReq.EndDateTime, zoomLinkReq.TimeZone)
	gEndDate, _ := goment.New(endDate)
	eDate := gEndDate.EndOf("d")
	tEndDate := eDate.ToTime()
	durationRecurrence := GetDurationByWeek(startTime, endDate)
	weeklyDay, _ := ConvertWeekDayGoToWeekDayZoom(startTime, zoomLinkReq.TimeZone)

	maximumRepeatInterval := int(MaximumRepeatInterval)
	totalLinkShouldBeGen := 1
	if durationRecurrence > MaximumRepeatInterval {
		totalLinkShouldBeGen = int(math.Ceil(durationRecurrence / float64(maximumRepeatInterval)))
	}
	requests := make([]*ZoomGenerateMeetingRequest, 0, totalLinkShouldBeGen)
	durationMeeting := int(duration.Minutes())
	for i := 0; i < totalLinkShouldBeGen; i++ {
		nextStartTime := startTime.AddDate(0, 0, i*7*maximumRepeatInterval)
		requests = append(requests, &ZoomGenerateMeetingRequest{
			StartTime: support.ConvertToStringRFCFormat(nextStartTime),
			Duration:  durationMeeting,
			TypeZoom:  RecurringFixedMeeting,
			RecurrenceSetting: &RecurrenceSetting{
				EndDateTime: support.ConvertToStringRFCFormat(tEndDate),
				Type:        Weekly,
				WeeklyDays:  weeklyDay,
			},
		})
	}
	return requests, nil
}
