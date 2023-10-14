package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/stretchr/testify/assert"
)

func Test_ToConversationMemberDomain(t *testing.T) {
	t.Parallel()

	t.Run("map successful", func(t *testing.T) {
		now := time.Now()

		conversationMemberDTOs := ConversationMemberPgDTOs{
			{
				ID:             database.Text("conversation-member-id-1"),
				ConversationID: database.Text("conversation-id-1"),
				UserID:         database.Text("user-id-1"),
				Status:         database.Text(string(common.ConversationMemberStatusActive)),
				SeenAt:         database.Timestamptz(now),
				CreatedAt:      database.Timestamptz(now),
				UpdatedAt:      database.Timestamptz(now),
			},
			{
				ID:             database.Text("conversation-member-id-2"),
				ConversationID: database.Text("conversation-id-1"),
				UserID:         database.Text("user-id-2"),
				Status:         database.Text(string(common.ConversationMemberStatusActive)),
				SeenAt:         database.Timestamptz(now),
				CreatedAt:      database.Timestamptz(now),
				UpdatedAt:      database.Timestamptz(now),
			},
		}
		expectedConversationMemberDomains := []*domain.ConversationMember{
			{
				ID:             "conversation-member-id-1",
				ConversationID: "conversation-id-1",
				User: domain.ChatVendorUser{
					UserID: "user-id-1",
				},
				Status:    common.ConversationMemberStatusActive,
				SeenAt:    now,
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:             "conversation-member-id-2",
				ConversationID: "conversation-id-1",
				User: domain.ChatVendorUser{
					UserID: "user-id-2",
				},
				Status:    common.ConversationMemberStatusActive,
				SeenAt:    now,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		actualConversationMemberDomains := conversationMemberDTOs.ToConversationMemberDomain()
		assert.Equal(t, expectedConversationMemberDomains, actualConversationMemberDomains)
	})

	t.Run("should nil", func(t *testing.T) {
		conversationMemberDTOs := ConversationMemberPgDTOs{}
		expectedConversationMemberDomains := []*domain.ConversationMember{}
		actualConversationMemberDomains := conversationMemberDTOs.ToConversationMemberDomain()
		assert.Equal(t, expectedConversationMemberDomains, actualConversationMemberDomains)
	})
}
