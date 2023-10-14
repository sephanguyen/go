package core

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamps "google.golang.org/protobuf/types/known/timestamppb"
)

type ChatReader struct {
	Logger *zap.Logger
	DB     database.Ext

	ConversationMemberRepo interface {
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (mapUserID map[pgtype.Text]core.ConversationMembers, err error)
		FindByConversationIDAndStatus(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, status pgtype.Text) (mapUserID map[pgtype.Text]core.ConversationMembers, err error)
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (mapConversationID map[pgtype.Text][]*core.ConversationMembers, err error)
	}
	ConversationRepo interface {
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*core.Conversation, error)
		FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]core.ConversationFull, error)
	}
	MessageRepo interface {
		GetLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz, includeSystemMsg bool) ([]*core.Message, error)
		Create(context.Context, database.QueryExecer, *core.Message) error
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *core.Message, err error)
		FindAllMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, limit uint, endAt pgtype.Timestamptz) ([]*core.Message, error)
		CountMessagesSince(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, since *pgtype.Timestamptz) (int, error)
		GetLatestMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (*core.Message, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, userID, id pgtype.Text) error
		FindMessages(ctx context.Context, db database.QueryExecer, args *core.FindMessagesArgs) ([]*core.Message, error)
		FindLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *core.FindMessagesArgs) ([]*core.Message, error)
		FindPrivateLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *core.FindMessagesArgs) ([]*core.Message, error)
	}
}

// Switch to using GetConversationV2, this function will be removed when moving tom_dart_client to manabuf is done
func (rcv *ChatReader) GetConversation(ctx context.Context, req *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	var (
		err            error
		conversationID = req.ConversationId
		userID         = interceptors.UserIDFromContext(ctx)
	)

	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty conversation id in request")
	}

	conversation, err := rcv.ConversationRepo.FindByID(ctx, rcv.DB, database.Text(conversationID))
	if err != nil {
		return nil, fmt.Errorf("rcv.conversationRepo.FindByID: %w", err)
	}
	var (
		latestMsg    *core.Message
		latestMsgErr error
	)
	// TODO: normal conversation vs session conversation, no more lesson type in chat
	if conversation.ConversationType.String == pb.CONVERSATION_LESSON.String() {
		isPrivateConversation := false
		latestMsg, latestMsgErr = rcv.getConversationLessonLatestMessage(ctx, conversation.ID.String, isPrivateConversation)
	} else {
		latestMsg, latestMsgErr = rcv.MessageRepo.GetLatestMessageByConversation(ctx, rcv.DB, database.Text(conversationID))
	}

	if latestMsgErr != nil {
		if errors.Is(latestMsgErr, pgx.ErrNoRows) {
			latestMsg = nil
		} else {
			return nil, fmt.Errorf("rcv.messageRepo.GetLatestMessageByConversation: %w", latestMsgErr)
		}
	}

	conversationMembers, err := rcv.ConversationMemberRepo.FindByConversationIDAndStatus(ctx, rcv.DB, database.Text(conversationID), pgtype.Text{Status: pgtype.Null})
	if err != nil {
		return nil, fmt.Errorf("rcv.conversationMemberRepo.FindByConversationID: %w", err)
	}

	users := make([]*pb.Conversation_User, 0, len(conversationMembers))
	var seen bool
	for _, u := range conversationMembers {
		if u.Status.String == core.ConversationStatusActive && u.UserID.String == userID {
			if latestMsg != nil {
				seen = u.SeenAt.Time.After(latestMsg.CreatedAt.Time)
			}
		}
		seenAt, err := types.TimestampProto(u.SeenAt.Time)
		if err != nil {
			return nil, fmt.Errorf("types.TimestampProto: %w", err)
		}
		users = append(users, &pb.Conversation_User{
			Id:        u.UserID.String,
			Group:     u.Role.String,
			IsPresent: u.Status.String == core.ConversationStatusActive,
			SeenAt:    seenAt,
		})
	}
	if latestMsg == nil {
		seen = true
	}
	var latestMsgPb *pb.MessageResponse
	if latestMsg != nil {
		latestMsgPb = toMessageResponse(latestMsg)
	}

	return &pb.GetConversationResponse{
		Conversation: &pb.Conversation{
			ConversationId:   conversationID,
			Seen:             seen,
			LastMessage:      latestMsgPb,
			Status:           pb.ConversationStatus(pb.ConversationStatus_value[conversation.Status.String]),
			Users:            users,
			ConversationType: pb.ConversationType(pb.ConversationType_value[conversation.ConversationType.String]),
			ConversationName: conversation.Name.String,
		},
	}, nil
}

// this function is temporary, after DDD, no more conversation lesson in tom
func (rcv *ChatReader) getConversationLessonLatestMessage(ctx context.Context, conversationID string, isPrivateConversation bool) (*core.Message, error) {
	args := &core.FindMessagesArgs{
		EndAt:            database.Timestamptz(time.Now()),
		Limit:            1,
		IncludeSystemMsg: false,
	}
	var err error
	var messages []*core.Message
	if isPrivateConversation {
		messages, err = rcv.MessageRepo.FindPrivateLessonMessages(ctx, rcv.DB, database.Text(conversationID), args)
	} else {
		messages, err = rcv.MessageRepo.FindLessonMessages(ctx, rcv.DB, database.Text(conversationID), args)
	}

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		rcv.Logger.Error("MessageRepo.FindByID", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	var latestMsg *core.Message
	if len(messages) == 1 {
		latestMsg = messages[0]
	}
	return latestMsg, nil
}

func (rcv *ChatReader) ConversationDetail(ctx context.Context, req *pb.ConversationDetailRequest) (*pb.ConversationDetailResponse, error) {
	logger := ctxzap.Extract(ctx)

	conversationID := database.Text(req.ConversationId)

	var endAtTime time.Time
	if req.GetEndAt() != nil {
		endAtTime = time.Unix(req.EndAt.Seconds, 0).UTC()
	} else {
		endAtTime = time.Now().UTC()
	}
	endAt := database.Timestamptz(endAtTime)

	args := &core.FindMessagesArgs{
		ConversationID:   conversationID,
		EndAt:            endAt,
		Limit:            req.Limit,
		IncludeSystemMsg: false,
	}

	messages, err := rcv.MessageRepo.FindMessages(ctx, rcv.DB, args)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Unknown, err.Error())
	}

	resp := new(pb.ConversationDetailResponse)
	for _, message := range messages {
		resp.Messages = append(resp.Messages, toMessageResponse(message))
	}
	return resp, nil
}

func (rcv *ChatReader) GetConversationV2(ctx context.Context, req *tpb.GetConversationV2Request) (*tpb.GetConversationV2Response, error) {
	var (
		err            error
		conversationID = req.ConversationId
		userID         = interceptors.UserIDFromContext(ctx)
	)

	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty conversation id in request")
	}

	conversation, err := rcv.ConversationRepo.FindByID(ctx, rcv.DB, database.Text(conversationID))
	if err != nil {
		return nil, fmt.Errorf("rcv.conversationRepo.FindByID: %w", err)
	}
	var (
		latestMsg    *core.Message
		latestMsgErr error
	)

	switch conversation.ConversationType.String {
	case tpb.ConversationType_CONVERSATION_LESSON.String():
		isPrivateConversation := false
		latestMsg, latestMsgErr = rcv.getConversationLessonLatestMessage(ctx, conversation.ID.String, isPrivateConversation)
	case tpb.ConversationType_CONVERSATION_LESSON_PRIVATE.String():
		isPrivateConversation := true
		latestMsg, latestMsgErr = rcv.getConversationLessonLatestMessage(ctx, conversation.ID.String, isPrivateConversation)
	default:
		latestMsg, latestMsgErr = rcv.MessageRepo.GetLatestMessageByConversation(ctx, rcv.DB, database.Text(conversationID))
	}

	if latestMsgErr != nil {
		if errors.Is(latestMsgErr, pgx.ErrNoRows) {
			latestMsg = nil
		} else {
			return nil, fmt.Errorf("rcv.messageRepo.GetLatestMessageByConversation: %w", latestMsgErr)
		}
	}

	conversationMembers, err := rcv.ConversationMemberRepo.FindByConversationIDAndStatus(ctx, rcv.DB, database.Text(conversationID), pgtype.Text{Status: pgtype.Null})
	if err != nil {
		return nil, fmt.Errorf("rcv.conversationMemberRepo.FindByConversationID: %w", err)
	}

	users := make([]*tpb.Conversation_User, 0, len(conversationMembers))
	var seen bool
	for _, u := range conversationMembers {
		if u.Status.String == core.ConversationStatusActive && u.UserID.String == userID {
			if latestMsg != nil {
				seen = u.SeenAt.Time.After(latestMsg.CreatedAt.Time)
			}
		}
		users = append(users, &tpb.Conversation_User{
			Id:        u.UserID.String,
			Group:     cpb.UserGroup(cpb.UserGroup_value[u.Role.String]),
			IsPresent: u.Status.String == core.ConversationStatusActive,
			SeenAt:    timestamps.New(u.SeenAt.Time),
		})
	}
	if latestMsg == nil {
		seen = true
	}
	var latestMsgPb *tpb.MessageResponse
	if latestMsg != nil {
		latestMsgPb = toMessageResponseV2(latestMsg)
	}

	return &tpb.GetConversationV2Response{
		Conversation: &tpb.Conversation{
			ConversationId:   conversationID,
			Seen:             seen,
			LastMessage:      latestMsgPb,
			Status:           tpb.ConversationStatus(tpb.ConversationStatus_value[conversation.Status.String]),
			Users:            users,
			ConversationType: tpb.ConversationType(tpb.ConversationType_value[conversation.ConversationType.String]),
			ConversationName: conversation.Name.String,
		},
	}, nil
}
