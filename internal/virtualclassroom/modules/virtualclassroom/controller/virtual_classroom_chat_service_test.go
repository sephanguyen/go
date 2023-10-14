package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_clients "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/clients"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVirtualClassroomChatService_GetConversationID(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	conClient := &mock_clients.MockConversationClient{}
	conRepo := &mock_repositories.MockLiveLessonConversationRepo{}

	teacherID := "teacher-id1"
	lessonID := "lesson-id1"
	conversationID := "12345678"

	participants := []string{"user-id1", "teacher-id1"}
	dupParticipants := []string{"user-id1", "user-id1", "user-id1", "teacher-id1"}
	dupParticipantsWithNew := []string{"user-id1", "user-id1", "user-id1", "teacher-id1", "teacher-id2", "user-id2"}
	participantsWithNew := []string{"user-id1", "teacher-id1", "teacher-id2", "user-id2"}

	tcs := []struct {
		name      string
		reqUserID string
		req       *vpb.GetConversationIDRequest
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "teacher gets conversation ID for private chat successfully and conversation is not yet existing",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  dupParticipants,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PRIVATE,
			},
			setup: func(ctx context.Context) {
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, participants, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)

				conClient.On("CreateConversation", ctx, mock.Anything).Run(func(args mock.Arguments) {
					req := args.Get(1).(*cpb.CreateConversationRequest)
					assert.Equal(t, req.Name, lessonID+"-"+"private")
					assert.Equal(t, req.MemberIds, participants)
				}).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: conversationID,
				}, nil)

				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					con := args.Get(2).(domain.LiveLessonConversation)
					assert.Equal(t, con.ConversationID, conversationID)
					assert.Equal(t, con.LessonID, lessonID)
					assert.Equal(t, con.ParticipantList, participants)
					assert.Equal(t, con.ConversationType, domain.LiveLessonConversationTypePrivate)
				}).Once().Return(nil)
			},
		},
		{
			name:      "teacher gets conversation ID for private chat successfully and conversation already existing",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  dupParticipants,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PRIVATE,
			},
			setup: func(ctx context.Context) {
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, participants, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("12345678", nil)
			},
		},
		{
			name:      "teacher gets conversation ID for public chat successfully and conversation is not yet existing",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  dupParticipants,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC,
			},
			setup: func(ctx context.Context) {
				conRepo.On("GetConversationByLessonIDAndConvType", ctx, mock.Anything, lessonID, string(domain.LiveLessonConversationTypePublic)).Once().
					Return(domain.LiveLessonConversation{}, domain.ErrNoConversationFound)

				conClient.On("CreateConversation", ctx, mock.Anything).Run(func(args mock.Arguments) {
					req := args.Get(1).(*cpb.CreateConversationRequest)
					assert.Equal(t, req.Name, lessonID+"-"+"public")
					assert.Equal(t, req.MemberIds, participants)
				}).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: conversationID,
				}, nil)

				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					con := args.Get(2).(domain.LiveLessonConversation)
					assert.Equal(t, con.ConversationID, conversationID)
					assert.Equal(t, con.LessonID, lessonID)
					assert.Equal(t, con.ParticipantList, participants)
					assert.Equal(t, con.ConversationType, domain.LiveLessonConversationTypePublic)
				}).Once().Return(nil)
			},
		},
		{
			name:      "teacher gets conversation ID for public chat successfully and conversation already existing with the same participants",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  dupParticipants,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC,
			},
			setup: func(ctx context.Context) {
				conRepo.On("GetConversationByLessonIDAndConvType", ctx, mock.Anything, lessonID, string(domain.LiveLessonConversationTypePublic)).Once().
					Return(domain.LiveLessonConversation{
						ConversationID:   "12345678",
						LessonID:         lessonID,
						ParticipantList:  participants,
						ConversationType: domain.LiveLessonConversationTypePrivate,
					}, nil)
			},
		},
		{
			name:      "teacher gets conversation ID for public chat successfully and conversation already existing with new participants",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  dupParticipantsWithNew,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC,
			},
			setup: func(ctx context.Context) {
				conRepo.On("GetConversationByLessonIDAndConvType", ctx, mock.Anything, lessonID, string(domain.LiveLessonConversationTypePublic)).Once().
					Return(domain.LiveLessonConversation{
						ConversationID:   "12345678",
						LessonID:         lessonID,
						ParticipantList:  participants,
						ConversationType: domain.LiveLessonConversationTypePrivate,
					}, nil)

				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					con := args.Get(2).(domain.LiveLessonConversation)
					assert.Equal(t, con.ConversationID, "12345678")
					assert.Equal(t, con.LessonID, lessonID)
					assert.Equal(t, con.ParticipantList, participantsWithNew)
					assert.Equal(t, con.ConversationType, domain.LiveLessonConversationTypePublic)
				}).Once().Return(nil)
			},
		},
		{
			name:      "lesson ID is empty",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         "",
				ParticipantList:  dupParticipants,
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
		{
			name:      "participants is empty",
			reqUserID: teacherID,
			req: &vpb.GetConversationIDRequest{
				LessonId:         lessonID,
				ParticipantList:  []string{},
				ConversationType: vpb.LiveLessonConversationType_LIVE_LESSON_CONVERSATION_TYPE_PUBLIC,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			command := commands.ChatServiceCommand{
				LessonmgmtDB:               db,
				ConversationClient:         conClient,
				LiveLessonConversationRepo: conRepo,
			}

			service := &VirtualClassroomChatService{
				ChatServiceCommand: command,
			}

			res, err := service.GetConversationID(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, res.GetConversationId())
			}
			mock.AssertExpectationsForObjects(t, db, conRepo)
		})
	}
}

func TestVirtualClassroomChatService_GetPrivateConversationIDs(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	conClient := &mock_clients.MockConversationClient{}
	conRepo := &mock_repositories.MockLiveLessonConversationRepo{}

	lessonID := "lesson-id1"
	userID1 := "user-id1"
	userID2 := "user-id2"
	teacherID1 := "teacher-id1"
	teacherID2 := "teacher-id2"

	dupParticipants := []string{userID1, userID1, userID2, teacherID2}
	dupParticipantsWithCurrentUser := []string{userID1, userID1, userID2, teacherID2, teacherID1}

	tcs := []struct {
		name           string
		reqUserID      string
		req            *vpb.GetPrivateConversationIDsRequest
		res            *vpb.GetPrivateConversationIDsResponse
		setup          func(ctx context.Context)
		hasError       bool
		skipCheckEqual bool
	}{
		{
			name:      "teacher gets private conversation IDs successfully",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: dupParticipants,
			},
			setup: func(ctx context.Context) {
				// userID1
				set1 := []string{teacherID1, userID1}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set1, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				// userID2
				set2 := []string{teacherID1, userID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set2, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				// teacherID2
				set3 := []string{teacherID1, teacherID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set3, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("12345678", nil)
			},
			res: &vpb.GetPrivateConversationIDsResponse{
				ParticipantConversationMap: map[string]string{
					userID1:    "random-conversation-ID",
					userID2:    "random-conversation-ID",
					teacherID2: "random-conversation-ID",
				},
			},
		},
		{
			name:      "teacher gets private conversation IDs successfully with current user in the participant list",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: dupParticipantsWithCurrentUser,
			},
			setup: func(ctx context.Context) {
				// userID1
				set1 := []string{teacherID1, userID1}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set1, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				// userID2
				set2 := []string{teacherID1, userID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set2, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				// teacherID2
				set3 := []string{teacherID1, teacherID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set3, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("12345678", nil)
			},
			res: &vpb.GetPrivateConversationIDsResponse{
				ParticipantConversationMap: map[string]string{
					userID1:    "random-conversation-ID",
					userID2:    "random-conversation-ID",
					teacherID2: "random-conversation-ID",
				},
			},
		},
		{
			name:      "teacher gets some failed conversation IDs",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: dupParticipantsWithCurrentUser,
			},
			setup: func(ctx context.Context) {
				// userID1
				set1 := []string{teacherID1, userID1}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set1, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(nil)

				// userID2
				set2 := []string{teacherID1, userID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set2, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(errors.New("error"))

				// teacherID2
				set3 := []string{teacherID1, teacherID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set3, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("12345678", nil)
			},
			res: &vpb.GetPrivateConversationIDsResponse{
				ParticipantConversationMap: map[string]string{
					userID1:    "random-conversation-ID",
					teacherID2: "random-conversation-ID",
				},
				FailedPrivConv: &vpb.GetPrivateConversationIDsResponse_FailedPrivateConversation{
					LessonId:       lessonID,
					ParticipantIds: []string{userID2},
					ErrorMsg:       "this-expected-error-field-is-not-checked",
				},
			},
			skipCheckEqual: true,
		},
		{
			name:      "teacher gets all failed conversation IDs",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: dupParticipants,
			},
			setup: func(ctx context.Context) {
				// userID1
				set1 := []string{teacherID1, userID1}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set1, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(nil, errors.New("error"))

				// userID2
				set2 := []string{teacherID1, userID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set2, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", domain.ErrNoConversationFound)
				conClient.On("CreateConversation", ctx, mock.Anything).Once().Return(&cpb.CreateConversationResponse{
					ConversationId: "random-conversation-ID",
				}, nil)
				conRepo.On("UpsertConversation", ctx, mock.Anything, mock.Anything).Once().Return(errors.New("error"))

				// teacherID2
				set3 := []string{teacherID1, teacherID2}
				conRepo.On("GetConversationIDByExactInfo", ctx, mock.Anything, lessonID, set3, string(domain.LiveLessonConversationTypePrivate)).Once().
					Return("", errors.New("error"))
			},
			res: &vpb.GetPrivateConversationIDsResponse{
				ConversationIds: []string{},
				FailedPrivConv: &vpb.GetPrivateConversationIDsResponse_FailedPrivateConversation{
					LessonId:       lessonID,
					ParticipantIds: []string{userID1, userID2, teacherID2},
					ErrorMsg:       "this-expected-error-field-is-not-checked",
				},
			},
		},
		{
			name:      "teacher gets error due to no more participant list left after cleanup",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: []string{teacherID1, teacherID1, teacherID1},
			},
			setup:    func(ctx context.Context) {},
			res:      &vpb.GetPrivateConversationIDsResponse{},
			hasError: true,
		},
		{
			name:      "teacher gets error due to empty lesson ID",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       " ",
				ParticipantIds: []string{teacherID1, teacherID1, teacherID1},
			},
			setup:    func(ctx context.Context) {},
			res:      &vpb.GetPrivateConversationIDsResponse{},
			hasError: true,
		},
		{
			name:      "teacher gets error due to empty participant IDs",
			reqUserID: teacherID1,
			req: &vpb.GetPrivateConversationIDsRequest{
				LessonId:       lessonID,
				ParticipantIds: []string{},
			},
			setup:    func(ctx context.Context) {},
			res:      &vpb.GetPrivateConversationIDsResponse{},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ctx = interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctx)

			command := commands.ChatServiceCommand{
				LessonmgmtDB:               db,
				ConversationClient:         conClient,
				LiveLessonConversationRepo: conRepo,
			}

			service := &VirtualClassroomChatService{
				ChatServiceCommand: command,
				Logger:             ctxzap.Extract(ctx),
			}

			res, err := service.GetPrivateConversationIDs(ctx, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				if len(res.FailedPrivConv.GetParticipantIds()) == 0 {
					require.NoError(t, err)
				} else {
					require.Equal(t, res.FailedPrivConv.GetLessonId(), tc.req.GetLessonId())
					require.Equal(t, len(res.FailedPrivConv.GetParticipantIds()), len(tc.res.FailedPrivConv.GetParticipantIds()))
					require.NotEmpty(t, res.FailedPrivConv.GetErrorMsg())
				}

				if len(res.GetParticipantConversationMap()) > 0 {
					require.Equal(t, len(res.GetParticipantConversationMap()), len(tc.res.GetParticipantConversationMap()))

					// skip check equality of some test as the upsert conversation mock function is randomized to different users
					// example: in test run 1, user 1 can get the mock with error and not in test run 2
					if !tc.skipCheckEqual {
						// check if all participants can be found in the expected response is in the actual response
						for partID := range tc.res.ParticipantConversationMap {
							_, ok := res.ParticipantConversationMap[partID]
							require.True(t, ok)
						}
					}

				}
			}
			mock.AssertExpectationsForObjects(t, db, conRepo)

		})
	}
}
