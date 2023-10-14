package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/app/support"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type StaffUpsertSubscription struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	ChatModifier *support.ChatModifier
}

func (rcv *StaffUpsertSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStaff, constants.DurableUpsertStaffTom),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUpsertStaffTom),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "Staff.HandleUpsertStaff",
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUpsertStaff, constants.QueueUpsertStaffTom, opts, rcv.Handle)
	if err != nil {
		return fmt.Errorf("Staff.JSM.QueueSubscribe: %v", err)
	}

	return nil
}

func (rcv *StaffUpsertSubscription) Handle(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &upb.EvtUpsertStaff{}
	err := proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	retry, err := rcv.ChatModifier.HandleUpsertStaff(ctx, req)
	return retry, err
}
