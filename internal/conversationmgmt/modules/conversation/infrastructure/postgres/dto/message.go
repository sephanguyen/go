package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type MessageTypePgDTO string

const (
	MessageTypeText  MessageTypePgDTO = "text"
	MessageTypeImage MessageTypePgDTO = "image"
	MessageTypeVideo MessageTypePgDTO = "video"
	MessageTypeFile  MessageTypePgDTO = "file"
)

type MessageMediaPgDTO struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MessagePgDTO struct {
	ConversationID  string              `json:"conversation_id"`
	VendorMessageID string              `json:"vendor_message_id"`
	Message         string              `json:"message"`
	UserID          string              `json:"user_id"`
	VendorUserID    string              `json:"vendor_user_id"`
	Type            MessageTypePgDTO    `json:"type"`
	SentTime        time.Time           `json:"sent_time"`
	IsDeleted       bool                `json:"is_deleted"`
	Media           []MessageMediaPgDTO `json:"media"`
}

func NewMessagePgDTOFromJSONB(payload pgtype.JSONB) (*MessagePgDTO, error) {
	if payload.Status == pgtype.Null {
		// Careful with this
		return nil, nil
	}

	msg := &MessagePgDTO{}
	err := payload.AssignTo(msg)
	return msg, err
}

func NewMessagePgDTOFromDomain(message *domain.Message) *MessagePgDTO {
	msg := &MessagePgDTO{
		ConversationID:  message.ConversationID,
		VendorMessageID: message.VendorMessageID,
		VendorUserID:    message.VendorUserID,
		UserID:          message.UserID,
		Message:         message.Message,
		Type:            MessageTypePgDTO(message.Type),
		SentTime:        message.SentTime,
		IsDeleted:       message.IsDeleted,
		Media:           []MessageMediaPgDTO{},
	}

	for _, media := range message.Media {
		msg.Media = append(msg.Media, MessageMediaPgDTO{
			Name: media.Name,
			URL:  media.URL,
		})
	}

	return msg
}

func (m *MessagePgDTO) ToJSONB() pgtype.JSONB {
	if m == nil {
		return database.JSONB(nil)
	}
	return database.JSONB(m)
}

func (m *MessagePgDTO) ToMessageDomain() *domain.Message {
	if m == nil {
		return nil
	}

	msg := &domain.Message{
		ConversationID:  m.ConversationID,
		VendorMessageID: m.VendorMessageID,
		VendorUserID:    m.VendorUserID,
		UserID:          m.UserID,
		Message:         m.Message,
		Type:            domain.MessageType(m.Type),
		SentTime:        m.SentTime,
		IsDeleted:       m.IsDeleted,
		Media:           []domain.MessageMedia{},
	}

	for _, media := range m.Media {
		msg.Media = append(msg.Media, domain.MessageMedia{
			Name: media.Name,
			URL:  media.URL,
		})
	}

	return msg
}
