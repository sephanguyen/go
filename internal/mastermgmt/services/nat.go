package services

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	org_subscription "github.com/manabie-com/backend/internal/mastermgmt/modules/location/subscriptions"

	"go.uber.org/zap"
)

type OrganizationSubscription struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	OrganizationModifier *org_subscription.OrganizationModifier
}

func (rcv *OrganizationSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamOrganization, constants.DurableOrganizationCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverOrganizationCreated),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectOrganizationCreated, constants.QueueOrganizationCreated, opts,
		rcv.OrganizationModifier.HandleOrganizationCreated)
	if err != nil {
		return fmt.Errorf("OrganizationSubscription.JSM.QueueSubscribe: %v", err)
	}

	return nil
}
