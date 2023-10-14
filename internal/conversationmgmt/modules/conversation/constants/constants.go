package constants

type WebhookEventType string

const (
	WebhookEventTypeChat        WebhookEventType = "chat"
	WebhookEventTypeChatOffline WebhookEventType = "chat_offline"
)

type WebhookChatType string

const (
	WebhookChatTypeGroupChat WebhookChatType = "groupchat"
	// For future
	WebhookChatTypeRecall WebhookChatType = "recall"
)

type WebhookHandlerType string

const (
	WebhookHandlerTypeNewMessage     WebhookHandlerType = "new_message"
	WebhookHandlerTypeOfflineMessage WebhookHandlerType = "offline_message"
	// For future
	WebhookHandlerTypeDeleteMessage WebhookHandlerType = "deleted_message"
)

const (
	ResponseWebhookOK     = "OK:200"
	ResponseWebhookFAILED = "FAILED:500"
)
