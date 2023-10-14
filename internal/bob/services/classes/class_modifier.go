package classes

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services/log"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/agoratokenbuilder"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricLabel string

const (
	MetricLabelUpdateRoomState MetricLabel = "update_room_state"
	MetricLabelGetRoomState    MetricLabel = "get_room_state"
)

type VirtualClassroomHandledMetric struct {
	HandleRoomState  *prometheus.HistogramVec
	RealLiveTime     *prometheus.HistogramVec
	TotalAttendees   *prometheus.HistogramVec
	TotalActiveRooms *prometheus.GaugeVec
}

func RegisterVirtualClassroomHandledMetric(collector metrics.MetricCollector) *VirtualClassroomHandledMetric {
	VCrMetric := &VirtualClassroomHandledMetric{
		HandleRoomState: collector.RegisterHistogram(metrics.MetricOpt{
			Name:       "backend_virtual_classroom_handled_total",
			Help:       "Total number of calling completed of every room, include update and get room state action.",
			LabelNames: []string{"action"},
		}, []float64{10, 20, 50, 100, 200, 500, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000}), // TODO: tracking metric and adjust these params after
		RealLiveTime: collector.RegisterHistogram(metrics.MetricOpt{
			Name: "backend_virtual_classroom_total_real_live_time_minutes",
			Help: "Total real live time of every room",
		}, prometheus.LinearBuckets(10, 10, 11)), // TODO: tracking metric and adjust these params after
		TotalAttendees: collector.RegisterHistogram(metrics.MetricOpt{
			Name: "backend_virtual_classroom_attendees_total",
			Help: "Total number of attendees who really joined room",
		}, prometheus.LinearBuckets(5, 5, 10)), // TODO: tracking metric and adjust these params after
		TotalActiveRooms: collector.RegisterGauge(metrics.MetricOpt{
			Name: "backend_virtual_classroom_active_rooms_total",
			Help: "Total active rooms",
		}),
	}

	return VCrMetric
}

func (v *VirtualClassroomHandledMetric) ObserveWhenEndRoom(ctx context.Context, log *entities.VirtualClassRoomLog) {
	defer func() {
		if err := recover(); err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"VirtualClassroomHandledMetric.ObserveWhenEndRoom: panic: ",
				zap.Error(err.(error)),
			)
		}
	}()
	v.HandleRoomState.WithLabelValues(string(MetricLabelUpdateRoomState)).Observe(float64(log.TotalTimesUpdatingRoomState.Int))
	v.HandleRoomState.WithLabelValues(string(MetricLabelGetRoomState)).Observe(float64(log.TotalTimesGettingRoomState.Int))
	v.TotalAttendees.WithLabelValues().Observe(float64(len(log.AttendeeIDs.Elements)))
	v.RealLiveTime.WithLabelValues().Observe(log.UpdatedAt.Time.Sub(log.CreatedAt.Time).Minutes())
	v.TotalActiveRooms.WithLabelValues().Dec()
}

func (v *VirtualClassroomHandledMetric) ObserveWhenStartRoom(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"VirtualClassroomHandledMetric.ObserveWhenStartRoom: panic: ",
				zap.Error(err.(error)),
			)
		}
	}()
	v.TotalActiveRooms.WithLabelValues().Inc()
}

type ClassModifierService struct {
	bpb.UnimplementedClassModifierServiceServer
	VirtualClassRoomLogService *log.VirtualClassRoomLogService

	DB database.Ext

	ConversionTaskRepo interface {
		CreateTasks(ctx context.Context, db database.QueryExecer, tasks []*entities.ConversionTask) error
	}

	ConversionSvc interface {
		CreateConversionTasks(ctx context.Context, resourceURLs []string) ([]string, error)
	}

	LessonRoomStateRepo interface {
		UpdateRecordingState(ctx context.Context, db database.QueryExecer, lessonID string, recording *domain.CompositeRecordingState) error
		UpsertCurrentPollingState(ctx context.Context, db database.QueryExecer, lessonID string, polling *domain.CurrentPolling) error
	}
	LessonMgmtRoomStateRepo interface {
		GetLessonRoomStateByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*lesson_domain.LessonRoomState, error)
	}

	VCrMetric       *VirtualClassroomHandledMetric
	OldClassService *ClassServiceABAC
}

func (c *ClassModifierService) ConvertMedia(ctx context.Context, req *bpb.ConvertMediaRequest) (*bpb.ConvertMediaResponse, error) {
	var urls []string
	for _, media := range req.Media {
		if media.Type == bpb.MediaType_MEDIA_TYPE_PDF {
			urls = append(urls, media.Resource)
		}
	}
	if len(urls) == 0 {
		return &bpb.ConvertMediaResponse{}, nil
	}
	urls = golibs.Uniq(urls)

	tasks, err := c.ConversionSvc.CreateConversionTasks(ctx, urls)
	if err != nil {
		return nil, err
	}

	taskEntities := make([]*entities.ConversionTask, 0, len(tasks))
	for i, task := range tasks {
		taskEntities = append(taskEntities, toConversionTaskEntity(task, urls[i]))
	}

	if err := c.ConversionTaskRepo.CreateTasks(ctx, c.DB, taskEntities); err != nil {
		return nil, err
	}

	return &bpb.ConvertMediaResponse{}, nil
}

func toConversionTaskEntity(taskUUID, resourceURL string) *entities.ConversionTask {
	e := new(entities.ConversionTask)
	database.AllNullEntity(e)
	e.TaskUUID.Set(taskUUID)
	e.ResourceURL.Set(resourceURL)
	e.Status.Set(bpb.ConversionTaskStatus_CONVERSION_TASK_STATUS_WAITING.String())

	return e
}

func (c *ClassModifierService) JoinLesson(ctx context.Context, req *bpb.JoinLessonRequest) (*bpb.JoinLessonResponse, error) {
	ol := c.OldClassService
	resp, err := ol.JoinLesson(ctx, &pb.JoinLessonRequest{
		LessonId: req.LessonId,
	})
	if err != nil {
		return nil, err
	}

	userID := interceptors.UserIDFromContext(ctx)
	rtmToken, err := agoratokenbuilder.BuildRTMToken(ol.Cfg.Agora.AppID,
		ol.Cfg.Agora.Cert,
		userID+req.LessonId,
		uint32(timeutil.Now().Unix()+3600))
	if err != nil {
		return nil, status.Error(codes.Internal, "agoratokenbuilder.BuildRTMToken: could not generate RTM token: "+err.Error())
	}

	userGroup, err := ol.UserRepo.UserGroup(ctx, c.DB, database.Text(userID))
	if err != nil {
		return nil, err
	}

	var shareForRecordingToken string
	if userGroup != entities.UserGroupStudent {
		shareForRecordingToken, err = generateAgoraStreamToken(&ol.Cfg.Agora, req.LessonId, userID+"-streamforcloudrecording", agoratokenbuilder.RolePublisher)
		if err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"agoratokenbuilder.BuildRTMToken: could not generate RTM token streamforcloudrecording: ",
				zap.String("lesson_id", req.LessonId),
				zap.String("user_ID", userID),
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "")
		}
	}

	// log for virtual classroom
	createdNewLog, err := c.VirtualClassRoomLogService.LogWhenAttendeeJoin(ctx, database.Text(req.LessonId), database.Text(userID))
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenAttendeeJoin: could not log this activity",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}
	if createdNewLog {
		c.VCrMetric.ObserveWhenStartRoom(ctx)
	}

	return &bpb.JoinLessonResponse{
		RoomId:               resp.RoomId,
		StreamToken:          resp.StreamToken,
		WhiteboardToken:      resp.WhiteboardToken,
		VideoToken:           resp.VideoToken,
		StmToken:             rtmToken,
		AgoraAppId:           ol.Cfg.Agora.AppID,
		WhiteboardAppId:      ol.Cfg.Whiteboard.AppID,
		ScreenRecordingToken: shareForRecordingToken,
	}, nil
}

func (c *ClassModifierService) RetrieveWhiteboardToken(ctx context.Context, req *bpb.RetrieveWhiteboardTokenRequest) (*bpb.RetrieveWhiteboardTokenResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	userGroup, err := c.OldClassService.UserRepo.UserGroup(ctx, c.DB, database.Text(userID))
	if err != nil {
		return nil, fmt.Errorf("rcv.UserRepo.UserGroup: %w", err)
	}
	if userGroup == entities.UserGroupStudent {
		studentSubscribePermission, err := c.OldClassService.StreamSubscriberPermission(ctx, req.LessonId, userID)
		if err != nil {
			return nil, err
		}
		if !studentSubscribePermission {
			return nil, status.Error(codes.PermissionDenied, "student not allowed to join lesson")
		}
	}

	lessons, err := c.OldClassService.LessonRepo.Find(ctx, c.DB, &repositories.LessonFilter{
		LessonID:  database.TextArray([]string{req.LessonId}),
		TeacherID: pgtype.TextArray{Status: pgtype.Null},
		CourseID:  pgtype.TextArray{Status: pgtype.Null},
	})
	if err != nil {
		return nil, err
	}
	if len(lessons) == 0 {
		return nil, fmt.Errorf("invalid lesson: %q", req.LessonId)
	}

	roomUUID := lessons[0].RoomID.String
	if len(roomUUID) == 0 {
		room, err := c.OldClassService.WhiteboardSvc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
			Name:     lessons[0].LessonID.String,
			IsRecord: false,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create a new room for lesson %s: %v", lessons[0].LessonID.String, err)
		}
		roomUUID = room.UUID
		if err = c.OldClassService.LessonRepo.UpdateRoomID(ctx, c.DB, lessons[0].LessonID, database.Text(roomUUID)); err != nil {
			return nil, fmt.Errorf("could not update room id: LessonRepo.UpdateRoomID: %w", err)
		}
	}
	whiteBoardToken := ""
	retryCount := 0
	for {
		retryCount += 1
		whiteBoardToken, err = c.OldClassService.WhiteboardSvc.FetchRoomToken(ctx, roomUUID)
		if err == nil || retryCount > 5 {
			break
		}
		ctxzap.Extract(ctx).Warn("cannot fetch whiteboard room token ", zap.Error(err))
		time.Sleep(time.Duration(200*retryCount) * time.Millisecond)
		ctxzap.Extract(ctx).Warn(fmt.Sprintf("retry fetch whiteboard room token %d time", retryCount))
	}
	if err != nil {
		return nil, fmt.Errorf("cannot fetch whiteboard room token: %q", err)
	}

	return &bpb.RetrieveWhiteboardTokenResponse{
		RoomId:          roomUUID,
		WhiteboardToken: whiteBoardToken,
		WhiteboardAppId: c.OldClassService.Cfg.Whiteboard.AppID,
	}, nil
}

func (c *ClassModifierService) EndLiveLesson(ctx context.Context, req *bpb.EndLiveLessonRequest) (*bpb.EndLiveLessonResponse, error) {
	_, err := c.OldClassService.EndLiveLesson(ctx, &pb.EndLiveLessonRequest{
		LessonId: req.LessonId,
	})
	if err != nil {
		return nil, err
	}
	_, err = c.LessonMgmtRoomStateRepo.GetLessonRoomStateByLessonID(ctx, c.DB, database.Text(req.GetLessonId()))
	if err != nil && err != lesson_domain.ErrNotFound {
		return nil, fmt.Errorf("l.LessonMgmtRoomStateRepo.GetLessonRoomStateByLessonID: %w", err)
	}
	if err != lesson_domain.ErrNotFound {
		err = c.LessonRoomStateRepo.UpdateRecordingState(ctx, c.DB, req.GetLessonId(), nil)
		if err != nil {
			return nil, fmt.Errorf("l.LessonRoomStateRepo.UpdateRecordingState: %w", err)
		}

		err = c.LessonRoomStateRepo.UpsertCurrentPollingState(ctx, c.DB, req.GetLessonId(), nil)
		if err != nil {
			return nil, fmt.Errorf("l.LessonRoomStateRepo.UpdateRecordingState: %w", err)
		}
	}

	// logging and observe metrics
	// make sure not observe one log twice
	exceptedID := ""
	log, _ := c.VirtualClassRoomLogService.GetCompletedLogByLesson(ctx, database.Text(req.LessonId))
	if log != nil {
		exceptedID = log.LogID.String
	}
	// log for virtual classroom
	if err = c.VirtualClassRoomLogService.LogWhenEndRoom(ctx, database.Text(req.LessonId)); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenEndRoom: could not log this activity",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", interceptors.UserIDFromContext(ctx)),
			zap.Error(err),
		)
	}

	log, err = c.VirtualClassRoomLogService.GetCompletedLogByLesson(ctx, database.Text(req.LessonId))
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.GetCompletedLogByLesson: could not get log",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", interceptors.UserIDFromContext(ctx)),
			zap.Error(err),
		)
	}
	if log != nil && exceptedID != log.LogID.String {
		c.VCrMetric.ObserveWhenEndRoom(ctx, log)
	}

	return &bpb.EndLiveLessonResponse{}, nil
}

func (c *ClassModifierService) LeaveLesson(ctx context.Context, req *bpb.LeaveLessonRequest) (*bpb.LeaveLessonResponse, error) {
	_, err := c.OldClassService.LeaveLesson(ctx, &pb.LeaveLessonRequest{
		LessonId: req.LessonId,
	})
	if err != nil {
		return nil, err
	}

	return &bpb.LeaveLessonResponse{}, nil
}

func (c *ClassModifierService) RegisterMetric(collector metrics.MetricCollector) {
	c.VCrMetric = RegisterVirtualClassroomHandledMetric(collector)
}
