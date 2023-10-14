package controller

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonAllocationReaderService struct {
	GetLessonAllocationHandler                    application.GetLessonAllocationHandler
	GetLessonScheduleByStudentSubscriptionHandler application.GetLessonScheduleByStudentSubscriptionHandler
}

func NewLessonAllocationReaderService(
	wrapperConnection *support.WrapperDBConnection,
	lessonAllocationRepo infrastructure.LessonAllocationRepo,
	academicWeekRepo infrastructure.AcademicWeekRepo,
	studentSubscriptionRepo infrastructure.StudentSubscriptionRepo,
	studentSubscriptionAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo,
	courseLocationSchedule infrastructure.CourseLocationScheduleRepo,
	academicYearRepo infrastructure.AcademicYearRepo,
	studentCourseRepo infrastructure.StudentCourseRepo,
) *LessonAllocationReaderService {
	return &LessonAllocationReaderService{
		GetLessonAllocationHandler: application.GetLessonAllocationHandler{
			AcademicYearRepo:     academicYearRepo,
			LessonAllocationRepo: lessonAllocationRepo,
			WrapperConnection:    wrapperConnection,
		},
		GetLessonScheduleByStudentSubscriptionHandler: application.GetLessonScheduleByStudentSubscriptionHandler{
			LessonAllocationRepo:              lessonAllocationRepo,
			WrapperConnection:                 wrapperConnection,
			StudentSubscriptionRepo:           studentSubscriptionRepo,
			StudentSubscriptionAccessPathRepo: studentSubscriptionAccessPathRepo,
			AcademicWeekRepo:                  academicWeekRepo,
			CourseLocationScheduleRepo:        courseLocationSchedule,
			StudentCourseRepo:                 studentCourseRepo,
		},
	}
}

func (a *LessonAllocationReaderService) GetLessonAllocation(ctx context.Context, in *lpb.GetLessonAllocationRequest) (*lpb.GetLessonAllocationResponse, error) {
	teachingMethods := in.Filter.GetTeachingMethods()
	teachingMethod := make([]string, len(teachingMethods))
	for _, tm := range teachingMethods {
		teachingMethod = append(teachingMethod, tm.String())
	}
	resp, err := a.GetLessonAllocationHandler.GetLessonAllocation(ctx, &application.GetLessonAllocationRequest{
		Filter: struct {
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
		}{
			in.Filter.GetCourseIds(),
			in.Filter.GetCourseTypeIds(),
			in.Filter.GetLocationIds(),
			teachingMethod,
			in.Filter.GetStartTime().AsTime(),
			in.Filter.GetEndTime().AsTime(),
			in.Filter.GetIsReallocationOnly(),
			in.Filter.GetIsClassUnassigned(),
			in.GetFilter().GetAllocationStatus().String(),
			in.Filter.GetProductId(),
		},
		LocationSettings: in.GetLocationIds(),
		KeySearch:        in.GetKeyword(),
		Paging: support.Paging[int]{
			Limit:  int(in.Paging.GetLimit()),
			Offset: int(in.Paging.GetOffsetInteger()),
		},
		Timezone: in.GetTimezone(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	preOffset := in.Paging.GetOffsetInteger()
	if in.Paging.GetOffsetInteger() > 0 {
		preOffset = in.Paging.GetOffsetInteger() - int64(in.Paging.GetLimit())
	}
	return &lpb.GetLessonAllocationResponse{
		Items:      toAllocationPb(resp),
		TotalItems: resp.Total,
		NextPage: &cpb.Paging{
			Limit: in.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: in.Paging.GetOffsetInteger() + int64(in.Paging.GetLimit()),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: in.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: preOffset,
			},
		},
		TotalOfNoneAssigned:      resp.TotalOfNoneAssigned,
		TotalOfPartiallyAssigned: resp.TotalOfPartiallyAssigned,
		TotalOfFullyAssigned:     resp.TotalOfFullyAssigned,
		TotalOfOverAssigned:      resp.TotalOfOverAssigned,
	}, nil
}

func toAllocationPb(resp *application.GetLessonAllocationResponse) []*lpb.GetLessonAllocationResponse_AllocationListInfo {
	results := []*lpb.GetLessonAllocationResponse_AllocationListInfo{}
	for _, student := range resp.Items {
		results = append(results, &lpb.GetLessonAllocationResponse_AllocationListInfo{
			StudentSubscriptionId: student.StudentSubscriptionID,
			StudentId:             student.StudentID,
			CourseId:              student.CourseID,
			LocationId:            student.LocationID,
			StartTime:             timestamppb.New(student.StartTime),
			EndTime:               timestamppb.New(student.EndTime),
			AllocationStatus:      lpb.LessonAllocationStatus(lpb.LessonAllocationStatus_value[student.AllocationStatus]),
			PurchasedSlot:         student.PurchasedSlot,
			AssignedSlot:          student.AssignedSlot,
			IsWeeklySchedule:      student.IsWeeklySchedule,
			PackageTypeSchedule:   lpb.PackageTypeSchedule(lpb.PackageTypeSchedule_value[student.PackageTypeSchedule]),
		})
	}
	return results
}

func (a *LessonAllocationReaderService) GetLessonScheduleByStudentSubscription(ctx context.Context, req *lpb.GetLessonScheduleByStudentSubscriptionRequest) (*lpb.GetLessonScheduleByStudentSubscriptionResponse, error) {
	return a.GetLessonScheduleByStudentSubscriptionHandler.GetLessonScheduleByStudentSubscription(ctx, req)
}
