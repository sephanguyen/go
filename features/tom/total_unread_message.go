package tom

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"go.uber.org/multierr"
)

func (s *suite) tomMustReturnsTotalUnreadMessageInLocations(ctx context.Context, expect int, locLabels string) (context.Context, error) {
	locIDs := s.getLocations(locLabels)
	ret, err := tpb.NewChatReaderServiceClient(s.Conn).RetrieveTotalUnreadConversationsWithLocations(contextWithToken(ctx, s.TeacherToken), &tpb.RetrieveTotalUnreadConversationsWithLocationsRequest{
		LocationIds: locIDs,
	})
	if err != nil {
		return ctx, err
	}
	if expect != int(ret.GetTotalUnreadConversations()) {
		return ctx, fmt.Errorf("teacher has %d total unread conversations with locations %v, expect %d", ret.GetTotalUnreadConversations(), locIDs, expect)
	}
	return ctx, nil
}

func (s *suite) chatsEachHasNewMessageFromStudentOrParent(ctx context.Context, chatLabels string) (context.Context, error) {
	labels := strings.Split(chatLabels, ",")
	for _, label := range labels {
		info := s.getChatInfoFromPool(label)
		user := info.getOneUserID()
		m := &entities.Message{}
		database.AllNullEntity(m)
		err := multierr.Combine(
			m.ID.Set(idutil.ULIDNow()),
			m.ConversationID.Set(info.id),
			m.UserID.Set(user),
			m.Message.Set("Hello student"),
			m.Type.Set(pb.MESSAGE_TYPE_TEXT.String()),
			m.CreatedAt.Set(time.Now()),
			m.UpdatedAt.Set(time.Now()),
		)
		if err != nil {
			return ctx, err
		}
		_, err = repositories.Insert(ctx, m, s.DB.Exec)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) studentSendMessageToTeacher(ctx context.Context, messageNum int) (context.Context, error) {
	return s.hackOneMessageInAboveConversation(ctx, messageNum)
}
func (s *suite) teacherJoinedSomeConversationInSchool(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.aSignedAsATeacher,
		s.randomNewConversationsCreated,
		s.teacherJoinAllConversations,
	)
}
func (s *suite) tomMustReturnsTotalUnreadMessage(ctx context.Context, expectedTotalMessage int) (context.Context, error) {
	rsp := s.Response.(*tpb.RetrieveTotalUnreadMessageResponse)
	if expectedTotalMessage != int(rsp.TotalUnreadMessages) {
		return ctx, fmt.Errorf("expected %d messages got %d", expectedTotalMessage, rsp.TotalUnreadMessages)
	}
	return ctx, nil
}
func (s *suite) teacherReadAllMessages(ctx context.Context) (context.Context, error) {
	return s.aSeeAllTheMessages(ctx, "teacher")
}
func (s *suite) getTotalUnreadMessage(ctx context.Context, user string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var userID, token string
	switch user {
	case "teacher":
		userID = s.teacherID
		token = s.TeacherToken
	case "student":
		userID = s.studentID
		token = s.studentToken
	default:
		userID = ""
		token = ""
	}

	req := &tpb.RetrieveTotalUnreadMessageRequest{
		UserId: userID,
	}

	s.Request = req
	res, err := tpb.NewChatReaderServiceClient(s.Conn).RetrieveTotalUnreadMessage(contextWithToken(ctx2, token), req)
	if err != nil {
		return ctx, err
	}
	s.Response = res
	return ctx, nil
}
func (s *suite) hackOneMessageInAboveConversation(ctx context.Context, count int) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	for i := 0; i < count; i++ {
		m := &entities.Message{}
		database.AllNullEntity(m)
		err := multierr.Combine(
			m.ID.Set(idutil.ULIDNow()),
			m.ConversationID.Set(s.ConversationIDs[0]),
			m.UserID.Set(s.teacherID),
			m.Message.Set("Hello student"),
			m.Type.Set(pb.MESSAGE_TYPE_TEXT.String()),
			m.CreatedAt.Set(time.Now()),
			m.UpdatedAt.Set(time.Now()),
		)
		if err != nil {
			return ctx, err
		}
		_, err = repositories.Insert(ctx2, m, s.DB.Exec)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}
