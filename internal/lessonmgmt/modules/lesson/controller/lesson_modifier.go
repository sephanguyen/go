package controller

import (
	"context"
	"fmt"
	"time"

	calendar_infras "github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/producers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	infra_lesson_report "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
	zoom_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure"
	zoom_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonModifierService struct {
	wrapperConnection     *support.WrapperDBConnection
	RetrieveLessonCommand application.RetrieveLessonCommand
	LessonCommandHandler  commands.LessonCommandHandler
	LessonProducer        producers.LessonProducer

	MasterDataPort  infrastructure.MasterDataPort
	UserModulePort  infrastructure.UserModulePort
	MediaModulePort infrastructure.MediaModulePort
	DateInfoRepo    infrastructure.DateInfoRepo
	ClassroomRepo   infrastructure.ClassroomRepo
}

func NewLessonModifierService(
	wrapperConnection *support.WrapperDBConnection,
	jsm nats.JetStreamManagement,
	lessonRepo infrastructure.LessonRepo,
	masterDataPort infrastructure.MasterDataPort,
	userModulePort infrastructure.UserModulePort,
	mediaModulePort infrastructure.MediaModulePort,
	dateInfoRepo infrastructure.DateInfoRepo,
	classroomRepo infrastructure.ClassroomRepo,
	lessonReportRepo infra_lesson_report.LessonReportRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
	schedulerRepo calendar_infras.SchedulerPort,
	studentSubscriptionRepo user_infras.StudentSubscriptionRepo,
	reallocationRepo infrastructure.ReallocationRepo,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	zoomService zoom_service.ZoomServiceInterface,
	zoomAccountRepo zoom_repo.ZoomAccountRepo,
	userAccessPath infrastructure.UserAccessPathPort,
	studentEnrollmentHistory infrastructure.StudentEnrollmentStatusHistoryPort,
	schedulerClient clients.SchedulerClientInterface,
	lessonPublisher infrastructure.LessonPublisher,
) *LessonModifierService {
	return &LessonModifierService{
		wrapperConnection: wrapperConnection,
		RetrieveLessonCommand: application.RetrieveLessonCommand{
			WrapperConnection: wrapperConnection,
			LessonRepo:        lessonRepo,
		},
		LessonCommandHandler: commands.LessonCommandHandler{
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
			ClassroomRepo:                classroomRepo,
			ReallocationRepo:             reallocationRepo,
			ZoomService:                  zoomService,
			ZoomAccountRepo:              zoomAccountRepo,
			LessonMemberRepo:             lessonMemberRepo,
			UserAccessPathRepo:           userAccessPath,
			StudentEnrollmentHistoryRepo: studentEnrollmentHistory,
			SchedulerClient:              schedulerClient,
			LessonPublisher:              lessonPublisher,
			JSM:                          jsm,
			MasterDataPort:               masterDataPort,
		},
		LessonProducer: producers.LessonProducer{
			JSM: jsm,
		},
		MasterDataPort:  masterDataPort,
		UserModulePort:  userModulePort,
		MediaModulePort: mediaModulePort,
		DateInfoRepo:    dateInfoRepo,
		ClassroomRepo:   classroomRepo,
	}
}

func (l *LessonModifierService) BulkUpdateLessonSchedulingStatus(ctx context.Context, req *lpb.BulkUpdateLessonSchedulingStatusRequest) (*lpb.BulkUpdateLessonSchedulingStatusResponse, error) {
	foundLessons, err := l.RetrieveLessonCommand.GetLessonByIDs(ctx, req.GetLessonIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := l.LessonCommandHandler.BulkUpdateLessonSchedulingStatus(ctx, &commands.BulkUpdateLessonSchedulingStatusCommandRequest{Lessons: foundLessons, Action: req.Action})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for _, ls := range res.UpdatedLessons {
		teacherIds := ls.GetTeacherIDs()
		startAt := timestamppb.New(ls.StartTime)
		endAt := timestamppb.New(ls.EndTime)
		pbLiveLessons := &bpb.EvtLesson_UpdateLesson{
			LessonId:               ls.LessonID,
			LearnerIds:             ls.GetLearnersIDs(),
			LocationIdBefore:       ls.LocationID,
			LocationIdAfter:        ls.LocationID,
			StartAtBefore:          startAt,
			StartAtAfter:           startAt,
			EndAtBefore:            endAt,
			EndAtAfter:             endAt,
			TeacherIdsBefore:       teacherIds,
			TeacherIdsAfter:        teacherIds,
			SchedulingStatusBefore: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(ls.PreSchedulingStatus)]),
			SchedulingStatusAfter:  cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(ls.SchedulingStatus)]),
		}
		if l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
			Message: &bpb.EvtLesson_UpdateLesson_{
				UpdateLesson: pbLiveLessons,
			},
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &lpb.BulkUpdateLessonSchedulingStatusResponse{}, err
}

func (l *LessonModifierService) UpdateLessonSchedulingStatus(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error) {
	res, err := l.LessonCommandHandler.UpdateLessonSchedulingStatus(ctx, &commands.UpdateLessonStatusCommandRequest{
		LessonID:         req.LessonId,
		SchedulingStatus: req.SchedulingStatus.String(),
		SavingType:       req.SavingType,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, ls := range res.UpdatedLesson {
		teacherIds := ls.GetTeacherIDs()
		startAt := timestamppb.New(ls.StartTime)
		endAt := timestamppb.New(ls.EndTime)
		pbLiveLessons := &bpb.EvtLesson_UpdateLesson{
			LessonId:               ls.LessonID,
			LearnerIds:             ls.GetLearnersIDs(),
			LocationIdBefore:       ls.LocationID,
			LocationIdAfter:        ls.LocationID,
			StartAtBefore:          startAt,
			StartAtAfter:           startAt,
			EndAtBefore:            endAt,
			EndAtAfter:             endAt,
			TeacherIdsBefore:       teacherIds,
			TeacherIdsAfter:        teacherIds,
			SchedulingStatusBefore: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(ls.PreSchedulingStatus)]),
			SchedulingStatusAfter:  cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(ls.SchedulingStatus)]),
		}
		if l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
			Message: &bpb.EvtLesson_UpdateLesson_{
				UpdateLesson: pbLiveLessons,
			},
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &lpb.UpdateLessonSchedulingStatusResponse{}, nil
}

func (l *LessonModifierService) CreateLesson(ctx context.Context, req *lpb.CreateLessonRequest) (*lpb.CreateLessonResponse, error) {
	lesson, err := l.buildArgsFromRequest(req)
	if err != nil {
		return nil, err
	}
	if req.SavingOption == nil {
		return nil, status.Error(codes.InvalidArgument, "saving_option could not be empty")
	}
	lessons := []*domain.Lesson{}
	switch req.SavingOption.Method {
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME:
		payload := commands.CreateLesson{Lesson: lesson, TimeZone: req.TimeZone}
		lesson, err = l.LessonCommandHandler.CreateLessonOneTime(ctx, payload)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		lessons = append(lessons, lesson)
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE:
		var endDate time.Time
		if req.SavingOption.Recurrence != nil {
			endDate = golibs.TimestamppbToTime(req.SavingOption.Recurrence.EndDate)
		}
		ruleCmd := commands.CreateRecurringLesson{
			Lesson: lesson,
			RRuleCmd: commands.RecurrenceRuleCommand{
				StartTime: golibs.TimestamppbToTime(req.StartTime),
				EndTime:   golibs.TimestamppbToTime(req.EndTime),
				UntilDate: endDate,
			},
			TimeZone: req.TimeZone,
		}
		zoomInfo := req.GetZoomInfo()
		if zoomInfo != nil {
			ruleCmd.ZoomInfo = &commands.ZoomInfo{
				ZoomAccountID: zoomInfo.ZoomAccountOwner,
				ZoomLink:      zoomInfo.ZoomLink,
				ZoomID:        zoomInfo.ZoomId,
				ZoomOccurrences: sliceutils.Map(zoomInfo.Occurrences, func(occurrence *lpb.ZoomInfo_OccurrenceZoom) *commands.ZoomOccurrence {
					return &commands.ZoomOccurrence{
						OccurrenceID: occurrence.OccurrenceId,
						StartTime:    occurrence.StartTime,
					}
				}),
			}
		}
		recurLesson, err := l.LessonCommandHandler.CreateRecurringLesson(ctx, ruleCmd)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		lessons = recurLesson.Lessons
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf(`unexpected saving option method %T`, req.SavingOption.Method))
	}

	// Publish lesson event
	if err = l.PublishEventCreatedLesson(ctx, lessons); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var selectedLesson string
	if len(lessons) > 0 {
		selectedLesson = lessons[0].LessonID
	}
	return &lpb.CreateLessonResponse{
		Id: selectedLesson,
	}, nil
}

func (l *LessonModifierService) buildArgsFromRequest(req *lpb.CreateLessonRequest) (*domain.Lesson, error) {
	now := time.Now()
	builder := domain.NewLesson().
		WithLocationID(req.LocationId).
		WithTimeRange(golibs.TimestamppbToTime(req.StartTime), golibs.TimestamppbToTime(req.EndTime)).
		WithTeachingMedium(domain.LessonTeachingMedium(req.TeachingMedium.String())).
		WithTeachingMethod(domain.LessonTeachingMethod(req.TeachingMethod.String())).
		WithTeacherIDs(req.TeacherIds).
		WithClassroomIDs(req.ClassroomIds).
		WithMasterDataPort(l.MasterDataPort).
		WithUserModulePort(l.UserModulePort).
		WithModificationTime(now, now).
		WithMediaModulePort(l.MediaModulePort).
		WithDateInfoRepo(l.DateInfoRepo).
		WithClassroomRepo(l.ClassroomRepo).
		WithSchedulingStatus(domain.LessonSchedulingStatus(req.SchedulingStatus.String())).
		WithLessonCapacity(int32(req.GetLessonCapacity())).
		WithPreparationTime(int32(-1)).
		WithBreakTime(int32(-1))
	if req.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP {
		builder.WithClassID(req.ClassId).
			WithCourseID(req.CourseId)
	}
	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		builder.WithZoomAccountID(zoomInfo.GetZoomAccountOwner()).
			WithZoomLink(zoomInfo.GetZoomLink()).
			WithZoomID(zoomInfo.GetZoomId())
	}
	classDoInfo := req.GetClassDoInfo()
	if classDoInfo != nil {
		builder.
			WithClassDoOwnerID(classDoInfo.GetClassDoOwnerId()).
			WithClassDoLink(classDoInfo.GetClassDoLink()).
			WithClassDoRoomID(classDoInfo.GetClassDoRoomId())
	}
	// only get mediaIDs from request to add new lesson
	mediaIDs, err := getMediaIDsFromMaterials(req.Materials)
	if err != nil {
		if errors.Is(err, fmt.Errorf("not yet support this type")) {
			return nil, status.Error(codes.InvalidArgument, "currently not yet allowed create media when creating lesson")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	builder.WithMaterials(mediaIDs)

	for _, studentInfo := range req.StudentInfoList {
		learner := domain.NewLessonLearner(
			studentInfo.GetStudentId(),
			studentInfo.GetCourseId(),
			studentInfo.GetLocationId(),
			studentInfo.GetAttendanceStatus().String(),
			studentInfo.GetAttendanceNotice().String(),
			studentInfo.GetAttendanceReason().String(),
			studentInfo.GetAttendanceNote(),
		)
		if studentInfo.Reallocate != nil {
			learner.AddReallocate(studentInfo.Reallocate.GetOriginalLessonId())
		}
		err := learner.Validate()
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		builder.AddLearner(learner)
	}
	return builder.BuildDraft(), nil
}

func getMediaIDsFromMaterials(materials []*lpb.Material) ([]string, error) {
	if len(materials) == 0 {
		return nil, nil
	}
	res := make([]string, 0, len(materials))
	for _, material := range materials {
		switch resource := material.Resource.(type) {
		case *lpb.Material_BrightcoveVideo_:
			return nil, fmt.Errorf("not yet support this type")
		case *lpb.Material_MediaId:
			res = append(res, resource.MediaId)
		default:
			return nil, fmt.Errorf(`unexpected material's type %T`, resource)
		}
	}

	return res, nil
}

func (l *LessonModifierService) DeleteLesson(ctx context.Context, req *lpb.DeleteLessonRequest) (*lpb.DeleteLessonResponse, error) {
	if req.LessonId == "" {
		return nil, status.Error(codes.InvalidArgument, "lessonID could not be empty")
	}
	var IsDeletedRecurring bool
	if req.SavingOption != nil {
		switch req.SavingOption.Method {
		case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME:
		case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE:
			IsDeletedRecurring = true
		}
	}
	lessonIDs, err := l.LessonCommandHandler.DeleteLesson(ctx, commands.DeleteLessonRequest{
		LessonID:           req.LessonId,
		IsDeletedRecurring: IsDeletedRecurring,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := l.LessonProducer.PublishLessonEvt(ctx,
		&bpb.EvtLesson{
			Message: &bpb.EvtLesson_DeletedLessons_{
				DeletedLessons: &bpb.EvtLesson_DeletedLessons{
					LessonIds: lessonIDs,
				},
			},
		}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &lpb.DeleteLessonResponse{}, nil
}

func (l *LessonModifierService) UpdateLesson(ctx context.Context, req *lpb.UpdateLessonRequest) (*lpb.UpdateLessonResponse, error) {
	builder := domain.NewLesson().
		WithID(req.LessonId).
		WithLocationID(req.LocationId).
		WithTimeRange(golibs.TimestamppbToTime(req.StartTime), golibs.TimestamppbToTime(req.EndTime)).
		WithTeachingMedium(domain.LessonTeachingMedium(req.TeachingMedium.String())).
		WithTeachingMethod(domain.LessonTeachingMethod(req.TeachingMethod.String())).
		WithTeacherIDs(req.TeacherIds).
		WithClassroomIDs(req.ClassroomIds).
		WithMasterDataPort(l.MasterDataPort).
		WithUserModulePort(l.UserModulePort).
		WithMediaModulePort(l.MediaModulePort).
		WithLessonRepo(l.LessonCommandHandler.LessonRepo).
		WithClassID(req.ClassId).
		WithCourseID(req.CourseId).
		WithDateInfoRepo(l.DateInfoRepo).
		WithClassroomRepo(l.ClassroomRepo).
		WithSchedulingStatus(domain.LessonSchedulingStatus(req.SchedulingStatus.String())).
		WithLessonCapacity(int32(req.GetLessonCapacity())).
		WithPreparationTime(int32(-1)).
		WithBreakTime(int32(-1))

	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		builder.WithZoomAccountID(zoomInfo.GetZoomAccountOwner()).
			WithZoomLink(zoomInfo.GetZoomLink()).
			WithZoomID(zoomInfo.GetZoomId())
	}
	classDoInfo := req.GetClassDoInfo()
	if classDoInfo != nil {
		builder.
			WithClassDoOwnerID(classDoInfo.GetClassDoOwnerId()).
			WithClassDoLink(classDoInfo.GetClassDoLink()).
			WithClassDoRoomID(classDoInfo.GetClassDoRoomId())
	}
	// only get mediaIDs from request to add new lesson
	mediaIDs, err := getMediaIDsFromMaterials(req.Materials)
	if err != nil {
		if errors.Is(err, fmt.Errorf("not yet support this type")) {
			return nil, status.Error(codes.InvalidArgument, "currently not yet allowed create media when creating lesson")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	builder = builder.WithMaterials(mediaIDs)

	for _, studentInfo := range req.StudentInfoList {
		learner := domain.NewLessonLearner(
			studentInfo.GetStudentId(),
			studentInfo.GetCourseId(),
			studentInfo.GetLocationId(),
			studentInfo.GetAttendanceStatus().String(),
			studentInfo.GetAttendanceNotice().String(),
			studentInfo.GetAttendanceReason().String(),
			studentInfo.GetAttendanceNote(),
		)
		if studentInfo.Reallocate != nil {
			learner.AddReallocate(studentInfo.Reallocate.GetOriginalLessonId())
		}
		if learner.Validate() != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		builder = builder.AddLearner(learner)
	}
	lesson := builder.BuildDraft()
	currentLesson, err := l.RetrieveLessonCommand.GetLessonByID(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if req.SavingOption == nil {
		return nil, status.Error(codes.InvalidArgument, "saving_option could not be empty")
	}
	var (
		updatedLessons []*domain.Lesson
		createdLessons []*domain.Lesson
	)
	var mapOldLessons map[string]*domain.Lesson

	switch req.SavingOption.Method {
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME:
		cmdRequest := commands.UpdateLessonOneTimeCommandRequest{
			Lesson:        lesson,
			CurrentLesson: currentLesson,
			TimeZone:      req.TimeZone,
		}
		lesson, err = l.LessonCommandHandler.UpdateLessonOneTime(ctx, cmdRequest)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		updatedLessons = append(updatedLessons, lesson)
	case lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE:
		var endDate time.Time
		if req.SavingOption.Recurrence != nil {
			endDate = golibs.TimestamppbToTime(req.SavingOption.Recurrence.EndDate)
		}
		cmdRequest := commands.UpdateRecurringLessonCommandRequest{
			SelectedLesson: lesson,
			CurrentLesson:  currentLesson,
			RRuleCmd: commands.RecurrenceRuleCommand{
				StartTime: golibs.TimestamppbToTime(req.StartTime),
				EndTime:   golibs.TimestamppbToTime(req.EndTime),
				UntilDate: endDate,
			},
			TimeZone: req.TimeZone,
		}
		zoomInfo := req.GetZoomInfo()
		if zoomInfo != nil {
			cmdRequest.ZoomInfo = &commands.ZoomInfo{
				ZoomAccountID: zoomInfo.ZoomAccountOwner,
				ZoomLink:      zoomInfo.ZoomLink,
				ZoomID:        zoomInfo.ZoomId,
				ZoomOccurrences: sliceutils.Map(zoomInfo.Occurrences, func(occurrence *lpb.ZoomInfo_OccurrenceZoom) *commands.ZoomOccurrence {
					return &commands.ZoomOccurrence{
						OccurrenceID: occurrence.OccurrenceId,
						StartTime:    occurrence.StartTime,
					}
				}),
			}
		}
		upsertedLessons, lessonMap, err := l.LessonCommandHandler.UpdateRecurringLesson(ctx, cmdRequest)
		mapOldLessons = lessonMap
		for _, ls := range upsertedLessons {
			if ls.Persisted {
				updatedLessons = append(updatedLessons, ls)
			} else {
				createdLessons = append(createdLessons, ls)
			}
		}
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf(`unexpected saving option method %T`, req.SavingOption.Method))
	}

	// Publish lesson event
	learnerIds := make([]string, 0, len(req.StudentInfoList))
	for _, student := range req.StudentInfoList {
		learnerIds = append(learnerIds, student.StudentId)
	}

	for _, updatedLesson := range updatedLessons {
		oldData, ok := mapOldLessons[updatedLesson.LessonID]
		if !ok {
			oldData = currentLesson
		}
		pbLiveLessons := &bpb.EvtLesson_UpdateLesson{
			LessonId:               updatedLesson.LessonID,
			ClassName:              "",
			LearnerIds:             learnerIds,
			LocationIdBefore:       oldData.LocationID,
			LocationIdAfter:        updatedLesson.LocationID,
			StartAtBefore:          timestamppb.New(oldData.StartTime),
			StartAtAfter:           timestamppb.New(updatedLesson.StartTime),
			EndAtBefore:            timestamppb.New(oldData.EndTime),
			EndAtAfter:             timestamppb.New(updatedLesson.EndTime),
			TeacherIdsBefore:       oldData.GetTeacherIDs(),
			TeacherIdsAfter:        updatedLesson.GetTeacherIDs(),
			SchedulingStatusBefore: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(oldData.SchedulingStatus)]),
			SchedulingStatusAfter:  cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(updatedLesson.SchedulingStatus)]),
		}
		if err = l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
			Message: &bpb.EvtLesson_UpdateLesson_{
				UpdateLesson: pbLiveLessons,
			},
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	lessonEvt := make([]*bpb.EvtLesson_Lesson, 0, len(createdLessons))
	for _, ls := range createdLessons {
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
	if len(lessonEvt) > 0 {
		if err = l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
			Message: &bpb.EvtLesson_CreateLessons_{
				CreateLessons: &bpb.EvtLesson_CreateLessons{
					Lessons: lessonEvt,
				},
			},
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &lpb.UpdateLessonResponse{}, nil
}

func (l *LessonModifierService) UpdateToRecurrence(ctx context.Context, req *lpb.UpdateToRecurrenceRequest) (*lpb.UpdateToRecurrenceResponse, error) {
	builder := domain.NewLesson().
		WithID(req.LessonId).
		WithLocationID(req.LocationId).
		WithTimeRange(golibs.TimestamppbToTime(req.StartTime), golibs.TimestamppbToTime(req.EndTime)).
		WithTeachingMedium(domain.LessonTeachingMedium(req.TeachingMedium.String())).
		WithTeachingMethod(domain.LessonTeachingMethod(req.TeachingMethod.String())).
		WithTeacherIDs(req.TeacherIds).
		WithClassroomIDs(req.ClassroomIds).
		WithMasterDataPort(l.MasterDataPort).
		WithUserModulePort(l.UserModulePort).
		WithMediaModulePort(l.MediaModulePort).
		WithLessonRepo(l.LessonCommandHandler.LessonRepo).
		WithClassID(req.ClassId).
		WithCourseID(req.CourseId).
		WithDateInfoRepo(l.DateInfoRepo).
		WithClassroomRepo(l.ClassroomRepo).
		WithSchedulingStatus(domain.LessonSchedulingStatus(req.SchedulingStatus.String())).
		WithLessonCapacity(int32(req.GetLessonCapacity())).
		WithPreparationTime(int32(-1)).
		WithBreakTime(int32(-1))

	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		builder.WithZoomAccountID(zoomInfo.GetZoomAccountOwner()).
			WithZoomLink(zoomInfo.GetZoomLink()).
			WithZoomID(zoomInfo.GetZoomId())
	}
	classDoInfo := req.GetClassDoInfo()
	if classDoInfo != nil {
		builder.
			WithClassDoOwnerID(classDoInfo.GetClassDoOwnerId()).
			WithClassDoLink(classDoInfo.GetClassDoLink()).
			WithClassDoRoomID(classDoInfo.GetClassDoRoomId())
	}

	mediaIDs, err := getMediaIDsFromMaterials(req.Materials)
	if err != nil {
		if errors.Is(err, fmt.Errorf("not yet support this type")) {
			return nil, status.Error(codes.InvalidArgument, "currently not yet allowed create media when creating lesson")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	builder = builder.WithMaterials(mediaIDs)

	for _, studentInfo := range req.StudentInfoList {
		learner := domain.NewLessonLearner(
			studentInfo.GetStudentId(),
			studentInfo.GetCourseId(),
			studentInfo.GetLocationId(),
			studentInfo.GetAttendanceStatus().String(),
			studentInfo.GetAttendanceNotice().String(),
			studentInfo.GetAttendanceReason().String(),
			studentInfo.GetAttendanceNote(),
		)
		if studentInfo.Reallocate != nil {
			learner.AddReallocate(studentInfo.Reallocate.GetOriginalLessonId())
		}
		if learner.Validate() != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		builder = builder.AddLearner(learner)
	}

	lesson := builder.BuildDraft()

	var endDate time.Time
	if req.SavingOption.Recurrence != nil {
		endDate = golibs.TimestamppbToTime(req.SavingOption.Recurrence.GetEndDate())
	}
	ruleCmd := commands.CreateRecurringLesson{
		Lesson: lesson,
		RRuleCmd: commands.RecurrenceRuleCommand{
			StartTime: golibs.TimestamppbToTime(req.StartTime),
			EndTime:   golibs.TimestamppbToTime(req.EndTime),
			UntilDate: endDate,
		},
		TimeZone: req.TimeZone,
	}
	if zoomInfo != nil {
		ruleCmd.ZoomInfo = &commands.ZoomInfo{
			ZoomAccountID: zoomInfo.ZoomAccountOwner,
			ZoomLink:      zoomInfo.ZoomLink,
			ZoomID:        zoomInfo.ZoomId,
			ZoomOccurrences: sliceutils.Map(zoomInfo.Occurrences, func(occurrence *lpb.ZoomInfo_OccurrenceZoom) *commands.ZoomOccurrence {
				return &commands.ZoomOccurrence{
					OccurrenceID: occurrence.OccurrenceId,
					StartTime:    occurrence.StartTime,
				}
			}),
		}
	}
	currentLesson, err := l.RetrieveLessonCommand.GetLessonByID(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	recurringLesson, err := l.LessonCommandHandler.CreateRecurringLesson(ctx, ruleCmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	lessons := recurringLesson.Lessons
	if len(lessons) > 0 {
		if err = l.PublishEventUpdatedLesson(ctx, lessons[0], currentLesson); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err = l.PublishEventCreatedLesson(ctx, lessons[1:]); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &lpb.UpdateToRecurrenceResponse{}, nil
}

func (l *LessonModifierService) PublishEventCreatedLesson(ctx context.Context, lessons []*domain.Lesson) error {
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
	return l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_CreateLessons_{
			CreateLessons: &bpb.EvtLesson_CreateLessons{
				Lessons: lessonEvt,
			},
		},
	})
}

func (l *LessonModifierService) PublishEventUpdatedLesson(ctx context.Context, updatedLesson, oldLesson *domain.Lesson) error {
	updatedLessonEvent := &bpb.EvtLesson_UpdateLesson{
		LessonId:               updatedLesson.LessonID,
		LearnerIds:             updatedLesson.GetLearnersIDs(),
		LocationIdBefore:       oldLesson.LocationID,
		LocationIdAfter:        updatedLesson.LocationID,
		StartAtBefore:          timestamppb.New(oldLesson.StartTime),
		StartAtAfter:           timestamppb.New(updatedLesson.StartTime),
		EndAtBefore:            timestamppb.New(oldLesson.EndTime),
		EndAtAfter:             timestamppb.New(updatedLesson.EndTime),
		TeacherIdsBefore:       oldLesson.GetTeacherIDs(),
		TeacherIdsAfter:        updatedLesson.GetTeacherIDs(),
		SchedulingStatusBefore: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(oldLesson.SchedulingStatus)]),
		SchedulingStatusAfter:  cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(updatedLesson.SchedulingStatus)]),
	}
	return l.LessonProducer.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_UpdateLesson_{
			UpdateLesson: updatedLessonEvent,
		},
	})
}

func (l *LessonModifierService) MarkStudentAsReallocate(ctx context.Context, request *lpb.MarkStudentAsReallocateRequest) (*lpb.MarkStudentAsReallocateResponse, error) {
	member := &domain.LessonMember{
		LessonID:         request.LessonId,
		StudentID:        request.StudentId,
		AttendanceStatus: string(domain.StudentAttendStatusReallocate),
	}
	reAllocations := &domain.Reallocation{
		OriginalLessonID: request.LessonId,
		StudentID:        request.StudentId,
	}
	req := &commands.MarkStudentAsReallocateRequest{
		Member:        member,
		ReAllocations: reAllocations,
	}

	err := l.LessonCommandHandler.MarkStudentAsReallocate(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &lpb.MarkStudentAsReallocateResponse{}, nil
}
