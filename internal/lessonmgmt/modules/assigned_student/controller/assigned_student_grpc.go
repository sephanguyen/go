package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/infrastructure"
	lesson_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UnderAssignedStatus = "Under assigned"
	JustAssignedStatus  = "Just assigned"
	OverAssignedStatus  = "Over assigned"
)

type AssignedStudentGRPCService struct {
	QueryHandler queries.AssignedStudentQueryHandler

	env              string
	unleashClientIns unleashclient.ClientInstance
}

func NewAssignedStudentGRPCService(
	wrapperConnection *support.WrapperDBConnection,
	assignedStudentRepo infrastructure.AssignedStudentRepo,
	reallocationRepo lesson_infras.ReallocationRepo,
	studentSubscriptionRepo user_infras.StudentSubscriptionRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
	academicYearRepo infrastructure.AcademicYearRepo,
) *AssignedStudentGRPCService {
	return &AssignedStudentGRPCService{
		QueryHandler: queries.AssignedStudentQueryHandler{
			WrapperConnection:       wrapperConnection,
			AssignedStudentRepo:     assignedStudentRepo,
			ReallocationRepo:        reallocationRepo,
			StudentSubscriptionRepo: studentSubscriptionRepo,
			AcademicYearRepo:        academicYearRepo,
		},
		env:              env,
		unleashClientIns: unleashClientIns,
	}
}

func (a *AssignedStudentGRPCService) GetAssignedStudentList(ctx context.Context, req *lpb.GetAssignedStudentListRequest) (*lpb.GetAssignedStudentListResponse, error) {
	if err := validateGetAssignedStudentListRequest(req); err != nil {
		return nil, err
	}
	args, isEmptyIntersectedLocation := filterArgsFromRequestPayload(ctx, req)

	if isEmptyIntersectedLocation {
		return &lpb.GetAssignedStudentListResponse{
			Items:      []*lpb.AssignedStudentInfo{},
			TotalItems: uint32(0),
			NextPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
			PreviousPage: &cpb.Paging{
				Limit: req.Paging.Limit,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "",
				},
			},
		}, nil
	}

	result := a.QueryHandler.GetAssignedStudentList(ctx, &args)

	if result.Error != nil {
		return nil, result.Error
	}

	items := []*lpb.AssignedStudentInfo{}
	for _, assignedStudent := range result.AsgStudents {
		items = append(items, toAsgStudentPb(assignedStudent))
	}
	lastItem := ""
	studentLen := len(result.AsgStudents)
	if studentLen > 0 {
		lastItem = result.AsgStudents[studentLen-1].StudentSubscriptionID
		if req.GetPurchaseMethod() == lpb.PurchaseMethod_PURCHASE_METHOD_RECURRING {
			lastItem = result.AsgStudents[studentLen-1].StudentSubscriptionID + "_" + strings.Split(result.AsgStudents[studentLen-1].Duration, " - ")[0]
		}
	}

	return &lpb.GetAssignedStudentListResponse{
		Items:      items,
		TotalItems: result.Total,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lastItem,
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: result.OffsetID,
			},
		},
	}, nil
}

func (a *AssignedStudentGRPCService) GetStudentAttendance(ctx context.Context, req *lpb.GetStudentAttendanceRequest) (*lpb.GetStudentAttendanceResponse, error) {
	limit := req.Paging.GetLimit()
	offset := req.Paging.GetOffsetInteger()
	timezone := req.GetTimezone()
	if len(timezone) == 0 {
		timezone = "Asia/Ho_Chi_Minh"
	}
	loc, _ := time.LoadLocation(timezone)
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	attendanceStatuses := make([]string, 0, len(req.Filter.GetAttendanceStatus()))
	for _, value := range req.Filter.GetAttendanceStatus() {
		attendanceStatuses = append(attendanceStatuses, value.String())
	}

	isFilterByCurrentYear, err := a.unleashClientIns.IsFeatureEnabledOnOrganization("Lesson_LessonManagement_FilterAttendanceListByCurrentYear", a.env, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("a.connectToUnleash: %w", err)
	}

	resp, err := a.QueryHandler.GetStudentAttendance(ctx, &queries.GetStudentAttendanceRequest{
		SearchKey: req.GetSearchKey(),
		Timezone:  timezone,
		Filter: domain.Filter{
			StudentID:    req.Filter.GetStudentId(),
			CourseID:     req.Filter.GetCourseId(),
			LocationID:   req.Filter.GetLocationId(),
			AttendStatus: attendanceStatuses,
			StartDate:    support.StartOfDate(req.Filter.GetStartDate().AsTime().In(loc)),
			EndDate:      support.EndOfDate(req.Filter.GetEndDate().AsTime().In(loc)),
		},
		Paging: support.Paging[int]{
			Limit:  int(limit),
			Offset: int(offset),
		},
		IsFilterByCurrentYear: isFilterByCurrentYear,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*lpb.GetStudentAttendanceResponse_StudentAttendance, 0, len(resp.StudentAttendance))
	for _, sa := range resp.StudentAttendance {
		items = append(items, &lpb.GetStudentAttendanceResponse_StudentAttendance{
			LessonId:            sa.LessonID,
			StudentId:           sa.StudentID,
			CourseId:            sa.CourseID,
			LocationId:          sa.LocationID,
			ReallocatedLessonId: sa.ReallocatedLessonID,
			AttendanceStatus:    lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[sa.AttendStatus]),
		})
	}
	preOffset := offset
	if offset > 0 {
		preOffset = offset - int64(limit)
	}
	apiResp := &lpb.GetStudentAttendanceResponse{
		Items: items,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: offset + int64(limit),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: preOffset,
			},
		},
		TotalItems: resp.Total,
	}
	return apiResp, nil
}

func filterArgsFromRequestPayload(ctx context.Context, req *lpb.GetAssignedStudentListRequest) (args payloads.GetAssignedStudentListArg, isEmptyIntersectedLocation bool) {
	tz := req.GetTimezone()
	if tz == "" {
		tz = "UTC"
	}

	args = payloads.GetAssignedStudentListArg{
		PurchaseMethod:        req.GetPurchaseMethod().Enum().String(),
		CourseIDs:             req.GetFilter().GetCourseIds(),
		StudentIDs:            req.GetFilter().GetStudentIds(),
		Limit:                 req.Paging.GetLimit(),
		KeyWord:               req.GetKeyword(),
		StudentSubscriptionID: req.Paging.GetOffsetString(),
		SchoolID:              golibs.ResourcePathFromCtx(ctx),
		Timezone:              tz,
	}

	locations := req.GetLocationIds()
	filterLocations := req.GetFilter().GetLocationIds()
	locationsLen := len(locations)
	filterLocationsLen := len(filterLocations)
	location, _ := time.LoadLocation(tz)

	if locationsLen > 0 && filterLocationsLen > 0 {
		intersect, _, _ := golibs.Compare(locations, filterLocations)
		if len(intersect) == 0 {
			return args, true
		}
		args.LocationIDs = intersect
	} else {
		switch {
		case filterLocationsLen != 0:
			args.LocationIDs = filterLocations
		case locationsLen != 0:
			args.LocationIDs = locations
		}
	}

	if fromDate := req.Filter.GetStartDate(); fromDate != nil {
		args.FromDate = fromDate.AsTime().In(location)
	}

	if toDate := req.Filter.GetEndDate(); toDate != nil {
		args.ToDate = toDate.AsTime().In(location)
	}

	if status := req.GetFilter().GetStatuses(); len(status) > 0 {
		assignedStudentStatuses := make([]domain.AssignedStudentStatus, 0, len(status))
		for _, v := range status {
			assignedStudentStatuses = append(assignedStudentStatuses, domain.AssignedStudentStatus(v.String()))
		}
		args.AssignedStudentStatuses = assignedStudentStatuses
	}

	return args, false
}

func toAsgStudentPb(st *domain.AssignedStudent) *lpb.AssignedStudentInfo {
	assignedStatus := st.AssignedStatus
	switch assignedStatus {
	case UnderAssignedStatus:
		assignedStatus = domain.AssignedStudentStatusUnderAssigned
	case JustAssignedStatus:
		assignedStatus = domain.AssignedStudentStatusJustAssigned
	case OverAssignedStatus:
		assignedStatus = domain.AssignedStudentStatusOverAssigned
	default:
		assignedStatus = domain.AssignedStudentStatusUnderAssigned
	}

	return &lpb.AssignedStudentInfo{
		StudentId:     st.StudentID,
		CourseId:      st.CourseID,
		LocationId:    st.LocationID,
		Duration:      st.Duration,
		Status:        lpb.AssignedStudentStatus(lpb.AssignedStudentStatus_value[string(assignedStatus)]),
		PurchasedSlot: st.PurchasedSlot,
		AssignedSlot:  st.AssignedSlot,
		SlotGap:       st.SlotGap,
	}
}

func validateGetAssignedStudentListRequest(req *lpb.GetAssignedStudentListRequest) error {
	if req.GetPaging() == nil {
		return status.Error(codes.Internal, "missing paging info")
	}
	return nil
}
