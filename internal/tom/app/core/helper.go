package core

import (
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	tompb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	timestamps "google.golang.org/protobuf/types/known/timestamppb"
)

func toMessageResponse(m *entities.Message) *tompb.MessageResponse {
	createdAt, _ := types.TimestampProto(m.CreatedAt.Time)
	isDeleted := !m.DeletedAt.Time.IsZero()

	var content string
	var urlMedia string

	if !isDeleted {
		content = m.Message.String
		urlMedia = m.UrlMedia.String
	}

	return &tompb.MessageResponse{
		MessageId:      m.ID.String,
		ConversationId: m.ConversationID.String,
		UserId:         m.UserID.String,
		Content:        content,
		UrlMedia:       urlMedia,
		Type:           tompb.MessageType(tompb.MessageType_value[m.Type.String]),
		CreatedAt:      createdAt,
		TargetUser:     m.TargetUser.String,
		IsDeleted:      isDeleted,
		DeletedBy:      m.DeletedBy.String,
	}
}

var (
	lessonIgnoredSystemMessages = []string{
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_JOINED_LESSON.String(),
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_LEFT_LESSON.String(),
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_END_LIVE_LESSON.String(),
	}
)

// TODO: add is_silent field to conversation
func isSilentConversation(c *entities.Conversation) bool {
	conversationTypes := []string{
		tpb.ConversationType_CONVERSATION_LESSON.String(),
		tpb.ConversationType_CONVERSATION_LESSON_PRIVATE.String(),
	}
	return slices.Contains(conversationTypes, c.ConversationType.String)
}
func warnningIfError(logger *zap.Logger, msg string, err error) {
	if err != nil {
		logger.Warn(msg, zap.Error(err))
	}
}

func toMessageResponseV2(m *entities.Message) *tpb.MessageResponse {
	isDeleted := !m.DeletedAt.Time.IsZero()
	var content, urlMedia string
	if !isDeleted {
		content = m.Message.String
		urlMedia = m.UrlMedia.String
	}

	return &tpb.MessageResponse{
		MessageId:      m.ID.String,
		ConversationId: m.ConversationID.String,
		UserId:         m.UserID.String,
		Content:        content,
		UrlMedia:       urlMedia,
		Type:           tpb.MessageType(tpb.MessageType_value[m.Type.String]),
		CreatedAt:      timestamps.New(m.CreatedAt.Time),
		TargetUser:     m.TargetUser.String,
		IsDeleted:      isDeleted,
		DeletedBy:      m.DeletedBy.String,
	}
}
