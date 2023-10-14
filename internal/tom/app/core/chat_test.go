package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_core "github.com/manabie-com/backend/mock/tom/app/core"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	natsJS "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	customCtx    func(context.Context) context.Context
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestChatService_SendMessageToConversations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)

	msgPusher := &mock_core.ChatInfra{}

	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		ChatInfra:              msgPusher,
		logger:                 zap.NewNop(),
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	messageContent := "hello world"
	cid1 := "conversation-1"
	cid2 := "conversation-2"
	sendMsgRequests := []*pb.SendMessageRequest{
		{
			ConversationId: cid1,
			Message:        messageContent,
		},
		{
			ConversationId: cid2,
			Message:        messageContent,
		},
	}
	cid1Members := randomConversationMembers(cid1, 2)
	cid2Members := randomConversationMembers(cid2, 1)
	conversationIDs := []string{cid1, cid2}

	testCases := []TestCase{
		{
			name:        "cannot find conversation members",
			ctx:         ctx,
			req:         domain.MessageToConversationOpts{Persist: false},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).
					Once().Return(map[pgtype.Text][]*entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:        "cannot find conversation info",
			ctx:         ctx,
			req:         domain.MessageToConversationOpts{Persist: false},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).
					Once().Return(map[pgtype.Text][]*entities.ConversationMembers{}, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).Once().Return(map[pgtype.Text]entities.ConversationFull{}, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err persisting message",
			ctx:         ctx,
			req:         domain.MessageToConversationOpts{Persist: true},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				messageRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.MatchedBy(func(msges []*domain.Message) bool {
					return msges[0].ConversationID.String == cid1 && msges[1].ConversationID.String == cid2
				})).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req:  domain.MessageToConversationOpts{Persist: true},
			setup: func(ctx context.Context) {
				messageRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.MatchedBy(func(msges []*domain.Message) bool {
					return msges[0].ConversationID.String == cid1 && msges[1].ConversationID.String == cid2
				})).Once().Return(nil)
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).
					Once().Return(map[pgtype.Text][]*entities.ConversationMembers{
					database.Text(cid1): cid1Members,
					database.Text(cid2): cid2Members,
				}, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).
					Once().Return(map[pgtype.Text]entities.ConversationFull{
					database.Text(cid1): domain.ConversationFull{},
					// database.Text(cid2): domain.ConversationFull{}, // let this not returned, so we assert PushMessage only called Once
				}, nil)
				msgPusher.On("PushMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SendMessageToConversations(testCase.ctx, sendMsgRequests, testCase.req.(domain.MessageToConversationOpts))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func TestChatService_SendMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)

	msgPusher := &mock_core.ChatInfra{}

	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		ChatInfra:              msgPusher,
		logger:                 zap.NewNop(),
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	conversationID := idutil.ULIDNow()
	fakeConversationName := "fake conversation"

	var (
		cID pgtype.Text
		uID pgtype.Text
	)
	_ = cID.Set(conversationID)
	_ = uID.Set(userID)
	fakeConversation := &entities.Conversation{
		ID:               cID,
		ConversationType: database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()),
		Name:             database.Text(fakeConversationName),
	}

	testCases := []TestCase{
		{
			name:         "err conversation not exist in db",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrNoRows)
			},
		},
		{
			name:         "err db when find conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:         "err not found user in conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, nil)
			},
		},
		{
			name:         "err create message",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {},
				}, nil)

				messageRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:         "err mark seen for sender",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {},
				}, nil)

				messageRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				msgPusher.On("PushMessage", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Enabled:      true,
						IgnoredUsers: []string{userID},
					},
				}).Once().Return(nil)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(pgx.ErrTxClosed)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)

				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "success send text message",
			ctx:  ctx,
			req: &pb.SendMessageRequest{
				ConversationId: conversationID,
				Message:        "",
				Type:           pb.MESSAGE_TYPE_TEXT,
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {
						UserID:         database.Text(userID),
						ConversationID: database.Text(conversationID),
					},
				}, nil)

				messageRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)

				// title = name of msg sender, tokens = result returned by userTokenRepo
				msgPusher.On("PushMessage", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Silence:      false,
						Title:        fakeConversationName,
						Enabled:      true,
						IgnoredUsers: []string{userID},
					},
				}).Once().Return(nil)
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Twice().Return(
					fakeConversation, nil)
			},
		},
		{
			name: "success send image",
			ctx:  ctx,
			req: &pb.SendMessageRequest{
				ConversationId: conversationID,
				Message:        "",
				Type:           pb.MESSAGE_TYPE_IMAGE,
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {
						UserID:         database.Text(userID),
						ConversationID: database.Text(conversationID),
					},
				}, nil)

				messageRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)

				// title = name of msg sender, tokens = result returned by userTokenRepo
				msgPusher.On("PushMessage", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Title:        fakeConversationName,
						Enabled:      true,
						IgnoredUsers: []string{userID},
					},
				}).Once().Return(nil)

				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Twice().Return(
					fakeConversation, nil)
			},
		},
		{
			name: "success send file",
			ctx:  ctx,
			req: &pb.SendMessageRequest{
				ConversationId: conversationID,
				Message:        "",
				Type:           pb.MESSAGE_TYPE_FILE,
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {
						UserID:         database.Text(userID),
						ConversationID: database.Text(conversationID),
					},
				}, nil)

				messageRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)

				// title = name of msg sender, tokens = result returned by userTokenRepo
				msgPusher.On("PushMessage", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Title:        fakeConversationName,
						Enabled:      true,
						IgnoredUsers: []string{userID},
					},
				}).Once().Return(nil)

				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Twice().Return(
					fakeConversation, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.SendMessage(testCase.ctx, testCase.req.(*pb.SendMessageRequest))
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatService_broadcastMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)
	onlineUserRepo := new(mock_repositories.MockOnlineUserRepo)

	msgPusher := &mock_core.ChatInfra{}
	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		logger:                 zap.NewNop(),
		ChatInfra:              msgPusher,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	conversationID := idutil.ULIDNow()

	var (
		cID pgtype.Text
		uID pgtype.Text
	)
	_ = cID.Set(conversationID)
	_ = uID.Set(userID)

	testCases := []TestCase{
		{
			name:         "err conversation not exist in db",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrNoRows)
			},
		},
		{
			name:         "err db when find conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:         "err not found user in conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrNoRows)
			},
		},
		{
			name:         "success",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {},
				}, nil)

				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)

				msgPusher.On("PushMessage", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Enabled:      true,
						IgnoredUsers: []string{""},
					},
				}).Once().Return(nil)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)
				onlineUserRepo.On("Find", ctx, mock.Anything, mock.AnythingOfType("pgtype.TextArray"), mock.AnythingOfType("pgtype.Timestamptz"), mock.Anything).
					Once().Return(map[pgtype.Text][]string{}, nil)

				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.broadcastMessage(testCase.ctx, "", testCase.req.(*pb.SendMessageRequest))
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatService_SeenMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)

	msgPusher := &mock_core.ChatInfra{}
	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		logger:                 zap.NewNop(),
		ChatInfra:              msgPusher,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	conversationID := idutil.ULIDNow()

	var (
		cID pgtype.Text
		uID pgtype.Text
	)
	_ = cID.Set(conversationID)
	_ = uID.Set(userID)

	testCases := []TestCase{
		{
			name:         "err db Upsert conversation status",
			ctx:          ctx,
			req:          &pb.SeenMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("markSeenConversation: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockBroadcastMessage(conversationMemberRepo, messageRepo, jsm, msgPusher, conversationRepo, cID, uID, ctx)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name:         "success",
			ctx:          ctx,
			req:          &pb.SeenMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockBroadcastMessage(conversationMemberRepo, messageRepo, jsm, msgPusher, conversationRepo, cID, uID, ctx)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.SeenMessage(testCase.ctx, testCase.req.(*pb.SeenMessageRequest))
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func mockBroadcastMessage(conversationMemberRepo *mock_repositories.MockConversationMemberRepo,
	messageRepo *mock_repositories.MockMessageRepo,
	jsm *mock_nats.JetStreamManagement,
	msgPusher *mock_core.ChatInfra,
	conversationRepo *mock_repositories.MockConversationRepo, cID pgtype.Text, uID pgtype.Text, ctx context.Context) {
	conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
		uID: {},
	}, nil)

	jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)
	msgPusher.On("PushMessage", mock.Anything, []string{uID.String}, mock.Anything, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			IgnoredUsers: []string{uID.String},
			Enabled:      true,
		},
	}).Once().Return(nil)

	conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
}

func mockBroadcastMessageDeleted(
	conversationMemberRepo *mock_repositories.MockConversationMemberRepo,
	jsm *mock_nats.JetStreamManagement,
	msgPusher *mock_core.ChatInfra,
	cID pgtype.Text, uID pgtype.Text, ctx context.Context,
) {
	conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
		uID: {},
	}, nil)

	jsm.On("PublishContext", constants.SubjectChatMessageDeleted, mock.Anything).Return(&natsJS.PubAck{}, nil)
	msgPusher.On("PushMessageDeleted", mock.Anything, []string{uID.String}, mock.Anything, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			IgnoredUsers: []string{uID.String},
			Enabled:      true,
		},
	}).Once().Return(nil)
}

func TestChatService_DeleteMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)

	msgPusher := &mock_core.ChatInfra{}
	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		MessageRepo:            messageRepo,
		logger:                 zap.NewNop(),
		ChatInfra:              msgPusher,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		db:                     db,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	teacherCtx, teacherCtxCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer teacherCtxCancel()

	teacherCtx = interceptors.ContextWithUserID(ctx, userID)
	teacherCtx = interceptors.ContextWithUserGroup(ctx, constant.RoleTeacher)

	conversationID := idutil.ULIDNow()
	messageID := idutil.ULIDNow()

	var (
		cID pgtype.Text
		uID pgtype.Text
		mID pgtype.Text
	)
	_ = cID.Set(conversationID)
	_ = uID.Set(userID)
	_ = mID.Set(messageID)

	studentConversation := &entities.Conversation{
		ID:               cID,
		ConversationType: database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String()),
	}

	lessonConversation := &entities.Conversation{
		ID:               cID,
		ConversationType: database.Text(tpb.ConversationType_CONVERSATION_LESSON.String()),
	}

	message := &entities.Message{UserID: uID, ConversationID: cID}
	request := &tpb.DeleteMessageRequest{MessageId: messageID}
	expectedResp := &tpb.DeleteMessageResponse{}

	testCases := []TestCase{
		{
			name:         "delete message success",
			ctx:          ctx,
			req:          request,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(message, nil)
				messageRepo.On("SoftDelete", ctx, db, mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Text")).Once().Return(nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(studentConversation, nil)
			},
		},
		{
			name:         "delete message success with teacher role",
			ctx:          teacherCtx,
			req:          request,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(message, nil)
				messageRepo.On("SoftDelete", ctx, db, mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Text")).Once().Return(nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(lessonConversation, nil)
			},
		},
		{
			name:        "err no message found",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.NotFound, "not found message"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(&entities.Message{}, pgx.ErrNoRows)
			},
		},
		{
			name:        "err no conversation found",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(message, nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(&entities.Conversation{}, pgx.ErrNoRows)
			},
		},
		{
			name:        "err db when find message",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.Unknown, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(&entities.Message{}, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err db when find conversation",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.Unknown, "tx is closed"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(message, nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err user permission in student chat",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.PermissionDenied, "permission denied: current user is not the owner"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(&entities.Message{ID: mID, UserID: database.Text(idutil.ULIDNow()), ConversationID: cID}, nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(studentConversation, nil)
			},
		},
		{
			name:        "err user permission in lesson chat",
			ctx:         ctx,
			req:         request,
			expectedErr: status.Error(codes.PermissionDenied, "permission denied: current user is not either staff or the owner"),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(&entities.Message{ID: mID, UserID: database.Text(idutil.ULIDNow()), ConversationID: cID}, nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(lessonConversation, nil)
			},
		},
		{
			name:        "err db when soft delete",
			ctx:         ctx,
			req:         request,
			expectedErr: fmt.Errorf("deleteMessage: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockBroadcastMessageDeleted(conversationMemberRepo, jsm, msgPusher, cID, uID, ctx)
				messageRepo.On("FindByID", ctx, db, mID).Once().Return(message, nil)
				conversationRepo.On("FindByID", ctx, db, cID).Once().Return(lessonConversation, nil)
				messageRepo.On("SoftDelete", ctx, db, mock.AnythingOfType("pgtype.Text"), mock.AnythingOfType("pgtype.Text")).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.DeleteMessage(testCase.ctx, testCase.req.(*tpb.DeleteMessageRequest))
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatService_broadcastMessageDeleted(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)

	msgPusher := &mock_core.ChatInfra{}
	jsm := &mock_nats.JetStreamManagement{}

	s := &ChatServiceImpl{
		JSM:                    jsm,
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		logger:                 zap.NewNop(),
		ChatInfra:              msgPusher,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	conversationID := idutil.ULIDNow()
	messageID := idutil.ULIDNow()

	var (
		cID pgtype.Text
		uID pgtype.Text
		mID pgtype.Text
	)
	_ = cID.Set(conversationID)
	_ = uID.Set(userID)
	_ = mID.Set(messageID)

	testCases := []TestCase{
		{
			name:         "err conversation not exist in db",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrNoRows)
			},
		},
		{
			name:         "err db when find conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:         "err not found user in conversation",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  status.Error(codes.NotFound, "not found conversation"),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrNoRows)
			},
		},
		{
			name:         "success",
			ctx:          ctx,
			req:          &pb.SendMessageRequest{ConversationId: conversationID},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, cID).Once().Return(map[pgtype.Text]entities.ConversationMembers{
					uID: {},
				}, nil)

				conversationMemberRepo.On("SetSeenAt", ctx, nil, mock.AnythingOfType("pgtype.Text"), pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)

				msgPusher.On("PushMessageDeleted", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{
					Notification: domain.NotificationOpts{
						Enabled:      true,
						IgnoredUsers: []string{userID},
					},
				}).Once().Return(nil)

				jsm.On("PublishContext", constants.SubjectSendChatMessageCreated, mock.Anything).Return(&natsJS.PubAck{}, nil)
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.broadcastMessageDeleted(testCase.ctx, userID, conversationID, messageID)
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
