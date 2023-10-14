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

type UserUpdateSubscription struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	ChatModifier        *support.ChatModifier
	DeviceTokenModifier *support.DeviceTokenModifier
}

func (rcv *UserUpdateSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUserDeviceToken, constants.DurableUserDeviceTokenUpdated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUserDeviceTokenUpdated),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "UserDeviceTokenSub.HandleUserDeviceTokenMsg",
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectUserDeviceTokenUpdated, constants.QueueGroupUserDeviceTokenUpdated, opts, rcv.Handle)
	if err != nil {
		return fmt.Errorf("UserInfoSubscription.JSM.QueueSubscribe: %v", err)
	}
	opts = nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurableUserUpdatedTom),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUserUpdatedTom),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "UserEventSub.HandleNatsMessageCreateConversation",
	}
	_, err = rcv.JSM.QueueSubscribe(constants.SubjectUserUpdated, constants.QueueUserUpdatedTom, opts, rcv.HandleUpdateStudent)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %v", err)
	}

	return nil
}

func (rcv *UserUpdateSubscription) HandleUpdateStudent(ctx context.Context, raw []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req := &upb.EvtUser{}
	err := proto.Unmarshal(raw, req)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return false, err
	}

	switch req.Message.(type) {
	case *upb.EvtUser_UpdateStudent_:
		msg := req.GetUpdateStudent()
		newReq := &upb.EvtUserInfo{
			UserId:            msg.GetStudentId(),
			DeviceToken:       msg.GetDeviceToken(),
			AllowNotification: msg.GetAllowNotification(),
			Name:              msg.GetName(),
			LocationIds:       msg.GetLocationIds(),
		}
		retry, err := rcv.DeviceTokenModifier.HandleEvtUserInfo(ctx, newReq, true)
		return retry, err
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
	default:
		return false, fmt.Errorf("invalid msg type for %T", req.Message)
	}
	return false, nil
}

func (rcv *UserUpdateSubscription) Handle(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &upb.EvtUserInfo{}
	if err := proto.Unmarshal(data, req); err != nil {
		rcv.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}
	retry, err := rcv.DeviceTokenModifier.HandleEvtUserInfo(ctx, req, false)
	return retry, err
}
