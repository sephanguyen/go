package lesson

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_core "github.com/manabie-com/backend/mock/tom/app/core"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	mock_services "github.com/manabie-com/backend/mock/tom/services"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_HandleEventUpdateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	s := &ChatModifier{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		DB:                     mockDB,
	}

	existLearner := "exist-learner"
	addedLearner := "added-learner"
	tobeRemovedLearner := "removed-learner"

	lessonName := idutil.ULIDNow()
	lessonID := idutil.ULIDNow()
	convID := idutil.ULIDNow()
	lessonConv := &lentities.ConversationLesson{
		LessonID:       database.Text(lessonID),
		ConversationID: database.Text(convID),
	}

	testCases := []TestCase{
		{
			name: "update removing one student",
			ctx:  ctx,
			req: &bpb.EvtLesson_UpdateLesson{
				LessonId:   lessonID,
				ClassName:  lessonName,
				LearnerIds: []string{existLearner},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, database.Text(lessonID)).Once().Return(lessonConv, nil)
				currentMembers := map[pgtype.Text]domain.ConversationMembers{
					dbText(existLearner):       makeConvMember(existLearner, convID),
					dbText(tobeRemovedLearner): makeConvMember(tobeRemovedLearner, convID),
				}
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, database.Text(convID)).Once().Return(currentMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(newMembers []*domain.ConversationMembers) bool {
					// upserting inactive member
					if len(newMembers) != 1 {
						return false
					}
					if newMembers[0].UserID.String != tobeRemovedLearner || newMembers[0].Status.String != domain.ConversationStatusInActive {
						return false
					}
					return true
				})).Once().Return(nil)
			},
		},
		{
			name: "update adding one student",
			ctx:  ctx,
			req: &bpb.EvtLesson_UpdateLesson{
				LessonId:  lessonID,
				ClassName: lessonName,
				LearnerIds: []string{
					existLearner, addedLearner,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, database.Text(lessonID)).Once().Return(lessonConv, nil)
				currentMembers := map[pgtype.Text]domain.ConversationMembers{
					dbText(existLearner): makeConvMember(existLearner, convID),
				}
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, database.Text(convID)).Once().Return(currentMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(newMembers []*domain.ConversationMembers) bool {
					// upserting inactive member
					if len(newMembers) != 1 {
						return false
					}
					if newMembers[0].UserID.String != addedLearner || newMembers[0].Status.String != domain.ConversationStatusActive {
						return false
					}
					return true
				})).Once().Return(nil)
			},
		},
		{
			name: "error bulk upserting",
			ctx:  ctx,
			req: &bpb.EvtLesson_UpdateLesson{
				LessonId:  lessonID,
				ClassName: lessonName,
				LearnerIds: []string{
					addedLearner,
				},
			},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, database.Text(lessonID)).Once().Return(lessonConv, nil)
				currentMembers := map[pgtype.Text]domain.ConversationMembers{}
				conversationMemberRepo.On("FindByConversationID", ctx, mock.Anything, database.Text(convID)).Once().Return(currentMembers, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(newMembers []*domain.ConversationMembers) bool {
					// upserting inactive member
					if len(newMembers) != 1 {
						return false
					}
					if newMembers[0].UserID.String != addedLearner || newMembers[0].Status.String != domain.ConversationStatusActive {
						return false
					}
					return true
				})).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleEventUpdateLesson(testCase.ctx, testCase.req.(*bpb.EvtLesson_UpdateLesson))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func Test_HandleEventCreateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	s := &ChatModifier{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		DB:                     mockDB,
	}

	learner1 := "student-1"
	learner2 := "student-2"

	lessonName := idutil.ULIDNow()
	lessonID := idutil.ULIDNow()

	lessonsReq := []*bpb.EvtLesson_Lesson{
		{
			Name:       lessonName,
			LessonId:   lessonID,
			LearnerIds: []string{learner1, learner2},
		},
	}

	testCases := []TestCase{
		{
			name: "success",
			ctx:  ctx,
			req: &bpb.EvtLesson_CreateLessons{
				Lessons: lessonsReq,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				recentConvID := ""
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(convs []*domain.Conversation) bool {
					recentConvID = convs[0].ID.String
					return len(convs) == 1
				})).Once().Return(nil)
				// correct conversation lesson
				conversationLessonRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(clessons []*lentities.ConversationLesson) bool {
					if len(clessons) != 1 {
						return false
					}
					return clessons[0].ConversationID.String == recentConvID &&
						clessons[0].LessonID.String == lessonID
				})).Once().Return(nil)
				// correct student membership inserted

				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(es []*domain.ConversationMembers) bool {
					for _, e := range es {
						if e.UserID.String != learner1 && e.UserID.String != learner2 {
							return false
						}
						if e.ConversationID.String != recentConvID {
							return false
						}
					}
					return true
				})).Once().Return(nil)
			},
		},
		{
			name: "err upsert conversation",
			ctx:  ctx,
			req: &bpb.EvtLesson_CreateLessons{
				Lessons: lessonsReq,
			},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleEventCreateLesson(testCase.ctx, testCase.req.(*bpb.EvtLesson_CreateLessons))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func multiMockAnything(count int) []interface{} {
	ret := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		ret = append(ret, mock.Anything)
	}
	return ret
}

func Test_HandleEventLeaveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)
	chatsvc := &mock_services.ChatService{}

	s := &ChatModifier{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		ChatService:            chatsvc,
	}

	learner1 := "student-1"
	teacher := "teacher"

	lessonID := idutil.ULIDNow()
	convID := idutil.ULIDNow()

	lessonConversation := &lentities.ConversationLesson{
		ConversationID: database.Text(convID),
		LessonID:       database.Text(lessonID),
	}

	testCases := []TestCase{
		{
			name: "student can't leave lesson",
			ctx:  ctx,
			req: &bpb.EvtLesson_LeaveLesson{
				LessonId: lessonID,
				UserId:   learner1,
			},
			expectedErr: errOnlyTeacherCanLeaveLesson,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)
				studentMembership := makeConvMember(learner1, convID)
				studentMembership.Role.Set(cpb.UserGroup_USER_GROUP_STUDENT.String())
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, mock.Anything, dbText(convID), dbText(learner1)).Once().Return(&studentMembership, nil)
			},
		},
		{
			name: "teacher successfully leave lesson",
			ctx:  ctx,
			req: &bpb.EvtLesson_LeaveLesson{
				LessonId: lessonID,
				UserId:   teacher,
			},
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)

				studentMembership := makeConvMember(teacher, convID)
				studentMembership.Role.Set(cpb.UserGroup_USER_GROUP_TEACHER.String())
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, mock.Anything, dbText(convID), dbText(teacher)).Once().Return(&studentMembership, nil)

				conversationMemberRepo.On("SetStatus", ctx, mock.Anything, dbText(convID), database.TextArray([]string{teacher}), dbText(domain.ConversationStatusInActive)).
					Once().Return(nil)

				chatsvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)
			},
		},
		{
			name: "error updating membership",
			ctx:  ctx,
			req: &bpb.EvtLesson_LeaveLesson{
				LessonId: lessonID,
				UserId:   teacher,
			},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)

				studentMembership := makeConvMember(teacher, convID)
				studentMembership.Role.Set(cpb.UserGroup_USER_GROUP_TEACHER.String())
				conversationMemberRepo.On("FindByCIDAndUserID", ctx, mock.Anything, dbText(convID), dbText(teacher)).Once().Return(&studentMembership, nil)

				conversationMemberRepo.On("SetStatus", ctx, mock.Anything, dbText(convID), database.TextArray([]string{teacher}), dbText(domain.ConversationStatusInActive)).
					Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.HandleEventLeaveLesson(testCase.ctx, testCase.req.(*bpb.EvtLesson_LeaveLesson))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func makeConvMember(userID string, convID string) domain.ConversationMembers {
	return domain.ConversationMembers{
		UserID:         database.Text(userID),
		ConversationID: database.Text(convID),
		Role:           database.Text(cpb.UserGroup_USER_GROUP_STUDENT.String()),
	}
}

func Test_HandleEventEndLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)

	chatsvc := &mock_services.ChatService{}
	s := &ChatModifier{
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		ChatService:            chatsvc,
	}

	learner1 := "student-1"
	teacher := "teacher"

	lessonID := idutil.ULIDNow()
	convID := idutil.ULIDNow()

	lessonConversation := &lentities.ConversationLesson{
		ConversationID: database.Text(convID),
		LessonID:       database.Text(lessonID),
	}

	testCases := []TestCase{
		{
			name: "successfully end live lesson",
			ctx:  ctx,
			req: &bpb.EvtLesson_EndLiveLesson{
				LessonId: lessonID,
				UserId:   learner1,
			},
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)

				// persist and send system message
				chatsvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)
			},
		},
		{
			name: "error finding lesson conversation",
			ctx:  ctx,
			req: &bpb.EvtLesson_EndLiveLesson{
				LessonId: lessonID,
				UserId:   teacher,
			},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleEventEndLiveLesson(testCase.ctx, testCase.req.(*bpb.EvtLesson_EndLiveLesson))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func Test_HandleEventJoinLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)

	chatsvc := &mock_services.ChatService{}
	s := &ChatModifier{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		ChatService:            chatsvc,
	}

	learner1 := "student-1"
	teacher := "teacher"

	lessonID := idutil.ULIDNow()
	convID := idutil.ULIDNow()

	lessonConversation := &lentities.ConversationLesson{
		ConversationID: database.Text(convID),
		LessonID:       database.Text(lessonID),
	}

	testCases := []TestCase{
		{
			name: "student cannot join lesson",
			ctx:  ctx,
			req: &bpb.EvtLesson_JoinLesson{
				LessonId:  lessonID,
				UserId:    learner1,
				UserGroup: bpb.UserGroup(cpb.UserGroup_USER_GROUP_STUDENT),
			},
			expectedErr: nil,
			setup:       func(ctx context.Context) {},
		},
		{
			name: "teacher successfully join lesson",
			ctx:  ctx,
			req: &bpb.EvtLesson_JoinLesson{
				LessonId:  lessonID,
				UserId:    teacher,
				UserGroup: bpb.UserGroup(cpb.UserGroup_USER_GROUP_TEACHER),
			},
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)

				conversationMemberRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(member *domain.ConversationMembers) bool {
					return member.UserID.String == teacher && member.ConversationID.String == convID
				})).Once().Return(nil)
				// persist and send system message
				chatsvc.On("SendMessageToConversation", mock.Anything, mock.Anything, domain.MessageToConversationOpts{
					Persist: true,
				}).Once().Return(&pb.MessageResponse{}, nil)
			},
		},
		{
			name: "error creating membership",
			ctx:  ctx,
			req: &bpb.EvtLesson_JoinLesson{
				LessonId:  lessonID,
				UserId:    teacher,
				UserGroup: bpb.UserGroup(cpb.UserGroup_USER_GROUP_TEACHER),
			},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonID", ctx, mock.Anything, dbText(lessonID)).Once().Return(lessonConversation, nil)

				conversationMemberRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(member *domain.ConversationMembers) bool {
					return member.UserID.String == teacher && member.ConversationID.String == convID
				})).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.HandleEventJoinLesson(testCase.ctx, testCase.req.(*bpb.EvtLesson_JoinLesson))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func Test_SyncLessonConversationStudents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLessonRepo := new(mock_repositories.MockConversationLessonRepo)

	chatsvc := &mock_services.ChatService{}
	s := &ChatModifier{
		ConversationMemberRepo: conversationMemberRepo,
		ConversationRepo:       conversationRepo,
		ConversationLessonRepo: conversationLessonRepo,
		Logger:                 zap.NewNop(),
		ChatService:            chatsvc,
	}

	learner1 := "student-1"
	learner2 := "student-2"

	lessonID := idutil.ULIDNow()
	nonExistLessonID := idutil.ULIDNow()
	convID := idutil.ULIDNow()

	lessonConversation := &lentities.ConversationLesson{
		ConversationID: database.Text(convID),
		LessonID:       database.Text(lessonID),
	}

	testCases := []TestCase{
		{
			name: "err bulk upsert from db",
			ctx:  ctx,
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					StudentId:  learner1,
					LessonIds:  []string{lessonID},
				},
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  learner2,
					LessonIds:  []string{lessonID},
				},
			},
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonIDs", ctx, mock.Anything, database.TextArray([]string{lessonID}), false).Once().
					Return([]*lentities.ConversationLesson{lessonConversation}, nil)

				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "sync non-exist lesson",
			ctx:  ctx,
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					StudentId:  learner1,
					LessonIds:  []string{lessonID, nonExistLessonID},
				},
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  learner2,
					LessonIds:  []string{lessonID},
				},
			},
			expectedErr: errLessonConversationDoesNotExist,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonIDs", ctx, mock.Anything, database.TextArray([]string{lessonID, nonExistLessonID}), false).Once().
					Return([]*lentities.ConversationLesson{lessonConversation}, nil)
			},
		},
		{
			name: "successful sync",
			ctx:  ctx,
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					StudentId:  learner1,
					LessonIds:  []string{lessonID},
				},
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  learner2,
					LessonIds:  []string{lessonID},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationLessonRepo.On("FindByLessonIDs", ctx, mock.Anything, database.TextArray([]string{lessonID}), false).Once().
					Return([]*lentities.ConversationLesson{lessonConversation}, nil)

				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(memberships []*domain.ConversationMembers) bool {
					if len(memberships) != 2 {
						return false
					}
					for _, membership := range memberships {
						switch membership.UserID.String {
						case learner1:
							if membership.Status.String != domain.ConversationStatusInActive {
								return false
							}
						case learner2:
							if membership.Status.String != domain.ConversationStatusActive {
								return false
							}
						default:
							return false
						}
					}
					return true
				})).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.SyncLessonConversationStudents(testCase.ctx, testCase.req.([]*npb.EventSyncUserCourse_StudentLesson))
			assert.ErrorIs(t, err, testCase.expectedErr)
		})
	}
}

func Test_CreateLiveLessonPrivateConversation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	privateConversationLessonRepo := new(mock_repositories.MockPrivateConversationLessonRepo)
	messageRepo := new(mock_repositories.MockMessageRepo)
	usersRepo := new(mock_repositories.MockUsersRepo)
	chatInfra := &mock_core.ChatInfra{}

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	s := &ChatModifier{
		ConversationMemberRepo:        conversationMemberRepo,
		ConversationRepo:              conversationRepo,
		PrivateConversationLessonRepo: privateConversationLessonRepo,
		Logger:                        zap.NewNop(),
		ChatInfra:                     chatInfra,
		MessageRepo:                   messageRepo,
		UserRepo:                      usersRepo,
		DB:                            mockDB,
	}

	teacherID := "memberB"
	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	studentId := "memberA"
	usersMapRes := map[string]*core.User{}
	usersMapRes[studentId] = &core.User{
		UserID:    pgtype.Text{String: studentId},
		UserGroup: pgtype.Text{String: cpb.UserGroup_USER_GROUP_STUDENT.String()},
	}
	usersMapRes[teacherID] = &core.User{
		UserID:    pgtype.Text{String: teacherID},
		UserGroup: pgtype.Text{String: cpb.UserGroup_USER_GROUP_TEACHER.String()},
	}
	lessonID := idutil.ULIDNow()
	userIds := []string{studentId}

	testCaseReq := &tpb.CreateLiveLessonPrivateConversationRequest{
		LessonId: lessonID,
		UserIds:  userIds,
	}
	invalidReq := &tpb.CreateLiveLessonPrivateConversationRequest{}
	txError := errors.New("database.ExecInTx: rcv.conversationRepo.BulkUpsert: tx is closed")
	expectedResp := &tpb.CreateLiveLessonPrivateConversationResponse{
		Conversation: &tpb.Conversation{
			Seen:             false,
			Status:           tpb.ConversationStatus_CONVERSATION_STATUS_NONE,
			ConversationType: tpb.ConversationType_CONVERSATION_LESSON_PRIVATE,
			ConversationName: "",
		},
	}
	testCases := []TestCase{
		{
			name:         "success",
			ctx:          ctx,
			req:          testCaseReq,
			expectedErr:  nil,
			expectedResp: expectedResp,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				chatInfra.On("PushMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				messageRepo.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(usersMapRes, nil)
				var conversationId string
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(conversations []*domain.Conversation) bool {
					conversationId = conversations[0].ID.String
					expectedResp.Conversation.ConversationId = conversationId
					return len(conversations) == 1
				})).Once().Return(nil)
				// correct private conversation lesson
				privateConversationLessonRepo.On("Create", ctx, mock.Anything, mock.MatchedBy(func(privateConversation *lentities.PrivateConversationLesson) bool {
					flattenUserIds := "memberA_memberB"
					return privateConversation.ConversationID.String == conversationId &&
						flattenUserIds == privateConversation.FlattenUserIds.String &&
						privateConversation.LessonID.String == lessonID
				})).Once().Return(nil)
				// correct student membership inserted
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.MatchedBy(func(es []*domain.ConversationMembers) bool {
					if len(es) != 2 {
						return false
					}
					convUsers := make([]*tpb.Conversation_User, 0, 2)
					for _, e := range es {
						if e.UserID.String != teacherID && e.UserID.String != studentId {
							return false
						}
						if e.ConversationID.String != conversationId {
							return false
						}
						convUsers = append(convUsers, &tpb.Conversation_User{
							Id:        e.UserID.String,
							Group:     cpb.UserGroup(cpb.UserGroup_value[e.Role.String]),
							IsPresent: e.Status.String == domain.ConversationStatusActive,
							SeenAt:    timestamppb.New(e.SeenAt.Time),
						})
					}
					expectedResp.Conversation.Users = convUsers

					return true
				})).Once().Return(nil)

			},
		},
		{
			name:        "err upsert conversation",
			ctx:         ctx,
			req:         testCaseReq,
			expectedErr: txError,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(usersMapRes, nil)
				conversationRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "err create private conversation lesson",
			ctx:         ctx,
			req:         testCaseReq,
			expectedErr: txError,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(usersMapRes, nil)
				privateConversationLessonRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "err create conversation member",
			ctx:         ctx,
			req:         testCaseReq,
			expectedErr: txError,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(usersMapRes, nil)
				conversationMemberRepo.On("BulkUpsert", ctx, mock.Anything, mock.Anything).Return(pgx.ErrTxClosed)
			},
		},
		{
			name:        "err invalid memberships",
			ctx:         ctx,
			req:         invalidReq,
			expectedErr: errors.New("error invalid memberships"),
			setup: func(ctx context.Context) {
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(usersMapRes, nil)
			},
		},
		{
			name:        "not found user",
			ctx:         ctx,
			req:         testCaseReq,
			expectedErr: errors.New("error not found user"),
			setup: func(ctx context.Context) {
				usersRepo.On("FindByUserIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.CreateLiveLessonPrivateConversation(testCase.ctx, testCase.req.(*tpb.CreateLiveLessonPrivateConversationRequest))
			if err != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}
			if testCase.expectedResp != nil {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
