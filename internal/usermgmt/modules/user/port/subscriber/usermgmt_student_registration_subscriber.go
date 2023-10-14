package subscriber

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
)

type StudentRegistrationSubscriber struct {
	JSM nats.JetStreamManagement

	StudentRegistrationService interface {
		SyncOrderHandler(ctx context.Context, data []byte) (bool, error)
	}
}

func (u *StudentRegistrationSubscriber) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
			nats.Bind(constants.StreamOrderEventLog, constants.DurableOrderEventLogCreated),
			nats.DeliverSubject(constants.DeliverOrderEventLogCreated),
		},
	}

	_, err := u.JSM.QueueSubscribe(constants.SubjectOrderEventLogCreated,
		constants.QueueOrderEventLogCreated, option, u.StudentRegistrationService.SyncOrderHandler)
	if err != nil {
		return fmt.Errorf("syncOrderSub.Subscribe: %w", err)
	}

	return nil
}
