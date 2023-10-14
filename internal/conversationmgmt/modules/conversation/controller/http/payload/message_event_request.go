package payload

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/constants"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

// This is Agora message payload in webhook request
type Message struct {
	Message string `json:"msg"`
	Type    string `json:"type"`
}

// Custom media field in Manabie
type MessageMedia struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Custom extension message for each webhook request
type MessageExtension struct {
	ResourcePath       string         `json:"resource_path"`
	ManabieMessageType string         `json:"manabie_msg_type"`
	Paths              []MessageMedia `json:"paths"`
}

type MessageEventPayload struct {
	Extension MessageExtension `json:"ext"`
	Messages  []Message        `json:"bodies"`
	Type      string           `json:"type"`
}

type MessageEventRequest struct {
	EventType      string              `json:"eventType"`
	RecallID       string              `json:"recall_id"`
	ChatType       string              `json:"chat_type"`
	ConversationID string              `json:"group_id"`
	MessageID      string              `json:"msg_id"`
	From           string              `json:"from"`
	To             string              `json:"to"`
	Timestamp      uint64              `json:"timestamp"`
	Payload        MessageEventPayload `json:"payload"`
}

func NewMessageEventRequestFromJSONBytes(rawReq []byte) (*MessageEventRequest, error) {
	obj := new(MessageEventRequest)
	err := json.Unmarshal(rawReq, obj)

	if err != nil {
		return nil, fmt.Errorf("failed parse json request: [%+v]", err)
	}

	return obj, nil
}

func (req *MessageEventRequest) ToMessageDomain(userID string) *domain.Message {
	conversationID := req.ConversationID
	// Incase recall -> conversation_id is stored in `to` field
	if conversationID == "" {
		conversationID = req.To
	}

	message := &domain.Message{
		VendorMessageID: req.MessageID,
		VendorUserID:    req.From,
		UserID:          userID,
		ConversationID:  conversationID,
		Type:            domain.MessageType(req.Payload.Extension.ManabieMessageType),
		IsDeleted:       false,
		SentTime:        time.UnixMilli(int64(req.Timestamp)),
	}

	if len(req.Payload.Messages) > 0 {
		message.Message = req.Payload.Messages[0].Message
	}

	for _, media := range req.Payload.Extension.Paths {
		message.Media = append(message.Media, domain.MessageMedia{
			Name: media.Name,
			URL:  media.URL,
		})
	}

	if req.RecallID != "" {
		message.IsDeleted = true
	}

	return message
}

func (req *MessageEventRequest) ToOfflineMessageDomain(userID string) *domain.OfflineMessage {
	conversationID := req.ConversationID

	message := &domain.OfflineMessage{
		Message: domain.Message{
			VendorMessageID: req.MessageID,
			VendorUserID:    req.From,
			UserID:          userID,
			ConversationID:  conversationID,
			Type:            domain.MessageType(req.Payload.Extension.ManabieMessageType),
			IsDeleted:       false,
			SentTime:        time.UnixMilli(int64(req.Timestamp)),
		},
		OfflineVendorUserID: req.To,
	}

	if len(req.Payload.Messages) > 0 {
		message.Message.Message = req.Payload.Messages[0].Message
	}

	for _, media := range req.Payload.Extension.Paths {
		message.Media = append(message.Media, domain.MessageMedia{
			Name: media.Name,
			URL:  media.URL,
		})
	}

	return message
}

func (req *MessageEventRequest) GetWebhookHandlerTypeAndConversationIDAndVendorUserID() (constants.WebhookHandlerType, string, string) {
	eventType, chatType := constants.WebhookEventType(req.EventType), constants.WebhookChatType(req.ChatType)

	if eventType == constants.WebhookEventTypeChat && chatType == constants.WebhookChatTypeGroupChat {
		return constants.WebhookHandlerTypeNewMessage, req.ConversationID, req.From
	}

	if eventType == constants.WebhookEventTypeChatOffline && chatType == constants.WebhookChatTypeGroupChat {
		return constants.WebhookHandlerTypeOfflineMessage, req.ConversationID, req.To
	}

	if eventType == constants.WebhookEventTypeChat && chatType == constants.WebhookChatTypeRecall {
		return constants.WebhookHandlerTypeDeleteMessage, req.To, req.From
	}

	return constants.WebhookHandlerType(""), "", ""
}
