package chat

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (rcv *Server) SubscribeV2(req *pb.SubscribeV2Request, srv pb.ChatService_SubscribeV2Server) error {
	srvCtx := srv.Context()
	userID := interceptors.UserIDFromContext(srvCtx)
	rcv.logger.Info("chatService.SubscribeV2", zap.String("userID", userID))
	resourcePath, err := interceptors.ResourcePathFromContext(srvCtx)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	clientConn := rcv.newClientConnV2(srv, userID, resourcePath)
	err = rcv.HubRegister(clientConn)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	err = clientConn.SetOnline()
	if err != nil {
		rcv.HubUnregister(clientConn)
		return status.Error(codes.Internal, err.Error())
	}

	rcv.PushLocal(clientConn.Context, []string{userID}, &pb.Event{
		Event: &pb.Event_EventPing{
			EventPing: &pb.EventPing{
				SessionId: clientConn.SessionID,
			},
		},
	})

	return clientConn.PumpSubscribeV2()
}

func (rcv *Server) PingSubscribeV2(ctx context.Context, req *pb.PingSubscribeV2Request) (*pb.PingSubscribeV2Response, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "sessionId is required")
	}

	userID := interceptors.UserIDFromContext(ctx)
	// 1. check last online in db
	pgSince := database.Timestamptz(time.Now().Add(HEALTHCHECK_TICKER * -2))

	users, err := rcv.onlineUserRepo.Find(ctx, rcv.db, database.TextArray([]string{userID}), pgSince, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(users) == 0 {
		return nil, status.Error(codes.NotFound, "user does not online")
	}

	// 2. broadcast to cluster
	rcv.PushCluster(ctx, []string{userID}, &pb.Event{
		Event: &pb.Event_EventPing{
			EventPing: &pb.EventPing{
				SessionId: req.SessionId,
			},
		},
	})

	return &pb.PingSubscribeV2Response{}, nil
}

func (rcv *Server) RetrievePushedNotificationMessages(ctx context.Context, req *pb.RetrievePushedNotificationMessageRequest) (*pb.RetrievePushedNotificationMessageResponse, error) {
	messages, err := rcv.notification.Pusher.RetrievePushedMessages(ctx, req.DeviceToken, int(req.Limit), req.Since)
	if err != nil {
		return nil, errors.Wrap(err, "rcv.notification.Pusher.RetrievePushedMessages")
	}

	ret := make([]*pb.PushedNotificationMessage, 0, len(messages))
	for _, m := range messages {
		data := m.Data

		protobufFields := map[string]*types.Value{}
		for key := range data {
			protobufFields[key] = &types.Value{
				Kind: &types.Value_StringValue{StringValue: data[key]},
			}
		}
		msg := &pb.PushedNotificationMessage{
			Data: &types.Struct{
				Fields: protobufFields,
			},
			PushedAt: nil,
		}
		if m.Notification != nil {
			msg.Title = m.Notification.Title
			msg.Body = m.Notification.Body
		}

		ret = append(ret, msg)
	}
	return &pb.RetrievePushedNotificationMessageResponse{Messages: ret}, nil
}
