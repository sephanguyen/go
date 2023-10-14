package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	serviceConstants "github.com/manabie-com/backend/internal/virtualclassroom/constants"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type LiveLessonCommand struct {
	WrapperDBConnection      *support.WrapperDBConnection
	VideoTokenSuffix         string
	MaximumLearnerStreamings int

	WhiteboardSvc     infrastructure.WhiteboardPort
	AgoraTokenSvc     infrastructure.AgoraTokenPort
	VirtualLessonRepo infrastructure.VirtualLessonRepo
	LessonMemberRepo  infrastructure.LessonMemberRepo
	ActivityLogRepo   infrastructure.ActivityLogRepo
	CourseRepo        infrastructure.CourseRepo
	StudentsRepo      infrastructure.StudentsRepo
}

type JoinLiveLessonResponse struct {
	StreamToken          string
	WhiteboardToken      string
	RoomID               string
	StmToken             string
	VideoToken           string
	ScreenRecordingToken string
	UserGroup            string
}

func (l *LiveLessonCommand) JoinLiveLesson(ctx context.Context, lessonID, userID string) (*JoinLiveLessonResponse, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	// validate lesson
	lesson, err := l.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID)
	if err != nil {
		return nil, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", lessonID, err)
	}

	isUserAStudent, err := l.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
	}
	userGroup := constant.UserGroupTeacher
	if isUserAStudent {
		userGroup = constant.UserGroupStudent
	}

	// check user permission
	if userGroup == constant.UserGroupStudent {
		studentSubscribePermission, err := l.checkStreamSubscriberPermission(ctx, conn, lessonID, userID)
		if err != nil {
			return nil, err
		}
		if !studentSubscribePermission {
			return nil, fmt.Errorf("student %s not allowed to join lesson %s", userID, lessonID)
		}
	} else {
		userPublishPermission, err := l.checkStreamPublisherPermission(ctx, conn, lesson.CourseID)
		if err != nil {
			return nil, err
		}
		if !userPublishPermission {
			return nil, fmt.Errorf("user %s cannot retrieve stream token for lesson %s due to course (%s) is not valid", userID, lessonID, lesson.CourseID)
		}
	}

	// generate room ID and get whiteboard token
	roomUUID, whiteboardToken, err := l.getRoomIDAndWhiteboardToken(ctx, conn, lessonID, lesson.RoomID)
	if err != nil {
		return nil, fmt.Errorf("error in getRoomIDAndWhiteboardToken: %w", err)
	}

	// get tokens
	var streamToken, videoToken, shareForRecordingToken, rtmToken string
	if userGroup == constant.UserGroupStudent {
		streamToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(lessonID, userID, domain.RoleSubscriber)
		if err != nil {
			return nil, fmt.Errorf("error retrieve subscribe token (user: %s, lesson:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, lessonID, err)
		}
	} else {
		streamToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(lessonID, userID, domain.RolePublisher)
		if err != nil {
			return nil, fmt.Errorf("error retrieve broadcast token (user: %s, lesson:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, lessonID, err)
		}
		if streamToken == "" {
			return nil, fmt.Errorf("cannot get token for room uuid: %q", roomUUID)
		}

		videoToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(lessonID, userID+l.VideoTokenSuffix, domain.RolePublisher)
		if err != nil {
			return nil, fmt.Errorf("error retrieve video broadcast token (user: %s, lesson:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID+l.VideoTokenSuffix, lessonID, err)
		}

		shareForRecordingToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(lessonID, userID+"-streamforcloudrecording", domain.RolePublisher)
		if err != nil {
			logger := ctxzap.Extract(ctx)
			logger.Error(
				"could not generate token streamforcloudrecording: AgoraTokenSrv.GenerateAgoraStreamToken: ",
				zap.String("lesson_id", lessonID),
				zap.String("user_ID", userID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("could not generate token streamforcloudrecording (user: %s, lesson:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, lessonID, err)
		}
	}

	rtmToken, err = l.AgoraTokenSvc.BuildRTMToken(lessonID, userID)
	if err != nil {
		return nil, fmt.Errorf("could not generate RTM token (user: %s, lesson:%s): AgoraTokenSrv.BuildRTMToken: %w", userID, lessonID, err)
	}

	return &JoinLiveLessonResponse{
		StreamToken:          streamToken,
		WhiteboardToken:      whiteboardToken,
		RoomID:               roomUUID,
		VideoToken:           videoToken,
		StmToken:             rtmToken,
		ScreenRecordingToken: shareForRecordingToken,
		UserGroup:            userGroup, // user group return is only either teacher or student
	}, nil
}

func (l *LiveLessonCommand) CanLeaveLiveLesson(ctx context.Context, lessonID, userID string) (bool, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	if _, err := l.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID); err != nil {
		return false, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", lessonID, err)
	}

	isUserAStudent, err := l.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
	if err != nil {
		return false, fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
	}

	contextUserID := interceptors.UserIDFromContext(ctx)
	if isUserAStudent && userID != contextUserID {
		return false, nil
	}

	return true, nil
}

func (l *LiveLessonCommand) CanEndLiveLesson(ctx context.Context, lessonID string) (bool, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	lesson, err := l.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID)
	if err != nil {
		return false, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", lessonID, err)
	}

	return l.checkStreamPublisherPermission(ctx, conn, lesson.CourseID)
}

func (l *LiveLessonCommand) EndLiveLesson(ctx context.Context, lessonID string) error {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	if err := l.VirtualLessonRepo.EndLiveLesson(ctx, conn, lessonID, time.Now()); err != nil {
		return fmt.Errorf("error in VirtualLessonRepo.EndLiveLesson, lesson %s: %w", lessonID, err)
	}

	return nil
}

func (l *LiveLessonCommand) GetRoomIDAndWhiteboardToken(ctx context.Context, lessonID string) (string, string, error) {
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return "", "", err
	}
	lesson, err := l.VirtualLessonRepo.GetVirtualLessonByID(ctx, conn, lessonID)
	if err != nil {
		return "", "", fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", lessonID, err)
	}

	userID := interceptors.UserIDFromContext(ctx)
	isUserAStudent, err := l.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
	if err != nil {
		return "", "", fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
	}

	// check user permission
	if isUserAStudent {
		studentSubscribePermission, err := l.checkStreamSubscriberPermission(ctx, conn, lessonID, userID)
		if err != nil {
			return "", "", err
		}
		if !studentSubscribePermission {
			return "", "", fmt.Errorf("student %s not allowed to join lesson %s", userID, lessonID)
		}
	}

	return l.getRoomIDAndWhiteboardToken(ctx, conn, lessonID, lesson.RoomID)
}

func (l *LiveLessonCommand) getRoomIDAndWhiteboardToken(ctx context.Context, db database.QueryExecer, lessonID, roomID string) (string, string, error) {
	var err error

	// generate room ID
	roomUUID := roomID
	if len(roomUUID) == 0 {
		createRoomReq := &whiteboard.CreateRoomRequest{
			Name:     lessonID,
			IsRecord: false,
		}

		room, err := l.WhiteboardSvc.CreateRoom(ctx, createRoomReq)
		if err != nil {
			return "", "", fmt.Errorf("could not create a new room for lesson %s: %v", lessonID, err)
		}
		roomUUID = room.UUID
		if err = l.VirtualLessonRepo.UpdateRoomID(ctx, db, lessonID, roomUUID); err != nil {
			return "", "", fmt.Errorf("could not update room id %s for lesson %s: VirtualLessonRepo.UpdateRoomID: %w", roomUUID, lessonID, err)
		}
	}

	// get whiteboard token
	whiteboardToken := ""
	retryCount := 0
	for {
		retryCount++
		whiteboardToken, err = l.WhiteboardSvc.FetchRoomToken(ctx, roomUUID)
		if err == nil || retryCount > 5 {
			break
		}
		ctxzap.Extract(ctx).Warn("cannot fetch whiteboard room token", zap.Error(err))

		time.Sleep(time.Duration(200*retryCount) * time.Millisecond)
		ctxzap.Extract(ctx).Warn(fmt.Sprintf("retry fetch whiteboard room token %d time", retryCount))
	}
	if err != nil {
		return "", "", fmt.Errorf("cannot fetch whiteboard room token from room ID %s: %q", roomUUID, err)
	}

	return roomUUID, whiteboardToken, nil
}

func (l *LiveLessonCommand) checkStreamSubscriberPermission(ctx context.Context, db database.QueryExecer, lessonID, userID string) (bool, error) {
	studentCourseIDs, err := l.LessonMemberRepo.GetCourseAccessible(ctx, db, userID)
	if err != nil {
		return false, fmt.Errorf("error in LessonMemberRepo.GetCourseAccessible, user %s: %w", userID, err)
	}

	lessons, err := l.VirtualLessonRepo.GetVirtualLessonByLessonIDsAndCourseIDs(ctx, db, []string{lessonID}, studentCourseIDs)
	if err != nil {
		return false, fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByLessonIDsAndCourseIDs, lesson %s: %w", lessonID, err)
	}
	if len(lessons) == 0 {
		return false, nil
	}

	return true, nil
}

func (l *LiveLessonCommand) checkStreamPublisherPermission(ctx context.Context, db database.QueryExecer, courseID string) (bool, error) {
	if len(courseID) == 0 {
		return true, nil
	}

	courses, err := l.CourseRepo.GetValidCoursesByCourseIDsAndStatus(ctx, db, []string{courseID}, domain.StatusActive)
	if err != nil {
		return false, fmt.Errorf("error in CourseRepo.GetValidCoursesByCourseIDsAndStatus, course %s: %w", courseID, err)
	}
	if len(courses) == 0 {
		return false, nil
	}

	return true, nil
}

func (l *LiveLessonCommand) PreparePublish(ctx context.Context, lessonID, learnerID string) (domain.PrepareToPublishStatus, error) {
	publishStatus := domain.PublishStatusNone
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return "", err
	}

	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		learnerIDs, e := l.VirtualLessonRepo.GetStreamingLearners(ctx, tx, lessonID, true)
		if e != nil {
			return fmt.Errorf("error in VirtualLessonRepo.GetStreamingLearners, lesson %s: %w", lessonID, e)
		}
		if sliceutils.Contains(learnerIDs, learnerID) {
			publishStatus = domain.PublishStatusPreparedBefore
			return fmt.Errorf("prepared before")
		}

		if e = l.VirtualLessonRepo.IncreaseNumberOfStreaming(ctx, tx, lessonID, learnerID, l.MaximumLearnerStreamings); e != nil {
			if e.Error() == serviceConstants.NoRowsUpdatedError {
				publishStatus = domain.PublishStatusMaxLimit
				return e
			}
			return fmt.Errorf("error in VirtualLessonRepo.IncreaseNumberOfStreaming, lesson %s learner %s: %w", lessonID, learnerID, e)
		}

		payload := map[string]interface{}{
			"lesson_id": lessonID,
		}
		if e = l.ActivityLogRepo.Create(ctx, tx, learnerID, constant.LogActionTypePublish, payload); e != nil {
			return fmt.Errorf("error in ActivityLogRepo.Create, lesson %s user %s: %w", lessonID, learnerID, e)
		}

		return nil
	})
	if err != nil {
		if err.Error() != serviceConstants.NoRowsUpdatedError && err.Error() != "prepared before" {
			return "", fmt.Errorf("ExecInTx: %w", err)
		}
	}

	return publishStatus, nil
}

func (l *LiveLessonCommand) Unpublish(ctx context.Context, lessonID, learnerID string) (domain.UnpublishStatus, error) {
	unpublishStatus := domain.UnpublishStatsNone
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return "", err
	}
	err = database.ExecInTx(ctx, conn, func(ctx context.Context, tx pgx.Tx) error {
		// used to check if the lesson exists
		_, err := l.VirtualLessonRepo.GetStreamingLearners(ctx, tx, lessonID, false)
		if err != nil {
			return fmt.Errorf("error in VirtualLessonRepo.GetStreamingLearners, lesson %s: %w", lessonID, err)
		}

		if err := l.VirtualLessonRepo.DecreaseNumberOfStreaming(ctx, tx, lessonID, learnerID); err != nil {
			if err.Error() == serviceConstants.NoRowsUpdatedError {
				unpublishStatus = domain.UnpublishStatsUnpublishedBefore
				return err
			}
			return fmt.Errorf("error in VirtualLessonRepo.DecreaseNumberOfStreaming, lesson %s learner %s: %w", lessonID, learnerID, err)
		}

		payload := map[string]interface{}{
			"lesson_id": lessonID,
		}
		if err := l.ActivityLogRepo.Create(ctx, tx, learnerID, constant.LogActionTypeUnpublish, payload); err != nil {
			return fmt.Errorf("error in ActivityLogRepo.Create, lesson %s user %s: %w", lessonID, learnerID, err)
		}

		return nil
	})
	if err != nil {
		if err.Error() != serviceConstants.NoRowsUpdatedError {
			return "", fmt.Errorf("ExecInTx: %w", err)
		}
	}

	return unpublishStatus, nil
}
