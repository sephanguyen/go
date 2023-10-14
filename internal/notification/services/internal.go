package services

import (
	"context"

	"github.com/manabie-com/backend/internal/notification/infra"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

func NewInternalService(pushNotificationService infra.PushNotificationService) *InternalService {
	return &InternalService{
		PushNotificationService: pushNotificationService,
	}
}

type InternalService struct {
	infra.PushNotificationService
}

func (rcv *InternalService) RetrievePushedNotificationMessages(ctx context.Context, req *npb.RetrievePushedNotificationMessageRequest) (*npb.RetrievePushedNotificationMessageResponse, error) {
	since := types.Timestamp{
		Seconds: req.Since.Seconds,
		Nanos:   req.Since.Nanos,
	}

	messages, err := rcv.PushNotificationService.RetrievePushedMessages(ctx, req.DeviceToken, int(req.Limit), &since)

	if err != nil {
		return nil, errors.Wrap(err, "rcv.NotificationPusher.RetrievePushedMessages")
	}

	ret := make([]*npb.PushedNotificationMessage, 0, len(messages))
	for _, m := range messages {
		data := m.Data

		protobufFields := map[string]*structpb.Value{}
		for key := range data {
			protobufFields[key] = &structpb.Value{
				Kind: &structpb.Value_StringValue{StringValue: data[key]},
			}
		}
		msg := &npb.PushedNotificationMessage{
			Data: &structpb.Struct{
				Fields: protobufFields,
			},
			PushedAt: nil,
		}
		msg.Title = m.Title
		msg.Body = m.Body

		ret = append(ret, msg)
	}

	return &npb.RetrievePushedNotificationMessageResponse{Messages: ret}, nil
}
