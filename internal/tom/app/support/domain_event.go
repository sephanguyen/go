package support

import (
	"context"
	"fmt"

	cconstants "github.com/manabie-com/backend/internal/golibs/constants"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *ChatModifier) HandleCoreMessageSent(ctx context.Context, idata interface{}) error {
	evt, ok := idata.(domain.MessageSentEvent)
	if !ok {
		return nil
	}
	convType := evt.ConversationType

	if convType != tpb.ConversationType_CONVERSATION_STUDENT.String() &&
		convType != tpb.ConversationType_CONVERSATION_PARENT.String() {
		return nil
	}

	event := tpb.ConversationInternal{
		TriggeredAt: timestamppb.Now(),
		Message: &tpb.ConversationInternal_MessageSent{
			MessageSent: &tpb.ConversationInternal_MessageSentToConversation{
				ConversationId: evt.ConversationID,
			},
		},
	}
	b, err := proto.Marshal(&event)
	if err != nil {
		return fmt.Errorf("proto.Marshal %w", err)
	}
	_, err = c.JSM.TracedPublish(ctx, "HandleCoreMessageSent", cconstants.SubjectChatMessageCreated, b)

	if err != nil {
		c.Logger.Warn("c.Jsm.TracedPublish", zap.Error(err))
	}

	return nil
}
