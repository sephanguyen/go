package lesson

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"
	"github.com/manabie-com/backend/internal/tom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatReader struct {
	DB database.Ext
	// ChatService      *ChatService
	ConversationRepo interface {
		FindByLessonID(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (*domain.Conversation, error)
		FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]domain.ConversationFull, error)
	}
	MessageRepo interface {
		FindLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *domain.FindMessagesArgs) ([]*domain.Message, error)
		FindPrivateLessonMessages(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, args *domain.FindMessagesArgs) ([]*domain.Message, error)
		FindByID(ctx context.Context, db database.QueryExecer, msgID pgtype.Text) (*domain.Message, error)
		GetLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz, includeSystemMsg bool) ([]*domain.Message, error)
	}
	ConversationMemberRepo interface {
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (mapConversationID map[pgtype.Text][]*domain.ConversationMembers, err error)
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (mapUserID map[pgtype.Text]domain.ConversationMembers, err error)
	}
	ConversationLessonRepo interface {
		FindByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs pgtype.TextArray, includeSoftdeleted bool) ([]*lentities.ConversationLesson, error)
		UpdateLatestStartTime(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, latestStartTime pgtype.Timestamptz) error
	}
	PrivateConversationLessonRepo interface {
		UpdateLatestStartTime(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, latestStartTime pgtype.Timestamptz) error
	}
}

func NewLessonChatReader(db database.Ext) *ChatReader {
	return &ChatReader{
		DB:                            db,
		ConversationRepo:              &repositories.ConversationRepo{},
		MessageRepo:                   &repositories.MessageRepo{},
		ConversationLessonRepo:        &repositories.ConversationLessonRepo{},
		ConversationMemberRepo:        &repositories.ConversationMemberRepo{},
		PrivateConversationLessonRepo: &repositories.PrivateConversationLessonRepo{},
	}
}
func (s *ChatReader) ListConversationByLessons(ctx context.Context, req *tpb.ListConversationByLessonsRequest) (*tpb.ListConversationByLessonsResponse, error) {
	cl, err := s.ConversationLessonRepo.FindByLessonIDs(ctx, s.DB, database.TextArray(req.LessonIds), false)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("s.ConversationLessonRepo.FindByLessonIDs: %s", err.Error()))
	}
	conversationIDs := make([]string, 0, len(cl))
	convLessonMap := map[string]string{}
	for _, item := range cl {
		conversationIDs = append(conversationIDs, item.ConversationID.String)
		convLessonMap[item.ConversationID.String] = item.LessonID.String
	}
	mapConversationMembers, err := s.ConversationMemberRepo.FindByConversationIDs(ctx, s.DB, database.TextArray(conversationIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationMemberRepo.FindByConversationIDs: %w", err).Error())
	}

	conversationMap, err := s.ConversationRepo.FindByIDsReturnMapByID(ctx, s.DB, database.TextArray(conversationIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationRepo.FindByIDsReturnMapByID: %w", err).Error())
	}
	var pgEndAt pgtype.Timestamptz
	_ = pgEndAt.Set(time.Now())
	messages, err := s.MessageRepo.GetLastMessageByConversationIDs(ctx, s.DB, database.TextArray(conversationIDs), uint(len(conversationIDs)), pgEndAt, false)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("MessageRepo.GetLastMessageByConversationIDs: %w", err).Error())
	}
	messageMap := convertToMapMessages(messages)

	resp := map[string]*tpb.Conversation{}
	for _, c := range conversationMap {
		lesson := convLessonMap[c.Conversation.ID.String]
		conversationMembers := mapConversationMembers[c.Conversation.ID]
		var users = make([]*tpb.Conversation_User, 0, len(conversationMembers))
		for _, u := range conversationMembers {
			users = append(users, &tpb.Conversation_User{
				Id:        u.UserID.String,
				Group:     cpb.UserGroup(cpb.UserGroup_value[u.Role.String]),
				IsPresent: u.Status.String == domain.ConversationStatusActive,
			})
		}
		var lastmessage *tpb.MessageResponse
		if m, ok := messageMap[c.Conversation.ID]; ok {
			lastmessage = toMessagePb(m)
		}
		resp[lesson] = &tpb.Conversation{
			ConversationId:   c.Conversation.ID.String,
			LastMessage:      lastmessage,
			Status:           tpb.ConversationStatus(tpb.CodesMessageType_value[c.Conversation.Status.String]),
			ConversationType: tpb.ConversationType(tpb.ConversationType_value[c.Conversation.ConversationType.String]),
			Users:            users,
			ConversationName: c.Conversation.Name.String,
			IsReplied:        c.IsReply.Bool,
			Owner:            c.Conversation.Owner.String,
			StudentId:        c.StudentID.String,
		}
	}
	return &tpb.ListConversationByLessonsResponse{
		Conversations: resp,
	}, nil
}

func (s *ChatReader) RefreshLiveLessonSession(ctx context.Context, req *tpb.RefreshLiveLessonSessionRequest) (*tpb.RefreshLiveLessonSessionResponse, error) {
	logger := ctxzap.Extract(ctx)
	lessonID := req.GetLessonId()
	if lessonID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty lesson id")
	}

	now := database.Timestamptz(time.Now())

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		er := s.ConversationLessonRepo.UpdateLatestStartTime(ctx, s.DB, database.Text(lessonID), now)
		if er != nil {
			if errors.Is(er, pgx.ErrNoRows) {
				return status.Error(codes.NotFound, "conversation for lesson does not exist")
			}
			return status.Error(codes.Internal, er.Error())
		}

		er = s.PrivateConversationLessonRepo.UpdateLatestStartTime(ctx, s.DB, database.Text(lessonID), now)
		if er != nil {
			return status.Error(codes.Internal, er.Error())
		}

		return nil
	})
	if err != nil {
		logger.Error("ConversationLessonRepo.FindAndUpdateLatestCallID", zap.Error(err))
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}

	return &tpb.RefreshLiveLessonSessionResponse{}, nil
}

func (s *ChatReader) LiveLessonConversationDetail(ctx context.Context, req *tpb.LiveLessonConversationDetailRequest) (*tpb.LiveLessonConversationDetailResponse, error) {
	userInReq := interceptors.UserIDFromContext(ctx)
	logger := ctxzap.Extract(ctx)
	lessonID := req.GetLessonId()
	if lessonID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty lesson id")
	}
	conversation, err := s.ConversationRepo.FindByLessonID(ctx, s.DB, database.Text(lessonID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logger.Error("ConversationRepo.FindByLessonID", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		return nil, status.Error(codes.NotFound, "not found conversation")
	}
	members, err := s.ConversationMemberRepo.FindByConversationID(ctx, s.DB, database.Text(conversation.ID.String))

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	args := &domain.FindMessagesArgs{
		EndAt:                database.Timestamptz(time.Now()),
		Limit:                1,
		IncludeMessageTypes:  pgtype.TextArray{Status: pgtype.Null},
		ExcludeMessagesTypes: database.TextArray(lessonIgnoredSystemMessages), // ignore message type system
	}

	msges, err := s.MessageRepo.FindLessonMessages(ctx, s.DB, database.Text(conversation.ID.String), args)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		logger.Error("MessaageRepo.FindByID", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	var users = make([]*tpb.Conversation_User, 0, len(members))
	var seen, foundUser bool
	var latestMsg *domain.Message
	if len(msges) == 1 {
		latestMsg = msges[0]
	}

	for _, u := range members {
		if u.Status.String == domain.ConversationStatusActive && u.UserID.String == userInReq {
			foundUser = true
			if latestMsg != nil {
				seen = u.SeenAt.Time.After(latestMsg.CreatedAt.Time)
			} else {
				seen = true
			}
		}
		users = append(users, &tpb.Conversation_User{
			Id:        u.UserID.String,
			Group:     cpb.UserGroup(cpb.UserGroup_value[u.Role.String]),
			IsPresent: u.Status.String == domain.ConversationStatusActive,
			SeenAt:    timestamppb.New(u.SeenAt.Time),
		})
	}
	if !foundUser {
		return nil, status.Error(codes.PermissionDenied, "user in request is not member of conversation")
	}

	resp := new(tpb.LiveLessonConversationDetailResponse)
	resp.Conversation = &tpb.Conversation{
		ConversationId:   conversation.ID.String,
		Status:           tpb.ConversationStatus(tpb.ConversationStatus_value[conversation.Status.String]),
		ConversationType: tpb.ConversationType(tpb.ConversationType_value[conversation.ConversationType.String]),
		Users:            users,
		ConversationName: conversation.Name.String,
		Owner:            conversation.Owner.String,
		Seen:             seen,
	}
	if latestMsg != nil {
		resp.Conversation.LastMessage = toMessagePb(latestMsg)
	}
	return resp, nil
}
func (s *ChatReader) LiveLessonConversationMessages(ctx context.Context, req *tpb.LiveLessonConversationMessagesRequest) (*tpb.LiveLessonConversationMessagesResponse, error) {
	// extra request validation
	logger := ctxzap.Extract(ctx)
	conversationID := req.GetConversationId()
	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty conversation id")
	}

	paging := req.GetPaging()
	args := getFindMessageArgsByPaging(paging)

	msges, err := s.MessageRepo.FindLessonMessages(ctx, s.DB, database.Text(conversationID), args)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logger.Error("MessageRepo.FindLessonMessages", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := new(tpb.LiveLessonConversationMessagesResponse)
	for _, message := range msges {
		resp.Messages = append(resp.Messages, toMessagePb(message))
	}

	return resp, nil
}

func (s *ChatReader) LiveLessonPrivateConversationMessages(ctx context.Context, req *tpb.LiveLessonPrivateConversationMessagesRequest) (*tpb.LiveLessonPrivateConversationMessagesResponse, error) {
	logger := ctxzap.Extract(ctx)
	conversationID := req.GetConversationId()
	if conversationID == "" {
		return nil, status.Error(codes.InvalidArgument, "empty conversation id")
	}

	paging := req.GetPaging()
	args := getFindMessageArgsByPaging(paging)

	msges, err := s.MessageRepo.FindPrivateLessonMessages(ctx, s.DB, database.Text(conversationID), args)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			logger.Error("MessageRepo.FindPrivateLessonMessages", zap.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := new(tpb.LiveLessonPrivateConversationMessagesResponse)
	for _, message := range msges {
		resp.Messages = append(resp.Messages, toMessagePb(message))
	}

	return resp, nil
}

func getFindMessageArgsByPaging(paging *cpb.Paging) *domain.FindMessagesArgs {
	if paging == nil {
		paging = &cpb.Paging{
			Limit: 100,
		}
	}

	limit := paging.GetLimit()
	if limit == 0 {
		limit = 100
	}
	endAt := database.Timestamptz(time.Now())

	if offsetTime := paging.GetOffsetTime(); offsetTime != nil {
		endAt = database.TimestamptzFromPb(offsetTime)
	}

	return &domain.FindMessagesArgs{
		EndAt:                endAt,
		Limit:                limit,
		IncludeMessageTypes:  pgtype.TextArray{Status: pgtype.Null},
		ExcludeMessagesTypes: database.TextArray(lessonIgnoredSystemMessages), // ignore message type system
	}
}

func convertToMapMessages(ms []*domain.Message) map[pgtype.Text]*domain.Message {
	mapMessage := make(map[pgtype.Text]*domain.Message)
	for _, m := range ms {
		mapMessage[m.ConversationID] = m
	}
	return mapMessage
}
