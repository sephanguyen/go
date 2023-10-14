package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
)

type ConversationMember struct {
	ID             string
	ConversationID string
	User           ChatVendorUser
	Status         common.ConversationMemberStatus
	SeenAt         time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ConversationMemberOpt func(e *ConversationMember)

func WithConversationID(id string) ConversationMemberOpt {
	return func(e *ConversationMember) {
		e.ConversationID = id
	}
}

func WithStatus(status common.ConversationMemberStatus) ConversationMemberOpt {
	return func(e *ConversationMember) {
		e.Status = status
	}
}
