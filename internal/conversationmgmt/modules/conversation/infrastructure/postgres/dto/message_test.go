package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/stretchr/testify/assert"
)

func Test_NewMessagePgDTOFromJSONB(t *testing.T) {
	t.Parallel()

	t.Run("should create new object successful", func(t *testing.T) {
		rawJSONB := database.JSONB([]byte(`{
			"conversation_id": "convo-id",
			"vendor_message_id": "vendor-msg-id",
			"message": "msg",
			"vendor_user_id": "vendor-user-id",
			"user_id": "user-id",
			"type": "text",
			"sent_time": "2022-03-01T09:30:00.00Z",
			"is_deleted": false,
			"media": [
			  {
				"name": "media-1.ext",
				"url": "example.com"
			  }
			]
		}`))

		dto, err := NewMessagePgDTOFromJSONB(rawJSONB)

		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T09:30:00.00Z")
		expectedDTO := &MessagePgDTO{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []MessageMediaPgDTO{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		assert.NoError(t, err)
		assert.Equal(t, expectedDTO, dto)
	})
}

func Test_NewMessagePgDTOFromDomain(t *testing.T) {
	t.Parallel()

	t.Run("should create new object successful", func(t *testing.T) {
		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T09:30:00.00Z")
		domain := &domain.Message{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            domain.MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []domain.MessageMedia{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		dto := NewMessagePgDTOFromDomain(domain)

		expectedDTO := &MessagePgDTO{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []MessageMediaPgDTO{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		assert.Equal(t, expectedDTO, dto)
	})
}

func Test_ToJSONB(t *testing.T) {
	t.Parallel()

	t.Run("should convert successful", func(t *testing.T) {
		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T09:30:00.00Z")
		dto := &MessagePgDTO{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []MessageMediaPgDTO{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		expectedJSONB := database.JSONB([]byte(`{"conversation_id":"convo-id","vendor_message_id":"vendor-msg-id","message":"msg","user_id":"user-id","vendor_user_id":"vendor-user-id","type":"text","sent_time":"2022-03-01T09:30:00Z","is_deleted":false,"media":[{"name":"media-1.ext","url":"example.com"}]}`))
		jsonB := dto.ToJSONB()

		assert.Equal(t, expectedJSONB, jsonB)
	})
	t.Run("nil", func(t *testing.T) {
		dto := &MessagePgDTO{}
		dto = nil
		expectedJSONB := database.JSONB(nil)
		jsonB := dto.ToJSONB()
		assert.Equal(t, expectedJSONB, jsonB)
	})
}

func Test_ToMessageDomain(t *testing.T) {
	t.Parallel()

	t.Run("should create new object successful", func(t *testing.T) {
		expectedSentTime, _ := time.Parse(time.RFC3339, "2022-03-01T09:30:00.00Z")
		dto := &MessagePgDTO{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []MessageMediaPgDTO{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		domainResult := dto.ToMessageDomain()
		expectedDomain := &domain.Message{
			ConversationID:  "convo-id",
			VendorMessageID: "vendor-msg-id",
			Message:         "msg",
			UserID:          "user-id",
			Type:            domain.MessageTypeText,
			SentTime:        expectedSentTime,
			VendorUserID:    "vendor-user-id",
			IsDeleted:       false,
			Media: []domain.MessageMedia{
				{
					Name: "media-1.ext",
					URL:  "example.com",
				},
			},
		}

		assert.Equal(t, expectedDomain, domainResult)
	})
	t.Run("nil", func(t *testing.T) {
		dto := &MessagePgDTO{}
		dto = nil
		domainResult := dto.ToMessageDomain()
		assert.Nil(t, domainResult)
	})
}
