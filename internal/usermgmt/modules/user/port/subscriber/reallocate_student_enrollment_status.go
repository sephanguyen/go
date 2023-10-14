package subscriber

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
)

type ReallocateStudentEnrollmentStatusSubscriber struct {
	JSM nats.JetStreamManagement

	StudentRegistrationService interface {
		ReallocateStudentEnrollmentStatus(ctx context.Context, data []byte) (bool, error)
	}
}

func (u *ReallocateStudentEnrollmentStatusSubscriber) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
			nats.Bind(constants.StreamEnrollmentStatusAssignment, constants.DurableEnrollmentStatusAssignment),
			nats.DeliverSubject(constants.DeliverEnrollmentStatusAssignment),
		},
	}

	_, err := u.JSM.QueueSubscribe(constants.SubjectEnrollmentStatusAssignmentCreated,
		constants.QueueEnrollmentStatusAssignment, option, u.StudentRegistrationService.ReallocateStudentEnrollmentStatus)
	if err != nil {
		return fmt.Errorf("syncOrderSub.Subscribe: %w", err)
	}

	return nil
}
