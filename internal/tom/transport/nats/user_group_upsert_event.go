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

type UserGroupUpsertSubscription struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	ChatModifier *support.ChatModifier
}

func (rcv *UserGroupUpsertSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUserGroup, constants.DurableUpserUserGroupTom),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUpsertUserGroupTom),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "Staff.HandleUpsertStaff",
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUpsertUserGroup, constants.QueueUpsertUserGroupTom, opts, rcv.Handle)
	if err != nil {
		return fmt.Errorf("Staff.JSM.QueueSubscribe: %v", err)
	}

	return nil
}

func (rcv *UserGroupUpsertSubscription) Handle(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &upb.EvtUpsertUserGroup{}
	err := proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	retry, err := rcv.ChatModifier.HandleUpsertUserGroup(ctx, req)
	return retry, err
}
