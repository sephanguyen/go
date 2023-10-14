package payload

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/constants"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"

	"github.com/stretchr/testify/assert"
)

func Test_NewMessageEventRequestFromJSONBytes(t *testing.T) {
	t.Parallel()
	t.Run("happy case", func(t *testing.T) {
		rawReqEvent := []byte(`{
			"callId": "call-id",
			"eventType": "chat",
			"chat_type": "groupchat",
			"recall_id": "recall-id",
			"payload": {
			  "ext": {
				"resource_path": "resource_path",
				"manabie_msg_type": "text",
				"paths": [
				  {
					"name": "file.ext",
					"url": "example.com"
				  }
				]
			  },
			  "bodies": [
				{
				  "msg": "msg",
				  "type": "txt"
				}
			  ],
			  "type": "groupchat"
			},
			"group_id": "convo-id",
			"from": "from-id",
			"to": "to-id",
			"msg_id": "msg-id",
			"timestamp": 1646101800000
		}`)

		reqEvent, err := NewMessageEventRequestFromJSONBytes(rawReqEvent)
		assert.NoError(t, err)
		expectedEvt := &MessageEventRequest{
			ConversationID: "convo-id",
			EventType:      "chat",
			RecallID:       "recall-id",
			ChatType:       "groupchat",
			MessageID:      "msg-id",
			From:           "from-id",
			To:             "to-id",
			Timestamp:      1646101800000,
			Payload: MessageEventPayload{
				Extension: MessageExtension{
					ResourcePath:       "resource_path",
					ManabieMessageType: "text",
					Paths: []MessageMedia{
						{
							Name: "file.ext",
							URL:  "example.com",
						},
					},
				},
				Messages: []Message{
					{
						Message: "msg",
						Type:    "txt",
					},
				},
				Type: "groupchat",
			},
		}

		assert.Equal(t, expectedEvt, reqEvent)
	})
}

func Test_ToMessageDomain(t *testing.T) {
	t.Parallel()
	t.Run("happy case chat - chat group event", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType:      "chat",
			ConversationID: "convo-id",
			ChatType:       "groupchat",
			MessageID:      "msg-id",
			From:           "from-id",
			To:             "convo-id",
			Timestamp:      1646101800000,
			Payload: MessageEventPayload{
				Extension: MessageExtension{
					ResourcePath:       "resource-path",
					ManabieMessageType: "text",
					Paths: []MessageMedia{
						{
							Name: "name",
							URL:  "example.com",
						},
					},
				},
				Messages: []Message{
					{
						Message: "msg",
						Type:    "text",
					},
				},
				Type: "txt",
			},
		}

		domainResult := reqEvent.ToMessageDomain("user-id")

		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T02:30:00.00Z")
		expectedDomain := &domain.Message{
			ConversationID:  "convo-id",
			VendorMessageID: "msg-id",
			Message:         "msg",
			VendorUserID:    "from-id",
			UserID:          "user-id",
			Type:            domain.MessageTypeText,
			SentTime:        expectedSentTime.Local(),
			IsDeleted:       false,
			Media: []domain.MessageMedia{
				{
					Name: "name",
					URL:  "example.com",
				},
			},
		}

		assert.Equal(t, expectedDomain, domainResult)
	})

	t.Run("happy case chat - recall event", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType: "chat",
			RecallID:  "recall-id",
			ChatType:  "recall",
			MessageID: "msg-id",
			From:      "from-id",
			To:        "convo-id",
			Timestamp: 1646101800000,
			Payload: MessageEventPayload{
				Extension: MessageExtension{
					ResourcePath: "resource-path",
				},
				Type: "recall",
			},
		}

		domainResult := reqEvent.ToMessageDomain("user-id")

		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T02:30:00.00Z")
		expectedDomain := &domain.Message{
			ConversationID:  "convo-id",
			VendorMessageID: "msg-id",
			VendorUserID:    "from-id",
			UserID:          "user-id",
			SentTime:        expectedSentTime.Local(),
			IsDeleted:       true,
		}

		assert.Equal(t, expectedDomain, domainResult)
	})
}

func Test_ToOfflineMessageDomain(t *testing.T) {
	t.Parallel()
	t.Run("happy case chat_offline", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType:      "chat_offline",
			ConversationID: "convo-id",
			ChatType:       "groupchat",
			MessageID:      "msg-id",
			From:           "from-id",
			To:             "to-id",
			Timestamp:      1646101800000,
			Payload: MessageEventPayload{
				Extension: MessageExtension{
					ResourcePath:       "resource-path",
					ManabieMessageType: "text",
					Paths: []MessageMedia{
						{
							Name: "name",
							URL:  "example.com",
						},
					},
				},
				Messages: []Message{
					{
						Message: "msg",
						Type:    "text",
					},
				},
				Type: "txt",
			},
		}

		domainResult := reqEvent.ToOfflineMessageDomain("user-id")

		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T02:30:00.00Z")
		expectedDomain := &domain.OfflineMessage{
			Message: domain.Message{
				ConversationID:  "convo-id",
				VendorMessageID: "msg-id",
				Message:         "msg",
				VendorUserID:    "from-id",
				UserID:          "user-id",
				Type:            domain.MessageTypeText,
				SentTime:        expectedSentTime.Local(),
				IsDeleted:       false,
				Media: []domain.MessageMedia{
					{
						Name: "name",
						URL:  "example.com",
					},
				},
			},
			OfflineVendorUserID: "to-id",
		}

		assert.Equal(t, expectedDomain, domainResult)
	})
}

func Test_GetWebhookHandlerTypeAndConversationID(t *testing.T) {
	t.Parallel()
	t.Run("happy case new message", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType:      "chat",
			ConversationID: "convo-id",
			ChatType:       "groupchat",
			From:           "from-id",
			To:             "convo-id",
		}

		handlerType, conversationID, userID := reqEvent.GetWebhookHandlerTypeAndConversationIDAndVendorUserID()

		assert.Equal(t, constants.WebhookHandlerTypeNewMessage, handlerType)
		assert.Equal(t, reqEvent.ConversationID, conversationID)
		assert.Equal(t, reqEvent.From, userID)
	})
	t.Run("happy case offline message", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType:      "chat_offline",
			ConversationID: "convo-id",
			ChatType:       "groupchat",
			From:           "from-id",
			To:             "to-id",
		}

		handlerType, conversationID, userID := reqEvent.GetWebhookHandlerTypeAndConversationIDAndVendorUserID()

		assert.Equal(t, constants.WebhookHandlerTypeOfflineMessage, handlerType)
		assert.Equal(t, reqEvent.ConversationID, conversationID)
		assert.Equal(t, reqEvent.To, userID)
	})
	t.Run("happy case deleted message", func(t *testing.T) {
		reqEvent := &MessageEventRequest{
			EventType:      "chat",
			ConversationID: "",
			ChatType:       "recall",
			From:           "from-id",
			To:             "convo-id",
		}

		handlerType, conversationID, userID := reqEvent.GetWebhookHandlerTypeAndConversationIDAndVendorUserID()
		assert.Equal(t, constants.WebhookHandlerTypeDeleteMessage, handlerType)
		assert.Equal(t, reqEvent.To, conversationID)
		assert.Equal(t, reqEvent.From, userID)
	})
}
