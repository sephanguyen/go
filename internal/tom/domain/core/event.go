package core

// TODO: update more fields after migrating Es indexing logic from Yasuo
type MessageSentEvent struct {
	ConversationType string
	ConversationID   string
}

const (
	MessageSentEventStr = "MessageSentEvent"
)
