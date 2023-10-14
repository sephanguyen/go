package subscriber

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
)

type UserRegistrationSubscriber struct {
	JSM nats.JetStreamManagement

	UserRegistrationService interface {
		SyncStaffHandler(ctx context.Context, data []byte) (bool, error)
		SyncStudentHandler(ctx context.Context, data []byte) (bool, error)
	}
}

func (u *UserRegistrationSubscriber) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
	}

	optionStudent := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncUserRegistration, constants.DurableSyncStudent),
			nats.DeliverSubject(constants.DeliverSyncUserRegistrationStudent)),
	}
	_, err := u.JSM.QueueSubscribe(constants.SubjectUserRegistrationNatsJS,
		constants.QueueSyncStudent, optionStudent, u.UserRegistrationService.SyncStudentHandler)
	if err != nil {
		return fmt.Errorf("syncStudentSub.Subscribe: %w", err)
	}

	optionStaff := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncUserRegistration, constants.DurableSyncStaff),
			nats.DeliverSubject(constants.DeliverSyncUserRegistrationStaff)),
	}
	_, err = u.JSM.QueueSubscribe(constants.SubjectUserRegistrationNatsJS,
		constants.QueueSyncStaff, optionStaff, u.UserRegistrationService.SyncStaffHandler)
	if err != nil {
		return fmt.Errorf("syncStaffSub.Subscribe: %w", err)
	}

	return nil
}
