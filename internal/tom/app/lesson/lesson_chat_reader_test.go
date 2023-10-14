package lesson

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestChatReader_LiveLessonConversationDetail(t *testing.T) {
	t.Parallel()

	userID := idutil.ULIDNow()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = interceptors.ContextWithUserID(ctx, userID)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)

	s := &ChatReader{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
	}
	lessonID := idutil.ULIDNow()
	conversationID := idutil.ULIDNow()

	messages := []*entities.Message{}
	respMsges := []*tpb.MessageResponse{}
	for i := 0; i < 10; i++ {
		newMsg := &entities.Message{
			ID:             database.Text(idutil.ULIDNow()),
			ConversationID: database.Text(conversationID),
		}
		messages = append(messages, newMsg)
		respMsges = append(respMsges, toMessagePb(newMsg))
	}
	lessonName := idutil.ULIDNow()
	schoolID := idutil.ULIDNow()
	seenMsg := &entities.Message{
		ID:             randomDBText(),
		ConversationID: dbText(conversationID),
		CreatedAt:      database.Timestamptz(time.Now().Add(-time.Hour)),
	}

	conversation := &entities.Conversation{
		ID:               database.Text(conversationID),
		Name:             database.Text(lessonName),
		Status:           database.Text(tpb.ConversationStatus_CONVERSATION_STATUS_NONE.String()),
		ConversationType: database.Text(tpb.ConversationType_CONVERSATION_LESSON.String()),
		Owner:            database.Text(schoolID),
		LastMessageID:    dbText(seenMsg.ID.String),
	}
	member := &entities.ConversationMembers{
		ID:     randomDBText(),
		UserID: dbText(userID),
		Role:   database.Text(cpb.UserGroup_USER_GROUP_TEACHER.String()),
		SeenAt: dbNow(),
		Status: database.Text(entities.ConversationStatusActive),
	}
	respUsers := []*tpb.Conversation_User{
		{
			Id:        member.UserID.String,
			Group:     cpb.UserGroup_USER_GROUP_TEACHER,
			IsPresent: true,
			SeenAt:    timestamppb.New(member.SeenAt.Time),
		},
	}
	respConversation := &tpb.Conversation{
		ConversationId:   conversationID,
		Status:           tpb.ConversationStatus_CONVERSATION_STATUS_NONE,
		ConversationType: tpb.ConversationType_CONVERSATION_LESSON,
		Users:            respUsers,
		Owner:            schoolID,
		ConversationName: lessonName,
		Seen:             true,
		LastMessage:      toMessagePb(seenMsg),
	}
	respConversationWithNoLatestMsg := &tpb.Conversation{
		ConversationId:   conversationID,
		Status:           tpb.ConversationStatus_CONVERSATION_STATUS_NONE,
		ConversationType: tpb.ConversationType_CONVERSATION_LESSON,
		Users:            respUsers,
		Owner:            schoolID,
		ConversationName: lessonName,
		Seen:             true,
		LastMessage:      nil,
	}

	testCases := map[string]TestCase{
		"invalid request: no lesson_id": {
			ctx:          ctx,
			req:          &tpb.LiveLessonConversationDetailRequest{},
			expectedErr:  status.Error(codes.InvalidArgument, "empty lesson id"),
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
		},
		"error query db": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationDetailRequest{
				LessonId: lessonID,
			},
			expectedErr: status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByLessonID", mock.Anything, mock.Anything, dbText(lessonID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"success, has nil latest message": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationDetailRequest{
				LessonId: lessonID,
			},
			expectedErr: nil,
			expectedResp: &tpb.LiveLessonConversationDetailResponse{
				Conversation: respConversationWithNoLatestMsg,
			},
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByLessonID", mock.Anything, mock.Anything, dbText(lessonID)).Once().Return(conversation, nil)
				members := map[pgtype.Text]entities.ConversationMembers{
					dbText(member.ID.String): *member,
				}
				conversationMemberRepo.On("FindByConversationID", mock.Anything, mock.Anything, database.Text(conversationID)).
					Once().Return(members, nil)
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(conversationID),
					mock.MatchedBy(func(args *core.FindMessagesArgs) bool {
						if args.Limit != 1 {
							return false
						}
						if !cmp.Equal(lessonIgnoredSystemMessages, database.FromTextArray(args.ExcludeMessagesTypes)) {
							return false
						}
						return true
					})).Once().Return(nil, nil)
			},
		},
		"success, have a latest message": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationDetailRequest{
				LessonId: lessonID,
			},
			expectedErr: nil,
			expectedResp: &tpb.LiveLessonConversationDetailResponse{
				Conversation: respConversation,
			},
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByLessonID", mock.Anything, mock.Anything, dbText(lessonID)).Once().Return(conversation, nil)
				members := map[pgtype.Text]entities.ConversationMembers{
					dbText(member.ID.String): *member,
				}
				conversationMemberRepo.On("FindByConversationID", mock.Anything, mock.Anything, database.Text(conversationID)).
					Once().Return(members, nil)
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(conversationID),
					mock.MatchedBy(func(args *core.FindMessagesArgs) bool {
						if args.Limit != 1 {
							return false
						}
						if !cmp.Equal(lessonIgnoredSystemMessages, database.FromTextArray(args.ExcludeMessagesTypes)) {
							return false
						}
						return true
					})).Once().Return([]*entities.Message{seenMsg}, nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.LiveLessonConversationDetail(testCase.ctx, testCase.req.(*tpb.LiveLessonConversationDetailRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
func TestChatReader_RefreshLiveLessonSession(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)
	privateConversationRepo := new(mock_repositories.MockPrivateConversationLessonRepo)

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	s := &ChatReader{
		ConversationLessonRepo:        conversationLessonRepo,
		PrivateConversationLessonRepo: privateConversationRepo,
		DB:                            mockDB,
	}
	lessonID := idutil.ULIDNow()
	conversationID := idutil.ULIDNow()

	messages := []*entities.Message{}
	respMsges := []*tpb.MessageResponse{}
	for i := 0; i < 10; i++ {
		newMsg := &entities.Message{
			ID:             database.Text(idutil.ULIDNow()),
			ConversationID: database.Text(conversationID),
		}
		messages = append(messages, newMsg)
		respMsges = append(respMsges, toMessagePb(newMsg))
	}

	testCases := map[string]TestCase{
		"invalid request: no lesson_id": {
			ctx:          ctx,
			req:          &tpb.RefreshLiveLessonSessionRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty lesson id"),
			setup:        func(ctx context.Context) {},
		},
		"err calling db with conversation lesson repo": {
			ctx: ctx,
			req: &tpb.RefreshLiveLessonSessionRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.ExecInTx: %w", status.Error(codes.Internal, pgx.ErrTxClosed.Error())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)

				conversationLessonRepo.On("UpdateLatestStartTime", mock.Anything, mock.Anything, dbText(lessonID), mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err calling db with private conversation lesson repo": {
			ctx: ctx,
			req: &tpb.RefreshLiveLessonSessionRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("database.ExecInTx: %w", status.Error(codes.Internal, pgx.ErrTxClosed.Error())),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)

				conversationLessonRepo.On("UpdateLatestStartTime", mock.Anything, mock.Anything, dbText(lessonID), mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)
				privateConversationRepo.On("UpdateLatestStartTime", mock.Anything, mock.Anything, dbText(lessonID), mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx: ctx,
			req: &tpb.RefreshLiveLessonSessionRequest{
				LessonId: lessonID,
			},
			expectedResp: &tpb.RefreshLiveLessonSessionResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)

				conversationLessonRepo.On("UpdateLatestStartTime", mock.Anything, mock.Anything, dbText(lessonID), mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)
				privateConversationRepo.On("UpdateLatestStartTime", mock.Anything, mock.Anything, dbText(lessonID), mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RefreshLiveLessonSession(testCase.ctx, testCase.req.(*tpb.RefreshLiveLessonSessionRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatReader_LiveLessonConversationMessages(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)

	s := &ChatReader{
		MessageRepo:            messageRepo,
		ConversationLessonRepo: conversationLessonRepo,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
	}
	conversationID := idutil.ULIDNow()
	paging := &cpb.Paging{
		Limit: 100,
		Offset: &cpb.Paging_OffsetTime{
			OffsetTime: timestamppb.Now(),
		},
	}

	messages := []*entities.Message{}
	respMsges := []*tpb.MessageResponse{}
	for i := 0; i < 10; i++ {
		newMsg := &entities.Message{
			ID:             database.Text(idutil.ULIDNow()),
			ConversationID: database.Text(conversationID),
		}
		messages = append(messages, newMsg)
		respMsges = append(respMsges, toMessagePb(newMsg))
	}

	testCases := map[string]TestCase{
		"invalid request: no lesson_id": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationMessagesRequest{
				Paging: paging,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty conversation id"),
			setup:        func(ctx context.Context) {},
		},
		"err callind db": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationMessagesRequest{
				Paging:         paging,
				ConversationId: conversationID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"return empty if receive ErrNoRows": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationMessagesRequest{
				Paging:         paging,
				ConversationId: conversationID,
			},
			expectedResp: &tpb.LiveLessonConversationMessagesResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success": {
			ctx: ctx,
			req: &tpb.LiveLessonConversationMessagesRequest{
				ConversationId: conversationID,
				Paging:         paging,
			},
			expectedResp: &tpb.LiveLessonConversationMessagesResponse{
				Messages: respMsges,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(messages, nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.LiveLessonConversationMessages(testCase.ctx, testCase.req.(*tpb.LiveLessonConversationMessagesRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatReader_LiveLessonPrivateConversationMessages(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)

	s := &ChatReader{
		MessageRepo:            messageRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		ConversationMemberRepo: conversationMemberRepo,
	}
	conversationID := idutil.ULIDNow()
	paging := &cpb.Paging{
		Limit: 20,
		Offset: &cpb.Paging_OffsetTime{
			OffsetTime: timestamppb.Now(),
		},
	}

	messages := []*entities.Message{}
	respMsges := []*tpb.MessageResponse{}
	for i := 0; i < 20; i++ {
		newMsg := &entities.Message{
			ID:             database.Text(idutil.ULIDNow()),
			ConversationID: database.Text(conversationID),
		}
		messages = append(messages, newMsg)
		respMsges = append(respMsges, toMessagePb(newMsg))
	}

	testCases := map[string]TestCase{
		"invalid request: no conversation_id": {
			ctx: ctx,
			req: &tpb.LiveLessonPrivateConversationMessagesRequest{
				Paging: paging,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty conversation id"),
			setup:        func(ctx context.Context) {},
		},
		"err calling db": {
			ctx: ctx,
			req: &tpb.LiveLessonPrivateConversationMessagesRequest{
				Paging:         paging,
				ConversationId: conversationID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				messageRepo.On("FindPrivateLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"return empty if receive ErrNoRows": {
			ctx: ctx,
			req: &tpb.LiveLessonPrivateConversationMessagesRequest{
				Paging:         paging,
				ConversationId: conversationID,
			},
			expectedResp: &tpb.LiveLessonPrivateConversationMessagesResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				messageRepo.On("FindPrivateLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success": {
			ctx: ctx,
			req: &tpb.LiveLessonPrivateConversationMessagesRequest{
				ConversationId: conversationID,
				Paging:         paging,
			},
			expectedResp: &tpb.LiveLessonPrivateConversationMessagesResponse{
				Messages: respMsges,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				messageRepo.On("FindPrivateLessonMessages", mock.Anything, mock.Anything, dbText(conversationID), mock.Anything).Once().Return(messages, nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.LiveLessonPrivateConversationMessages(testCase.ctx, testCase.req.(*tpb.LiveLessonPrivateConversationMessagesRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
