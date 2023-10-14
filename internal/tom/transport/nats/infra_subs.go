package nats

import (
	"context"
	"fmt"
	"os"
	"time"

	golib_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

type MessageChatModifierSubscription struct {
	HostName string
	JSM      nats.JetStreamManagement
	Logger   *zap.Logger

	ChatInfra interface {
		HandleInternalBroadcast(ctx context.Context, data []byte) (bool, error)
	}
}

func (rcv *MessageChatModifierSubscription) SubscribeMessageChatCreated() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("err when call os.Hostname(): %v", err)
	}
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.AckAll(),
			nats.Bind(golib_constants.StreamChatMessage, fmt.Sprintf("%s-%s", golib_constants.DurableChatMessageCreated, hostname)),
			nats.DeliverSubject(fmt.Sprintf("%s-%s", golib_constants.DeliverChatMessageCreated, hostname)),
			nats.AckWait(30 * time.Second),
			nats.MaxDeliver(-1),
		},
	}

	_, err = rcv.JSM.QueueSubscribe(golib_constants.SubjectSendChatMessageCreated, golib_constants.QueueChatMessageCreated, opts, rcv.ChatInfra.HandleInternalBroadcast)

	if err != nil {
		return fmt.Errorf("subSendMessageChat.Subscribe: %w", err)
	}

	return nil
}

func (rcv *MessageChatModifierSubscription) SubscribeMessageChatDeleted() error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("err when call os.Hostname(): %v", err)
	}
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.AckAll(),
			nats.Bind(golib_constants.StreamChatMessage, fmt.Sprintf("%s-%s", golib_constants.DurableChatMessageDeleted, hostname)),
			nats.DeliverSubject(fmt.Sprintf("%s-%s", golib_constants.DeliverChatMessageDeleted, hostname)),
			nats.AckWait(30 * time.Second),
			nats.MaxDeliver(-1),
		},
	}

	_, err = rcv.JSM.QueueSubscribe(golib_constants.SubjectChatMessageDeleted, golib_constants.QueueChatMessageDeleted, opts, rcv.ChatInfra.HandleInternalBroadcast)

	if err != nil {
		return fmt.Errorf("MessageChatDeleted.Subscribe: %w", err)
	}

	return nil
}
