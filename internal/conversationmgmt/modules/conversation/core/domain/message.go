package domain

import (
	"encoding/json"
	"time"
)

type MessageMedia struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
)

type Message struct {
	ConversationID  string         `json:"conversation_id"`
	VendorMessageID string         `json:"vendor_message_id"`
	Message         string         `json:"message"`
	UserID          string         `json:"user_id"`
	VendorUserID    string         `json:"vendor_user_id"`
	Type            MessageType    `json:"type"`
	SentTime        time.Time      `json:"sent_time"`
	IsDeleted       bool           `json:"is_deleted"`
	Media           []MessageMedia `json:"media"`
}

type OfflineMessage struct {
	Message
	OfflineVendorUserID string
}

func (m *Message) ToBytes() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}
