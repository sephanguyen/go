package controller

import (
	"context"

	calendar_infras "github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	infra_lesson_report "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	master_data_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	zoom_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonExecutorService struct {
	wrapperConnection                *support.WrapperDBConnection
	classroomQueryHandler            queries.ClassroomQueryHandler
	classroomCommandHandler          commands.ClassroomCommandHandler
	lessonQueryHandler               queries.LessonQueryHandler
	lessonCommandHandler             commands.LessonCommandHandler
	LessonProducer                   producers.LessonProducer
	MasterDataPort                   infrastructure.MasterDataPort
	DateInfoRepo                     infrastructure.DateInfoRepo
	zoomAccountService               zoom_service.ZoomAccountService
	exportUser                       queries.ExportUserHandler
	CourseTeachingTimeQueryHandler   queries.CourseTeachingTimeQueryHandler
	CourseTeachingTimeCommandHandler commands.CourseTeachingTimeCommandHandler
	UnleashClientIns                 unleashclient.ClientInstance
	Env                              string
}

func NewLessonExecutorService(
	wrapperConnection *support.WrapperDBConnection,
	classroomRepo infrastructure.ClassroomRepo,
	lessonRepo infrastructure.LessonRepo,
	jsm nats.JetStreamManagement,
	masterDataPort infrastructure.MasterDataPort,
	userModulePort infrastructure.UserModulePort,
	lessonReportRepo infra_lesson_report.LessonReportRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
	schedulerRepo calendar_infras.SchedulerPort,
	studentSubscriptionRepo user_infras.StudentSubscriptionRepo,
	reallocationRepo infrastructure.ReallocationRepo,
	dateInfoRepo infrastructure.DateInfoRepo,
	zoomAccountService zoom_service.ZoomAccountService,
	locationRepo master_data_domain.LocationRepository,
	courseRepo master_data_domain.CourseRepository,
	userBasicRepo user_infras.UserBasicInfoRepo,
	teacherRepo user_infras.TeacherRepo,
	schedulerClient clients.SchedulerClientInterface,
	lessonCourseRepo infrastructure.CourseRepo,
) *LessonExecutorService {
	return &LessonExecutorService{
		wrapperConnection: wrapperConnection,
		classroomQueryHandler: queries.ClassroomQueryHandler{
			WrapperConnection: wrapperConnection,
			ClassroomRepo:     classroomRepo,
			Env:               env,
			UnleashClientIns:  unleashClientIns,
		},
		classroomCommandHandler: commands.ClassroomCommandHandler{
			WrapperConnection: wrapperConnection,
			ClassroomRepo:     classroomRepo,
			MasterDataPort:    masterDataPort,
		},
		lessonQueryHandler: queries.LessonQueryHandler{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
			UnleashClientIns:  unleashClientIns,
			Env:               env,
		},
		lessonCommandHandler: commands.LessonCommandHandler{
			WrapperConnection:       wrapperConnection,
			LessonRepo:              lessonRepo,
			LessonReportRepo:        lessonReportRepo,
			SchedulerRepo:           schedulerRepo,
			Env:                     env,
			UnleashClientIns:        unleashClientIns,
			StudentSubscriptionRepo: studentSubscriptionRepo,
			LessonProducer: producers.LessonProducer{
				JSM: jsm,
			},
			ClassroomRepo:    classroomRepo,
			ReallocationRepo: reallocationRepo,
			MasterDataPort:   masterDataPort,
			DateInfoRepo:     dateInfoRepo,
			UserModulePort:   userModulePort,
			SchedulerClient:  schedulerClient,
		},
		LessonProducer: producers.LessonProducer{
			JSM: jsm,
		},
		MasterDataPort:     masterDataPort,
		DateInfoRepo:       dateInfoRepo,
		zoomAccountService: zoomAccountService,
		exportUser: queries.ExportUserHandler{
			WrapperConnection:       wrapperConnection,
			TeacherRepo:             teacherRepo,
			UserBasicInfoRepo:       userBasicRepo,
			LocationRepo:            locationRepo,
			CourseRepo:              courseRepo,
			StudentSubscriptionRepo: studentSubscriptionRepo,
			UnleashClient:           unleashClientIns,
			Env:                     env,
		},
		CourseTeachingTimeQueryHandler: queries.CourseTeachingTimeQueryHandler{
			WrapperConnection: wrapperConnection,
			CourseRepo:        lessonCourseRepo,
		},
		CourseTeachingTimeCommandHandler: commands.CourseTeachingTimeCommandHandler{
			WrapperConnection: wrapperConnection,
			CourseRepo:        lessonCourseRepo,
		},
		UnleashClientIns: unleashClientIns,
		Env:              env,
	}
}

func (l *LessonExecutorService) ExportClassrooms(ctx context.Context, req *lpb.ExportClassroomsRequest) (res *lpb.ExportClassroomsResponse, err error) {
	bytes, err := l.classroomQueryHandler.ExportClassrooms(ctx)
	if err != nil {
		return &lpb.ExportClassroomsResponse{}, status.Error(codes.Internal, err.Error())
	}
	res = &lpb.ExportClassroomsResponse{
		Data: bytes,
	}
	return res, nil
}

func (l *LessonExecutorService) GenerateLessonCSVTemplate(ctx context.Context, req *lpb.GenerateLessonCSVTemplateRequest) (res *lpb.GenerateLessonCSVTemplateResponse, err error) {
	bytes, err := l.lessonQueryHandler.GenerateLessonCSVTemplate(ctx)
	if err != nil {
		return &lpb.GenerateLessonCSVTemplateResponse{}, status.Error(codes.Internal, err.Error())
	}
	res = &lpb.GenerateLessonCSVTemplateResponse{
		Data: bytes,
	}
	return res, nil
}

func (l *LessonExecutorService) ImportLesson(ctx context.Context, req *lpb.ImportLessonRequest) (res *lpb.ImportLessonResponse, err error) {
	isEnabledOptimizingImportLesson, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_OptimizeImportLesson", l.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var lessons []*domain.Lesson
	var errorCSVs []*lpb.ImportLessonResponse_ImportLessonError

	if isEnabledOptimizingImportLesson {
		lessons, errorCSVs, err = l.lessonCommandHandler.ImportLessonV2(ctx, req)
	} else {
		lessons, errorCSVs, err = l.lessonCommandHandler.ImportLesson(ctx, req)
	}
	res = &lpb.ImportLessonResponse{
		Errors: errorCSVs,
	}
	if len(errorCSVs) > 0 {
		return res, nil
	}
	if err != nil {
		return res, status.Error(codes.Internal, err.Error())
	}

	// Publish lesson event
	lessonEvt := []*bpb.EvtLesson_Lesson{}
	for _, ls := range lessons {
		lessonEvt = append(lessonEvt, &bpb.EvtLesson_Lesson{
			LessonId:         ls.LessonID,
			LearnerIds:       ls.GetLearnersIDs(),
			TeacherIds:       ls.GetTeacherIDs(),
			LocationId:       ls.LocationID,
			StartAt:          timestamppb.New(ls.StartTime),
			EndAt:            timestamppb.New(ls.EndTime),
			SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(ls.SchedulingStatus)]),
		})
	}
	if err = l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_CreateLessons_{
			CreateLessons: &bpb.EvtLesson_CreateLessons{
				Lessons: lessonEvt,
			},
		},
	}); err != nil {
		return res, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (l *LessonExecutorService) ImportZoomAccount(ctx context.Context, req *lpb.ImportZoomAccountRequest) (res *lpb.ImportZoomAccountResponse, err error) {
	return l.zoomAccountService.ImportZoomAccount(ctx, req)
}

func (l *LessonExecutorService) ExportTeacher(ctx context.Context, req *lpb.ExportTeacherRequest) (*lpb.ExportTeacherResponse, error) {
	data, err := l.exportUser.ExportTeacher(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &lpb.ExportTeacherResponse{
		Data: data,
	}, nil
}

func (l *LessonExecutorService) ExportEnrolledStudent(ctx context.Context, req *lpb.ExportEnrolledStudentRequest) (*lpb.ExportEnrolledStudentResponse, error) {
	data, err := l.exportUser.ExportEnrolledStudent(ctx, req.Timezone)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &lpb.ExportEnrolledStudentResponse{
		Data: data,
	}, nil
}

func (l *LessonExecutorService) ImportClassroom(ctx context.Context, req *lpb.ImportClassroomRequest) (res *lpb.ImportClassroomResponse, err error) {
	return l.classroomCommandHandler.ImportClassroom(ctx, req)
}

func (l *LessonExecutorService) ExportCourseTeachingTime(ctx context.Context, _ *lpb.ExportCourseTeachingTimeRequest) (res *lpb.ExportCourseTeachingTimeResponse, err error) {
	data, err := l.CourseTeachingTimeQueryHandler.ExportCourseTeachingTime(ctx)
	if err != nil {
		return &lpb.ExportCourseTeachingTimeResponse{}, status.Error(codes.Internal, err.Error())
	}
	res = &lpb.ExportCourseTeachingTimeResponse{
		Data: data,
	}
	return res, nil
}

func (l *LessonExecutorService) ImportCourseTeachingTime(ctx context.Context, req *lpb.ImportCourseTeachingTimeRequest) (res *lpb.ImportCourseTeachingTimeResponse, err error) {
	courses, CSVErrs, err := l.CourseTeachingTimeCommandHandler.ImportCourseTeachingTime(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res = &lpb.ImportCourseTeachingTimeResponse{
		Errors: CSVErrs,
	}
	if len(CSVErrs) > 0 {
		return res, nil
	}

	// update future lesson with new course's teaching time value
	timezone := req.GetTimezone()
	if timezone == "" {
		timezone = "UTC"
	}
	err = l.lessonCommandHandler.UpdateFutureLessonsWhenCourseChanged(ctx, courses.GetCourseIDs(), timezone)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}
