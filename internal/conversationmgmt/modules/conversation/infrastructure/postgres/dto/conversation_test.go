package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func Test_ToConversationsDomain(t *testing.T) {
	t.Parallel()

	t.Run("should convert successful", func(t *testing.T) {
		rawLatestMessage := &domain.Message{
			ConversationID: "conv-1",
			Media:          []domain.MessageMedia{},
		}

		optConfig := struct {
			Config string `json:"config"`
		}{
			Config: "1",
		}
		byteOptConfig, _ := json.Marshal(optConfig)
		now := time.Now()
		dtoConversation := ConversationPgDTOs{
			&ConversationPgDTO{
				ID:                    database.Text("conv-1"),
				Name:                  database.Text("name-1"),
				LatestMessage:         NewMessagePgDTOFromDomain(rawLatestMessage).ToJSONB(),
				LatestMessageSentTime: database.Timestamptz(now),
				CreatedAt:             database.Timestamptz(now),
				UpdatedAt:             database.Timestamptz(now),
				OptionalConfig:        database.JSONB(optConfig),
			},
			&ConversationPgDTO{
				ID:                    database.Text("conv-2"),
				Name:                  database.Text("name-2"),
				LatestMessage:         database.JSONB(nil),
				LatestMessageSentTime: pgtype.Timestamptz{Status: pgtype.Null},
				CreatedAt:             database.Timestamptz(now),
				UpdatedAt:             database.Timestamptz(now),
				OptionalConfig:        database.JSONB(optConfig),
			},
		}

		expectedDomainEntities := []*domain.Conversation{
			{
				ID:                    "conv-1",
				Name:                  "name-1",
				LatestMessage:         rawLatestMessage,
				LatestMessageSentTime: &now,
				CreatedAt:             now,
				UpdatedAt:             now,
				OptionalConfig:        byteOptConfig,
			},
			{
				ID:                    "conv-2",
				Name:                  "name-2",
				LatestMessage:         nil,
				LatestMessageSentTime: nil,
				CreatedAt:             now,
				UpdatedAt:             now,
				OptionalConfig:        byteOptConfig,
			},
		}

		domains, err := dtoConversation.ToConversationsDomain()

		assert.Nil(t, err)
		assert.Equal(t, expectedDomainEntities, domains)
	})

	t.Run("should nil", func(t *testing.T) {
		dtoConversation := ConversationPgDTOs{}

		expectedDomainEntities := []*domain.Conversation{}

		domains, err := dtoConversation.ToConversationsDomain()

		assert.Nil(t, err)
		assert.Equal(t, expectedDomainEntities, domains)
	})
}
