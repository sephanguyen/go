package support

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	bobConst "github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	mock_services "github.com/manabie-com/backend/mock/tom/services"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getCtx(userID, role string) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			SchoolIDs: []string{strconv.Itoa(bobConst.ManabieSchool)},
		},
	})
	ctx = interceptors.ContextWithUserGroup(ctx, role)
	return ctx
}

func TestChatService_HandleParentRemovedFromStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	jsm := &mock_nats.JetStreamManagement{}

	// messageRepo := new(mock_repositories.MockMessageRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)

	chatSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                      mockDB,
		ChatService:             chatSvc,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		ConversationStudentRepo: conversationStudentRepo,
		JSM:                     jsm,
	}
	parentID := "parent-id"
	studentID := "student-id"

	validReq := &upb.EvtUser_ParentRemovedFromStudent{
		StudentId: studentID,
		ParentId:  parentID,
	}
	convID := "conversation-id"
	member := domain.ConversationMembers{
		UserID:         database.Text(parentID),
		ConversationID: database.Text(convID),
	}

	testCases := map[string]TestCase{
		"invalid req": {
			ctx: ctx,
			req: &upb.EvtUser_ParentRemovedFromStudent{
				ParentId: parentID,
			},
			expectedErr: errEmptyStudent,
			setup: func(ctx context.Context) {
			},
		},
		"fail bulk upsert": {
			ctx:         ctx,
			req:         validReq,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				deactivatedMember := member
				deactivatedMember.Status.Set(entities.ConversationStatusInActive)
				deactivatedMembers := []*entities.ConversationMembers{&deactivatedMember}

				// find parent conv
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, tx, database.TextArray([]string{studentID}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{convID}, nil)

				// find membership
				conversationMemberRepo.On("FindByCIDsAndUserID",
					mock.Anything, tx, database.TextArray([]string{convID}), database.Text(parentID)).
					Once().Return(deactivatedMembers, nil)

				// soft delete
				conversationMemberRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					if len(members) != 1 {
						return false
					}
					calledMem := members[0]
					return calledMem.Status.String == entities.ConversationStatusInActive &&
						calledMem.UserID.String == parentID
				})).Once().Return(pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx: ctx,
			req: validReq,
			setup: func(ctx context.Context) {
				deactivatedMember := member
				deactivatedMember.Status.Set(entities.ConversationStatusInActive)
				deactivatedMembers := []*entities.ConversationMembers{&deactivatedMember}

				// find parent conv
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, tx, database.TextArray([]string{studentID}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{convID}, nil)

				// find membership
				conversationMemberRepo.On("FindByCIDsAndUserID",
					mock.Anything, tx, database.TextArray([]string{convID}), database.Text(parentID)).
					Once().Return(deactivatedMembers, nil)

				// soft delete
				conversationMemberRepo.On("BulkUpsert", mock.Anything, tx, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					if len(members) != 1 {
						return false
					}
					calledMem := members[0]
					return calledMem.Status.String == entities.ConversationStatusInActive &&
						calledMem.UserID.String == parentID
				})).Once().Return(nil)
				// system message to users
				chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)

				// system message to this parent
				chatSvc.On("SendMessageToUsers", mock.Anything, []string{deactivatedMember.UserID.String}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleParentRemovedFromStudent(testCase.ctx, testCase.req.(*upb.EvtUser_ParentRemovedFromStudent))

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			}
		})
	}
}

func TestChatService_HandleEventCreateStudentConversation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	// messageRepo := new(mock_repositories.MockMessageRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	locationRepo := new(mock_repositories.MockConversationLocationRepo)

	jsm := &mock_nats.JetStreamManagement{}
	mockSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                       mockDB,
		ChatService:              mockSvc,
		JSM:                      jsm,
		ConversationMemberRepo:   conversationMemberRepo,
		ConversationRepo:         conversationRepo,
		ConversationStudentRepo:  conversationStudentRepo,
		ConversationLocationRepo: locationRepo,
	}
	locations := []string{"loc-1", "loc-2"}

	validReq := &upb.EvtUser_CreateStudent{
		StudentId:   "student-id",
		StudentName: "student-name",
		SchoolId:    "school-id",
		LocationIds: locations,
	}

	testCases := map[string]TestCase{
		"err upsert conversation": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.conversationRepo.BulkCreate: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err create conversation member": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.ConversationMemberRepo.Create: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(locs []domain.ConversationLocation) bool {
					for _, item := range locs {
						if !stringIn(item.LocationID.String, locations) {
							return false
						}
					}
					return true
				})).Once().Return(nil)

				conversationMemberRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err create conversation student ": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.ConversationStudentRepo.BulkUpsert: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(locs []domain.ConversationLocation) bool {
					for _, item := range locs {
						if !stringIn(item.LocationID.String, locations) {
							return false
						}
					}
					return true
				})).Once().Return(nil)
				conversationStudentRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				locationRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(locs []domain.ConversationLocation) bool {
					for _, item := range locs {
						if !stringIn(item.LocationID.String, locations) {
							return false
						}
					}
					return true
				})).Once().Return(nil)
				conversationStudentRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatCreated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				mockSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleEventCreateStudentConversation(testCase.ctx, testCase.req.(*upb.EvtUser_CreateStudent))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestChatService_HandleEventCreateParentConversation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)

	jsm := &mock_nats.JetStreamManagement{}
	chatSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                       mockDB,
		ChatService:              chatSvc,
		JSM:                      jsm,
		ConversationMemberRepo:   conversationMemberRepo,
		ConversationRepo:         conversationRepo,
		ConversationStudentRepo:  conversationStudentRepo,
		ConversationLocationRepo: conversationLocationRepo,
	}

	validReq := &upb.EvtUser_ParentAssignedToStudent{
		StudentId: "student-id",
		ParentId:  "parent-id",
	}

	member := entities.ConversationMembers{}
	database.AllNullEntity(&member)
	member.ConversationID.Set("student-conversation-id")
	member.ID.Set("parent-id")
	member.Role.Set(cpb.UserGroup_USER_GROUP_PARENT)
	cid := database.Text("student-conversation-id")

	conversationMembers := make(map[pgtype.Text]entities.ConversationMembers)
	conversationMembers[member.ID] = member
	conv := &entities.Conversation{
		ID:   cid,
		Name: database.Text("student name"),
	}
	studentLoc := "loc-1"
	studenLocations := map[string][]entities.ConversationLocation{
		cid.String: {
			{
				LocationID:     dbText(studentLoc),
				ConversationID: cid,
			},
		},
	}
	testCases := map[string]TestCase{
		"err upsert conversation": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.conversationRepo.BulkUpsert: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				// empty conv-student type parent, force create new one
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{}, nil)
				conversationLocationRepo.On("FindByConversationIDs", ctx, mockDB, database.TextArray([]string{"student-conversation-id"})).
					Once().Return(studenLocations, nil)
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String())).
					Once().Return([]string{cid.String}, nil)
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, cid).
					Once().Return(conv, nil)
				conversationMemberRepo.On("FindByConversationID", mock.Anything, mock.Anything, database.Text("student-conversation-id")).
					Once().Return(conversationMembers, nil)
				conversationRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err create conversation member": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{}, nil)
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String())).
					Once().Return([]string{"student-conversation-id"}, nil)
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, cid).
					Once().Return(conv, nil)
				conversationLocationRepo.On("FindByConversationIDs", ctx, mockDB, database.TextArray([]string{"student-conversation-id"})).
					Once().Return(studenLocations, nil)
				conversationLocationRepo.On("BulkUpsert", ctx, tx, mock.MatchedBy(func(upsertedLocs []entities.ConversationLocation) bool {
					return len(upsertedLocs) == 1 && upsertedLocs[0].LocationID.String == studentLoc
				})).Return(nil)

				conversationMemberRepo.On("FindByConversationID", mock.Anything, mock.Anything, database.Text("student-conversation-id")).
					Once().Return(conversationMembers, nil)
				conversationRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				conversationStudentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("BulkUpsert", mock.Anything, tx, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err create conversation student ": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("c.ConversationStudentRepo.BulkUpsert: %v", pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				// find conv id of student
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String())).
					Once().Return([]string{"student-conversation-id"}, nil)
				// check if conv id of parent exist
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{}, nil)
				// get student conv to copy name to parent conv
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, cid).
					Once().Return(conv, nil)
				// copy teachers of student conv to parent conv
				conversationMemberRepo.On("FindByConversationID", mock.Anything, mock.Anything, database.Text("student-conversation-id")).
					Once().Return(conversationMembers, nil)
				// copy student locations to parent locations
				conversationLocationRepo.On("FindByConversationIDs", ctx, mockDB, database.TextArray([]string{"student-conversation-id"})).
					Once().Return(studenLocations, nil)
				conversationLocationRepo.On("BulkUpsert", ctx, tx, mock.MatchedBy(func(upsertedLocs []entities.ConversationLocation) bool {
					return len(upsertedLocs) == 1 && upsertedLocs[0].LocationID.String == studentLoc
				})).Return(nil)
				conversationRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				conversationStudentRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"success adding parent to existing conversation": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				convID := "parent-conversation-id"
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{convID}, nil)

				// upsert new membership
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(mems []*domain.ConversationMembers) bool {
					given := mems[0]
					return given.UserID.String == validReq.ParentId && given.ConversationID.String == convID
				})).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
		"success creating new conversation": {
			ctx:          ctx,
			req:          validReq,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_PARENT.String())).
					Once().Return([]string{}, nil)
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray([]string{"student-id"}), database.Text(tpb.ConversationType_CONVERSATION_STUDENT.String())).
					Once().Return([]string{"student-conversation-id"}, nil)
				conversationRepo.On("FindByID", mock.Anything, mock.Anything, cid).
					Once().Return(conv, nil)
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, database.Text("student-conversation-id")).
					Once().Return(conversationMembers, nil)
				// student locations = parent location
				conversationLocationRepo.On("FindByConversationIDs", ctx, mockDB, database.TextArray([]string{"student-conversation-id"})).
					Once().Return(studenLocations, nil)
				conversationLocationRepo.On("BulkUpsert", ctx, tx, mock.MatchedBy(func(ents []entities.ConversationLocation) bool {
					return len(ents) == 1 && ents[0].LocationID.String == studentLoc
				})).Return(nil)

				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				conversationStudentRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatCreated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCases[name].setup(testCase.ctx)
			err := s.HandleEventCreateParentConversation(testCase.ctx, testCase.req.(*upb.EvtUser_ParentAssignedToStudent))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockDB, chatSvc, jsm, conversationMemberRepo, conversationRepo, conversationStudentRepo)
		})
	}
}

func TestChatService_UserJoinConversation(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("Rollback", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)

	jsm := &mock_nats.JetStreamManagement{}

	mockChatSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                      mockDB,
		ChatService:             mockChatSvc,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		ConversationStudentRepo: conversationStudentRepo,
		JSM:                     jsm,
	}
	convID := idutil.ULIDNow()

	validReq := &tpb.JoinConversationsRequest{
		ConversationIds: []string{convID},
	}
	userID := idutil.ULIDNow()

	testCases := map[string]TestCase{
		"success with role teacher": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					return len(members) == 1 &&
						members[0].ConversationID.String == convID &&
						members[0].Status.String == entities.ConversationStatusActive &&
						members[0].UserID.String == userID
				})).Once().Return(nil)
				// send and persist system message
				mockChatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(nil, nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		"success with role school admin": {
			ctx:          getCtx(userID, constant.RoleSchoolAdmin),
			req:          validReq,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					return len(members) == 1 &&
						members[0].ConversationID.String == convID &&
						members[0].Status.String == entities.ConversationStatusActive &&
						members[0].UserID.String == userID
				})).Once().Return(nil)
				// send and persist system message
				mockChatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(nil, nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		"err when create conversation member": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err when user have empty group": {
			ctx:          getCtx(userID, ""),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "user do not have permission with role  to join conversations"),
			setup:        func(ctx context.Context) {},
		},
		"err when user is in parent group": {
			ctx:          getCtx(userID, constant.RoleParent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to join conversations", constant.UserGroupParent)),
			setup:        func(ctx context.Context) {},
		},
		"err when user is in student group": {
			ctx:          getCtx(userID, constant.RoleStudent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to join conversations", constant.UserGroupStudent)),
			setup:        func(ctx context.Context) {},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.JoinConversations(testCase.ctx, testCase.req.(*tpb.JoinConversationsRequest))

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatService_UserJoinAllConversationWithLocations(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	locationRepo := new(mock_repositories.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)

	chatSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                      mockDB,
		ChatService:             chatSvc,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		ConversationStudentRepo: conversationStudentRepo,
		LocationRepo:            locationRepo,
		JSM:                     jsm,
	}
	locIDs := []string{"loc1", "loc2"}
	accessPaths := []string{"org/loca/loc1", "org/locb/loc2"}

	validReq := &tpb.JoinAllConversationRequest{
		LocationIds: locIDs,
	}

	member := entities.ConversationMembers{}
	database.AllNullEntity(&member)
	member.ConversationID.Set("student-conversation-id")
	member.ID.Set("teacher-id")
	member.Role.Set(cpb.UserGroup_USER_GROUP_TEACHER)
	userID := idutil.ULIDNow()
	schoolID := strconv.Itoa(bobConst.ManabieSchool)

	// context with user isn't in group teacher

	conversationMembers := make(map[pgtype.Text]entities.ConversationMembers)
	conversationMembers[member.ID] = member
	unjoinedConvID := "conv-1"
	unjoinedConv := &domain.Conversation{
		ID: dbText(unjoinedConvID),
	}
	testCases := map[string]TestCase{
		"success with role teacher": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: &tpb.JoinAllConversationResponse{},
			setup: func(ctx context.Context) {
				locationRepo.On("FindAccessPaths", mock.Anything, mockDB, locIDs).Once().Return(accessPaths, nil)
				conversationRepo.On("ListConversationUnjoinedInLocations", mock.Anything, mockDB, &domain.ListConversationUnjoinedFilter{
					UserID:      dbText(userID),
					OwnerIDs:    database.TextArray([]string{schoolID}),
					AccessPaths: database.TextArray(accessPaths),
				}).Return([]*domain.Conversation{unjoinedConv}, nil)
				conversationMemberRepo.On("BulkUpsert", mock.Anything, mockDB, mock.MatchedBy(func(memberships []*domain.ConversationMembers) bool {
					if len(memberships) != 1 {
						return false
					}
					mem := memberships[0]
					return mem.UserID.String == userID && mem.Status.String == domain.ConversationMemberStatusActive && mem.ConversationID.String == unjoinedConvID
				})).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(nil, nil)
				chatSvc.On("SendMessageToConversations", mock.Anything, mock.MatchedBy(func(reqs []*pb.SendMessageRequest) bool {
					if len(reqs) != 1 {
						return false
					}
					return reqs[0].ConversationId == unjoinedConvID
				}), domain.MessageToConversationOpts{Persist: true, AsUser: false}).Once().Return(nil)
			},
		},
		"success with role school admin": {
			ctx:          getCtx(userID, constant.RoleSchoolAdmin),
			req:          validReq,
			expectedResp: &tpb.JoinAllConversationResponse{},
			setup: func(ctx context.Context) {
				locationRepo.On("FindAccessPaths", mock.Anything, mockDB, locIDs).Once().Return(accessPaths, nil)
				conversationRepo.On("ListConversationUnjoinedInLocations", mock.Anything, mockDB, &domain.ListConversationUnjoinedFilter{
					UserID:      dbText(userID),
					OwnerIDs:    database.TextArray([]string{schoolID}),
					AccessPaths: database.TextArray(accessPaths),
				}).Return([]*domain.Conversation{unjoinedConv}, nil)
				conversationMemberRepo.On("BulkUpsert", mock.Anything, mockDB, mock.MatchedBy(func(memberships []*domain.ConversationMembers) bool {
					if len(memberships) != 1 {
						return false
					}
					mem := memberships[0]
					return mem.UserID.String == userID && mem.Status.String == domain.ConversationMemberStatusActive && mem.ConversationID.String == unjoinedConvID
				})).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(nil, nil)
				chatSvc.On("SendMessageToConversations", mock.Anything, mock.MatchedBy(func(reqs []*pb.SendMessageRequest) bool {
					if len(reqs) != 1 {
						return false
					}
					return reqs[0].ConversationId == unjoinedConvID
				}), domain.MessageToConversationOpts{Persist: true, AsUser: false}).Once().Return(nil)
			},
		},
		"err when user have empty group": {
			ctx:          getCtx(userID, ""),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "ChatModifierService.JoinAllConversationsWithLocations: user do not have permission with role "),
			setup:        func(ctx context.Context) {},
		},
		"err when user is student group": {
			ctx:          getCtx(userID, constant.RoleStudent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversationsWithLocations: user do not have permission with role %s", constant.UserGroupStudent)),
			setup:        func(ctx context.Context) {},
		},
		"err when user is parent group": {
			ctx:          getCtx(userID, constant.RoleParent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversationsWithLocations: user do not have permission with role %s", constant.UserGroupParent)),
			setup:        func(ctx context.Context) {},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.JoinAllConversationsWithLocations(testCase.ctx, testCase.req.(*tpb.JoinAllConversationRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatService_UserJoinAllConversation(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)

	chatSvc := &mock_services.ChatService{}

	s := &ChatModifier{
		DB:                      mockDB,
		ChatService:             chatSvc,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		ConversationStudentRepo: conversationStudentRepo,
		JSM:                     jsm,
	}

	validReq := &tpb.JoinAllConversationRequest{}

	member := entities.ConversationMembers{}
	database.AllNullEntity(&member)
	member.ConversationID.Set("student-conversation-id")
	member.ID.Set("teacher-id")
	member.Role.Set(cpb.UserGroup_USER_GROUP_TEACHER)

	userID := idutil.ULIDNow()
	conversationMembers := make(map[pgtype.Text]entities.ConversationMembers)
	conversationMembers[member.ID] = member

	testCases := map[string]TestCase{
		"success with role teacher": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationRepo.On("ListConversationUnjoined", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Conversation{{ID: pgtype.Text{String: "01F31YQD08V96XJXZ34N895482"}}}, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				chatSvc.On("SendMessageToConversations", mock.Anything, mock.MatchedBy(func(reqs []*pb.SendMessageRequest) bool {
					if len(reqs) != 1 {
						return false
					}
					return reqs[0].ConversationId == "01F31YQD08V96XJXZ34N895482"
				}), domain.MessageToConversationOpts{Persist: true, AsUser: false}).Once().Return(nil)
			},
		},
		"success with role school admin": {
			ctx:          getCtx(userID, constant.RoleSchoolAdmin),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationRepo.On("ListConversationUnjoined", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Conversation{{ID: pgtype.Text{String: "01F31YQD08V96XJXZ34N895482"}}}, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				chatSvc.On("SendMessageToConversations", mock.Anything, mock.MatchedBy(func(reqs []*pb.SendMessageRequest) bool {
					if len(reqs) != 1 {
						return false
					}
					return reqs[0].ConversationId == "01F31YQD08V96XJXZ34N895482"
				}), domain.MessageToConversationOpts{Persist: true, AsUser: false}).Once().Return(nil)
			},
		},
		"err fetch conversations with schoolIds": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("ConversationRepo.ListConversationUnjoined: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationRepo.On("ListConversationUnjoined", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"err when create conversation member": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("makeUserJoinConversationsd: %s", fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %w", pgx.ErrTxClosed))),
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationRepo.On("ListConversationUnjoined", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.Conversation{{ID: pgtype.Text{String: "01F31YQD08V96XJXZ34N895482"}}}, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("c.ConversationMemberRepo.BulkUpsert: %v", pgx.ErrTxClosed.Error()))
			},
		},
		"err when user have empty group": {
			ctx:          getCtx(userID, ""),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "ChatModifierService.JoinAllConversations: user do not have permission with role "),
			setup:        func(ctx context.Context) {},
		},
		"err when user is student group": {
			ctx:          getCtx(userID, constant.RoleStudent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversations: user do not have permission with role %s", constant.UserGroupStudent)),
			setup:        func(ctx context.Context) {},
		},
		"err when user is parent group": {
			ctx:          getCtx(userID, constant.RoleParent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("ChatModifierService.JoinAllConversations: user do not have permission with role %s", constant.UserGroupParent)),
			setup:        func(ctx context.Context) {},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.JoinAllConversations(testCase.ctx, testCase.req.(*tpb.JoinAllConversationRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatService_UserLeaveConversations(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	onlineUserRepo := new(mock_repositories.MockOnlineUserRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	jsm := &mock_nats.JetStreamManagement{}

	chatSvc := &mock_services.ChatService{}
	s := &ChatModifier{
		DB:                      mockDB,
		JSM:                     jsm,
		ChatService:             chatSvc,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		ConversationStudentRepo: conversationStudentRepo,
	}

	conversationID := idutil.ULIDNow()

	validReq := &tpb.LeaveConversationsRequest{
		ConversationIds: []string{conversationID},
	}
	invalidReq := &tpb.LeaveConversationsRequest{}
	userID := idutil.ULIDNow()

	member := entities.ConversationMembers{}
	database.AllNullEntity(&member)
	member.ConversationID.Set(conversationID)
	member.ID.Set("teacher-id")
	member.Role.Set(cpb.UserGroup_USER_GROUP_TEACHER)
	member.UserID.Set(userID)

	conversationMembers := make(map[pgtype.Text]entities.ConversationMembers)
	conversationMembers[member.ID] = member
	testCases := map[string]TestCase{
		"err empty conversation ids in request": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          invalidReq,
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, "empty conversation ids"),
			setup:        func(ctx context.Context) {},
		},
		"err upsert conversation member": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				deactivatedMember := member
				deactivatedMember.Status.Set(entities.ConversationStatusInActive)
				deactivatedMembers := []*entities.ConversationMembers{&deactivatedMember}
				conversationMemberRepo.On("FindByCIDsAndUserID",
					mock.Anything, mock.Anything, database.TextArray([]string{conversationID}), database.Text(userID)).
					Once().Return(deactivatedMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		"err user have empty group": {
			ctx:          getCtx(userID, ""),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, "user do not have permission with role  to leave conversations"),
			setup:        func(ctx context.Context) {},
		},
		"err when user is student group": {
			ctx:          getCtx(userID, constant.RoleStudent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to leave conversations", constant.UserGroupStudent)),
			setup:        func(ctx context.Context) {},
		},
		"err when user is parent group": {
			ctx:          getCtx(userID, constant.RoleParent),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, fmt.Sprintf("user do not have permission with role %s to leave conversations", constant.UserGroupParent)),
			setup:        func(ctx context.Context) {},
		},
		"successfully with teacher group": {
			ctx:          getCtx(userID, constant.RoleTeacher),
			req:          validReq,
			expectedResp: &tpb.LeaveConversationsResponse{},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				deactivatedMember := member
				deactivatedMember.Status.Set(entities.ConversationStatusInActive)
				deactivatedMembers := []*entities.ConversationMembers{&deactivatedMember}

				conversationMemberRepo.On("FindByCIDsAndUserID",
					mock.Anything, mock.Anything, database.TextArray([]string{conversationID}), database.Text(userID)).
					Once().Return(deactivatedMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					if len(members) != 1 {
						return false
					}
					calledMem := members[0]
					return calledMem.Status.String == entities.ConversationStatusInActive &&
						calledMem.UserID.String == userID
				})).Once().Return(nil)
				// system message to users
				chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)

				// system message to user who left
				chatSvc.On("SendMessageToUsers", mock.Anything, []string{deactivatedMember.UserID.String}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Return(nil, nil)

				// teacher is no longer member, but service must also send leave system message to this user
				onlineUserRepo.On("Find", mock.Anything, mock.Anything, database.TextArray([]string{userID}), mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		"successfully with school admin group": {
			ctx:          getCtx(userID, constant.RoleSchoolAdmin),
			req:          validReq,
			expectedResp: &tpb.LeaveConversationsResponse{},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				deactivatedMember := member
				deactivatedMember.Status.Set(entities.ConversationStatusInActive)
				deactivatedMembers := []*entities.ConversationMembers{&deactivatedMember}

				conversationMemberRepo.On("FindByCIDsAndUserID",
					mock.Anything, mock.Anything, database.TextArray([]string{conversationID}), database.Text(userID)).
					Once().Return(deactivatedMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(members []*entities.ConversationMembers) bool {
					if len(members) != 1 {
						return false
					}
					calledMem := members[0]
					return calledMem.Status.String == entities.ConversationStatusInActive &&
						calledMem.UserID.String == userID
				})).Once().Return(nil)
				// system message to users
				chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)

				// system message to user who left
				chatSvc.On("SendMessageToUsers", mock.Anything, []string{deactivatedMember.UserID.String}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)
				jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Return(nil, nil)

				// teacher is no longer member, but service must also send leave system message to this user
				onlineUserRepo.On("Find", mock.Anything, mock.Anything, database.TextArray([]string{userID}), mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.LeaveConversations(testCase.ctx, testCase.req.(*tpb.LeaveConversationsRequest))

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatService_HandleUpsertStaff(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	grantedPermissionsRepo := new(mock_repositories.MockGrantedPermissionsRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)

	jsm := &mock_nats.JetStreamManagement{}

	chatSvc := &mock_services.ChatService{}
	s := &ChatModifier{
		DB:                       mockDB,
		JSM:                      jsm,
		Logger:                   zap.NewNop(),
		ChatService:              chatSvc,
		ConversationMemberRepo:   conversationMemberRepo,
		ConversationRepo:         conversationRepo,
		ConversationStudentRepo:  conversationStudentRepo,
		ConversationLocationRepo: conversationLocationRepo,
		GrantedPermissionRepo:    grantedPermissionsRepo,
	}

	conversationIDs := []string{"conv 1", "conv 2"}
	staffID := idutil.ULIDNow()

	req := &upb.EvtUpsertStaff{
		StaffId: staffID,
	}

	emptyConvLocationsMap := map[string][]domain.ConversationLocation{
		"conv 1": []entities.ConversationLocation{},
		"conv 2": []entities.ConversationLocation{},
	}

	nonEmptyConvLocationsMap := map[string][]domain.ConversationLocation{
		"conv 1": []entities.ConversationLocation{
			{
				ConversationID: database.Text("conv 1"),
				LocationID:     database.Text("location-1"),
			},
		},
		"conv 2": []entities.ConversationLocation{
			{
				ConversationID: database.Text("conv 2"),
				LocationID:     database.Text("location-2"),
			},
		},
	}

	emptyGrantedPermissionMap := map[string][]*domain.GrantedPermission{
		staffID: []*entities.GrantedPermission{},
	}

	nonEmptyGrantedPermissionMap := map[string][]*domain.GrantedPermission{
		staffID: []*entities.GrantedPermission{
			{
				UserID:     database.Text(staffID),
				LocationID: database.Text("location-1"),
			},
		},
	}

	testCases := map[string]TestCase{
		"err empty staffID": {
			ctx:         ctx,
			req:         &upb.EvtUpsertStaff{},
			expectedErr: errors.New("no staffID found"),
			setup:       func(ctx context.Context) {},
		},
		"err FindByStaffIDs": {
			ctx:         ctx,
			req:         req,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err FindByConversationIDs": {
			ctx:          ctx,
			req:          req,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err FindByUserIDAndPermissionName": {
			ctx:          ctx,
			req:          req,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(emptyConvLocationsMap, nil)
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray([]string{staffID}), database.Text("master.location.read")).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"error with empty conversation locations": {
			ctx:         ctx,
			req:         req,
			expectedErr: errors.New("no inactivating conversations found"),
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(emptyConvLocationsMap, nil)
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray([]string{staffID}), database.Text("master.location.read")).Once().Return(nonEmptyGrantedPermissionMap, nil)
			},
		},
		"success with empty granted locations": {
			ctx:          ctx,
			req:          req,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(nonEmptyConvLocationsMap, nil)
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray([]string{staffID}), database.Text("master.location.read")).Once().Return(emptyGrantedPermissionMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, mock.Anything, database.TextArray([]string{staffID}), database.Text(domain.ConversationStatusInActive)).Once().Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(2).(pgtype.TextArray)
					assert.ElementsMatch(t, arg.Elements, database.TextArray(conversationIDs).Elements)
				})

				for range conversationIDs {
					// system message to users
					chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
						Persist: true,
					}).Once().Return(&pb.MessageResponse{}, nil)

					// system message to this parent
					chatSvc.On("SendMessageToUsers", mock.Anything, []string{staffID}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)

					jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				}
			},
		},
		"success with non-empty granted locations": {
			ctx:          ctx,
			req:          req,
			expectedResp: nil,
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(nonEmptyConvLocationsMap, nil)
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray([]string{staffID}), database.Text("master.location.read")).Once().Return(nonEmptyGrantedPermissionMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, database.TextArray([]string{"conv 2"}), database.TextArray([]string{staffID}), database.Text(domain.ConversationStatusInActive)).Once().Return(nil)
				for range conversationIDs {
					// system message to users
					chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
						Persist: true,
					}).Once().Return(&pb.MessageResponse{}, nil)

					// system message to this parent
					chatSvc.On("SendMessageToUsers", mock.Anything, []string{staffID}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)

					jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				}
			},
		},
		"error SetStatusByConversationAndUserIDs": {
			ctx:         ctx,
			req:         req,
			expectedErr: fmt.Errorf("ExecInTx: ConversationMemberRepo.SetStatusByConversationAndUserIDs: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStaffIDs", mock.Anything, mockDB, database.TextArray([]string{staffID})).Once().Return(conversationIDs, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, database.TextArray(conversationIDs)).Once().Return(nonEmptyConvLocationsMap, nil)
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray([]string{staffID}), database.Text("master.location.read")).Once().Return(nonEmptyGrantedPermissionMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, database.TextArray([]string{"conv 2"}), database.TextArray([]string{staffID}), database.Text(domain.ConversationStatusInActive)).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.HandleUpsertStaff(testCase.ctx, testCase.req.(*upb.EvtUpsertStaff))

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChatService_HandleUpsertUserGroup(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	mockDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("Commit", mock.Anything).Return(nil)

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	grantedPermissionsRepo := new(mock_repositories.MockGrantedPermissionsRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	userGroupMemberRepo := new(mock_repositories.MockUserGroupMembersRepo)

	jsm := &mock_nats.JetStreamManagement{}

	chatSvc := &mock_services.ChatService{}
	s := &ChatModifier{
		DB:                       mockDB,
		JSM:                      jsm,
		Logger:                   zap.NewNop(),
		ChatService:              chatSvc,
		ConversationMemberRepo:   conversationMemberRepo,
		ConversationRepo:         conversationRepo,
		ConversationStudentRepo:  conversationStudentRepo,
		ConversationLocationRepo: conversationLocationRepo,
		GrantedPermissionRepo:    grantedPermissionsRepo,
		UserGroupMemberRepo:      userGroupMemberRepo,
	}

	userGroupID := idutil.ULIDNow()

	req := &upb.EvtUpsertUserGroup{
		UserGroupId: userGroupID,
	}

	userGroupMemberIDs := []string{"user-1", "user-2"}

	conversationMembersMap := map[string][]string{
		"user-1": {"conv-1", "conv-2"},
		"user-2": {"conv-2"},
	}

	conversationIDs := []string{"conv-1", "conv-2"}
	conversationLocationsMap := map[string][]domain.ConversationLocation{
		"conv-1": {
			{
				LocationID: database.Text("loc-1"),
			},
		},
		"con-2": {
			{
				LocationID: database.Text("loc-1"),
			}, {
				LocationID: database.Text("loc-2"),
			},
		},
	}

	grantedLocationIDsMap := map[string][]*domain.GrantedPermission{
		"user-1": []*entities.GrantedPermission{
			{
				UserID:     database.Text("user-1"),
				LocationID: database.Text("loc-1"),
			},
			{
				UserID:     database.Text("user-2"),
				LocationID: database.Text("loc-2"),
			},
		},
	}

	testCases := map[string]TestCase{
		"err empty user group id": {
			req:         &upb.EvtUpsertUserGroup{},
			expectedErr: errors.New("no userGroupID found"),
			setup:       func(ctx context.Context) {},
		},
		"err FindUserIDsByUserGroupID": {
			req:         req,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err when empty user group member ids": {
			req:         req,
			expectedErr: errors.New(fmt.Sprintf("no user group members found with ID %s", userGroupID)),
			setup: func(ctx context.Context) {
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return([]string{}, nil)
			},
		},
		"err FindUserIDConversationIDsMapByUserIDs": {
			req:          req,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(userGroupMemberIDs, nil)
				conversationMemberRepo.On("FindUserIDConversationIDsMapByUserIDs", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs)).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err FindByConversationIDs": {
			req:          req,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(userGroupMemberIDs, nil)
				conversationMemberRepo.On("FindUserIDConversationIDsMapByUserIDs", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs)).Once().Return(conversationMembersMap, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err FindByUserIDAndPermissionName": {
			req:          req,
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(userGroupMemberIDs, nil)
				conversationMemberRepo.On("FindUserIDConversationIDsMapByUserIDs", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs)).Once().Return(conversationMembersMap, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, mock.Anything).Once().Return(conversationLocationsMap, nil).Run(func(args mock.Arguments) {
					arg1 := args.Get(2).(pgtype.TextArray)
					assert.ElementsMatch(t, arg1.Elements, database.TextArray(conversationIDs).Elements)
				})
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs), database.Text("master.location.read")).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err SetStatusByConversationAndUserIDs": {
			req:          req,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("ExecInTx: retrying to find InactiveConversations"),
			setup: func(ctx context.Context) {
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(userGroupMemberIDs, nil)
				conversationMemberRepo.On("FindUserIDConversationIDsMapByUserIDs", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs)).Once().Return(conversationMembersMap, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, mock.Anything).Once().Return(conversationLocationsMap, nil).Run(func(args mock.Arguments) {
					arg1 := args.Get(2).(pgtype.TextArray)
					assert.ElementsMatch(t, arg1.Elements, database.TextArray(conversationIDs).Elements)
				})
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs), database.Text("master.location.read")).Once().Return(grantedLocationIDsMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", ctx, tx, mock.Anything, mock.Anything, database.Text(domain.ConversationStatusInActive)).Once().Return(nil)
			},
		},
		"successful": {
			req:          req,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				updatedGrantedLocationIDsMap := map[string][]*domain.GrantedPermission{
					"user-1": []*entities.GrantedPermission{
						{
							UserID:     database.Text("user-1"),
							LocationID: database.Text("loc-3"),
						},
						{
							UserID:     database.Text("user-2"),
							LocationID: database.Text("loc-4"),
						},
					},
				}
				userGroupMemberRepo.On("FindUserIDsByUserGroupID", mock.Anything, mockDB, database.Text(userGroupID)).Once().Return(userGroupMemberIDs, nil)
				conversationMemberRepo.On("FindUserIDConversationIDsMapByUserIDs", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs)).Once().Return(conversationMembersMap, nil)
				conversationLocationRepo.On("FindByConversationIDs", mock.Anything, mockDB, mock.Anything).Once().Return(conversationLocationsMap, nil).Run(func(args mock.Arguments) {
					arg1 := args.Get(2).(pgtype.TextArray)
					assert.ElementsMatch(t, arg1.Elements, database.TextArray(conversationIDs).Elements)
				})
				grantedPermissionsRepo.On("FindByUserIDAndPermissionName", mock.Anything, mockDB, database.TextArray(userGroupMemberIDs), database.Text("master.location.read")).Once().Return(updatedGrantedLocationIDsMap, nil)
				conversationMemberRepo.On("SetStatusByConversationAndUserIDs", mock.Anything, tx, mock.Anything, mock.Anything, database.Text(domain.ConversationStatusInActive)).Twice().Return(nil)

				for userID := range conversationMembersMap {
					// system message to users
					chatSvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
						Persist: true,
					}).Once().Return(&pb.MessageResponse{}, nil)

					// system message to this parent
					chatSvc.On("SendMessageToUsers", mock.Anything, []string{userID}, mock.Anything, domain.MessageToUserOpts{}).Once().Return(nil)

					jsm.On("TracedPublish", mock.Anything, mock.Anything, constants.SubjectChatMembersUpdated, mock.Anything).Once().Return(&nats.PubAck{}, nil)
				}
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			_, err := s.HandleUpsertUserGroup(ctx, testCase.req.(*upb.EvtUpsertUserGroup))

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
