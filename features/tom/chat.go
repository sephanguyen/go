package tom

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	legacybpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
)

func (s *suite) aInvalidConversationID(ctx context.Context) (context.Context, error) {
	s.conversationID = idutil.ULIDNow()

	return ctx, nil
}
func (s *suite) aSendAChatMessageToConversation(ctx context.Context, user string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var token string
	switch user {
	case "student":
		token = s.studentToken
	case "teacher":
		token = s.TeacherToken
	default:
	}
	time.Sleep(time.Second * 5)

	s.RequestAt = time.Now()
	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).SendMessage(contextWithToken(ctx2, token), s.Request.(*pb.SendMessageRequest))

	return ctx, nil
}
func (s *suite) aSendMessageRequest(ctx context.Context) (context.Context, error) {
	s.Request = &pb.SendMessageRequest{
		ConversationId: s.conversationID,
		Message:        "Hello",
		UrlMedia:       "",
		Type:           pb.MESSAGE_TYPE_TEXT,
		LocalMessageId: idutil.ULIDNow(),
	}

	return ctx, nil
}
func (s *suite) aUserGoToChat(ctx context.Context) (context.Context, error) {
	s.StreamClient, s.ResponseErr = pb.NewChatServiceClient(s.Conn).Subscribe(metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", s.studentToken), &pb.SubscribeRequest{})
	return ctx, nil
}
func (s *suite) aValidConversationID(ctx context.Context) (context.Context, error) {
	s.conversationID = idutil.ULIDNow()

	conversation := new(entities.Conversation)
	database.AllNullEntity(conversation)
	err := multierr.Combine(
		conversation.ID.Set(s.conversationID),
		conversation.Status.Set(pb.CONVERSATION_STATUS_NONE.String()),
	)
	if err != nil {
		return ctx, err
	}

	conversationRepo := &repositories.ConversationRepo{}
	err = conversationRepo.Create(ctx, s.DB, conversation)
	if err != nil {
		return ctx, err
	}

	e := &entities.ConversationMembers{}
	err = multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.UserID.Set(s.studentID),
		e.Role.Set(entities.ConversationRoleStudent),
		e.Status.Set(entities.ConversationStatusActive),
		e.SeenAt.Set(nil),
		e.LastNotifyAt.Set(nil),
	)
	if err != nil {
		return ctx, err
	}

	e.ConversationID = conversation.ID
	conversationMemberRepo := &repositories.ConversationMemberRepo{}
	err = conversationMemberRepo.Create(ctx, s.DB, e)
	if err != nil {
		return ctx, err
	}

	e = &entities.ConversationMembers{}
	err = multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.UserID.Set(s.teacherID),
		e.Role.Set(entities.ConversationRoleTeacher),
		e.Status.Set(entities.ConversationStatusActive),
		e.SeenAt.Set(nil),
		e.LastNotifyAt.Set(nil),
	)
	if err != nil {
		return ctx, err
	}

	e.ConversationID = conversation.ID

	err = conversationMemberRepo.Create(ctx, s.DB, e)
	if err != nil {
		return ctx, err
	}

	s.Request = conversation

	return ctx, nil
}
func (s *suite) aDeviceTokenIsExistedInDB(ctx context.Context, arg1 string) (context.Context, error) {
	rand.Seed(time.Now().UnixNano())

	var userID string
	// nolint
	switch arg1 {
	case "student":
		userID = s.studentID
	}

	repo := &repositories.UserDeviceTokenRepo{}
	if err := repo.Upsert(ctx, s.DB, &entities.UserDeviceToken{
		UserID:            pgtype.Text{String: userID, Status: pgtype.Present},
		Token:             pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
		AllowNotification: pgtype.Bool{Bool: true, Status: pgtype.Present},
		UserName:          pgtype.Text{String: arg1 + userID, Status: pgtype.Present},
	}); err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) userHasNotSeenTheMessageInADuration(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second) // buffer +3s to make sure the code has a chance to update to the DB
	return ctx, nil
}
func (s *suite) tomShouldPushNotificationToThe(ctx context.Context, arg1 string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var userID string
	// nolint
	switch arg1 {
	case "student":
		userID = s.studentID
	}

	lastNotifyAt := new(pgtype.Timestamptz)
	row := s.DB.QueryRow(ctx2, "SELECT last_notify_at FROM conversation_members WHERE user_id = $1", userID)

	if err := row.Scan(lastNotifyAt); err != nil {
		return ctx, err
	}
	if lastNotifyAt.Status != pgtype.Present {
		return ctx, errors.New("last_notify_at must be preset")
	}

	return ctx, nil
}
func (s *suite) tomWillCloseStreamWhenAUserResubscribe(ctx context.Context) (context.Context, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	var err error
	go func() {
		defer wg.Done()
		for {
			_, err = s.StreamClient.Recv()
			if err == io.EOF {
				err = nil
				break
			}
		}
	}()

	_, _ = pb.NewChatServiceClient(s.Conn).Subscribe(metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", s.studentToken), &pb.SubscribeRequest{})

	wg.Wait()
	return ctx, err
}
func (s *suite) aHasNotSeenTheMessageInADuration(ctx context.Context, arg1 string) (context.Context, error) {
	time.Sleep(3 * time.Second) // buffer +3s to make sure the code has a chance to update to the DB
	return ctx, nil
}
func (s *suite) aSeeAllTheMessages(ctx context.Context, arg1 string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var (
		userID string
		cond   string
		token  string
	)
	switch arg1 {
	case "student":
		cond = "user_id = $1 AND role = '" + entities.ConversationRoleStudent + "'"
		userID = s.studentID
		token = s.studentToken
	case "teacher":
		cond = "user_id = $1 AND role = '" + entities.ConversationRoleTeacher + "'"
		userID = s.teacherID
		token = s.TeacherToken
	}

	query := fmt.Sprintf("SELECT conversation_id FROM conversation_members WHERE %s", cond)
	rows, err := s.DB.Query(ctx2, query, userID)
	if err != nil {
		return ctx, err
	}
	defer rows.Close()

	var convIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return ctx, err
		}
		convIDs = append(convIDs, id)
	}

	for _, id := range convIDs {
		if _, err := pb.NewChatServiceClient(s.Conn).SeenMessage(common.ValidContext(ctx, int(s.getSchool()), userID, token), &pb.SeenMessageRequest{
			ConversationId: id,
		}); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) tomShouldNotPushNotificationToThe(ctx context.Context, arg1 string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var userID string
	// nolint
	switch arg1 {
	case "student":
		userID = s.studentID
	}

	lastNotifyAt := new(pgtype.Timestamptz)
	row := s.DB.QueryRow(ctx2, "SELECT last_notify_at FROM conversation_members WHERE user_id = $1", userID)

	if err := row.Scan(lastNotifyAt); err != nil {
		return ctx, err
	}
	if lastNotifyAt.Status != pgtype.Null {
		return ctx, errors.New("expected last_notify_at is nil")
	}

	return ctx, nil
}
func (s *suite) aListOfMessagesWithTypes(ctx context.Context, msgTypes string) (context.Context, error) {
	_, err := s.aLessonConversationWithTeachersAndStudents(ctx, 0, 0)
	if err != nil {
		return ctx, err
	}
	id := idutil.ULIDNow()
	for _, typ := range strings.Split(msgTypes, ",") {
		switch typ {
		case pb.CODES_MESSAGE_TYPE_JOINED_LESSON.String():
			req := &legacybpb.EvtLesson{
				Message: &legacybpb.EvtLesson_JoinLesson_{
					JoinLesson: &legacybpb.EvtLesson_JoinLesson{
						LessonId:  s.lessonID,
						UserGroup: legacybpb.USER_GROUP_TEACHER,
						UserId:    id,
					},
				},
			}
			s.Request = req
		case pb.CODES_MESSAGE_TYPE_END_LIVE_LESSON.String():
			req := &legacybpb.EvtLesson{
				Message: &legacybpb.EvtLesson_EndLiveLesson_{
					EndLiveLesson: &legacybpb.EvtLesson_EndLiveLesson{
						LessonId: s.lessonID,
						UserId:   id,
					},
				},
			}
			s.Request = req
		case pb.CODES_MESSAGE_TYPE_LEFT_LESSON.String():
			req := &legacybpb.EvtLesson{
				Message: &legacybpb.EvtLesson_LeaveLesson_{
					LeaveLesson: &legacybpb.EvtLesson_LeaveLesson{
						LessonId: s.lessonID,
						UserId:   id,
					},
				},
			}
			s.Request = req
		}
		ctx, err = s.bobSendEventEvtLesson(ctx)
		if err != nil {
			return ctx, err
		}
		time.Sleep(1 * time.Second)
	}
	return ctx, nil
}
func (s *suite) responseDoesNotIncludeSystemMessage(ctx context.Context) (context.Context, error) {
	resp := s.Response.(*pb.ConversationDetailResponse)
	for _, msg := range resp.Messages {
		if msg.Type == pb.MessageType(tpb.MessageType_MESSAGE_TYPE_SYSTEM) {
			return ctx, fmt.Errorf("API still return system message")
		}
	}
	return ctx, nil
}
func (s *suite) clientCallingConversationDetail(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	ctx, err := s.aValidToken(ctx, "current teacher")

	if err != nil {
		return ctx, err
	}
	token := s.TeacherToken
	convID := s.LessonChatState.LessonConversationMap[s.lessonID]

	err = try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)

		s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).ConversationDetail(contextWithToken(ctx, token), &pb.ConversationDetailRequest{
			ConversationId: convID,
			Limit:          10,
		})

		if s.ResponseErr != nil {
			return attempt < 10, s.ResponseErr
		}

		return false, nil
	})

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) createAValidStudentConversationInDBWithATeacherAndAStudent(ctx context.Context) (context.Context, error) {
	ctx, err := s.aChatBetweenAStudentAndTeachers(ctx, 1)
	if err != nil {
		return ctx, fmt.Errorf("aChatBetweenAStudentAndTeachers %w", err)
	}
	if err != nil {
		return ctx, err
	}

	var cID pgtype.Text
	_ = cID.Set(s.conversationID)

	repo := &repositories.ConversationMemberRepo{}
	conversationMembers, err := repo.FindByConversationID(ctx, s.DB, cID)
	if err != nil {
		return ctx, err
	}

	for _, m := range conversationMembers {
		s.ConversationMembers[cID.String] = append(s.ConversationMembers[cID.String], m)
	}

	return ctx, nil
}
