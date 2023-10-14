package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type GetLessonAllocationHandler struct {
	LessonAllocationRepo infrastructure.LessonAllocationRepo
	AcademicYearRepo     infrastructure.AcademicYearRepo

	WrapperConnection *support.WrapperDBConnection
}

type (
	GetLessonAllocationRequest struct {
		LocationSettings []string
		KeySearch        string
		Timezone         string
		Filter           struct {
			CourseID               []string
			CourseTypeID           []string
			LocationID             []string
			TeachingMethod         []string
			StartDate              time.Time
			EndDate                time.Time
			IsOnlyReallocation     bool
			IsClassUnassigned      bool
			LessonAllocationStatus string
			ProductID              []string
		}
		Paging support.Paging[int]
	}

	GetLessonAllocationResponse struct {
		Items                    []Item
		Total                    uint32
		TotalOfNoneAssigned      uint32
		TotalOfPartiallyAssigned uint32
		TotalOfFullyAssigned     uint32
		TotalOfOverAssigned      uint32
	}
	Item struct {
		StudentSubscriptionID string
		StudentID             string
		CourseID              string
		LocationID            string
		StartTime             time.Time
		EndTime               time.Time
		AssignedSlot          int32
		PurchasedSlot         int32
		AllocationStatus      string
		IsWeeklySchedule      bool
		PackageTypeSchedule   string
	}
)

func (gah *GetLessonAllocationHandler) GetLessonAllocation(ctx context.Context, req *GetLessonAllocationRequest) (*GetLessonAllocationResponse, error) {
	conn, err := gah.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	locationID := req.LocationSettings
	if len(req.Filter.LocationID) > 0 {
		locationID = req.Filter.LocationID
	}
	academicYear, err := gah.AcademicYearRepo.GetCurrentAcademicYear(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("gah.AcademicYearRepo.GetCurrentAcademicYear: %w", err)
	}
	startDate := req.Filter.StartDate
	endDate := req.Filter.EndDate

	if req.Filter.StartDate.Unix() > 0 && req.Filter.EndDate.Unix() <= support.UnixToEnd {
		endDate = time.Now().UTC()
	}

	if academicYear.StartDate.Time.After(startDate) {
		startDate = academicYear.StartDate.Time
	}

	if endDate.Unix() <= support.UnixToEnd || endDate.After(academicYear.EndDate.Time) {
		endDate = academicYear.EndDate.Time
	}

	lessonAllocationStatus := domain.LessonAllocationStatus(req.Filter.LessonAllocationStatus)
	teachingMethods := make([]domain.CourseTeachingMethod, len(req.Filter.TeachingMethod))
	for _, tm := range req.Filter.TeachingMethod {
		teachingMethods = append(teachingMethods, domain.CourseTeachingMethod(tm))
	}
	allocatedStudent, lessonAllocationStatusMap, err := gah.LessonAllocationRepo.GetLessonAllocation(ctx, conn, domain.LessonAllocationFilter{
		CourseID:               req.Filter.CourseID,
		CourseTypeID:           req.Filter.CourseTypeID,
		TeachingMethod:         teachingMethods,
		StartDate:              startDate,
		EndDate:                endDate,
		IsOnlyReallocation:     req.Filter.IsOnlyReallocation,
		LocationID:             locationID,
		LessonAllocationStatus: domain.LessonAllocationStatus(req.Filter.LessonAllocationStatus),
		TimeZone:               req.Timezone,
		Limit:                  req.Paging.Limit,
		Offset:                 req.Paging.Offset,
		KeySearch:              strings.TrimSpace(req.KeySearch),
		IsClassUnassigned:      req.Filter.IsClassUnassigned,
		ProductID:              req.Filter.ProductID,
	})

	if err != nil {
		return nil, fmt.Errorf("gah.LessonAllocationRepo.GetLessonAllocation: %w", err)
	}
	items := make([]Item, 0, len(allocatedStudent))
	for _, as := range allocatedStudent {
		items = append(items, Item{
			StudentSubscriptionID: as.StudentSubscriptionID,
			StudentID:             as.StudentID,
			CourseID:              as.CourseID,
			LocationID:            as.LocationID,
			StartTime:             as.StartTime,
			EndTime:               as.EndTime,
			AssignedSlot:          as.AssignedSlot,
			PurchasedSlot:         as.PurchasedSlot,
			IsWeeklySchedule:      as.IsWeeklySchedule(),
			AllocationStatus:      as.AllocationStatus(),
			PackageTypeSchedule:   as.ProductTypeSchedule,
		})
	}
	return &GetLessonAllocationResponse{
		Items:                    items,
		TotalOfNoneAssigned:      lessonAllocationStatusMap[string(domain.NoneAssigned)],
		TotalOfPartiallyAssigned: lessonAllocationStatusMap[string(domain.PartiallyAssigned)],
		TotalOfFullyAssigned:     lessonAllocationStatusMap[string(domain.FullyAssigned)],
		TotalOfOverAssigned:      lessonAllocationStatusMap[string(domain.OverAssigned)],
		Total: func() (total uint32) {
			switch lessonAllocationStatus {
			case domain.NoneAssigned:
				return lessonAllocationStatusMap[string(domain.NoneAssigned)]
			case domain.PartiallyAssigned:
				return lessonAllocationStatusMap[string(domain.PartiallyAssigned)]
			case domain.OverAssigned:
				return lessonAllocationStatusMap[string(domain.OverAssigned)]
			case domain.FullyAssigned:
				return lessonAllocationStatusMap[string(domain.FullyAssigned)]
			default:
				for _, count := range lessonAllocationStatusMap {
					total += count
				}
				return total
			}
		}(),
	}, err
}
