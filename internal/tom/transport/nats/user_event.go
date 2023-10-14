package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/app/support"
	"github.com/manabie-com/backend/internal/tom/configurations"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type UserEventSubscription struct {
	Config       *configurations.Config
	Logger       *zap.Logger
	ChatModifier *support.ChatModifier
	JSM          nats.JetStreamManagement
}

func (rcv *UserEventSubscription) Subscribe() error {

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurableUserCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUserCreated),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "UserEventSub.HandleNatsMessageCreateConversation",
	}
	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUserCreated, constants.QueueUserCreated, opts, rcv.HandlerNatsMessageCreateConversation)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %v", err)
	}
	return nil
}

func (rcv *UserEventSubscription) HandlerNatsMessageCreateConversation(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req := &upb.EvtUser{}
	err := proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	switch req.Message.(type) {
	case *upb.EvtUser_CreateParent_:
		msg := req.GetCreateParent()
		err := rcv.ChatModifier.HandleEventCreateParentConversation(ctx, &upb.EvtUser_ParentAssignedToStudent{
			StudentId: msg.GetStudentId(),
			ParentId:  msg.GetParentId(),
		})
		if err != nil {
			return true, fmt.Errorf("err rcv.HandleEventCreateParentConversation: %w", err)
		}
	case *upb.EvtUser_CreateStudent_:
		msg := req.GetCreateStudent()
		err := rcv.ChatModifier.HandleEventCreateStudentConversation(ctx, msg)
		if err != nil {
			return true, fmt.Errorf("err rcv.HandleEventCreateStudentConversation: %w", err)
		}
	case *upb.EvtUser_ParentRemovedFromStudent_:
		msg := req.GetParentRemovedFromStudent()
		err := rcv.ChatModifier.HandleParentRemovedFromStudent(ctx, msg)
		if err != nil {
			return true, fmt.Errorf("err rcv.HandleEventParentRemovedFromStudent: %w", err)
		}
	case *upb.EvtUser_ParentAssignedToStudent_:
		msg := req.GetParentAssignedToStudent()
		err := rcv.ChatModifier.HandleEventCreateParentConversation(ctx, msg)
		if err != nil {
			return true, fmt.Errorf("err rcv.HandleEventCreateParentConversation: %w", err)
		}
	}
	return false, nil
}
