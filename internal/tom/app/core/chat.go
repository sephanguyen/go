package core

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	utils "github.com/manabie-com/backend/internal/tom/app"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: make interface method more generic
type ChatInfra interface {
	PushMessage(ctx context.Context, userIDs []string, msg *pb.Event, pushMsgOpts domain.MessageToUserOpts) error
	PushMessageDeleted(ctx context.Context, userIDs []string, msg *pb.Event, pushMsgOpts domain.MessageToUserOpts) error
}

// TODO: move to notification BC
type NotificationInfra interface {
	RetrievePushedNotificationMessages(ctx context.Context, req *pb.RetrievePushedNotificationMessageRequest) (*pb.RetrievePushedNotificationMessageResponse, error)
}

func NewChatService(
	logger *zap.Logger,
	notificationInfra NotificationInfra,
	chatInfra ChatInfra,
	wrapperDB database.Ext,
	jsm nats.JetStreamManagement,
) *ChatServiceImpl {
	svc := &ChatServiceImpl{
		db:                  wrapperDB,
		ChatInfra:           chatInfra,
		NotificatiionPusher: notificationInfra,
		logger:              logger,
		JSM:                 jsm,
	}
	return svc
}

type ChatService interface {
	// for other module to customize whether to persist or not
	SendMessageToConversation(ctx context.Context, req *pb.SendMessageRequest, opts domain.MessageToConversationOpts) (*pb.MessageResponse, error)
	SendMessageToConversations(ctx context.Context, reqs []*pb.SendMessageRequest, opts domain.MessageToConversationOpts) error

	// directly send message to specific users
	SendMessageToUsers(ctx context.Context, userIDs []string, msg *pb.Event, opts domain.MessageToUserOpts) error
}

type ChatServiceImpl struct {
	db                  database.Ext
	logger              *zap.Logger
	JSM                 nats.JetStreamManagement
	ChatInfra           ChatInfra
	NotificatiionPusher NotificationInfra

	ConversationRepo interface {
		Create(context.Context, database.QueryExecer, *domain.Conversation) error
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*domain.Conversation, error)
		FindByStudentQuestionID(context.Context, database.QueryExecer, pgtype.Text) (*domain.Conversation, error)
		FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]domain.ConversationFull, error)
		Update(ctx context.Context, db database.QueryExecer, c *domain.Conversation) error
		SetStatus(ctx context.Context, db database.QueryExecer, cID pgtype.Text, status pgtype.Text) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, conversations []*domain.Conversation) error
		FindBySchoolIDs(ctx context.Context, db database.QueryExecer, schoolIDs pgtype.TextArray, limit pgtype.Int4, offset pgtype.Text) ([]pgtype.Text, []pgtype.Timestamptz, error)
	}
	MessageRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, messages []*domain.Message) error
		Create(context.Context, database.QueryExecer, *domain.Message) error
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*domain.Message, error)
		FindAllMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, limit uint, endAt pgtype.Timestamptz) ([]*domain.Message, error)
		CountMessagesSince(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, since *pgtype.Timestamptz) (int, error)
		GetLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz, includeSystemMsg bool) ([]*domain.Message, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, userID, id pgtype.Text) error
		FindMessages(ctx context.Context, db database.QueryExecer, args *domain.FindMessagesArgs) ([]*domain.Message, error)
		FindLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *domain.FindMessagesArgs) ([]*domain.Message, error)
	}
	ConversationMemberRepo interface {
		Create(ctx context.Context, db database.QueryExecer, c *domain.ConversationMembers) error
		SetSeenAt(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text, seenAt pgtype.Timestamptz) error
		SetNotifyAt(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text, notifyAt pgtype.Timestamptz) error
		SetStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.TextArray, status pgtype.Text) error
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (mapConversationID map[pgtype.Text][]*domain.ConversationMembers, err error)
		FindUnseenSince(context.Context, database.QueryExecer, pgtype.Timestamptz) ([]*domain.ConversationMembers, error)
		GetSeenAt(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*pgtype.Timestamptz, error)
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (mapUserID map[pgtype.Text]domain.ConversationMembers, err error)
		Find(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, role, status pgtype.Text) (*domain.ConversationMembers, error)
		Update(ctx context.Context, db database.QueryExecer, c *domain.ConversationMembers) error
		FindByCIDAndUserID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, userID pgtype.Text) (c *domain.ConversationMembers, err error)
		SetStatusByConversationID(ctx context.Context, db database.QueryExecer, conversationID, status pgtype.Text) error
	}
}

func (rcv *ChatServiceImpl) DeleteMessage(ctx context.Context, req *tpb.DeleteMessageRequest) (*tpb.DeleteMessageResponse, error) {
	userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)

	pgUserID := database.Text(userID)
	pgMessageID := database.Text(req.MessageId)

	message, err := rcv.MessageRepo.FindByID(ctx, rcv.db, pgMessageID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "not found message")
		}

		return nil, status.Error(codes.Unknown, err.Error())
	}

	conversation, err := rcv.ConversationRepo.FindByID(ctx, rcv.db, database.Text(message.ConversationID.String))
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "not found conversation")
		}

		return nil, status.Error(codes.Unknown, err.Error())
	}

	errorPrefix := "permission denied"
	switch conversation.ConversationType.String {
	case tpb.ConversationType_CONVERSATION_LESSON.String():
		if !utils.IsStaff(userGroup) && userID != message.UserID.String {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%s: current user is not either staff or the owner", errorPrefix))
		}
	case tpb.ConversationType_CONVERSATION_STUDENT.String(), tpb.ConversationType_CONVERSATION_PARENT.String(), tpb.ConversationType_CONVERSATION_LESSON_PRIVATE.String():
		if userID != message.UserID.String {
			return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%s: current user is not the owner", errorPrefix))
		}
	default:
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("%s: invalid conversation type", errorPrefix))
	}

	conversationID := message.ConversationID.String

	err = rcv.MessageRepo.SoftDelete(ctx, rcv.db, pgUserID, pgMessageID)

	if err != nil {
		return nil, fmt.Errorf("deleteMessage: %w", err)
	}

	// only broadcast event seen to client, not save the message to db
	_, err = rcv.broadcastMessageDeleted(ctx, userID, conversationID, req.MessageId)

	if err != nil {
		return nil, fmt.Errorf("sendMessage: %w", err)
	}

	return &tpb.DeleteMessageResponse{}, nil
}

func (rcv *ChatServiceImpl) SeenMessage(ctx context.Context, req *pb.SeenMessageRequest) (*pb.SeenMessageResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)

	// only broadcast event seen to client, not save the message to db
	_, err := rcv.broadcastMessage(ctx, userID, &pb.SendMessageRequest{
		ConversationId: req.ConversationId,
		Message:        pb.CODES_MESSAGE_TYPE_SEEN_CONVERSATION.String(),
		Type:           pb.MESSAGE_TYPE_SYSTEM,
	})
	if err != nil {
		return nil, fmt.Errorf("seenMessage: %w", err)
	}

	var (
		pgCID pgtype.Text
		pgUID pgtype.Text
	)

	_ = pgCID.Set(req.ConversationId)
	_ = pgUID.Set(userID)

	// upsert table conversation status
	err = rcv.markSeenConversation(ctx, pgCID, pgUID)
	if err != nil {
		return nil, fmt.Errorf("markSeenConversation: %w", err)
	}

	return &pb.SeenMessageResponse{}, nil
}

func (rcv *ChatServiceImpl) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	message, err := rcv.SendMessageToConversation(ctx, req, domain.MessageToConversationOpts{Persist: true, AsUser: true})
	if err != nil {
		return nil, err
	}

	rcv.logger.Info(
		"ChatService.SendMessage send message",
		zap.String("user_id", userID),
		zap.String("conversation_id", message.ConversationId),
		zap.String("message_id", message.MessageId),
		zap.String("type", message.Type.String()),
	)

	var (
		pgCID pgtype.Text
		pgUID pgtype.Text
	)

	_ = pgCID.Set(req.ConversationId)
	_ = pgUID.Set(userID)

	if err := rcv.markSeenConversation(ctx, pgCID, pgUID); err != nil {
		return nil, errors.Wrap(err, "rcv.markSeenConversation")
	}

	return &pb.SendMessageResponse{
		MessageId:      message.MessageId,
		LocalMessageId: req.LocalMessageId,
	}, nil
}
func requestToMessageEntity(req *pb.SendMessageRequest, userID string) (*domain.Message, error) {
	message := new(domain.Message)
	database.AllNullEntity(message)

	messageID := idutil.ULIDNow()
	if userID != "" {
		err := message.UserID.Set(userID)
		if err != nil {
			return nil, err
		}
	} else {
		_ = message.UserID.Set(nil)
	}
	err := multierr.Combine(
		message.ID.Set(messageID),
		message.ConversationID.Set(req.ConversationId),
		message.Message.Set(req.Message),
		message.UrlMedia.Set(req.UrlMedia),
		message.Type.Set(req.Type.String()),
		message.TargetUser.Set(req.TargetUser),
	)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (rcv *ChatServiceImpl) persistMessage(ctx context.Context, req *pb.SendMessageRequest, userID string) (*domain.Message, error) {
	message, err := requestToMessageEntity(req, userID)
	if err != nil {
		return nil, err
	}

	// 2.2 save to db
	err = rcv.MessageRepo.Create(ctx, rcv.db, message)
	if err != nil {
		if e, isErr := errors.Cause(err).(*pgconn.PgError); isErr && e.Code == "23505" { // unique_violation
			return nil, status.Error(codes.AlreadyExists, e.Message)
		}

		return nil, status.Error(codes.Unknown, err.Error())
	}
	return message, nil
}

func (rcv *ChatServiceImpl) SendMessageToUsers(ctx context.Context, userIDs []string, msg *pb.Event, pushMsgOpts domain.MessageToUserOpts) error {
	return rcv.ChatInfra.PushMessage(ctx, userIDs, msg, pushMsgOpts)
}

func (rcv *ChatServiceImpl) SendMessageToConversations(ctx context.Context, reqs []*pb.SendMessageRequest, opts domain.MessageToConversationOpts) error {
	userID := ""
	if opts.AsUser {
		userID = interceptors.UserIDFromContext(ctx)
	}
	// 1. validate conversation
	// conversationID := database.Text(req.ConversationId)
	convIDs := make([]string, 0, len(reqs))
	for _, req := range reqs {
		convIDs = append(convIDs, req.ConversationId)
	}

	var persistedMsges []*domain.Message

	if opts.Persist {
		persistedMsges = make([]*domain.Message, 0, len(reqs))
		for _, req := range reqs {
			message, err := requestToMessageEntity(req, userID)
			if err != nil {
				return err
			}
			persistedMsges = append(persistedMsges, message)
		}
		err := rcv.MessageRepo.BulkUpsert(ctx, rcv.db, persistedMsges)

		// 2.2 save to db
		if err != nil {
			return fmt.Errorf("MessageRepo.BulkUpsert %w", err)
		}
	}
	mapConvMembers, err := rcv.ConversationMemberRepo.FindByConversationIDs(ctx, rcv.db, database.TextArray(convIDs))
	if err != nil {
		return fmt.Errorf("ConversationMemberRepo.FindByConversationIDs %w", err)
	}

	conversationInfoMap, err := rcv.ConversationRepo.FindByIDsReturnMapByID(ctx, rcv.db, database.TextArray(convIDs))
	if err != nil {
		return fmt.Errorf("ConversationRepo.FindByIDsReturnMapByID %w", err)
	}
	return rcv.bulkPushMessages(ctx, userID, reqs, persistedMsges, mapConvMembers, conversationInfoMap)
}

func (rcv *ChatServiceImpl) bulkPushMessages(
	ctx context.Context,
	userID string,
	originalReqs []*pb.SendMessageRequest,
	persistedMsges []*domain.Message,
	mapConvMembers map[pgtype.Text][]*domain.ConversationMembers,
	conversationInfoMap map[pgtype.Text]domain.ConversationFull,
) error {
	for idx, req := range originalReqs {
		conversationMembers, exist := mapConvMembers[database.Text(req.ConversationId)]
		if !exist {
			rcv.logger.Warn("not found conversation members for conversation", zap.String("conversation", req.ConversationId))
			continue
		}
		var (
			messageID string
			userIDs   = make([]string, 0, len(conversationMembers))
			found     bool
		)
		if len(persistedMsges) > idx {
			messageID = persistedMsges[idx].ID.String
		}

		for _, member := range conversationMembers {
			if member.Status.String == domain.ConversationStatusInActive {
				continue
			}
			if member.UserID.String == userID {
				found = true
			}

			userIDs = append(userIDs, member.UserID.String)
		}

		if userID != "" && !found {
			rcv.logger.Warn("user sending message is not a member of conversation", zap.String("conversation", req.ConversationId))
			continue
		}

		conversation, exist := conversationInfoMap[database.Text(req.ConversationId)]
		if !exist {
			rcv.logger.Warn("unable to find info of conversation %s for sending message", zap.String("conversation", req.ConversationId))
			continue
		}
		var (
			conversationName string
			conversationType string
			isSilence        bool
		)
		if exist {
			isSilence = isSilentConversation(&conversation.Conversation)
			conversationName = conversation.Conversation.Name.String
			conversationType = conversation.Conversation.ConversationType.String
		}

		// 3. broadcast message to all user in conversation
		// 3.1 prepare payload
		messageResponse := &pb.MessageResponse{
			MessageId:        messageID,
			ConversationId:   req.ConversationId,
			ConversationName: conversationName,
			UserId:           userID,
			Content:          req.Message,
			UrlMedia:         req.UrlMedia,
			Type:             req.Type,
			CreatedAt:        types.TimestampNow(),
			LocalMessageId:   req.LocalMessageId,
			TargetUser:       req.TargetUser,
		}

		userIDs = golibs.GetUniqueElementStringArray(userIDs)
		notiOpt := domain.NotificationOpts{
			Enabled: true,
			Silence: isSilence,
			Title:   conversationName,
		}
		if userID != "" {
			notiOpt.IgnoredUsers = []string{userID}
		}
		// TODO: consider putting this in transaction or not
		err := domain.DomainEvtBus.Publish(ctx, domain.MessageSentEventStr, domain.MessageSentEvent{
			ConversationType: conversationType,
			ConversationID:   req.ConversationId,
		})
		if err != nil {
			return fmt.Errorf("domain.DomainEvtBus.Publish %w", err)
		}

		err = rcv.ChatInfra.PushMessage(ctx, userIDs, &pb.Event{
			Event: &pb.Event_EventNewMessage{
				EventNewMessage: messageResponse,
			},
		}, domain.MessageToUserOpts{
			Notification: notiOpt,
		})
		if err != nil {
			rcv.logger.Error("rcv.MessagePusher.PushMessage", zap.Error(err))
		}
	}
	return nil
}

func (rcv *ChatServiceImpl) SendMessageToConversation(ctx context.Context, req *pb.SendMessageRequest, opts domain.MessageToConversationOpts) (*pb.MessageResponse, error) {
	userID := ""
	if opts.AsUser {
		userID = interceptors.UserIDFromContext(ctx)
	}
	// 1. validate conversation
	conversationID := database.Text(req.ConversationId)

	conversationMembers, err := rcv.ConversationMemberRepo.FindByConversationID(ctx, rcv.db, conversationID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "not found conversation")
		}

		return nil, status.Error(codes.Unknown, err.Error())
	}

	var (
		userIDs = make([]string, 0, len(conversationMembers))
		found   bool
	)

	for uID := range conversationMembers {
		if uID.String == userID {
			found = true
		}

		userIDs = append(userIDs, uID.String)
	}

	if userID != "" && !found {
		return nil, status.Error(codes.NotFound, "not found conversation")
	}
	var msgID string
	if opts.Persist {
		msg, err := rcv.persistMessage(ctx, req, userID)
		if err != nil {
			return nil, err
		}

		msgID = msg.ID.String
	}

	// publish message evt
	logger := ctxzap.Extract(ctx)
	conversation, err := rcv.ConversationRepo.FindByID(ctx, rcv.db, conversationID)
	if err != nil {
		logger.Warn("unable to get conversation conversationRepo.FindByID: ", zap.Error(err))
	}
	var (
		conversationName string
		conversationType string
		isSilence        bool
	)
	if conversation != nil {
		isSilence = isSilentConversation(conversation)
		conversationName = conversation.Name.String
		conversationType = conversation.ConversationType.String
	}

	// 3. broadcast message to all user in conversation
	// 3.1 prepare payload
	messageResponse := &pb.MessageResponse{
		MessageId:        msgID,
		ConversationId:   req.ConversationId,
		ConversationName: conversationName,
		UserId:           userID,
		Content:          req.Message,
		UrlMedia:         req.UrlMedia,
		Type:             req.Type,
		CreatedAt:        types.TimestampNow(),
		LocalMessageId:   req.LocalMessageId,
		TargetUser:       req.TargetUser,
	}

	userIDs = golibs.GetUniqueElementStringArray(userIDs)
	notiOpt := domain.NotificationOpts{
		Enabled: true,
		Silence: isSilence,
		Title:   conversationName,
	}
	if userID != "" {
		notiOpt.IgnoredUsers = []string{userID}
	}
	// TODO: consider putting this in transaction or not
	err = domain.DomainEvtBus.Publish(ctx, domain.MessageSentEventStr, domain.MessageSentEvent{
		ConversationType: conversationType,
		ConversationID:   conversationID.String,
	})
	if err != nil {
		return nil, err
	}

	err = rcv.ChatInfra.PushMessage(ctx, userIDs, &pb.Event{
		Event: &pb.Event_EventNewMessage{
			EventNewMessage: messageResponse,
		},
	}, domain.MessageToUserOpts{
		Notification: notiOpt,
	})
	if err != nil {
		rcv.logger.Error("rcv.MessagePusher.PushMessage", zap.Error(err))
	}

	return messageResponse, nil
}

func (rcv *ChatServiceImpl) markSeenConversation(ctx context.Context, conversationID pgtype.Text, userID pgtype.Text) error {
	var seenAt pgtype.Timestamptz
	_ = seenAt.Set(time.Now())

	return rcv.ConversationMemberRepo.SetSeenAt(ctx, rcv.db, conversationID, userID, seenAt)
}

func (rcv *ChatServiceImpl) findConversationMembersExceptCurrentUser(ctx context.Context, userID string, conversationID string) ([]string, error) {
	pgConversationID := database.Text(conversationID)

	conversationMembers, err := rcv.ConversationMemberRepo.FindByConversationID(ctx, rcv.db, pgConversationID)
	if err != nil {
		if errors.Cause(err) == pgx.ErrNoRows {
			return nil, status.Error(codes.NotFound, "not found conversation")
		}

		return nil, status.Error(codes.Unknown, err.Error())
	}

	var found bool
	userIDs := make([]string, 0)

	for uID := range conversationMembers {
		if uID.String == userID {
			found = true
		}

		userIDs = append(userIDs, uID.String)
	}

	if userID != "" && !found {
		return nil, status.Error(codes.NotFound, "not found conversation")
	}

	return userIDs, nil
}

func (rcv *ChatServiceImpl) broadcastMessageDeleted(ctx context.Context, userID string, conversationID string, messageID string) (*pb.Event_EventDeleteMessage, error) {
	userIDs, err := rcv.findConversationMembersExceptCurrentUser(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	deleteMessageEvent := &pb.Event_EventDeleteMessage{
		ConversationId: conversationID,
		MessageId:      messageID,
		DeletedBy:      userID,
	}

	err = rcv.ChatInfra.PushMessageDeleted(ctx, userIDs, &pb.Event{
		Event: &pb.Event_EventDeleteMessage_{
			EventDeleteMessage: deleteMessageEvent,
		},
	}, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{userID},
		},
	})
	if err != nil {
		rcv.logger.Error("chatInfra.PushMessageDeleted", zap.Error(err))
	}

	return deleteMessageEvent, nil
}

// broadcastMessage only broadcast to nodes
// not save the message to the database
func (rcv *ChatServiceImpl) broadcastMessage(ctx context.Context, userID string, req *pb.SendMessageRequest) (*pb.MessageResponse, error) {
	// 1. validate conversation
	userIDs, err := rcv.findConversationMembersExceptCurrentUser(ctx, userID, req.ConversationId)
	if err != nil {
		return nil, err
	}

	// 2. broadcast message to all user in conversation
	// 2.1 prepare payload
	messageResponse := &pb.MessageResponse{
		ConversationId: req.ConversationId,
		UserId:         userID,
		Content:        req.Message,
		UrlMedia:       req.UrlMedia,
		Type:           req.Type,
		CreatedAt:      types.TimestampNow(),
		LocalMessageId: req.LocalMessageId,
		TargetUser:     req.TargetUser,
	}
	// 2.2 broadcast to clients
	err = rcv.ChatInfra.PushMessage(ctx, userIDs, &pb.Event{
		Event: &pb.Event_EventNewMessage{
			EventNewMessage: messageResponse,
		},
	}, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{userID},
		},
	})
	if err != nil {
		rcv.logger.Error("chatInfra.PushMessage", zap.Error(err))
	}

	return messageResponse, nil
}

func (rcv *ChatServiceImpl) RetrievePushedNotificationMessages(ctx context.Context, req *pb.RetrievePushedNotificationMessageRequest) (*pb.RetrievePushedNotificationMessageResponse, error) {
	return rcv.NotificatiionPusher.RetrievePushedNotificationMessages(ctx, req)
}
