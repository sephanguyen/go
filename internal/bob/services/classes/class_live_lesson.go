package classes

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/agoratokenbuilder"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (rcv *ClassService) RetrieveBroadCastToken(ctx context.Context, lessonID string, userID string) (string, error) {
	result, err := generateAgoraStreamToken(&rcv.Cfg.Agora, lessonID, userID, agoratokenbuilder.RolePublisher)
	if err != nil {
		return "", fmt.Errorf("error create token: %v", err)
	}

	return result, nil
}

func (rcv *ClassService) RetrieveSubscribeToken(ctx context.Context, lessonID string, userID string) (string, error) {
	result, err := generateAgoraStreamToken(&rcv.Cfg.Agora, lessonID, userID, agoratokenbuilder.RoleSubscriber)
	if err != nil {
		return "", fmt.Errorf("error create token: %v", err)
	}
	return result, nil
}

func generateAgoraStreamToken(c *configurations.AgoraConfig, lessonID string, userID string, role agoratokenbuilder.Role) (string, error) {
	expireTimestamp := uint32(time.Now().UTC().Unix() + 21600)
	// Using lesson id channel name
	result, err := agoratokenbuilder.BuildStreamToken(c.AppID, c.Cert, lessonID, userID, role, expireTimestamp)
	if err != nil {
		return "", fmt.Errorf("error create token: %v", err)
	}
	return result, nil
}

func (rcv *ClassService) EndLiveLesson(ctx context.Context, req *pb.EndLiveLessonRequest) (*pb.EndLiveLessonResponse, error) {
	var endTime pgtype.Timestamptz
	err := endTime.Set(timeutil.Now())
	if err != nil {
		return nil, err
	}

	err = rcv.LessonRepo.EndLiveLesson(ctx, rcv.DB, database.Text(req.LessonId), endTime)
	if err != nil {
		return nil, services.ToStatusError(err)
	}

	if err := rcv.LessonModifierServices.ResetAllLiveLessonStatesInternal(ctx, req.LessonId); err != nil {
		return nil, services.ToStatusError(err)
	}

	if err := rcv.PublishLessonEvt(ctx, &pb.EvtLesson{
		Message: &pb.EvtLesson_EndLiveLesson_{
			EndLiveLesson: &pb.EvtLesson_EndLiveLesson{
				LessonId: req.LessonId,
				UserId:   interceptors.UserIDFromContext(ctx),
			},
		},
	}); err != nil {
		ctxzap.Extract(ctx).Warn("rcv.PublishLessonEvt", zap.Error(err))
	}
	return &pb.EndLiveLessonResponse{}, nil
}

func (rcv *ClassService) PublishLessonEvt(ctx context.Context, msg *pb.EvtLesson) error {
	var msgID string
	data, _ := msg.Marshal()

	msgID, err := rcv.JSM.PublishAsyncContext(ctx, constants.SubjectLessonUpdated, data)
	if err != nil {
		return services.HandlePushMsgFail(ctx, msg, fmt.Errorf("PublishLessonEvt rcv.JSM.PublishAsyncContext subject Lesson.Updated failed, msgID: %s, %w", msgID, err))
	}
	return err
}

func (rcv *ClassService) handleTeacherTokenPermission(ctx context.Context, lessonID string, userID string) (string, string, error) {
	allowedToBroadCast, err := rcv.LiveTeacherPermission(ctx, lessonID, userID)
	if err != nil {
		return "", "", err
	}
	if !allowedToBroadCast {
		return "", "", status.Error(codes.PermissionDenied, "teacher cannot retrieve stream token")
	}
	token, err := rcv.RetrieveBroadCastToken(ctx, lessonID, userID)
	if err != nil {
		return "", "", status.Error(codes.PermissionDenied, "error retrieve broad cast token")
	}
	videoToken, err := rcv.RetrieveBroadCastToken(ctx, lessonID, userID+rcv.Cfg.Agora.VideTokenSuffix)
	if err != nil {
		return "", "", status.Error(codes.PermissionDenied, "error retrieve video broad cast token")
	}
	return token, videoToken, nil
}

type CreateWhiteboardRoomRequest struct {
	Name  string `json:"name"`
	Limit int    `json:"limit"`
	Mode  string `json:"mode"`
}

type WhiteBoardRoom struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Limit     int64  `json:"limit"`
	AdminID   int64  `json:"adminId"`
	Mode      string `json:"mode"`
	Template  string `json:"template"`
	Region    string `json:"region"`
	Uuid      string `json:"uuid"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}

type CreateWhiteboardMsg struct {
	Room      WhiteBoardRoom `json:"room"`
	RoomToken string         `json:"roomToken"`
	Codes     int64          `json:"code"`
}

type CreateWhiteboardResponse struct {
	Code int64               `json:"code"`
	Msg  CreateWhiteboardMsg `json:"msg"`
}

func (rcv *ClassService) JoinLesson(ctx context.Context, req *pb.JoinLessonRequest) (*pb.JoinLessonResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)

	userGroup, err := rcv.UserRepo.UserGroup(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return nil, err
	}

	rsp := &pb.JoinLessonResponse{}
	lessons, err := rcv.LessonRepo.Find(ctx, rcv.DB, &repositories.LessonFilter{
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
		room, err := rcv.WhiteboardSvc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
			Name:     lessons[0].LessonID.String,
			IsRecord: false,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create a new room for lesson %s: %v", lessons[0].LessonID.String, err)
		}
		roomUUID = room.UUID
		if err = rcv.LessonRepo.UpdateRoomID(ctx, rcv.DB, lessons[0].LessonID, database.Text(roomUUID)); err != nil {
			return nil, fmt.Errorf("could not update room id: LessonRepo.UpdateRoomID: %w", err)
		}
	}
	whiteBoardToken := ""
	retryCount := 0
	for {
		retryCount += 1
		whiteBoardToken, err = rcv.WhiteboardSvc.FetchRoomToken(ctx, roomUUID)
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

	// return whiteboard token for both teacher and student
	rsp.WhiteboardToken = whiteBoardToken

	// return roomUUID for both teacher and student
	rsp.RoomId = roomUUID

	if userGroup == entities.UserGroupStudent {
		token, err := rcv.RetrieveSubscribeToken(ctx, req.LessonId, userID)
		if err != nil {
			return nil, err
		}
		rsp.StreamToken = token
	} else {
		token, videoToken, err := rcv.handleTeacherTokenPermission(ctx, req.LessonId, userID)
		if err != nil {
			return nil, err
		}

		if token == "" {
			return nil, fmt.Errorf("cannot get token for room uuid: %q", roomUUID)
		}

		rsp.StreamToken = token
		rsp.VideoToken = videoToken
	}

	err = rcv.PublishLessonEvt(ctx, &pb.EvtLesson{
		Message: &pb.EvtLesson_JoinLesson_{
			JoinLesson: &pb.EvtLesson_JoinLesson{
				LessonId:  req.LessonId,
				UserGroup: pb.UserGroup(pb.UserGroup_value[userGroup]),
				UserId:    userID,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error joining lesson: %w", err)
	}

	return rsp, nil
}

func (rcv *ClassService) LeaveLesson(ctx context.Context, req *pb.LeaveLessonRequest) (*pb.LeaveLessonResponse, error) {
	err := rcv.PublishLessonEvt(ctx, &pb.EvtLesson{
		Message: &pb.EvtLesson_LeaveLesson_{
			LeaveLesson: &pb.EvtLesson_LeaveLesson{
				LessonId: req.LessonId,
				UserId:   req.UserId,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error leave lesson: %w", err)
	}
	return &pb.LeaveLessonResponse{}, nil
}

func (rcv *ClassService) TeacherRetrieveStreamToken(ctx context.Context, req *pb.TeacherRetrieveStreamTokenRequest) (*pb.TeacherRetrieveStreamTokenResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	token, videoToken, err := rcv.handleTeacherTokenPermission(ctx, req.LessonId, userID)
	if err != nil {
		return nil, err
	}
	return &pb.TeacherRetrieveStreamTokenResponse{
		StreamToken: token,
		VideoToken:  videoToken,
	}, nil
}

func (rcv *ClassService) LiveTeacherPermission(ctx context.Context, lessonID string, teacherID string) (bool, error) {
	filter := &repositories.LessonFilter{}
	err := multierr.Combine(
		filter.LessonID.Set([]string{lessonID}),
		filter.TeacherID.Set(nil),
		filter.CourseID.Set(nil),
	)
	if err != nil {
		return false, err
	}
	lessons, err := rcv.LessonRepo.Find(ctx, rcv.DB, filter)
	if err != nil {
		return false, services.ToStatusError(err)
	}
	if len(lessons) != 1 {
		return false, nil
	}

	// Find courses belong to schools
	validCourses, err := rcv.CourseRepo.RetrieveCourses(ctx, rcv.DB, &repositories.CourseQuery{
		Status: pb.COURSE_STATUS_ACTIVE.String(),
		IDs:    []string{lessons[0].CourseID.String},
	})
	if err != nil {
		return false, err
	}
	found := len(validCourses) > 0
	return found, nil
}

func (rcv *ClassService) StreamSubscriberPermission(ctx context.Context, lessonID, userID string) (bool, error) {
	studentCourseIDs, err := rcv.LessonMemberRepo.CourseAccessible(ctx, rcv.DB, database.Text(userID))
	if err != nil {
		return false, fmt.Errorf("err rcv.LessonMemberRepo.CourseAccessible: %w", err)
	}

	filter := &repositories.LessonFilter{}
	err = multierr.Combine(
		filter.LessonID.Set([]string{lessonID}),
		filter.TeacherID.Set(nil),
		filter.CourseID.Set(studentCourseIDs),
	)

	if err != nil {
		return false, err
	}
	lessons, err := rcv.LessonRepo.Find(ctx, rcv.DB, filter)
	if err != nil {
		return false, fmt.Errorf("err rcv.LessonRepo.Find: %w", err)
	}
	if len(lessons) == 0 {
		return false, nil
	}
	return true, nil
}

func (rcv *ClassService) StudentRetrieveStreamToken(ctx context.Context, req *pb.StudentRetrieveStreamTokenRequest) (*pb.StudentRetrieveStreamTokenResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)

	token, err := rcv.RetrieveSubscribeToken(ctx, req.LessonId, userID)
	if err != nil {
		return nil, err
	}
	return &pb.StudentRetrieveStreamTokenResponse{
		StreamToken: token,
	}, nil
}

func (rcv *ClassService) SyncStudentLesson(ctx context.Context, req []*npb.EventSyncUserCourse_StudentLesson) error {
	for _, r := range req {
		switch r.ActionKind {
		case npb.ActionKind_ACTION_KIND_UPSERTED:
			rcv.JoinLesson(ctx, &pb.JoinLessonRequest{})
		case npb.ActionKind_ACTION_KIND_DELETED:
		}
	}

	return nil
}
