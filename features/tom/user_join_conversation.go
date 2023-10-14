package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) teacherJoinConversations(ctx context.Context) (context.Context, error) {
	var err error
	ctx, err = s.userJoinConversations(ctx, "teacher")
	if s.ResponseErr != nil {
		err = s.ResponseErr
	}
	return ctx, err
}

func (s *suite) userJoinConversations(ctx context.Context, role string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &tpb.JoinConversationsRequest{
		ConversationIds: []string{
			s.conversationID,
		},
	}
	s.ConversationIDs = req.ConversationIds
	s.Request = req

	var token string
	switch role {
	case "teacher":
		token = s.TeacherToken
	case "school admin":
		token = s.schoolAdminToken
	case "student":
		token = s.studentToken
	case "parent":
		token = s.parentToken
	default:
		return ctx, fmt.Errorf("not handle %s role to join conversation yet", role)
	}

	s.ResponseErr = try.Do(func(attempt int) (bool, error) {
		s.Response, s.ResponseErr = tpb.NewChatModifierServiceClient(s.Conn).
			JoinConversations(contextWithToken(ctx2, token), req)

		if s.ResponseErr != nil {
			time.Sleep(1 * time.Second)
			return attempt < 5, s.ResponseErr
		}
		return false, nil
	})

	return ctx, nil
}

func (s *suite) teacherMustBeMemberOfConversations(ctx context.Context) (context.Context, error) {
	return s.userMustBeMemberOfConversations(ctx, "teacher")
}

func (s *suite) userMustBeMemberOfConversations(ctx context.Context, role string) (context.Context, error) {
	var userID string
	switch role {
	case "teacher":
		userID = s.teacherID
	case "school admin":
		userID = s.schoolAdminID
	case "student":
		userID = s.studentID
	default:
		return ctx, fmt.Errorf("not handle %s role to check be member of conversation", role)
	}

	return ctx, try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(1 * time.Second)
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		req := s.Request.(*tpb.JoinConversationsRequest)

		query := `SELECT count(*) FROM conversation_members cm WHERE cm.conversation_id = $1 AND cm.user_id = $2 
	AND cm.status = 'CONVERSATION_STATUS_ACTIVE'`
		for _, conversationID := range req.ConversationIds {
			id := conversationID
			var count int64
			err := s.DB.QueryRow(ctx2, query, &id, &userID).Scan(&count)
			if err != nil {
				return attempt < 10, err
			}
			if count != 1 {
				return attempt < 10, fmt.Errorf("user %s who have role %s is not a member of conversation %s", userID, role, s.conversationID)
			}
		}
		return false, nil
	})
}

func (s *suite) systemMustSendConversationMessage(ctx context.Context, message string) (context.Context, error) {
	sysMes := ""
	if message == "joined" {
		sysMes = tpb.CodesMessageType_CODES_MESSAGE_TYPE_JOINED_CONVERSATION.String()
	}
	if message == "created" {
		sysMes = tpb.CodesMessageType_CODES_MESSAGE_TYPE_CREATED_CONVERSATION.String()
	}

	query := `SELECT count(*) FROM messages WHERE conversation_id = $1 AND message = $2 AND type ='MESSAGE_TYPE_SYSTEM'`
	for _, conversationID := range s.ConversationIDs {
		id := conversationID
		var count int64
		err := try.Do(func(attempt int) (retry bool, err error) {
			defer func() {
				if err != nil {
					time.Sleep(2 * time.Second)
				}
			}()

			ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			err = s.DB.QueryRow(ctx2, query, &id, &sysMes).Scan(&count)
			if err != nil {
				return attempt < 5, err
			}
			if count < 1 {
				return attempt < 5, fmt.Errorf("conversation %s does not have %s message", id, message)
			}
			return false, nil
		})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
