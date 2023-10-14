package migration

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	lentities "github.com/manabie-com/backend/internal/tom/domain/lesson"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResourcePathMigrator_MigrateUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}

	deviceTokenRepo := new(mock_repositories.MockUserDeviceTokenRepo)

	s := &ResourcePathMigrator{
		DB:                  mockDB,
		UserDeviceTokenRepo: deviceTokenRepo,
	}
	schoolID := "manabie"
	userIDs := []string{"user-1", "user-2"}

	testCases := map[string]TestCase{
		"success": {
			ctx: ctx,
			req: &tpb.ResourcePathMigration_Users{
				SchoolId: schoolID,
				UserIds:  userIDs,
			},
			setup: func(ctx context.Context) {
				deviceTokenRepo.On("BulkUpdateResourcePath", mock.Anything, mock.Anything, userIDs, schoolID).Once().Return(nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.MigrateUser(testCase.ctx, testCase.req.(*tpb.ResourcePathMigration_Users))
			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			}
		})
	}
}

func TestResourcePathMigrator_MigrateLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationLesson := new(mock_repositories.MockConversationLessonRepo)

	s := &ResourcePathMigrator{
		DB:                     mockDB,
		ConversationLessonRepo: conversationLesson,
		ConversationRepo:       conversationRepo,
	}
	schoolID := "manabie"
	lessonIDs := []string{"lesson-1", "lesson-2"}
	convs := []string{"conversation-1", "conversation-2"}
	lessonEntities := []*lentities.ConversationLesson{
		{
			LessonID:       database.Text("lesson-1"),
			ConversationID: database.Text("conversation-1"),
		},
		{
			LessonID:       database.Text("lesson-2"),
			ConversationID: database.Text("conversation-2"),
		},
	}

	testCases := map[string]TestCase{
		"success": {
			ctx: ctx,
			req: &tpb.ResourcePathMigration_Lessons{
				SchoolId:  schoolID,
				LessonIds: lessonIDs,
			},
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Commit", ctx).Once().Return(nil)
				conversationLesson.On("BulkUpdateResourcePath", mock.Anything, tx, lessonIDs, schoolID).Once().Return(nil)
				conversationLesson.On("FindByLessonIDs", mock.Anything, tx, database.TextArray(lessonIDs), true).Once().Return(lessonEntities, nil)
				conversationRepo.On("BulkUpdateResourcePath", ctx, tx, convs, schoolID).Once().Return(nil)
			},
		},
		"tx error": {
			ctx: ctx,
			req: &tpb.ResourcePathMigration_Lessons{
				SchoolId:  schoolID,
				LessonIds: lessonIDs,
			},
			expectedResp: nil,
			expectedErr:  pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				mockDB.On("Begin", ctx).Once().Return(tx, nil)
				tx.On("Rollback", ctx).Once().Return(nil)
				conversationLesson.On("BulkUpdateResourcePath", mock.Anything, tx, lessonIDs, schoolID).Once().Return(nil)
				conversationLesson.On("FindByLessonIDs", mock.Anything, tx, database.TextArray(lessonIDs), true).Once().Return(lessonEntities, nil)
				conversationRepo.On("BulkUpdateResourcePath", ctx, tx, convs, schoolID).Once().Return(pgx.ErrTxClosed)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.MigrateLesson(testCase.ctx, testCase.req.(*tpb.ResourcePathMigration_Lessons))

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			}
		})
	}
}

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	customCtx    func(context.Context) context.Context
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}
