package queries

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/infrastructure"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AssignedStudentQueryHandler struct {
	WrapperConnection *support.WrapperDBConnection
	// ports
	AssignedStudentRepo     infrastructure.AssignedStudentRepo
	ReallocationRepo        lesson_infras.ReallocationRepo
	StudentSubscriptionRepo user_infras.StudentSubscriptionRepo
	AcademicYearRepo        infrastructure.AcademicYearRepo
}

type GetAssignedStudentListResponse struct {
	AsgStudents []*domain.AssignedStudent
	Total       uint32
	OffsetID    string
	Error       error
}

type GetStudentAttendanceRequest struct {
	SearchKey             string
	Timezone              string
	Paging                support.Paging[int]
	Filter                domain.Filter
	IsFilterByCurrentYear bool
}

type GetStudentAttendanceResponse struct {
	StudentAttendance []*domain.StudentAttendance
	Total             uint32
}

func (a *AssignedStudentQueryHandler) GetAssignedStudentList(ctx context.Context, payload *payloads.GetAssignedStudentListArg) *GetAssignedStudentListResponse {
	var (
		preTotal    uint32
		asgStudents []*domain.AssignedStudent
		total       uint32
		offsetID    string
		err         error
	)
	conn, err := a.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return &GetAssignedStudentListResponse{Error: status.Error(codes.Internal, err.Error())}
	}
	asgStudents, total, offsetID, preTotal, err = a.AssignedStudentRepo.GetAssignedStudentList(ctx, conn, payload)
	if err != nil {
		return &GetAssignedStudentListResponse{Error: status.Error(codes.Internal, err.Error())}
	}

	if preTotal <= payload.Limit {
		offsetID = ""
	}

	return &GetAssignedStudentListResponse{
		AsgStudents: asgStudents,
		Total:       total,
		OffsetID:    offsetID,
		Error:       err,
	}
}

func (a *AssignedStudentQueryHandler) GetStudentAttendance(ctx context.Context, req *GetStudentAttendanceRequest) (*GetStudentAttendanceResponse, error) {
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	connDB, err := a.WrapperConnection.GetDB(resourcePath)
	loc, _ := time.LoadLocation(req.Timezone)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	academicYear, err := a.AcademicYearRepo.GetCurrentAcademicYear(ctx, connDB)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	attendanceStudent, total, err := a.AssignedStudentRepo.GetStudentAttendance(ctx, connDB, domain.GetStudentAttendanceParams{
		StartDate:                 req.Filter.StartDate,
		EndDate:                   req.Filter.EndDate,
		StudentID:                 req.Filter.StudentID,
		CourseID:                  req.Filter.CourseID,
		SearchKey:                 req.SearchKey,
		Timezone:                  req.Timezone,
		LocationID:                req.Filter.LocationID,
		AttendStatus:              req.Filter.AttendStatus,
		Limit:                     req.Paging.Limit,
		Offset:                    req.Paging.Offset,
		IsFilterByCurrentYear:     req.IsFilterByCurrentYear,
		YearStartDate:             academicYear.StartDate.Time.In(loc),
		YearEndDate:               academicYear.EndDate.Time.In(loc),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// reallocateLessonMembers = ["l1","s1","l1","s2","l2","s3"]
	var reallocateLessonMembers []string
	for _, as := range attendanceStudent {
		if as.AttendStatus == string(lesson_domain.StudentAttendStatusReallocate) {
			reallocateLessonMembers = append(reallocateLessonMembers, as.LessonID, as.StudentID)
		}
	}
	reallocationMap := make(map[string]string, 0)
	if len(reallocateLessonMembers) > 0 {
		reallocations, err := a.ReallocationRepo.GetReallocatedLesson(ctx, connDB, reallocateLessonMembers)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		for _, r := range reallocations {
			reallocationMap[r.GetKey()] = r.NewLessonID
		}
	}
	for _, as := range attendanceStudent {
		as.SetReallocatedLessonID(reallocationMap)
	}
	return &GetStudentAttendanceResponse{
		StudentAttendance: attendanceStudent,
		Total:             total,
	}, nil
}
