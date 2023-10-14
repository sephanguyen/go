package communication

import (
	"context"
	"fmt"
	"time"

	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"go.uber.org/multierr"
)

func (s *suite) readReplyMessageSteps() map[string]interface{} {
	return map[string]interface{}{
		`^"([^"]*)" is at the conversation screen$`:                                                      s.isAtTheConversationScreen,
		`^"([^"]*)" does not read message$`:                                                              s.doesNotReadMessage,
		`^"([^"]*)" has created student with parent info$`:                                               s.hasCreatedAStudentWithParentInfo,
		`^"([^"]*)" has joined parent chat group$`:                                                       s.hasJoinedParentChatGroup,
		`^"([^"]*)" has joined student chat group$`:                                                      s.hasJoinedStudentChatGroup,
		`^"([^"]*)" sees chat group with unread message is showed on top on Learner App$`:                s.seesChatGroupWithUnreadMessageIsShowedOnTopOnLearnerApp,
		`^"([^"]*)" sees "([^"]*)" icon next to chat group in Messages list on Learner App$`:             s.seesIconNextToChatGroupInMessagesListOnLearnerApp,
		`^teacher does not see "([^"]*)" status next to message in conversation on Teacher App$`:         s.teacherDoesNotSeeStatusNextToMessageInConversationOnTeacherApp,
		`^teacher sees "([^"]*)" icon next to chat group in Messages list on Teacher App$`:               s.teacherSeesIconNextToChatGroupInMessageListOnTeacherApp,
		`^teacher sends "([^"]*)" message to "([^"]*)"$`:                                                 s.teacherSendsMessageTo,
		`^"([^"]*)" reads the message$`:                                                                  s.readsTheMessage,
		`^"([^"]*)" sees "([^"]*)" icon next to chat group disappeared in Messages list on Learner App$`: s.seesIconNextToChatGroupDisappearedInMessagesListOnLearnerApp,
		`^teacher sees "([^"]*)" status next to message on Teacher App$`:                                 s.teacherSeesStatusNextToMessageOnTeacherApp,
		`^teacher does not read replies$`:                                                                s.teacherDoesNotReadReplies,
		`^"([^"]*)" replies to teacher$`:                                                                 s.repliesToTeacher,
		`^teacher does not see "([^"]*)" icon next to chat group in Messages list on Teacher App$`:       s.teacherDoesNotSeeIconNextToChatGroupInMessagesListOnTeacherApp,
		`^teacher sees chat group with unread message shown on top in Messages list on Teacher App$`:     s.teacherSeesChatGroupWithUnreadMessageShownOnTopInMessagesListOnTeacherApp,
		`^"([^"]*)" does not see "([^"]*)" status next to message on Learner App$`:                       s.doesNotSeeStatusNextToMessageOnLearnerApp,
		`^teacher read replies$`:                                                                         s.teacherReadReplies,
		`^teacher sees "([^"]*)" icon next to chat group disappeared in Messages list on Teacher App$`:   s.teacherSeesIconNextToChatGroupDisappearedInMessagesListOnTeacherApp,
	}
}

var (
	iconTypeUnread = "Unread"
)

func (s *suite) teacherSeesIconNextToChatGroupDisappearedInMessagesListOnTeacherApp(iconType string) error {
	if iconType != iconTypeUnread {
		return fmt.Errorf("want Unread icon, given %s", iconType)
	}
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}
	for _, chat := range s.studentChatState.teacherChats.Items {
		if chat.ConversationId == s.studentChatState.teacherChosenChat {
			if !chat.Seen {
				return fmt.Errorf("expect chat to be read, but was unread")
			}
			return nil
		}
	}
	return fmt.Errorf("cannot find chosen conversation after refreshing screen")
}

func (s *suite) teacherReadReplies() error {
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	_, err := legacytpb.NewChatServiceClient(s.tomConn).SeenMessage(
		ctx,
		&legacytpb.SeenMessageRequest{
			ConversationId: s.studentChatState.teacherChosenChat,
		},
	)
	return err
}

func (s *suite) doesNotSeeStatusNextToMessageOnLearnerApp(userAccount string, msgStatus string) error {
	// won't simulate: read status is not implemented on learner app
	return nil
}

func (s *suite) teacherSeesChatGroupWithUnreadMessageShownOnTopInMessagesListOnTeacherApp() error {
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}

	for _, item := range s.studentChatState.teacherChats.Items {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			if item.Seen {
				return fmt.Errorf("want conversation to be unseen, got seen")
			}
			return nil
		}
	}
	return fmt.Errorf("conversation not found after refreshing conversation list")
}

func (s *suite) teacherDoesNotSeeIconNextToChatGroupInMessagesListOnTeacherApp(replyStatus string) error {
	if replyStatus != "Replied" {
		return fmt.Errorf("want Replied, given %s", replyStatus)
	}
	// on mobile, if a new message is received for a conversation, its replied icon is remove, simulate this behaviour
	var newMsg *legacytpb.MessageResponse
	stream := s.studentChatState.teacherChannel
	var newMsgReceived bool
	for try := 0; try < 10; try++ {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("unexpected error: %s", err)
		}
		newMsg = resp.GetEvent().GetEventNewMessage()
		if newMsg == nil {
			continue
		}
		if newMsg.Type == legacytpb.MESSAGE_TYPE_SYSTEM {
			continue
		}
		if newMsg.ConversationId == s.studentChatState.teacherChosenChat {
			newMsgReceived = true
			break
		}
	}
	if !newMsgReceived {
		return fmt.Errorf("no signal from upstream notifying new message")
	}
	// refresh to check if replied status is actually removed
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}

	for _, item := range s.studentChatState.teacherChats.Items {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			if item.IsReplied {
				return fmt.Errorf("want conversation to be unreplied, got replied")
			}
			return nil
		}
	}
	return fmt.Errorf("conversation not found after refreshing conversation list")
}

func (s *suite) repliesToTeacher(userAccount string) error {
	return s.sendsToTheConversationOnLearnerApp(userAccount, sendMsgTypeText)
}

func (s *suite) teacherDoesNotReadReplies() error {
	return nil
}

func (s *suite) teacherSeesLastMessageAsReadAfterRefresh() error {
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}
	var seenAt time.Time
	var found bool

	for _, item := range s.studentChatState.teacherChats.Items {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			var localSeen time.Time
			for _, user := range item.Users {
				if user.Group == cpb.UserGroup_USER_GROUP_TEACHER {
					continue
				}
				found = true
				if user.SeenAt.AsTime().After(localSeen) {
					localSeen = user.SeenAt.AsTime()
				}
			}
			seenAt = localSeen
			break
		}
	}
	if !found {
		return fmt.Errorf("cannot find selected conversation after refreshing the screen")
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	details, err := legacytpb.NewChatServiceClient(s.tomConn).ConversationDetail(
		ctx,
		&legacytpb.ConversationDetailRequest{
			ConversationId: s.studentChatState.teacherChosenChat,
			Limit:          10,
		},
	)
	if err != nil {
		return err
	}
	lastMsgCreatedAt, err := types.TimestampFromProto(details.Messages[0].CreatedAt)
	if err != nil {
		return err
	}
	if lastMsgCreatedAt.After(seenAt) {
		return fmt.Errorf("expect last message to be read, but was unread")
	}
	return nil
}

func (s *suite) teacherSeesStatusNextToMessageOnTeacherApp(status string) error {
	if status != "Read" {
		return fmt.Errorf("want Read status, given %s", status)
	}
	var newMsg *legacytpb.MessageResponse

	stream := s.studentChatState.teacherChannel
	var receivedSeenMsg bool
	for try := 0; try < 10; try++ {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("unexpected error: %s", err)
		}
		newMsg = resp.GetEvent().GetEventNewMessage()
		if newMsg == nil {
			continue
		}
		if newMsg.Type == legacytpb.MESSAGE_TYPE_SYSTEM && newMsg.Content == legacytpb.CODES_MESSAGE_TYPE_SEEN_CONVERSATION.String() {
			receivedSeenMsg = true
			break
		}
	}
	if !receivedSeenMsg {
		return fmt.Errorf("no signal from upstream notifying seen message event")
	}
	// try refresh screen and check if msg is still seen
	return s.teacherSeesLastMessageAsReadAfterRefresh()
}

func (s *suite) seesIconNextToChatGroupDisappearedInMessagesListOnLearnerApp(userAccount string, iconType string) error {
	if iconType != iconTypeUnread {
		return fmt.Errorf("want Unread icon type, given %s", iconType)
	}
	// refresh screen
	err := s.isAtTheConversationScreen(userAccount)
	if err != nil {
		return err
	}
	chats, err := s.getUserAccountChat(userAccount)
	if err != nil {
		return err
	}
	for _, chat := range chats.Conversations {
		if chat.ConversationId == s.studentChatState.teacherChosenChat {
			if !chat.Seen {
				return fmt.Errorf("expect chat to be seen, but actually is not")
			}
			return nil
		}
	}
	return fmt.Errorf("cannot find conversation in %s chat", userAccount)
}

func (s *suite) getUserAccountChat(userAccount string) (*legacytpb.ConversationListResponse, error) {
	var chats *legacytpb.ConversationListResponse
	switch userAccount {
	case student:
		chats = s.studentChatState.studentChats
	case parent:
		chats = s.studentChatState.parentChats
	default:
		return nil, fmt.Errorf("unsupported useraccount %s", userAccount)
	}
	return chats, nil
}

func (s *suite) readsTheMessage(userAccount string) error {
	// refresh
	err := s.isAtTheConversationScreen(userAccount)
	if err != nil {
		return err
	}
	chats, err := s.getUserAccountChat(userAccount)
	if err != nil {
		return err
	}

	for _, item := range chats.GetConversations() {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(userAccount))
			defer cancel()
			_, err := legacytpb.NewChatServiceClient(s.tomConn).SeenMessage(
				ctx,
				&legacytpb.SeenMessageRequest{
					ConversationId: item.ConversationId,
				},
			)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("cannot find conversation in %s chat", userAccount)
}

func (s *suite) seesIconNextToChatGroupInMessagesListOnLearnerApp(userAccount string, iconType string) error {
	if iconType != iconTypeUnread {
		return fmt.Errorf("want unread icon, given %s", iconType)
	}
	// refresh to get new chats status
	err := s.isAtTheConversationScreen(userAccount)
	if err != nil {
		return err
	}

	chats, err := s.getUserAccountChat(userAccount)
	if err != nil {
		return err
	}
	for _, item := range chats.GetConversations() {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			if item.Seen {
				return fmt.Errorf("expect conversation to be unread, actual status is read")
			}
			return nil
		}
	}
	return fmt.Errorf("cannot find conversation in %s chat", userAccount)
}

func (s *suite) teacherSeesIconNextToChatGroupInMessageListOnTeacherApp(iconType string) error {
	var checkReplied, checkUnread bool
	switch iconType {
	case "Replied":
		checkReplied = true
	case iconTypeUnread:
		checkUnread = true
	default:
		return fmt.Errorf("unsupported icon type %s", iconType)
	}

	// refresh teacher screen to update replied status
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}
	for _, item := range s.studentChatState.teacherChats.Items {
		if item.ConversationId == s.studentChatState.teacherChosenChat {
			if checkReplied {
				if !item.IsReplied {
					return fmt.Errorf("expect conversation to show replied status but was not")
				}
			}
			if checkUnread {
				if item.Seen {
					return fmt.Errorf("expect conversation to be unread, but was read instead")
				}
			}

			return nil
		}
	}
	return fmt.Errorf("conversation not found after refreshing conversation list")
}

func (s *suite) teacherDoesNotSeeStatusNextToMessageInConversationOnTeacherApp(statusType string) error {
	return nil
}

func (s *suite) seesChatGroupWithUnreadMessageIsShowedOnTopOnLearnerApp(accountType string) error {
	var stream legacytpb.ChatService_SubscribeV2Client
	switch accountType {
	case student:
		stream = s.studentChatState.studentChannel
	case parent:
		stream = s.studentChatState.parentChannel
	default:
		return fmt.Errorf("unsupported account type %s", accountType)
	}
	var newMsg *legacytpb.MessageResponse

	for try := 0; try < 10; try++ {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("unexpected error: %s", err)
		}
		newMsg = resp.GetEvent().GetEventNewMessage()
		if newMsg == nil {
			continue
		}
		if newMsg.Type == legacytpb.MESSAGE_TYPE_SYSTEM {
			continue
		}
		break
	}
	if newMsg == nil {
		return fmt.Errorf("no signal from upstream notifying new message")
	}

	return nil
}

func (s *suite) teacherSendsMessageTo(msgType string, userAccount string) error {
	return multierr.Combine(
		s.teacherHasAccessedToTheConversationOfChatGroup(userAccount),
		s.teacherSendsToTheConversationOnTeacherApp(msgType),
	)
}

func (s *suite) hasCreatedAStudentWithParentInfo(ctx context.Context, role string) (context.Context, error) {
	if role != schoolAdmin {
		return ctx, fmt.Errorf("expect %s, got %s", schoolAdmin, role)
	}
	return s.schoolAdminHasCreatedStudentWithParentInfo(ctx)
}

func (s *suite) doesNotReadMessage(role string) error {
	return nil
}

func (s *suite) teacherJoinsConversations(ids []string) error {
	req := &tpb.JoinConversationsRequest{
		ConversationIds: ids,
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	_, err := tpb.NewChatModifierServiceClient(s.tomConn).
		JoinConversations(
			ctx,
			req,
		)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) hasJoinedParentChatGroup(role string) error {
	if role != teacher {
		return fmt.Errorf("expect %s, got %s", teacher, role)
	}
	return eventually(func() error {
		err := s.teacherIsAtTheConversationScreen()
		if err != nil {
			return err
		}

		for _, item := range s.studentChatState.teacherChats.Items {
			if item.ConversationType == tpb.ConversationType_CONVERSATION_PARENT {
				parentConv := item.GetConversationId()
				return s.teacherJoinsConversations([]string{parentConv})
			}
		}
		return fmt.Errorf("not found parent conversation in teacher screen")
	})
}

func (s *suite) hasJoinedStudentChatGroup(role string) error {
	if role != teacher {
		return fmt.Errorf("expect %s, got %s", teacher, role)
	}
	return eventually(func() error {
		err := s.teacherIsAtTheConversationScreen()
		if err != nil {
			return err
		}

		for _, item := range s.studentChatState.teacherChats.Items {
			if item.ConversationType == tpb.ConversationType_CONVERSATION_STUDENT {
				studentConv := item.GetConversationId()
				return s.teacherJoinsConversations([]string{studentConv})
			}
		}
		return fmt.Errorf("not found student conversation in teacher screen")
	})
}
