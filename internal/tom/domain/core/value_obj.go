package core

type MessageToUserOpts struct {
	Notification NotificationOpts
}
type MessageToConversationOpts struct {
	Persist bool
	AsUser  bool
}
type NotificationOpts struct {
	IgnoredUsers []string
	Enabled      bool
	Silence      bool
	Title        string
}

const (
	ConversationMemberStatusActive   = "CONVERSATION_STATUS_ACTIVE"
	ConversationMemberStatusInactive = "CONVERSATION_STATUS_INACTIVE"
)
