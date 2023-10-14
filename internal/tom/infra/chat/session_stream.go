package chat

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const SEND_QUEUE_SIZE = 256

type ClientConn struct {
	ResourcePath         string
	SessionID            string
	UserID               string
	Context              context.Context
	Send                 chan *pb.Event
	SubscribeServer      pb.ChatService_SubscribeServer      // deprecated
	StreamingEventServer pb.ChatService_StreamingEventServer //  deprecated
	SubscribeV2Server    pb.ChatService_SubscribeV2Server
	closeOnce            sync.Once
	endWritePump         chan struct{}
	endReadPump          chan struct{}
	pumpFinished         chan struct{}
	logger               *zap.Logger
	chatService          *Server
	lastActiveAt         time.Time
	pingEvent            chan bool
}

func (rcv *Server) newClientConnV2(srv pb.ChatService_SubscribeV2Server, userID string, resourcePath string) *ClientConn {
	return &ClientConn{
		ResourcePath:      resourcePath,
		UserID:            userID,
		Context:           srv.Context(),
		Send:              make(chan *pb.Event, SEND_QUEUE_SIZE),
		SubscribeV2Server: srv,
		closeOnce:         sync.Once{},
		endWritePump:      make(chan struct{}),
		endReadPump:       make(chan struct{}),
		pumpFinished:      make(chan struct{}),
		logger:            ctxzap.Extract(srv.Context()),
		chatService:       rcv,
		lastActiveAt:      time.Now(),
		pingEvent:         make(chan bool, 1),
	}
}

func (rcv *ClientConn) Close() {
	rcv.closeOnce.Do(func() {
		close(rcv.endWritePump)
	})
	close(rcv.endReadPump)

	<-rcv.pumpFinished
}

func (rcv *ClientConn) SetOnline() error {
	now := time.Now()

	e := &entities.OnlineUser{}
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.UserID.Set(rcv.UserID),
		e.NodeName.Set(rcv.chatService.hostName),
		e.LastActiveAt.Set(now),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	err = rcv.chatService.onlineUserRepo.Insert(rcv.Context, rcv.chatService.db, e)
	if err != nil {
		return err
	}

	rcv.SessionID = e.ID.String
	rcv.lastActiveAt = now

	return nil
}

func (rcv *ClientConn) SetActive() error {
	rcv.lastActiveAt = time.Now()
	err := rcv.chatService.onlineUserRepo.SetActive(rcv.Context, rcv.chatService.db, database.Text(rcv.SessionID))
	return err
}

func (rcv *ClientConn) PumpSubscribeV2() error {
	err := rcv.writePumpSubscribeV2(rcv.Context)

	close(rcv.pumpFinished)

	rcv.chatService.HubUnregister(rcv)
	// we don't want to use ctx in client, because it might have been canceled
	timedOutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctxWithRP := interceptors.ContextWithJWTClaims(timedOutCtx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{ResourcePath: rcv.ResourcePath},
	})
	err = multierr.Append(err, rcv.chatService.onlineUserRepo.Delete(ctxWithRP, rcv.chatService.db, database.Text(rcv.SessionID)))

	return err
}
func (rcv *ClientConn) recordServerDisconnect() {
	atomic.AddInt64(&rcv.chatService.metrics.serverSideDis, 1)
}

func (rcv *ClientConn) recordClientDisconnect() {
	atomic.AddInt64(&rcv.chatService.metrics.clientSideDis, 1)
}

func (rcv *ClientConn) writePumpSubscribeV2(ctx context.Context) error {
	ticker := time.NewTicker(HEALTHCHECK_TICKER)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-rcv.Send:
			if !ok {
				rcv.recordServerDisconnect()
				return errors.New("chatService.SubscribeV2 stream closed")
			}

			err := rcv.SendMsg(msg)
			if err != nil {
				rcv.recordServerDisconnect()
				return errors.Wrap(err, "rcv.Write")
			}

			n := len(rcv.Send)
			for i := 0; i < n; i++ {
				err := rcv.SendMsg(<-rcv.Send)
				if err != nil {
					rcv.recordServerDisconnect()
					return errors.Wrap(err, "rcv.Write")
				}
			}
		case <-rcv.pingEvent:
			err := rcv.SetActive()
			if err != nil {
				rcv.recordServerDisconnect()
				return fmt.Errorf("err SetActive: %w", err)
			}
		case <-ticker.C:
			now := time.Now()
			if now.Sub(rcv.lastActiveAt).Round(100*time.Millisecond) >= 3*HEALTHCHECK_TICKER {
				rcv.recordServerDisconnect()
				return fmt.Errorf("disconnected client did not ping")
			}
		case <-rcv.endWritePump:
			rcv.recordServerDisconnect()
			return nil
		case <-ctx.Done():
			rcv.recordClientDisconnect()
			return fmt.Errorf("ctx.Done: %w", ctx.Err())
		}
	}
}

func (rcv *ClientConn) SendMsg(msg *pb.Event) error {
	var err error
	newMsg := msg.GetEventNewMessage()
	conversationID := ""
	messageID := ""
	if newMsg != nil {
		conversationID = newMsg.GetConversationId()
		messageID = newMsg.GetMessageId()
	}
	rcv.logger.Info(
		"ClientConn.SendMsg response broadcast message to client",
		zap.String("host", rcv.chatService.hostName),
		zap.String("session_id", rcv.SessionID),
		zap.String("user_id", rcv.UserID),
		zap.String("conversation_id", conversationID),
		zap.String("message_id", messageID),
	)

	switch data := msg.Event.(type) {
	case *pb.Event_EventNewMessage:
		if rcv.SubscribeServer != nil { // for backward compatible
			err = rcv.SubscribeServer.Send(&pb.SubscribeResponse{Event: &pb.SubscribeResponse_MessageResponse{MessageResponse: data.EventNewMessage}})
		}

		if rcv.StreamingEventServer != nil { // for backward compatible
			err = rcv.StreamingEventServer.Send(&pb.StreamingEventResponse{Event: &pb.StreamingEventResponse_EventNewMessage{EventNewMessage: data.EventNewMessage}})
		}
	case *pb.Event_EventDeleteMessage_:
		if rcv.StreamingEventServer != nil { // for backward compatible
			err = rcv.StreamingEventServer.Send(&pb.StreamingEventResponse{
				Event: &pb.StreamingEventResponse_EventDeleteMessage_{
					EventDeleteMessage: &pb.StreamingEventResponse_EventDeleteMessage{
						ConversationId: data.EventDeleteMessage.ConversationId,
						MessageId:      data.EventDeleteMessage.MessageId,
						DeletedBy:      data.EventDeleteMessage.DeletedBy,
					},
				},
			})
		}
	case *pb.Event_EventPing:
		// special case, just broadcast for session by sessionID
		if rcv.SessionID != data.EventPing.SessionId {
			return nil
		}

		select {
		case rcv.pingEvent <- true:
		default:
		}
	}

	if rcv.SubscribeV2Server != nil {
		err = rcv.SubscribeV2Server.Send(&pb.SubscribeV2Response{
			Event: msg,
		})
	}

	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.OK:
			// noop
			return nil
		case codes.Unavailable, codes.Canceled, codes.DeadlineExceeded:
			return errors.New("chatService.Subscribe client terminated connection")
		default:
			return errors.Wrap(s.Err(), "chatService.Subscribe failed to send to client")
		}
	}

	return err
}
