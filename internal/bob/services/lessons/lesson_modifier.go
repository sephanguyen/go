package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	coursesRepo "github.com/manabie-com/backend/internal/bob/services/courses/repo"
	"github.com/manabie-com/backend/internal/bob/services/log"
	mediaRepo "github.com/manabie-com/backend/internal/bob/services/media/repo"
	topicsRepo "github.com/manabie-com/backend/internal/bob/services/topics/repo"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type LessonRepo interface {
	IncreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text, MaximumLearnerStreamings int) error
	DecreaseNumberOfStreaming(ctx context.Context, db database.QueryExecer, lessonID, learnerID pgtype.Text) error
	GetStreamingLearners(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, queryEnhancers ...repositories.QueryEnhancer) ([]string, error)
	Create(ctx context.Context, db database.Ext, lesson *entities.Lesson) (*entities.Lesson, error)
	FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
	FindEarliestAndLatestTimeLessonByCourses(ctx context.Context, db database.Ext, courseIDs pgtype.TextArray) (*entities.CourseAvailableRanges, error)
	GetCourseIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	Update(ctx context.Context, db database.Ext, lesson *entities.Lesson) error
	UpsertLessonTeachers(ctx context.Context, db database.Ext, lessonID pgtype.Text, teacherIDs pgtype.TextArray) error
	UpsertLessonMembers(ctx context.Context, db database.Ext, lessonID pgtype.Text, userIDs pgtype.TextArray) error
	UpsertLessonCourses(ctx context.Context, db database.Ext, lessonID pgtype.Text, courseIDs pgtype.TextArray) error
	Delete(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray) error
	DeleteLessonMembers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	DeleteLessonTeachers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	DeleteLessonCourses(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error

	Retrieve(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) ([]*entities.Lesson, uint32, string, uint32, error)
	FindPreviousPageOffset(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (string, error)
	CountLesson(ctx context.Context, db database.QueryExecer, args *repositories.ListLessonArgs) (int64, error)
	GetTeacherIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
	StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error

	// Course's find lesson endpoint
	FindLessonWithTime(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	FindLessonWithTimeAndLocations(ctx context.Context, db database.QueryExecer, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	FindLessonJoined(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
	FindLessonJoinedWithLocations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, courseIDs *pgtype.TextArray, startDate *pgtype.Timestamptz, endDate *pgtype.Timestamptz, locationIDs *pgtype.TextArray, limit int32, page int32, schedulingStatus pgtype.Text) ([]*repositories.LessonWithTime, pgtype.Int8, error)
}

type LessonGroupRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error
	Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error)
	UpdateMedias(ctx context.Context, db database.QueryExecer, e *entities.LessonGroup) error
}

type LessonMemberRepo interface {
	GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
	UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
	UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
	UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
}

type LessonPollingRepo interface {
	Create(ctx context.Context, db database.Ext, polling *entities.LessonPolling) (*entities.LessonPolling, error)
}

type ActivityLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.ActivityLog) error
}

type SchoolAdminRepo interface {
	Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error)
}

type TeacherRepo interface {
	Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, fields ...string) ([]entities.Teacher, error)
}

type UserRepo interface {
	Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.User, error)
	UserGroup(context.Context, database.QueryExecer, pgtype.Text) (string, error)
}

type StudentRepo interface {
	Retrieve(context.Context, database.QueryExecer, pgtype.TextArray) ([]repositories.StudentProfile, error)
}

type LessonRoomStateRepo interface {
	Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error
	UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error
	UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error
}

type LessonModifierServices struct {
	bpb.UnimplementedLessonModifierServiceServer
	Cfg                        configurations.Config
	DB                         database.Ext
	JSM                        nats.JetStreamManagement
	VirtualClassRoomLogService *log.VirtualClassRoomLogService
	ActivityLogRepo
	LessonRepo
	LessonGroupRepo
	LessonMemberRepo
	LessonPollingRepo
	mediaRepo.MediaRepo
	coursesRepo.CourseRepo
	coursesRepo.PresetStudyPlanRepo
	coursesRepo.PresetStudyPlanWeeklyRepo
	topicsRepo.TopicRepo
	UserRepo            UserRepo
	SchoolAdminRepo     SchoolAdminRepo
	TeacherRepo         TeacherRepo
	StudentRepo         StudentRepo
	LessonRoomStateRepo LessonRoomStateRepo
}

func NewLessonModifierServices(
	cfg configurations.Config,
	db database.Ext,
	mediaRp mediaRepo.MediaRepo,
	activityLogRp ActivityLogRepo,
	lessonRp LessonRepo,
	lessonGrRp LessonGroupRepo,
	courseRp coursesRepo.CourseRepo,
	pSpRp coursesRepo.PresetStudyPlanRepo,
	pSPWRp coursesRepo.PresetStudyPlanWeeklyRepo,
	topicRp topicsRepo.TopicRepo,
	userRp UserRepo,
	schoolAdminRp SchoolAdminRepo,
	teacherRp TeacherRepo,
	studentRp StudentRepo,
	lessonMemberRp LessonMemberRepo,
	lessonPollingRp LessonPollingRepo,
	lessonRoomStateRp LessonRoomStateRepo,
	jsm nats.JetStreamManagement,
	vcl *log.VirtualClassRoomLogService,
) *LessonModifierServices {
	return &LessonModifierServices{
		Cfg:                        cfg,
		DB:                         db,
		VirtualClassRoomLogService: vcl,
		MediaRepo:                  mediaRp,
		ActivityLogRepo:            activityLogRp,
		LessonRepo:                 lessonRp,
		LessonGroupRepo:            lessonGrRp,
		CourseRepo:                 courseRp,
		PresetStudyPlanRepo:        pSpRp,
		TopicRepo:                  topicRp,
		PresetStudyPlanWeeklyRepo:  pSPWRp,
		UserRepo:                   userRp,
		SchoolAdminRepo:            schoolAdminRp,
		TeacherRepo:                teacherRp,
		StudentRepo:                studentRp,
		LessonMemberRepo:           lessonMemberRp,
		LessonPollingRepo:          lessonPollingRp,
		LessonRoomStateRepo:        lessonRoomStateRp,
		JSM:                        jsm,
	}
}

func (l *LessonModifierServices) PreparePublish(ctx context.Context, req *bpb.PreparePublishRequest) (*bpb.PreparePublishResponse, error) {
	if req.GetLearnerId() == "" || req.GetLessonId() == "" {
		return nil, fmt.Errorf("LessonID or LearnerID must not empty")
	}
	response := &bpb.PreparePublishResponse{}
	err := database.ExecInTx(ctx, l.DB, func(ctx context.Context, tx pgx.Tx) error {
		learnerIds, err := l.LessonRepo.GetStreamingLearners(ctx, tx, database.Text(req.LessonId), repositories.WithUpdateLock())
		if err != nil {
			return err
		}
		if contains(learnerIds, req.LearnerId) {
			response.Status = bpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE
			return fmt.Errorf("prepared before")
		}
		if err := l.LessonRepo.IncreaseNumberOfStreaming(ctx, tx, database.Text(req.LessonId), database.Text(req.LearnerId), l.Cfg.Agora.MaximumLearnerStreamings); err != nil {
			if err == repositories.ErrUnAffected {
				response.Status = bpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT
				return err
			}
			return fmt.Errorf("s.LessonStreamRepo.IncreaseNumberOfStreaming: %w", err)
		}

		activityLog := &entities.ActivityLog{}
		database.AllNullEntity(activityLog)
		if err = multierr.Combine(
			activityLog.ID.Set(idutil.ULIDNow()),
			activityLog.Payload.Set(map[string]interface{}{
				"lesson_id": req.LessonId,
			}),
			activityLog.UserID.Set(req.LearnerId),
			activityLog.ActionType.Set(entities.LogActionTypePublish),
		); err != nil {
			return err
		}
		if err = l.ActivityLogRepo.Create(ctx, tx, activityLog); err != nil {
			return fmt.Errorf("s.ActivityLogRepo.Create: %w", err)
		}
		return nil
	})
	if err != nil {
		if err == repositories.ErrUnAffected || err.Error() == "prepared before" {
			return response, nil
		}
		return nil, fmt.Errorf("ExecInTx: %w", err)
	}

	return &bpb.PreparePublishResponse{}, nil
}

func (l *LessonModifierServices) Unpublish(ctx context.Context, req *bpb.UnpublishRequest) (*bpb.UnpublishResponse, error) {
	if req.GetLearnerId() == "" || req.GetLessonId() == "" {
		return nil, fmt.Errorf("LessonID or LearnerID must not empty")
	}
	response := &bpb.UnpublishResponse{}
	err := database.ExecInTx(ctx, l.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := l.LessonRepo.DecreaseNumberOfStreaming(ctx, tx, database.Text(req.LessonId), database.Text(req.LearnerId)); err != nil {
			if err == repositories.ErrUnAffected {
				response.Status = bpb.UnpublishStatus_UNPUBLISH_STATUS_UNPUBLISHED_BEFORE
				return err
			}
			return fmt.Errorf("s.LessonStreamRepo.DecreaseNumberOfStreaming: %w", err)
		}
		activityLog := &entities.ActivityLog{}
		database.AllNullEntity(activityLog)
		if err := multierr.Combine(
			activityLog.ID.Set(idutil.ULIDNow()),
			activityLog.Payload.Set(map[string]interface{}{
				"lesson_id": req.LessonId,
			}),
			activityLog.UserID.Set(req.LearnerId),
			activityLog.ActionType.Set(entities.LogActionTypePublish),
		); err != nil {
			return err
		}
		if err := l.ActivityLogRepo.Create(ctx, tx, activityLog); err != nil {
			return fmt.Errorf("s.ActivityLogRepo.Create: %w", err)
		}
		return nil
	})
	if err != nil {
		if err == repositories.ErrUnAffected {
			return response, nil
		}
		return nil, fmt.Errorf("ExecInTx: %w", err)
	}
	return &bpb.UnpublishResponse{}, nil
}

func contains(s []string, target string) bool {
	for _, val := range s {
		if target == val {
			return true
		}
	}
	return false
}

func (l *LessonModifierServices) CreateLiveLesson(ctx context.Context, req *bpb.CreateLiveLessonRequest) (*bpb.CreateLiveLessonResponse, error) {
	// convert payload to lesson entities
	lesson, err := createLiveLessonRequestToLesson(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = lesson.LessonType.Set(entities.LessonTypeOnline); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = lesson.TeachingMedium.Set(entities.LessonTeachingMediumOnline); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = lesson.TeachingMethod.Set(entities.LessonTeachingMethodIndividual); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err = lesson.SchedulingStatus.Set("LESSON_SCHEDULING_STATUS_PUBLISHED"); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	medias, err := materialsToMedias(req.Materials)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	admin, err := l.SchoolAdminRepo.Get(ctx, l.DB, database.Text(userID))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var schoolID *int32
	if admin != nil {
		schoolID = &admin.SchoolID.Int
	}

	builder := NewLessonBuilder(
		l.LessonRepo,
		l.LessonGroupRepo,
		l.CourseRepo,
		l.PresetStudyPlanRepo,
		l.PresetStudyPlanWeeklyRepo,
		l.TopicRepo,
		l.MediaRepo,
		l.UserRepo,
		l.SchoolAdminRepo,
		l.TeacherRepo,
		l.StudentRepo,
	)
	if err = builder.Create(
		ctx,
		l.DB,
		lesson,
		schoolID,
		WithMedia(medias),
	); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Publish lesson event
	pbLiveLessons := []*bpb.EvtLesson_Lesson{
		{
			LessonId:   builder.Lesson.LessonID.String,
			Name:       builder.Lesson.Name.String,
			LearnerIds: database.FromTextArray(builder.Lesson.LearnerIDs.LearnerIDs),
		},
	}
	if err = l.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_CreateLessons_{
			CreateLessons: &bpb.EvtLesson_CreateLessons{
				Lessons: pbLiveLessons,
			},
		},
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.CreateLiveLessonResponse{
		Id: lesson.LessonID.String,
	}, nil
}

func createLiveLessonRequestToLesson(inp *bpb.CreateLiveLessonRequest) (*entities.Lesson, error) {
	res := &entities.Lesson{}
	database.AllNullEntity(res)
	err := multierr.Combine(
		res.Name.Set(inp.Name),
		res.StartTime.Set(inp.StartTime.AsTime()),
		res.EndTime.Set(inp.EndTime.AsTime()),
		res.TeacherIDs.TeacherIDs.Set(inp.TeacherIds),
		res.CourseIDs.CourseIDs.Set(inp.CourseIds),
		res.LearnerIDs.LearnerIDs.Set(inp.LearnerIds),
		res.Status.Set(entities.LessonStatusDraft),
		res.IsLocked.Set(false),
	)
	if err != nil {
		return nil, err
	}

	if inp.StartTime == nil {
		res.StartTime = pgtype.Timestamptz{}
	}

	if inp.EndTime == nil {
		res.EndTime = pgtype.Timestamptz{}
	}

	return res, err
}

func materialsToMedias(materials []*bpb.Material) (entities.Medias, error) {
	if len(materials) == 0 {
		return nil, nil
	}
	res := make(entities.Medias, 0, len(materials))
	for _, material := range materials {
		t := entities.Media{}
		database.AllNullEntity(&t)
		switch resource := material.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			id, err := golibs.GetBrightcoveVideoIDFromURL(resource.BrightcoveVideo.Url)
			if err != nil {
				return nil, err
			}

			err = multierr.Combine(
				t.Name.Set(resource.BrightcoveVideo.Name),
				t.Resource.Set(id),
				t.Type.Set(string(entities.MediaTypeVideo)),
			)
			if err != nil {
				return nil, err
			}
		case *bpb.Material_MediaId:
			err := t.MediaID.Set(resource.MediaId)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf(`unexpected material's type %T`, resource)
		}
		res = append(res, &t)
	}

	return res, nil
}

func (l *LessonModifierServices) UpdateLiveLesson(ctx context.Context, req *bpb.UpdateLiveLessonRequest) (*bpb.UpdateLiveLessonResponse, error) {
	// Prepare the inputs
	lesson, err := updateLiveLessonRequestToLesson(req)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("updateLiveLessonRequestToLesson: %s", err))
	}

	medias, err := materialsToMedias(req.Materials)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("materialsToMedias: %s", err))
	}

	// Update the live lesson using a transaction
	builder := NewLessonBuilder(
		l.LessonRepo,
		l.LessonGroupRepo,
		l.CourseRepo,
		l.PresetStudyPlanRepo,
		l.PresetStudyPlanWeeklyRepo,
		l.TopicRepo,
		l.MediaRepo,
		l.UserRepo,
		l.SchoolAdminRepo,
		l.TeacherRepo,
		l.StudentRepo,
	)
	if err = builder.UpdateWithMedias(ctx, l.DB, lesson, medias); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("builder.UpdateWithMedias: %s", err))
	}

	// Publish lesson event
	pbLiveLessons := &bpb.EvtLesson_UpdateLesson{
		LessonId:   builder.Lesson.LessonID.String,
		ClassName:  builder.Lesson.Name.String,
		LearnerIds: database.FromTextArray(builder.Lesson.LearnerIDs.LearnerIDs),
	}

	if err = l.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_UpdateLesson_{
			UpdateLesson: pbLiveLessons,
		},
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.UpdateLiveLessonResponse{}, nil
}

func updateLiveLessonRequestToLesson(inp *bpb.UpdateLiveLessonRequest) (*entities.Lesson, error) {
	res := &entities.Lesson{}
	database.AllNullEntity(res)
	err := multierr.Combine(
		res.LessonID.Set(inp.Id),
		res.Name.Set(inp.Name),
		res.StartTime.Set(inp.StartTime.AsTime()),
		res.EndTime.Set(inp.EndTime.AsTime()),
		res.LessonType.Set(string(entities.LessonTypeOnline)),
		res.TeachingMedium.Set(string(entities.LessonTeachingMediumOnline)),
		res.TeacherIDs.TeacherIDs.Set(inp.TeacherIds),
		res.CourseIDs.CourseIDs.Set(inp.CourseIds),
		res.LearnerIDs.LearnerIDs.Set(inp.LearnerIds),
		res.SchedulingStatus.Set("LESSON_SCHEDULING_STATUS_PUBLISHED"),
		res.IsLocked.Set(false),
	)
	if err != nil {
		return nil, fmt.Errorf("multierr.Combine: %s", err)
	}

	if inp.StartTime == nil {
		res.StartTime = pgtype.Timestamptz{Status: pgtype.Null}
	}
	if inp.EndTime == nil {
		res.EndTime = pgtype.Timestamptz{Status: pgtype.Null}
	}
	return res, nil
}

func (l *LessonModifierServices) DeleteLiveLesson(ctx context.Context, req *bpb.DeleteLiveLessonRequest) (*bpb.DeleteLiveLessonResponse, error) {
	if req == nil || len(req.Id) == 0 {
		return nil, status.Error(codes.InvalidArgument, "request live lesson id could not be empty")
	}

	builder := NewLessonBuilder(
		l.LessonRepo,
		l.LessonGroupRepo,
		l.CourseRepo,
		l.PresetStudyPlanRepo,
		l.PresetStudyPlanWeeklyRepo,
		l.TopicRepo,
		l.MediaRepo,
		l.UserRepo,
		l.SchoolAdminRepo,
		l.TeacherRepo,
		l.StudentRepo,
	)
	err := builder.Delete(ctx, l.DB, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &bpb.DeleteLiveLessonResponse{}, nil
}

//nolint:interfacer
func (l *LessonModifierServices) PublishLessonEvt(ctx context.Context, msg *bpb.EvtLesson) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	switch msg.Message.(type) {
	case *bpb.EvtLesson_CreateLessons_:
		msgID, err := l.JSM.PublishAsyncContext(ctx, constants.SubjectLessonCreated, data)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishLessonEvt rcv.JSM.PublishAsyncContext Lesson.Created failed, msgID: %s, %w", msgID, err))
		}
	default:
		msgID, err := l.JSM.PublishAsyncContext(ctx, constants.SubjectLessonUpdated, data)
		if err != nil {
			return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishLessonEvt rcv.JSM.PublishAsyncContext Lesson.Updated failed, msgID: %s, %w", msgID, err))
		}
	}

	return nil
}

func (l *LessonModifierServices) ModifyLiveLessonState(ctx context.Context, req *bpb.ModifyLiveLessonStateRequest) (*bpb.ModifyLiveLessonStateResponse, error) {
	// get lesson
	reader := &LessonReaderServices{
		LessonRepo: l.LessonRepo,
		DB:         l.DB,
	}
	lesson, err := reader.getLiveLesson(ctx, database.Text(req.Id), includeLearnerIDs(), includeTeacherIDs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	var command StateModifyCommand
	switch req.Command.(type) {
	case *bpb.ModifyLiveLessonStateRequest_ShareAMaterial:
		t := &ShareMaterialCommand{
			State: &CurrentMaterial{
				MediaID: req.GetShareAMaterial().MediaId,
			},
		}

		switch req.GetShareAMaterial().State.(type) {
		case *bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState:
			tplState := req.GetShareAMaterial().GetVideoState()
			t.State.VideoState = &VideoState{
				PlayerState: PlayerState(tplState.PlayerState.String()),
			}
			if tplState.CurrentTime != nil {
				t.State.VideoState.CurrentTime = Duration(tplState.CurrentTime.AsDuration())
			}
		case *bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_PdfState:
			break
		case *bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_AudioState:
			audioState := req.GetShareAMaterial().GetAudioState()
			t.State.AudioState = &AudioState{
				PlayerState: PlayerState(audioState.PlayerState.String()),
			}
			if audioState.CurrentTime != nil {
				t.State.AudioState.CurrentTime = Duration(audioState.CurrentTime.AsDuration())
			}
		}
		command = t
	case *bpb.ModifyLiveLessonStateRequest_StopSharingMaterial:
		command = &StopSharingMaterialCommand{}
	case *bpb.ModifyLiveLessonStateRequest_FoldHandAll:
		command = &FoldHandAllCommand{}
	case *bpb.ModifyLiveLessonStateRequest_FoldUserHand:
		command = &UpdateHandsUpCommand{
			UserID: req.GetFoldUserHand(),
			State: &UserHandsUp{
				Value: false,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_RaiseHand:
		command = &UpdateHandsUpCommand{
			UserID: userID,
			State: &UserHandsUp{
				Value: true,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_HandOff:
		command = &UpdateHandsUpCommand{
			UserID: userID,
			State: &UserHandsUp{
				Value: false,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_AnnotationEnable:
		Learners := req.GetAnnotationEnable().Learners
		command = &UpdateAnnotationCommand{
			UserIDs: Learners,
			State: &UserAnnotation{
				Value: true,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_AnnotationDisable:
		Learners := req.GetAnnotationDisable().Learners
		command = &UpdateAnnotationCommand{
			UserIDs: Learners,
			State: &UserAnnotation{
				Value: false,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_StartPolling:
		Options := []*PollingOption{}
		for _, option := range req.GetStartPolling().Options {
			Options = append(Options, &PollingOption{
				Answer:    option.Answer,
				IsCorrect: option.IsCorrect,
			})
		}
		command = &StartPollingCommand{
			Options: Options,
		}
	case *bpb.ModifyLiveLessonStateRequest_StopPolling:
		command = &StopPollingCommand{}
	case *bpb.ModifyLiveLessonStateRequest_EndPolling:
		command = &EndPollingCommand{}
	case *bpb.ModifyLiveLessonStateRequest_SubmitPollingAnswer:
		command = &SubmitPollingAnswerCommand{
			UserID:  userID,
			Answers: req.GetSubmitPollingAnswer().StringArrayValue,
		}
	case *bpb.ModifyLiveLessonStateRequest_RequestRecording:
		command = &RequestRecordingCommand{}
	case *bpb.ModifyLiveLessonStateRequest_StopRecording:
		command = &StopRecordingCommand{}
	case *bpb.ModifyLiveLessonStateRequest_Spotlight_:
		spotlightedUser := req.GetSpotlight().UserId
		spotlightCommand := &SpotlightCommand{
			SpotlightedUser: spotlightedUser,
		}
		if req.GetSpotlight().GetIsSpotlight() {
			spotlightCommand.IsEnable = true
		}
		command = spotlightCommand
	case *bpb.ModifyLiveLessonStateRequest_WhiteboardZoomState_:
		zoomCommand := &WhiteboardZoomStateCommand{
			WhiteboardZoomState: getZoomWhiteboardFromRequest(req),
		}
		command = zoomCommand
	case *bpb.ModifyLiveLessonStateRequest_AnnotationDisableAll:
		command = &DisableAllAnnotationCommand{}
	case *bpb.ModifyLiveLessonStateRequest_ChatEnable:
		Learners := req.GetChatEnable().Learners
		command = &UpdateChatCommand{
			UserIDs: Learners,
			State: &UserChat{
				Value: true,
			},
		}
	case *bpb.ModifyLiveLessonStateRequest_ChatDisable:
		Learners := req.GetChatDisable().Learners
		command = &UpdateChatCommand{
			UserIDs: Learners,
			State: &UserChat{
				Value: false,
			},
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("unhandled state type %T", req.Command))
	}
	command.initBasicData(userID, req.Id)
	permissionChecker := &CommandPermissionChecker{
		lesson:   lesson,
		DB:       l.DB,
		UserRepo: l.UserRepo,
	}
	commandDp := &CommandDispatcher{
		cp:                  permissionChecker,
		DB:                  l.DB,
		LessonRepo:          l.LessonRepo,
		LessonGroupRepo:     l.LessonGroupRepo,
		LessonMemberRepo:    l.LessonMemberRepo,
		LessonPollingRepo:   l.LessonPollingRepo,
		LessonRoomStateRepo: l.LessonRoomStateRepo,
	}

	if err := commandDp.Execute(ctx, command); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// log for virtual classroom
	if err = l.VirtualClassRoomLogService.LogWhenUpdateRoomState(ctx, database.Text(req.Id)); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenUpdateRoomState: could not log this activity",
			zap.String("lesson_id", req.Id),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	return &bpb.ModifyLiveLessonStateResponse{}, nil
}

func getZoomWhiteboardFromRequest(req *bpb.ModifyLiveLessonStateRequest) *domain.WhiteboardZoomState {
	zoomState := req.GetWhiteboardZoomState()
	z := &domain.WhiteboardZoomState{
		PdfScaleRatio: zoomState.GetPdfScaleRatio(),
		CenterX:       zoomState.GetCenterX(),
		CenterY:       zoomState.GetCenterY(),
		PdfWidth:      zoomState.GetPdfWidth(),
		PdfHeight:     zoomState.GetPdfHeight(),
	}

	return z
}

func (l *LessonModifierServices) ResetAllLiveLessonStatesInternal(ctx context.Context, lessonID string) error {
	// get lesson
	reader := &LessonReaderServices{
		LessonRepo: l.LessonRepo,
		DB:         l.DB,
	}
	lesson, err := reader.getLiveLesson(ctx, database.Text(lessonID), includeLearnerIDs(), includeTeacherIDs())
	if err != nil {
		return err
	}

	permissionChecker := &CommandPermissionChecker{
		lesson:   lesson,
		DB:       l.DB,
		UserRepo: l.UserRepo,
	}

	commandDp := &CommandDispatcher{
		cp:                  permissionChecker,
		DB:                  l.DB,
		LessonRepo:          l.LessonRepo,
		LessonGroupRepo:     l.LessonGroupRepo,
		LessonMemberRepo:    l.LessonMemberRepo,
		LessonRoomStateRepo: l.LessonRoomStateRepo,
	}

	userID := interceptors.UserIDFromContext(ctx)
	if err := commandDp.Execute(
		ctx,
		&ResetAllStatesCommand{
			CommanderID: userID,
			LessonID:    lessonID,
		},
	); err != nil {
		return err
	}

	return nil
}
