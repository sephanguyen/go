package communication

import (
	"fmt"

	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

type studentchatState struct {
	teacherChats          *tpb.ListConversationsInSchoolResponse
	teacherChosenChat     string
	teacherChosenChatType string
	teacherChannel        legacytpb.ChatService_SubscribeV2Client

	studentChats   *legacytpb.ConversationListResponse
	studentChannel legacytpb.ChatService_SubscribeV2Client

	parentChats       *legacytpb.ConversationListResponse
	parentChannel     legacytpb.ChatService_SubscribeV2Client
	newMessageBuffers map[string][]message
}

type message struct {
	msgType       string
	content       string
	seenByTeacher bool
	seenByStudent bool
	seenByParent  bool
}

func (m *message) setSeen(person string) {
	switch person {
	case teacher:
		m.seenByTeacher = true
	case student:
		m.seenByStudent = true
	case parent:
		m.seenByParent = true
	default:
		panic(fmt.Sprintf("not expecting person %s", person))
	}
}

func (m *message) seenBy(person string) bool {
	switch person {
	case teacher:
		return m.seenByTeacher
	case student:
		return m.seenByStudent
	case parent:
		return m.seenByParent
	default:
		panic(fmt.Sprintf("not expecting person %s", person))
	}
}
