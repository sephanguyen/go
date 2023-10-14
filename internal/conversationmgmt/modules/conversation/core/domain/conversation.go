package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
)

type Conversation struct {
	ID   string
	Name string

	// Latest message can be null, careful when checking it
	LatestMessage         *Message
	LatestMessageSentTime *time.Time

	CreatedAt      time.Time
	UpdatedAt      time.Time
	OptionalConfig []byte

	Members []ConversationMember
}

func NewConversation(conversationID, name string, members []ChatVendorUser, optionalConfig []byte) Conversation {
	convoMembers := make([]ConversationMember, 0)

	for _, member := range members {
		convoMembers = append(convoMembers, ConversationMember{
			User: ChatVendorUser{
				UserID:       member.UserID,
				VendorUserID: member.VendorUserID,
			},
			Status:         common.ConversationMemberStatusActive,
			ConversationID: conversationID,
		})
	}

	return Conversation{
		ID:             conversationID,
		Name:           name,
		Members:        convoMembers,
		OptionalConfig: optionalConfig,
	}
}
