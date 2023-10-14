package chat

import (
	"context"
	"fmt"
	"hash/fnv"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	golib_constants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/constants"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"firebase.google.com/go/v4/messaging"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

const (
	BROADCAST_QUEUE_SIZE = 4096
	SESSION_CACHE_SIZE   = 35000                             // total connection 1 hub
	DEADLOCK_TICKER      = 15 * time.Second                  // check every 15 seconds
	HEALTHCHECK_TICKER   = 5 * time.Second                   // check every 5 seconds
	DEADLOCK_WARN        = (BROADCAST_QUEUE_SIZE * 99) / 100 // number of buffered messages before printing stack trace
)

func (rcv *Server) InvalidateCacheUserOnline(ctx context.Context) {
	ticker := time.NewTicker(HEALTHCHECK_TICKER)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rcv.onlineUserRepo.InvalidateCache(HEALTHCHECK_TICKER)
		}
	}
}

func (rcv *Server) NewHub() *Hub {
	return &Hub{
		connectionCount: 0,
		connectionIndex: 0,
		register:        make(chan *ClientConn, 1),
		unregister:      make(chan *ClientConn, 1),
		broadcast:       make(chan *Event, BROADCAST_QUEUE_SIZE),
		stop:            make(chan struct{}),
		didStop:         make(chan struct{}),
		ExplicitStop:    false,
		goroutineID:     0,
		chatService:     rcv,
	}
}

func (rcv *Server) TotalConnections() int {
	count := int64(0)
	for _, hub := range rcv.Hubs {
		count += atomic.LoadInt64(&hub.connectionCount)
	}

	return int(count)
}

func (rcv *Server) HubStart() {
	// Total number of hubs is twice the number of CPUs.
	numberOfHubs := runtime.NumCPU() * 2
	rcv.logger.Info("Starting hubs", zap.Int("number_of_hubs", numberOfHubs))

	rcv.Hubs = make([]*Hub, numberOfHubs)
	rcv.HubsStopCheckingForDeadlock = make(chan bool, 1)

	for i := 0; i < len(rcv.Hubs); i++ {
		rcv.Hubs[i] = rcv.NewHub()
		rcv.Hubs[i].connectionIndex = i
		rcv.Hubs[i].Start()
	}

	go func() {
		ticker := time.NewTicker(DEADLOCK_TICKER)

		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ticker.C:
				for _, hub := range rcv.Hubs {
					if len(hub.broadcast) >= DEADLOCK_WARN {
						rcv.logger.Error(
							"Hub processing might be deadlock with events in the buffer",
							zap.Int("hub", hub.connectionIndex),
							zap.Int("goroutine", hub.goroutineID),
							zap.Int("events", len(hub.broadcast)),
						)
						buf := make([]byte, 1<<16)
						runtime.Stack(buf, true)
						output := fmt.Sprintf("%s", buf)
						splits := strings.Split(output, "goroutine ")

						for _, part := range splits {
							if strings.Contains(part, fmt.Sprintf("%v", hub.goroutineID)) {
								rcv.logger.Error("Trace for possible deadlock goroutine", zap.String("trace", part))
							}
						}
					}
				}
				rcv.logger.Info("Total connection", zap.Int("total", rcv.TotalConnections()))
			case <-rcv.HubsStopCheckingForDeadlock:
				return
			}
		}
	}()
}

func (rcv *Server) HubStop() {
	rcv.logger.Info("stopping hub connections")

	select {
	case rcv.HubsStopCheckingForDeadlock <- true:
	default:
		rcv.logger.Warn("We appear to have already sent the stop checking for deadlocks command")
	}

	for _, hub := range rcv.Hubs {
		hub.Stop()
	}

	rcv.Hubs = []*Hub{}
}

func (rcv *Server) GetHubForUserID(userID string) *Hub {
	if len(rcv.Hubs) == 0 {
		return nil
	}

	hash := fnv.New32a()
	_, _ = hash.Write([]byte(userID))
	index := hash.Sum32() % uint32(len(rcv.Hubs))
	return rcv.Hubs[index]
}

func (rcv *Server) HubRegister(clientConn *ClientConn) error {
	hub := rcv.GetHubForUserID(clientConn.UserID)
	if hub == nil {
		return errors.New("no hub to register")
	}

	hub.Register(clientConn)
	return nil
}

func (rcv *Server) HubUnregister(clientConn *ClientConn) {
	hub := rcv.GetHubForUserID(clientConn.UserID)
	if hub != nil {
		hub.Unregister(clientConn)
	}
}

func (rcv *Server) PushLocal(ctx context.Context, userIDs []string, msg *pb.Event) {
	var newMsg interface {
		GetMessageId() string
		GetConversationId() string
	}

	switch data := msg.GetEvent().(type) {
	case *pb.Event_EventDeleteMessage_:
		newMsg = data.EventDeleteMessage

	case *pb.Event_EventNewMessage:
		newMsg = data.EventNewMessage

	case *pb.Event_EventPing:
		newMsg = nil

	default:
		return
	}

	msgID := ""
	conversationID := ""
	if newMsg != nil {
		msgID = newMsg.GetMessageId()
		conversationID = newMsg.GetConversationId()
	}
	rcv.logger.Info(
		"ChatService.PublishEvent",
		zap.Strings("user_ids", userIDs),
		zap.String("message_id", msgID),
		zap.String("conversation_id", conversationID),
	)
	rp, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		rcv.logger.Error(
			"getting resource path from ctx",
			zap.String("message_id", msgID),
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return
	}
	for _, userID := range userIDs {
		hub := rcv.GetHubForUserID(userID)
		if hub != nil {
			rcv.logger.Info(
				"ChatService.PublishEvent publish msg to user",
				zap.String("user_id", userID),
				zap.String("message_id", msgID),
				zap.String("conversation_id", conversationID),
			)
			hub.Broadcast(&Event{
				UserID:       userID,
				Data:         msg,
				ResourcePath: rp,
			})
		} else {
			rcv.logger.Error(
				"ChatService.PublishEvent publish msg to user error: cannot find hub of user",
				zap.String("user_id", userID),
				zap.String("message_id", msgID),
				zap.String("conversation_id", conversationID),
			)
		}
	}
}

func (rcv *Server) getMapNodeOnlineUserIDs(ctx context.Context, userIDs []string, msg *pb.Event) (map[pgtype.Text][]string, error) {
	var (
		pgUserIDs pgtype.TextArray
		pgSince   pgtype.Timestamptz
	)

	err := multierr.Combine(
		pgUserIDs.Set(userIDs),
		pgSince.Set(time.Now().Add(HEALTHCHECK_TICKER*-2)),
	)
	if err != nil {
		rcv.logger.Error("multierr.Combine", zap.Error(err))
		return nil, err
	}

	// get user online in another node
	mapNodeOnlineUserIDs, err := rcv.onlineUserRepo.Find(ctx, rcv.db, pgUserIDs, pgSince, msg)
	if err != nil {
		rcv.logger.Error("rcv.OnlineUserRepo.Find", zap.Error(err))
	}

	return mapNodeOnlineUserIDs, nil
}

func (rcv *Server) PushMessage(ctx context.Context, userIDs []string, msg *pb.Event, pushMsgOpts domain.MessageToUserOpts) error {
	mapNodeOnlineUserIDs, err := rcv.getMapNodeOnlineUserIDs(ctx, userIDs, msg)
	if err != nil || mapNodeOnlineUserIDs == nil {
		return nil
	}

	// publish localhost first
	usersOnlineOnThisNode := mapNodeOnlineUserIDs[database.Text(rcv.hostName)]
	rcv.PushLocal(ctx, usersOnlineOnThisNode, msg)

	// create a map for fast check userID
	notifiedUserChecklist := make(map[string]struct{})
	for _, userID := range userIDs {
		notifiedUserChecklist[userID] = struct{}{}
	}

	// delete user is online in mapUserID, the rest is offline
	for _, ids := range mapNodeOnlineUserIDs {
		for _, id := range ids {
			delete(notifiedUserChecklist, id)
		}
	}

	// if lists of ignored user provided, ignore them
	for _, id := range pushMsgOpts.Notification.IgnoredUsers {
		delete(notifiedUserChecklist, id)
	}
	usersToSendNoti := []string{}

	// get offline users
	for id := range notifiedUserChecklist {
		usersToSendNoti = append(usersToSendNoti, id)
	}

	if pushMsgOpts.Notification.Enabled && len(usersToSendNoti) > 0 && msg.GetEventNewMessage() != nil {
		err := rcv.pushNotification(ctx, usersToSendNoti, msg.GetEventNewMessage(), pushMsgOpts.Notification)
		if err != nil {
			rcv.logger.Error("rcv.pushNotify", zap.Error(err))
		}
	}

	// remove userIDs in this node
	delete(mapNodeOnlineUserIDs, database.Text(rcv.hostName))
	nodeReceive := make(map[string]*pb.InternalSendMessageRequest_UserIDs)
	for node, userIDs := range mapNodeOnlineUserIDs {
		nodeReceive[node.String] = &pb.InternalSendMessageRequest_UserIDs{
			Ids: userIDs,
		}
	}

	if len(nodeReceive) == 0 {
		return nil
	}

	internalSendMessageRequest := &pb.InternalSendMessageRequest{
		Event:       msg,
		NodeReceive: nodeReceive,
	}

	req, _ := internalSendMessageRequest.Marshal()
	_, err = rcv.JSM.PublishContext(ctx, golib_constants.SubjectSendChatMessageCreated, req)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("rcv.JSM.PublishContext with ChatMessage.Created is failed: %v", err))
	}
	return nil
}

func (rcv *Server) PushMessageDeleted(ctx context.Context, userIDs []string, msg *pb.Event, pushMsgOpts domain.MessageToUserOpts) error {
	mapNodeOnlineUserIDs, err := rcv.getMapNodeOnlineUserIDs(ctx, userIDs, msg)
	if err != nil || mapNodeOnlineUserIDs == nil {
		return nil
	}

	// publish localhost first
	usersOnlineOnThisNode := mapNodeOnlineUserIDs[database.Text(rcv.hostName)]
	rcv.PushLocal(ctx, usersOnlineOnThisNode, msg)

	// remove userIDs in this node
	delete(mapNodeOnlineUserIDs, database.Text(rcv.hostName))
	nodeReceive := make(map[string]*pb.InternalSendMessageRequest_UserIDs)
	for node, userIDs := range mapNodeOnlineUserIDs {
		nodeReceive[node.String] = &pb.InternalSendMessageRequest_UserIDs{
			Ids: userIDs,
		}
	}

	if len(nodeReceive) == 0 {
		return nil
	}

	internalSendMessageRequest := &pb.InternalSendMessageRequest{
		Event:       msg,
		NodeReceive: nodeReceive,
	}

	req, _ := internalSendMessageRequest.Marshal()
	_, err = rcv.JSM.PublishContext(ctx, golib_constants.SubjectChatMessageDeleted, req)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("rcv.JSM.PublishContext with ChatMessage.Created is failed: %v", err))
	}
	return nil
}

func (rcv *Server) pushNotification(ctx context.Context, offlineUserIDs []string, msg *pb.MessageResponse, notiOpt domain.NotificationOpts) error {
	var content string

	switch msg.Type {
	case pb.MESSAGE_TYPE_TEXT:
		content = msg.Content
	case pb.MESSAGE_TYPE_IMAGE:
		fallthrough
	case pb.MESSAGE_TYPE_FILE:
		content = constants.MessagingFileNotificationContent //TODO: align with mobile to use localization later
	default:
		return nil
	}

	var pgUserIDs pgtype.TextArray
	_ = pgUserIDs.Set(offlineUserIDs)
	deviceTokens, err := rcv.userDeviceTokenRepo.Find(ctx, rcv.db, pgUserIDs)
	if err != nil {
		return errors.Wrap(err, "rcv.userDeviceTokenRepo.Find")
	}

	if len(deviceTokens) == 0 {
		return errors.New("no device token")
	}

	data := map[string]string{
		constants.FcmKeyNotificationType: conversation.String(),
		constants.FcmKeyItemID:           msg.ConversationId,
		constants.FcmKeyClickAction:      clickAction,
		constants.FcmKeyConversationName: notiOpt.Title,
		constants.FcmKeyMessageContent:   content,
	}
	m := &messaging.MulticastMessage{
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound:            "default",
					ContentAvailable: true,
				},
			},
			Headers: map[string]string{
				"apns-priority": "5",
			},
		},
	}
	if !notiOpt.Silence {
		m.Notification = &messaging.Notification{
			Title: notiOpt.Title,
			Body:  content,
		}
	}

	rcv.logger.Info("rcv.pushNotify SendTokens", zap.Int("number_of_device_tokens", len(deviceTokens)))
	_, _, errPusher := rcv.notification.Pusher.SendTokens(ctx, m, deviceTokens)
	if errPusher != nil {
		if errPusher.BatchCombinedError != nil {
			rcv.logger.Error("error rcv.notification.Pusher.SendTokens batch error" + errPusher.BatchCombinedError.Error())
		}
		if errPusher.DirectError != nil {
			rcv.logger.Error("error rcv.notification.Pusher.SendTokens - SendMulticast: " + errPusher.DirectError.Error())
			return fmt.Errorf("rcv.notification.Pusher.SendTokens - SendMulticast: %w", errPusher.DirectError)
		}
	}

	return nil
}

func (rcv *Server) PushCluster(ctx context.Context, userIDs []string, msg *pb.Event) {
	err := rcv.PushMessage(ctx, userIDs, msg, domain.MessageToUserOpts{Notification: domain.NotificationOpts{Enabled: false}})
	if err != nil {
		rcv.logger.Error("rcv.Pushmessage", zap.Error(err))
	}
}
