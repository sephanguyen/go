package domain

import (
	"errors"
	"time"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

type (
	LiveLessonConversationType string
)

var (
	ErrNoConversationFound   = errors.New("no conversation found")
	ErrNoConversationCreated = errors.New("no conversation created")
)

const (
	LiveLessonConversationTypePublic  LiveLessonConversationType = "LIVE_LESSON_CONVERSATION_TYPE_PUBLIC"
	LiveLessonConversationTypePrivate LiveLessonConversationType = "LIVE_LESSON_CONVERSATION_TYPE_PRIVATE"
)

type LiveLessonConversation struct {
	LessonConversationID string
	ConversationID       string
	ConversationName     string
	LessonID             string
	ParticipantList      []string
	ConversationType     LiveLessonConversationType
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func NewLiveLessonConversation(lessonID string, participants []string, convType LiveLessonConversationType) LiveLessonConversation {
	convNameLabel := "public"
	if convType == LiveLessonConversationTypePrivate {
		convNameLabel = "private"
	}

	return LiveLessonConversation{
		LessonID:         lessonID,
		ParticipantList:  participants,
		ConversationType: convType,
		ConversationName: lessonID + "-" + convNameLabel,
	}
}

func (l *LiveLessonConversation) AddConversationID(conversationID string) {
	l.ConversationID = conversationID
}

func (l *LiveLessonConversation) UpdateParticipants(participants []string) {
	l.ParticipantList = participants
}

func (l *LiveLessonConversation) RemoveDuplicates() {
	l.ParticipantList = sliceutils.RemoveDuplicates(l.ParticipantList)
}
