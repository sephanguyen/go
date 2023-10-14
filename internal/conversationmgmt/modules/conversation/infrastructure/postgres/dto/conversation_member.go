package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ConversationMemberPgDTO struct {
	ID             pgtype.Text
	ConversationID pgtype.Text
	UserID         pgtype.Text
	Status         pgtype.Text
	SeenAt         pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (dto *ConversationMemberPgDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_member_id", "conversation_id", "user_id", "status", "seen_at", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&dto.ID, &dto.ConversationID, &dto.UserID, &dto.Status, &dto.SeenAt, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt}
	return
}

func (dto ConversationMemberPgDTO) TableName() string {
	return "conversation_member"
}

func (dto ConversationMemberPgDTO) ToConversationMemberDomain() *domain.ConversationMember {
	conversation := &domain.ConversationMember{
		ID:             dto.ID.String,
		ConversationID: dto.ConversationID.String,
		User: domain.ChatVendorUser{
			UserID: dto.UserID.String,
		},
		Status:    common.ConversationMemberStatus(dto.Status.String),
		SeenAt:    dto.SeenAt.Time,
		CreatedAt: dto.CreatedAt.Time,
		UpdatedAt: dto.UpdatedAt.Time,
	}

	return conversation
}

func NewConversationMemberDTOFromDomain(conversationMember *domain.ConversationMember) (*ConversationMemberPgDTO, error) {
	dto := &ConversationMemberPgDTO{}
	database.AllNullEntity(dto)

	if conversationMember.ID == "" {
		conversationMember.ID = idutil.ULIDNow()
	}

	err := multierr.Combine(
		dto.ID.Set(conversationMember.ID),
		dto.ConversationID.Set(conversationMember.ConversationID),
		dto.SeenAt.Set(conversationMember.SeenAt),
		dto.UserID.Set(conversationMember.User.UserID),
		dto.Status.Set(string(conversationMember.Status)),
		dto.CreatedAt.Set(conversationMember.CreatedAt),
		dto.UpdatedAt.Set(conversationMember.UpdatedAt),
		dto.DeletedAt.Set(nil),
	)

	now := time.Now()
	if conversationMember.CreatedAt.IsZero() {
		conversationMember.CreatedAt = now
		err = multierr.Combine(
			err,
			dto.CreatedAt.Set(now),
		)
	}

	if conversationMember.UpdatedAt.IsZero() {
		conversationMember.UpdatedAt = now
		err = multierr.Combine(
			err,
			dto.UpdatedAt.Set(now),
		)
	}

	return dto, err
}

func NewConversationMemberDTOsFromDomains(conversationMembers []*domain.ConversationMember) ([]*ConversationMemberPgDTO, error) {
	dtoArr := make([]*ConversationMemberPgDTO, 0)
	for _, entity := range conversationMembers {
		dto, err := NewConversationMemberDTOFromDomain(entity)
		if err != nil {
			return nil, err
		}

		dtoArr = append(dtoArr, dto)
	}

	return dtoArr, nil
}

type ConversationMemberPgDTOs []*ConversationMemberPgDTO

func (c ConversationMemberPgDTOs) ToConversationMemberDomain() []*domain.ConversationMember {
	conversationMemberDomains := make([]*domain.ConversationMember, 0, len(c))
	for _, e := range c {
		conversationMemberDomain := &domain.ConversationMember{
			ID:             e.ID.String,
			ConversationID: e.ConversationID.String,
			User: domain.ChatVendorUser{
				UserID: e.UserID.String,
			},
			Status:    common.ConversationMemberStatus(e.Status.String),
			SeenAt:    e.SeenAt.Time,
			CreatedAt: e.CreatedAt.Time,
			UpdatedAt: e.UpdatedAt.Time,
		}
		conversationMemberDomains = append(conversationMemberDomains, conversationMemberDomain)
	}
	return conversationMemberDomains
}
