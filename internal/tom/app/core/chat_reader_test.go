package core

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamps "google.golang.org/protobuf/types/known/timestamppb"
)

func TestChatService_GetConversation_Lesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cID := idutil.ULIDNow()
	conversationName := "conversation-name"
	seenAt := time.Now().Add(1 * time.Hour)

	teacherID := idutil.ULIDNow()
	// default has seen status
	seenAtPb, _ := types.TimestampProto(seenAt)
	members := map[pgtype.Text]entities.ConversationMembers{
		database.Text(teacherID): {
			UserID: database.Text(teacherID),
			Status: database.Text(entities.ConversationStatusActive),
			SeenAt: database.Timestamptz(seenAt),
			Role:   database.Text(cpb.UserGroup_USER_GROUP_TEACHER.String()),
		},
	}
	conversation := entities.Conversation{
		ID:               database.Text(cID),
		Name:             database.Text(conversationName),
		CreatedAt:        pgtype.Timestamptz{},
		UpdatedAt:        pgtype.Timestamptz{},
		ConversationType: database.Text(pb.ConversationType_name[int32(pb.CONVERSATION_LESSON)]),
	}
	latestMsg := entities.Message{
		ConversationID: conversation.ID,
		CreatedAt:      database.Timestamptz(time.Now()),
	}
	latestButNotSeenMsg := latestMsg
	latestButNotSeenMsg.CreatedAt = database.Timestamptz(seenAt.Add(1 * time.Hour))
	conversationMap := make(map[pgtype.Text]core.ConversationFull)
	conversationMap[database.Text(cID)] = core.ConversationFull{
		Conversation: conversation,
	}

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	db := new(mock_database.Ext)
	s := &ChatReader{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		DB:                     db,
		Logger:                 zap.NewNop(),
	}

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	type expectedResp = func() *pb.GetConversationResponse

	sampleRes := func() *pb.GetConversationResponse {
		return &pb.GetConversationResponse{
			Conversation: &pb.Conversation{
				ConversationId:   cID,
				ConversationName: conversationName,
				Status:           pb.CONVERSATION_STATUS_NONE,
				ConversationType: pb.CONVERSATION_LESSON,
				Seen:             true,
				Users: []*pb.Conversation_User{
					{
						Id:        teacherID,
						SeenAt:    seenAtPb,
						Group:     cpb.UserGroup_USER_GROUP_TEACHER.String(),
						IsPresent: true,
					},
				},
				LastMessage: toMessageResponse(&latestMsg),
			},
		}
	}

	testCases := []TestCase{
		{
			name:        "Success return non nil unseen latest message",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)

				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).
					Once().Return(members, nil)
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(cID),
					mock.MatchedBy(func(args *core.FindMessagesArgs) bool {
						if args.Limit != 1 {
							return false
						}
						return !args.IncludeSystemMsg
					})).Once().Return([]*entities.Message{&latestButNotSeenMsg}, nil)
			},
			expectedResp: func() *pb.GetConversationResponse {
				res := sampleRes()
				res.Conversation.Seen = false
				res.Conversation.LastMessage = toMessageResponse(&latestButNotSeenMsg)
				return res
			},
		},
		{
			name:        "Success return non nil seen latest message",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)
				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(cID),
					mock.MatchedBy(func(args *core.FindMessagesArgs) bool {
						if args.Limit != 1 {
							return false
						}
						return !args.IncludeSystemMsg
					})).Once().Return([]*entities.Message{&latestMsg}, nil)
			},
			expectedResp: func() *pb.GetConversationResponse {
				return sampleRes()
			},
		},
		{
			name:        "Success return nil latest message",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)

				messageRepo.On("FindLessonMessages", mock.Anything, mock.Anything, dbText(cID),
					mock.MatchedBy(func(args *core.FindMessagesArgs) bool {
						if args.Limit != 1 {
							return false
						}
						return !args.IncludeSystemMsg
					})).Once().Return(nil, nil)
			},
			expectedResp: func() *pb.GetConversationResponse {
				res := sampleRes()
				res.Conversation.Seen = true
				res.Conversation.LastMessage = nil
				return res
			},
		},
	}
	for _, tc := range testCases {
		tc.setup(tc.ctx)
		t.Run(tc.name, func(t *testing.T) {
			res, err := s.GetConversation(tc.ctx, tc.req.(*pb.GetConversationRequest))
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedResp != nil {
				expectedResp := tc.expectedResp.(func() *pb.GetConversationResponse)()
				assert.Equal(t, expectedResp, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

func TestChatService_ConversationDetail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)

	s := &ChatReader{
		MessageRepo:            messageRepo,
		ConversationMemberRepo: conversationMemberRepo,
		Logger:                 zap.NewNop(),
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	conversationID := idutil.ULIDNow()

	now := types.TimestampNow()
	testCases := []TestCase{
		{
			name:         "err db FindAllMessageByConversation",
			ctx:          ctx,
			req:          &pb.ConversationDetailRequest{ConversationId: conversationID, Limit: 10, EndAt: nil},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				var cID pgtype.Text
				_ = cID.Set(conversationID)
				messageRepo.On("FindMessages", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "joined member success",
			ctx:  ctx,
			req:  &pb.ConversationDetailRequest{ConversationId: conversationID, Limit: 10, EndAt: now},
			expectedResp: &pb.ConversationDetailResponse{
				Messages: []*pb.MessageResponse{
					{
						MessageId:      "",
						ConversationId: conversationID,
						UserId:         userID,
						Content:        "content",
						UrlMedia:       "",
						Type:           0,
						CreatedAt:      now,
						LocalMessageId: "",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				var cID pgtype.Text
				_ = cID.Set(conversationID)
				messageRepo.On("FindMessages", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Message{{
					ID:             pgtype.Text{},
					ConversationID: cID,
					UserID:         pgtype.Text{String: userID, Status: 2},
					Message:        pgtype.Text{String: "content", Status: 2},
					UrlMedia:       pgtype.Text{},
					Type:           pgtype.Text{},
					CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					UpdatedAt:      pgtype.Timestamptz{},
				}}, nil)
				fakeConversationMember := &entities.ConversationMembers{
					Status: database.Text(entities.ConversationStatusActive),
				}

				// is a member of this conversation
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, nil, cID, pgtype.Text{String: userID, Status: 2}).Once().Return(fakeConversationMember, nil)
				conversationMemberRepo.On("SetSeenAt", ctx, nil, cID, pgtype.Text{String: userID, Status: 2}, mock.AnythingOfType("pgtype.Timestamptz")).Once().Return(nil)
			},
		},
		{
			name: "unjoined member success",
			ctx:  ctx,
			req:  &pb.ConversationDetailRequest{ConversationId: conversationID, Limit: 10, EndAt: now},
			expectedResp: &pb.ConversationDetailResponse{
				Messages: []*pb.MessageResponse{
					{
						MessageId:      "",
						ConversationId: conversationID,
						UserId:         userID,
						Content:        "content",
						UrlMedia:       "",
						Type:           0,
						CreatedAt:      now,
						LocalMessageId: "",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				var cID pgtype.Text
				_ = cID.Set(conversationID)
				messageRepo.On("FindMessages", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Message{{
					ID:             pgtype.Text{},
					ConversationID: cID,
					UserID:         pgtype.Text{String: userID, Status: 2},
					Message:        pgtype.Text{String: "content", Status: 2},
					UrlMedia:       pgtype.Text{},
					Type:           pgtype.Text{},
					CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					UpdatedAt:      pgtype.Timestamptz{},
				}}, nil)

				// not a member of this conversation
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, nil, cID, pgtype.Text{String: userID, Status: 2}).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "filter messages by include enums",
			ctx:  ctx,
			req: &pb.ConversationDetailRequest{
				ConversationId: conversationID,
				Limit:          10,
				EndAt:          now,
			},
			expectedResp: &pb.ConversationDetailResponse{
				Messages: []*pb.MessageResponse{
					{
						MessageId:      "",
						ConversationId: conversationID,
						UserId:         userID,
						Content:        "content",
						UrlMedia:       "",
						Type:           0,
						CreatedAt:      now,
						LocalMessageId: "",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cID := database.Text(conversationID)
				args := &core.FindMessagesArgs{
					ConversationID: cID,
					EndAt:          pgtype.Timestamptz{Time: time.Unix(now.Seconds, 0).UTC(), Status: pgtype.Present},
					Limit:          10,
				}
				messageRepo.On("FindMessages", mock.Anything, mock.Anything, args).Once().Return([]*entities.Message{{
					ID:             pgtype.Text{},
					ConversationID: database.Text(conversationID),
					UserID:         pgtype.Text{String: userID, Status: 2},
					Message:        pgtype.Text{String: "content", Status: 2},
					UrlMedia:       pgtype.Text{},
					Type:           pgtype.Text{},
					CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					UpdatedAt:      pgtype.Timestamptz{},
				}}, nil)

				// not a member of this conversation
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, nil, cID, pgtype.Text{String: userID, Status: 2}).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "filter messages by exclude enums",
			ctx:  ctx,
			req: &pb.ConversationDetailRequest{
				ConversationId: conversationID,
				Limit:          10,
				EndAt:          now,
			},
			expectedResp: &pb.ConversationDetailResponse{
				Messages: []*pb.MessageResponse{
					{
						MessageId:      "",
						ConversationId: conversationID,
						UserId:         userID,
						Content:        "content",
						UrlMedia:       "",
						Type:           0,
						CreatedAt:      now,
						LocalMessageId: "",
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cID := database.Text(conversationID)
				args := &core.FindMessagesArgs{
					ConversationID: cID,
					EndAt:          pgtype.Timestamptz{Time: time.Unix(now.Seconds, 0).UTC(), Status: pgtype.Present},
					Limit:          10,
				}
				messageRepo.On("FindMessages", mock.Anything, mock.Anything, args).Once().Return([]*entities.Message{{
					ID:             pgtype.Text{},
					ConversationID: database.Text(conversationID),
					UserID:         pgtype.Text{String: userID, Status: 2},
					Message:        pgtype.Text{String: "content", Status: 2},
					UrlMedia:       pgtype.Text{},
					Type:           pgtype.Text{},
					CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					UpdatedAt:      pgtype.Timestamptz{},
				}}, nil)

				// not a member of this conversation
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, nil, cID, pgtype.Text{String: userID, Status: 2}).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ConversationDetail(testCase.ctx, testCase.req.(*pb.ConversationDetailRequest))
			assert.Equal(t, testCase.expectedErr, errors.Cause(err))
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

// This test will be removed when moving tom_dart_client to manabuf is done
func TestChatService_GetConversation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cID := idutil.ULIDNow()
	conversationName := "conversation-name"
	seenAt := time.Now().Add(1 * time.Hour)

	teacherID := idutil.ULIDNow()
	// default has seen status
	seenAtPb, _ := types.TimestampProto(seenAt)
	members := map[pgtype.Text]entities.ConversationMembers{
		database.Text(teacherID): {
			UserID: database.Text(teacherID),
			Status: database.Text(entities.ConversationStatusActive),
			SeenAt: database.Timestamptz(seenAt),
			Role:   database.Text(cpb.UserGroup_USER_GROUP_TEACHER.String()),
		},
	}
	conversation := entities.Conversation{
		ID:               database.Text(cID),
		Name:             database.Text(conversationName),
		CreatedAt:        pgtype.Timestamptz{},
		UpdatedAt:        pgtype.Timestamptz{},
		ConversationType: database.Text(pb.ConversationType_name[int32(pb.CONVERSATION_STUDENT)]),
	}
	latestMsg := entities.Message{
		ConversationID: conversation.ID,
		CreatedAt:      database.Timestamptz(time.Now()),
	}
	latestButNotSeenMsg := latestMsg
	latestButNotSeenMsg.CreatedAt = database.Timestamptz(seenAt.Add(1 * time.Hour))
	conversationMap := make(map[pgtype.Text]core.ConversationFull)
	conversationMap[database.Text(cID)] = core.ConversationFull{
		Conversation: conversation,
	}

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	db := new(mock_database.Ext)

	s := &ChatReader{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		DB:                     db,
		Logger:                 zap.NewNop(),
	}

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	type expectedResp = func() *pb.GetConversationResponse

	sampleRes := func() *pb.GetConversationResponse {
		return &pb.GetConversationResponse{
			Conversation: &pb.Conversation{
				ConversationId:   cID,
				ConversationName: conversationName,
				Status:           pb.CONVERSATION_STATUS_NONE,
				ConversationType: pb.CONVERSATION_STUDENT,
				Seen:             true,
				Users: []*pb.Conversation_User{
					{
						Id:        teacherID,
						SeenAt:    seenAtPb,
						Group:     cpb.UserGroup_USER_GROUP_TEACHER.String(),
						IsPresent: true,
					},
				},
				LastMessage: toMessageResponse(&latestMsg),
			},
		}
	}

	testCases := []TestCase{
		{
			name:        "empty conversation id",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "empty conversation id in request"),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "err from db when find conversation",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err from db when finding latest msg",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err from db finding conversation members",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:        "Success return with unseen status",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestButNotSeenMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)
			},
			expectedResp: func() *pb.GetConversationResponse {
				res := sampleRes()
				res.Conversation.Seen = false
				res.Conversation.LastMessage = toMessageResponse(&latestButNotSeenMsg)
				return res
			},
		},
		{
			name:        "Success return with seen status",
			ctx:         ctx,
			req:         &pb.GetConversationRequest{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)
			},
			expectedResp: func() *pb.GetConversationResponse {
				return sampleRes()
			},
		},
	}
	for _, tc := range testCases {
		tc.setup(tc.ctx)
		t.Run(tc.name, func(t *testing.T) {
			res, err := s.GetConversation(tc.ctx, tc.req.(*pb.GetConversationRequest))
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedResp != nil {
				expectedResp := tc.expectedResp.(func() *pb.GetConversationResponse)()
				assert.Equal(t, expectedResp, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}

func TestChatService_GetConversationV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cID := idutil.ULIDNow()
	conversationName := "conversation-name"
	seenAt := time.Now().Add(1 * time.Hour)

	teacherID := idutil.ULIDNow()
	// default has seen status
	seenAtPb := timestamps.New(seenAt)
	members := map[pgtype.Text]entities.ConversationMembers{
		database.Text(teacherID): {
			UserID: database.Text(teacherID),
			Status: database.Text(entities.ConversationStatusActive),
			SeenAt: database.Timestamptz(seenAt),
			Role:   database.Text(cpb.UserGroup_USER_GROUP_TEACHER.String()),
		},
	}
	conversation := entities.Conversation{
		ID:               database.Text(cID),
		Name:             database.Text(conversationName),
		CreatedAt:        pgtype.Timestamptz{},
		UpdatedAt:        pgtype.Timestamptz{},
		ConversationType: database.Text(tpb.ConversationType_name[int32(tpb.ConversationType_CONVERSATION_STUDENT)]),
	}
	latestMsg := entities.Message{
		ConversationID: conversation.ID,
		CreatedAt:      database.Timestamptz(time.Now()),
	}
	latestButNotSeenMsg := latestMsg
	latestButNotSeenMsg.CreatedAt = database.Timestamptz(seenAt.Add(1 * time.Hour))
	conversationMap := make(map[pgtype.Text]core.ConversationFull)
	conversationMap[database.Text(cID)] = core.ConversationFull{
		Conversation: conversation,
	}

	lessonPrivateConversation := entities.Conversation{
		ID:               database.Text(cID),
		Name:             database.Text(conversationName),
		CreatedAt:        pgtype.Timestamptz{},
		UpdatedAt:        pgtype.Timestamptz{},
		ConversationType: database.Text(tpb.ConversationType_name[int32(tpb.ConversationType_CONVERSATION_LESSON_PRIVATE)]),
	}

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	db := new(mock_database.Ext)

	s := &ChatReader{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		MessageRepo:            messageRepo,
		DB:                     db,
		Logger:                 zap.NewNop(),
	}

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	type expectedResp = func() *tpb.GetConversationV2Response

	sampleRes := func(conversationType tpb.ConversationType) *tpb.GetConversationV2Response {
		return &tpb.GetConversationV2Response{
			Conversation: &tpb.Conversation{
				ConversationId:   cID,
				ConversationName: conversationName,
				Status:           tpb.ConversationStatus_CONVERSATION_STATUS_NONE,
				ConversationType: conversationType,
				Seen:             true,
				Users: []*tpb.Conversation_User{
					{
						Id:        teacherID,
						SeenAt:    seenAtPb,
						Group:     cpb.UserGroup_USER_GROUP_TEACHER,
						IsPresent: true,
					},
				},
				LastMessage: toMessageResponseV2(&latestMsg),
			},
		}
	}

	testCases := []TestCase{
		{
			name:        "empty conversation id",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{},
			expectedErr: status.Error(codes.InvalidArgument, "empty conversation id in request"),
			setup:       func(ctx context.Context) {},
		},
		{
			name:        "err from db when find conversation",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err from db when finding latest msg",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "err from db finding conversation members",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(map[pgtype.Text]entities.ConversationMembers{}, pgx.ErrTxClosed)
			},
		},
		{
			name:        "Success return with unseen status",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestButNotSeenMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)
			},
			expectedResp: func() *tpb.GetConversationV2Response {
				res := sampleRes(tpb.ConversationType_CONVERSATION_STUDENT)
				res.Conversation.Seen = false
				res.Conversation.LastMessage = toMessageResponseV2(&latestButNotSeenMsg)
				return res
			},
		},
		{
			name:        "Success return with seen status",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&conversation, nil)
				messageRepo.On("GetLatestMessageByConversation", ctx, mock.Anything, database.Text(cID)).
					Once().Return(&latestMsg, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)
			},
			expectedResp: func() *tpb.GetConversationV2Response {
				return sampleRes(tpb.ConversationType_CONVERSATION_STUDENT)
			},
		},
		{
			name:        "Success return with lesson private chat",
			ctx:         ctx,
			req:         &tpb.GetConversationV2Request{ConversationId: cID},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("FindByID", ctx, mock.Anything, database.Text(cID)).Once().Return(&lessonPrivateConversation, nil)
				var messages []*core.Message
				messages = append(messages, &latestMsg)
				messageRepo.On("FindPrivateLessonMessages", mock.Anything, mock.Anything, dbText(cID), mock.Anything).Once().Return(messages, nil)
				conversationMemberRepo.On("FindByConversationIDAndStatus", ctx, mock.Anything, database.Text(cID), pgtype.Text{
					Status: pgtype.Null,
				}).Once().Return(members, nil)

			},
			expectedResp: func() *tpb.GetConversationV2Response {
				return sampleRes(tpb.ConversationType_CONVERSATION_LESSON_PRIVATE)
			},
		},
	}
	for _, tc := range testCases {
		tc.setup(tc.ctx)
		t.Run(tc.name, func(t *testing.T) {
			res, err := s.GetConversationV2(tc.ctx, tc.req.(*tpb.GetConversationV2Request))
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
			if tc.expectedResp != nil {
				expectedResp := tc.expectedResp.(func() *tpb.GetConversationV2Response)()
				assert.Equal(t, expectedResp, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}
