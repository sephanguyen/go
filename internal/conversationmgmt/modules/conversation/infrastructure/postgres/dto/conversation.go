package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type ConversationPgDTO struct {
	ID                    pgtype.Text
	Name                  pgtype.Text
	LatestMessage         pgtype.JSONB
	LatestMessageSentTime pgtype.Timestamptz
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
	OptionalConfig        pgtype.JSONB
}

func (dto *ConversationPgDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_id", "name", "latest_message", "latest_message_sent_time", "created_at", "updated_at", "deleted_at", "optional_config"}
	values = []interface{}{&dto.ID, &dto.Name, &dto.LatestMessage, &dto.LatestMessageSentTime, &dto.CreatedAt, &dto.UpdatedAt, &dto.DeletedAt, &dto.OptionalConfig}
	return
}

func (dto ConversationPgDTO) TableName() string {
	return "conversation"
}

func (dto ConversationPgDTO) ToConversationDomain() (*domain.Conversation, error) {
	latestMessage, err := NewMessagePgDTOFromJSONB(dto.LatestMessage)
	if err != nil {
		return nil, fmt.Errorf("cannot parse latest message: [%+v]", err)
	}

	conversation := &domain.Conversation{
		ID:             dto.ID.String,
		Name:           dto.Name.String,
		LatestMessage:  latestMessage.ToMessageDomain(),
		CreatedAt:      dto.CreatedAt.Time,
		UpdatedAt:      dto.UpdatedAt.Time,
		OptionalConfig: dto.OptionalConfig.Bytes,
	}

	if dto.LatestMessageSentTime.Status == pgtype.Present {
		conversation.LatestMessageSentTime = &dto.LatestMessageSentTime.Time
	} else {
		conversation.LatestMessageSentTime = nil
	}

	return conversation, nil
}

func NewConversationDTOFromDomain(conversation *domain.Conversation) (*ConversationPgDTO, error) {
	dto := &ConversationPgDTO{}
	database.AllNullEntity(dto)

	if conversation.ID == "" {
		conversation.ID = idutil.ULIDNow()
	}

	err := multierr.Combine(
		dto.ID.Set(conversation.ID),
		dto.Name.Set(conversation.Name),
		dto.CreatedAt.Set(conversation.CreatedAt),
		dto.UpdatedAt.Set(conversation.UpdatedAt),
		dto.OptionalConfig.Set(conversation.OptionalConfig),
		dto.DeletedAt.Set(nil),
	)

	now := time.Now()
	if conversation.CreatedAt.IsZero() {
		conversation.CreatedAt = now
		err = multierr.Combine(
			err,
			dto.CreatedAt.Set(now),
		)
	}

	if conversation.UpdatedAt.IsZero() {
		conversation.UpdatedAt = now
		err = multierr.Combine(
			err,
			dto.UpdatedAt.Set(now),
		)
	}

	if conversation.LatestMessageSentTime == nil || conversation.LatestMessageSentTime.IsZero() {
		err = multierr.Combine(
			err,
			dto.LatestMessageSentTime.Set(nil),
		)
	}

	if conversation.LatestMessage == nil {
		err = multierr.Combine(
			err,
			dto.LatestMessage.Set(database.JSONB(nil)),
		)
	} else {
		err = multierr.Combine(
			err,
			dto.LatestMessage.Set(NewMessagePgDTOFromDomain(conversation.LatestMessage).ToJSONB()),
		)
	}

	return dto, err
}

type ConversationPgDTOs []*ConversationPgDTO

func (c ConversationPgDTOs) ToConversationsDomain() ([]*domain.Conversation, error) {
	domainConversations := make([]*domain.Conversation, 0, len(c))
	for _, convoDTO := range c {
		convoDomain, err := convoDTO.ToConversationDomain()
		if err != nil {
			return nil, fmt.Errorf("cannot convert conversation DTO to domain object: [%+v]", err)
		}
		domainConversations = append(domainConversations, convoDomain)
	}

	return domainConversations, nil
}
