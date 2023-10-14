package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure"
	course_location_schedule_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
	master_data_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type GetLessonScheduleByStudentSubscriptionHandler struct {
	LessonAllocationRepo              infrastructure.LessonAllocationRepo
	StudentSubscriptionRepo           infrastructure.StudentSubscriptionRepo
	StudentSubscriptionAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo
	AcademicWeekRepo                  infrastructure.AcademicWeekRepo
	CourseLocationScheduleRepo        infrastructure.CourseLocationScheduleRepo
	StudentCourseRepo                 infrastructure.StudentCourseRepo

	WrapperConnection *support.WrapperDBConnection
}

func (g *GetLessonScheduleByStudentSubscriptionHandler) GetLessonScheduleByStudentSubscription(ctx context.Context, req *lpb.GetLessonScheduleByStudentSubscriptionRequest) (*lpb.GetLessonScheduleByStudentSubscriptionResponse, error) {
	dbConn, err := g.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	studentSubscription, err := g.StudentSubscriptionRepo.GetByStudentSubscriptionID(ctx, dbConn, req.GetStudentSubscriptionId())
	if err != nil {
		return nil, fmt.Errorf("g.StudentSubscriptionRepo.GetByStudentSubscriptionID: %w", err)
	}
	studentSubscriptionAccessPath, err := g.StudentSubscriptionAccessPathRepo.FindLocationsByStudentSubscriptionIDs(ctx, dbConn, []string{studentSubscription.StudentSubscriptionID})
	if err != nil {
		return nil, fmt.Errorf("g.StudentSubscriptionAccessPathRepo.FindLocationsByStudentSubscriptionIDs: %w", err)
	}

	locationID, exist := studentSubscriptionAccessPath[studentSubscription.StudentSubscriptionID]
	if !exist || len(locationID) == 0 {
		return nil, fmt.Errorf("student_subscription_id `%s` doesn't have location", studentSubscription.StudentSubscriptionID)
	}
	uniqueLocationID := locationID[0]

	courseLocationSchedule, err := g.CourseLocationScheduleRepo.GetByCourseIDAndLocationID(ctx, dbConn, studentSubscription.CourseID, uniqueLocationID)
	if err == course_location_schedule_domain.ErrorNotFound {
		return &lpb.GetLessonScheduleByStudentSubscriptionResponse{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("g.CourseLocationScheduleRepo.GetByCourseIDAndLocationID: %w", err)
	}

	academicWeeks, err := g.AcademicWeekRepo.GetByDateRange(
		ctx,
		dbConn,
		uniqueLocationID,
		courseLocationSchedule.AcademicWeeks,
		studentSubscription.StartAt,
		studentSubscription.EndAt)
	if err != nil {
		return nil, fmt.Errorf("g.AcademicWeekRepo.GetByDateRange: %w", err)
	}

	academicWeekID := make([]string, 0, len(academicWeeks))
	academicWeekMap := make(map[string]*master_data_domain.AcademicWeek, len(academicWeeks))
	for _, aw := range academicWeeks {
		academicWeekID = append(academicWeekID, aw.AcademicWeekID.String)
		academicWeekMap[aw.AcademicWeekID.String] = aw
	}

	offset := int(req.GetPaging().GetOffsetInteger())
	limit := int(req.GetPaging().GetLimit())

	end := limit + offset
	if end > len(academicWeekID) {
		end = len(academicWeekID)
	}
	academicWeekUse := academicWeekID[offset:end]

	studentAllocatedByWeek, err := g.LessonAllocationRepo.GetByStudentSubscriptionAndWeek(
		ctx, dbConn, studentSubscription.StudentID, studentSubscription.CourseID, academicWeekUse)
	if err != nil {
		return nil, fmt.Errorf("g.LessonAllocationRepo.GetByStudentSubscriptionAndWeek: %w", err)
	}

	assignedSlot, err := g.LessonAllocationRepo.CountAssignedSlotPerStudentCourse(ctx, dbConn, studentSubscription.StudentID, studentSubscription.CourseID)
	if err != nil {
		return nil, fmt.Errorf("g.LessonAllocationRepo.CountAssignedSlotPerStudentCourse: %w", err)
	}
	var (
		lessonTotal     uint32
		configFrequency uint32
	)
	switch {
	case courseLocationSchedule.IsFrequency():
		studentCourse, err := g.StudentCourseRepo.GetByStudentCourseID(ctx, dbConn, studentSubscription.StudentID, studentSubscription.CourseID, uniqueLocationID, studentSubscription.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("g.StudentCourseRepo.GetByStudentCourseID: %w", err)
		}
		configFrequency = uint32(studentCourse.CourseSlotPerWeek)
	case courseLocationSchedule.IsSchedule():
		configFrequency = uint32(*courseLocationSchedule.Frequency)
	case courseLocationSchedule.IsSlotBased():
		studentCourse, err := g.StudentCourseRepo.GetByStudentCourseID(ctx, dbConn, studentSubscription.StudentID, studentSubscription.CourseID, uniqueLocationID, studentSubscription.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("g.StudentCourseRepo.GetByStudentCourseID: %w", err)
		}
		lessonTotal = uint32(studentCourse.CourseSlot)
	case courseLocationSchedule.IsOneTime():
		lessonTotal = uint32(*courseLocationSchedule.TotalNoLesson)
	}
	if courseLocationSchedule.IsFrequency() || courseLocationSchedule.IsSchedule() {
		total, err := g.LessonAllocationRepo.CountPurchasedSlotPerStudentSubscription(
			ctx,
			dbConn,
			uint8(configFrequency),
			studentSubscription.StartAt,
			studentSubscription.EndAt,
			studentSubscription.CourseID,
			uniqueLocationID,
			studentSubscription.StudentID,
		)
		if err != nil {
			return nil, fmt.Errorf("g.LessonAllocationRepo.CountPurchasedSlotPerStudentSubscription: %w", err)
		}
		lessonTotal = total
	}

	items := make([]*lpb.GetLessonScheduleByStudentSubscriptionResponse_WeeklyLessonList, 0, len(academicWeekUse))
	for _, academicWeekID := range academicWeekUse {
		lessonAllocationInfo := studentAllocatedByWeek[academicWeekID]
		lessons := make([]*lpb.GetLessonScheduleByStudentSubscriptionResponse_Lesson, 0, len(lessonAllocationInfo))
		for _, ls := range lessonAllocationInfo {
			lessons = append(lessons, &lpb.GetLessonScheduleByStudentSubscriptionResponse_Lesson{
				LessonId:         ls.LessonID,
				StartTime:        timestamppb.New(ls.StartTime),
				EndTime:          timestamppb.New(ls.EndTime),
				LocationId:       ls.LocationID,
				AttendanceStatus: lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[string(ls.AttendanceStatus)]),
				LessonStatus:     lpb.LessonStatus(lpb.LessonStatus_value[string(ls.Status)]),
				TeachingMethod:   cpb.LessonTeachingMethod(cpb.LessonTeachingMethod_value[string(ls.TeachingMethod)]),
				ReportId:         ls.LessonReportID,
				IsLocked:         ls.IsLocked,
			})
		}
		awData := academicWeekMap[academicWeekID]
		items = append(items, &lpb.GetLessonScheduleByStudentSubscriptionResponse_WeeklyLessonList{
			AcademicWeekId: academicWeekID,
			WeekOrder:      uint32(awData.WeekOrder.Int),
			WeekName:       awData.Name.String,
			StartTime:      timestamppb.New(awData.StartDate.Time),
			EndTime:        timestamppb.New(awData.EndDate.Time),
			LocationId:     awData.LocationID.String,
			Frequency:      configFrequency,
			Lessons:        lessons,
		})
	}
	preOffset := req.Paging.GetOffsetInteger()
	if req.Paging.GetOffsetInteger() > 0 {
		preOffset = req.Paging.GetOffsetInteger() - int64(req.Paging.GetLimit())
	}
	return &lpb.GetLessonScheduleByStudentSubscriptionResponse{
		Items:      items,
		TotalItems: uint32(len(academicWeeks)),
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: req.Paging.GetOffsetInteger() + int64(req.Paging.GetLimit()),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: preOffset,
			},
		},
		AllocatedLessonsCount: assignedSlot,
		TotalLesson:           lessonTotal,
		CourseLocationSchedule: &lpb.GetLessonScheduleByStudentSubscriptionResponse_CourseLocationSchedule{
			CourseLocationScheduleId: courseLocationSchedule.ID,
			PackageTypeSchedule:      lpb.PackageTypeSchedule(lpb.PackageTypeSchedule_value[string(courseLocationSchedule.ProductTypeSchedule)]),
		},
	}, nil
}
