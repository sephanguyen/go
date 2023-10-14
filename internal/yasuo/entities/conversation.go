package entities

import "time"

// this entities for conversations index - Elasticsearch

type ESConversation struct {
	ConversationID           string    `json:"conversation_id,omitempty"`
	ConversationNameEnglish  string    `json:"conversation_name.english,omitempty"`
	ConversationNameJapanese string    `json:"conversation_name.japanese,omitempty"`
	CourseIDs                []string  `json:"course_ids"`
	UserIds                  []string  `json:"user_ids"`
	LastMessage              ESMessage `json:"last_message,omitempty"`
	IsReplied                bool      `json:"is_replied"`
	Owner                    string    `json:"owner"`
	ConversationType         string    `json:"conversation_type"`
}

type ESMessage struct {
	UpdatedAt time.Time `json:"updated_at"`
}
