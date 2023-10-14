package controller

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/calendar/application"
	"github.com/manabie-com/backend/internal/calendar/application/queries"
	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/calendar/support"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	commonpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonReaderService struct {
	wrapperConnection  *support.WrapperDBConnection
	lessonQueryHandler application.QueryLessonPort
}

func NewLessonReaderService(
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonPort,
	lessonTeacherRepo infrastructure.LessonTeacherPort,
	lessonMemberRepo infrastructure.LessonMemberPort,
	lessonClassroomRepo infrastructure.LessonClassroomPort,
	lessonGroupRepo infrastructure.LessonGroupPort,
	schedulerRepo infrastructure.SchedulerPort,
	userRepo infrastructure.UserPort,
	env string,
	unleashClient unleashclient.ClientInstance,
) *LessonReaderService {
	return &LessonReaderService{
		wrapperConnection: wrapperConnection,
		lessonQueryHandler: &queries.LessonQueryHandler{
			LessonRepo:          lessonRepo,
			LessonTeacherRepo:   lessonTeacherRepo,
			LessonMemberRepo:    lessonMemberRepo,
			LessonClassroomRepo: lessonClassroomRepo,
			LessonGroupRepo:     lessonGroupRepo,
			SchedulerRepo:       schedulerRepo,
			UserRepo:            userRepo,
			Env:                 env,
			UnleashClient:       unleashClient,
		},
	}
}

func (l *LessonReaderService) GetLessonDetailOnCalendar(ctx context.Context, req *cpb.GetLessonDetailOnCalendarRequest) (*cpb.GetLessonDetailOnCalendarResponse, error) {
	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request := &payloads.GetLessonDetailRequest{
		LessonID: req.LessonId,
	}

	if err = request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := l.lessonQueryHandler.GetLessonDetail(ctx, conn, request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return toLessonDetailOnCalendarPb(response.Lesson, response.Scheduler), nil
}

func toLessonDetailOnCalendarPb(l *lesson_domain.Lesson, s *dto.Scheduler) *cpb.GetLessonDetailOnCalendarResponse {
	lessonTeachersLength := len(l.Teachers)
	lessonTeachers := make([]*cpb.GetLessonDetailOnCalendarResponse_LessonTeacher, 0, lessonTeachersLength)
	for _, teacher := range l.Teachers {
		lessonTeachers = append(lessonTeachers, &cpb.GetLessonDetailOnCalendarResponse_LessonTeacher{
			TeacherId:   teacher.TeacherID,
			TeacherName: teacher.Name,
		})
	}

	lessonMembersLength := len(l.Learners)
	lessonMembers := make([]*cpb.GetLessonDetailOnCalendarResponse_LessonMember, 0, lessonMembersLength)
	for _, learner := range l.Learners {
		course := &cpb.GetLessonDetailOnCalendarResponse_LessonMember_Course{
			CourseId:   learner.CourseID,
			CourseName: learner.CourseName,
		}

		lessonMembers = append(lessonMembers, &cpb.GetLessonDetailOnCalendarResponse_LessonMember{
			StudentId:        learner.LearnerID,
			Grade:            learner.Grade,
			StudentName:      learner.LearnerName,
			Course:           course,
			AttendanceStatus: commonpb.StudentAttendStatus(commonpb.StudentAttendStatus_value[string(learner.AttendStatus)]),
			AttendanceReason: commonpb.StudentAttendanceReason(commonpb.StudentAttendanceReason_value[string(learner.AttendanceReason)]),
			AttendanceNotice: commonpb.StudentAttendanceNotice(commonpb.StudentAttendanceNotice_value[string(learner.AttendanceNotice)]),
			AttendanceNote:   learner.AttendanceNote,
		})
	}

	lessonClassroomsLength := len(l.Classrooms)
	lessonClassrooms := make([]*cpb.GetLessonDetailOnCalendarResponse_LessonClassroom, 0, lessonClassroomsLength)
	for _, classroom := range l.Classrooms {
		lessonClassrooms = append(lessonClassrooms, &cpb.GetLessonDetailOnCalendarResponse_LessonClassroom{
			ClassroomId:   classroom.ClassroomID,
			ClassroomName: classroom.ClassroomName,
		})
	}

	mediaIDs := []string{}
	if l.Material != nil {
		mediaIDs = l.Material.MediaIDs
	}

	return &cpb.GetLessonDetailOnCalendarResponse{
		IsLocked:         l.IsLocked,
		LessonId:         l.LessonID,
		LessonName:       l.Name,
		StartTime:        timestamppb.New(l.StartTime),
		EndTime:          timestamppb.New(l.EndTime),
		TeachingMedium:   commonpb.LessonTeachingMedium(commonpb.LessonTeachingMedium_value[string(l.TeachingMedium)]),
		TeachingMethod:   commonpb.LessonTeachingMethod(commonpb.LessonTeachingMethod_value[string(l.TeachingMethod)]),
		SchedulingStatus: commonpb.LessonSchedulingStatus(commonpb.LessonSchedulingStatus_value[string(l.SchedulingStatus)]),
		CourseId:         l.CourseID,
		CourseName:       l.CourseName,
		Location: &cpb.GetLessonDetailOnCalendarResponse_Location{
			LocationId:   l.LocationID,
			LocationName: l.LocationName,
		},
		Class: &cpb.GetLessonDetailOnCalendarResponse_Class{
			ClassId:   l.ClassID,
			ClassName: l.ClassName,
		},
		Scheduler: &cpb.GetLessonDetailOnCalendarResponse_Scheduler{
			SchedulerId: s.SchedulerID,
			StartDate:   timestamppb.New(s.StartDate),
			EndDate:     timestamppb.New(s.EndDate),
			Frequency:   cpb.Frequency(cpb.Frequency_value[strings.ToUpper(s.Frequency)]),
		},
		LessonTeachers:   lessonTeachers,
		LessonMembers:    lessonMembers,
		LessonClassrooms: lessonClassrooms,
		MediaIds:         mediaIDs,
		ZoomId:           l.ZoomID,
		ZoomLink:         l.ZoomLink,
		ZoomOwnerId:      l.ZoomOwnerID,
		ClassDoId:        l.ClassDoRoomID,
		ClassDoLink:      l.ClassDoLink,
		ClassDoOwnerId:   l.ClassDoOwnerID,
		LessonCapacity:   uint32(l.LessonCapacity),
	}
}

func (l *LessonReaderService) GetLessonIDsForBulkStatusUpdate(ctx context.Context, req *cpb.GetLessonIDsForBulkStatusUpdateRequest) (*cpb.GetLessonIDsForBulkStatusUpdateResponse, error) {
	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	request := &payloads.GetLessonIDsForBulkStatusUpdateRequest{
		LocationID: req.LocationId,
		Action:     lesson_domain.LessonBulkAction(req.GetAction().String()),
		StartDate:  req.StartDate.AsTime(),
		EndDate:    req.EndDate.AsTime(),
		Timezone:   req.Timezone,
	}
	if startTime := req.GetStartTime(); startTime != nil {
		request.StartTime = startTime.AsTime()
	}
	if endTime := req.GetEndTime(); endTime != nil {
		request.EndTime = endTime.AsTime()
	}

	if err := request.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := l.lessonQueryHandler.GetLessonIDsForBulkStatusUpdate(ctx, conn, request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	lessonIDsDetails := make([]*cpb.GetLessonIDsForBulkStatusUpdateResponse_LessonIDsDetail, 0, len(response))
	for _, detail := range response {
		lessonIDsDetails = append(lessonIDsDetails,
			&cpb.GetLessonIDsForBulkStatusUpdateResponse_LessonIDsDetail{
				SchedulingStatus:       commonpb.LessonSchedulingStatus(commonpb.LessonSchedulingStatus_value[string(detail.LessonStatus)]),
				ModifiableLessonsCount: detail.ModifiableLessonsCount,
				LessonsCount:           detail.LessonsCount,
				LessonIds:              detail.LessonIDs,
			},
		)
	}

	return &cpb.GetLessonIDsForBulkStatusUpdateResponse{
		LessonIdsDetails: lessonIDsDetails,
	}, nil
}
