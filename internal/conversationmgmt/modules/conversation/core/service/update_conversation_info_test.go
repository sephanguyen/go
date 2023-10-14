package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationModifierService_UpdateConversationInfo(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}

	mockConversationRepo := &mock_repositories.MockConversationRepo{}

	svc := &conversationModifierServiceImpl{
		DB:               mockDB,
		Environment:      "local",
		ConversationRepo: mockConversationRepo,
	}

	conversation := &domain.Conversation{
		ID:             "conversation-id-1",
		Name:           "conversation-name",
		OptionalConfig: []byte("optional-config"),
	}

	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		mockConversationRepo.On("UpdateConversationInfo", ctx, mockDB, mock.Anything).Once().Return(nil)
		_, err := svc.UpdateConversationInfo(ctx, conversation)
		assert.Nil(t, err)
	})

	t.Run("update conversation info error", func(t *testing.T) {
		mockConversationRepo.On("UpdateConversationInfo", ctx, mockDB, mock.Anything).Once().Return(puddle.ErrClosedPool)
		_, err := svc.UpdateConversationInfo(ctx, conversation)
		assert.Equal(t, fmt.Errorf("error when ConversationRepo.UpdateConversationInfo: [%v]", puddle.ErrClosedPool), err)
	})
}
