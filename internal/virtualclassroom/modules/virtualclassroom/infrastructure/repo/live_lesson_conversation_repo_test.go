package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveLessonConversationRepoWithSqlMock() (*LiveLessonConversationRepo, *testutil.MockDB) {
	r := &LiveLessonConversationRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveLessonConversationRepo_GetConversationByLessonIDAndConvType(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id1"
	conversationType := string(domain.LiveLessonConversationTypePrivate)
	mockDTO := &LiveLessonConversation{}
	fields, values := mockDTO.FieldMap()

	t.Run("successful", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, values)

		conversation, err := mockRepo.GetConversationByLessonIDAndConvType(ctx, mockDB.DB, lessonID, conversationType)
		assert.NoError(t, err)
		assert.NotNil(t, conversation)
	})

	t.Run("failed", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		conversation, err := mockRepo.GetConversationByLessonIDAndConvType(ctx, mockDB.DB, lessonID, conversationType)
		assert.True(t, errors.Is(err, domain.ErrNoConversationFound))
		assert.Equal(t, conversation, domain.LiveLessonConversation{})
	})
}

func TestLiveLessonConversationRepo_GetConversationIDByExactInfo(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonID := "lesson-id1"
	participants := []string{"user-id1", "user-id2"}
	conversationType := string(domain.LiveLessonConversationTypePrivate)
	mockDTO := &LiveLessonConversation{}
	fields, values := mockDTO.FieldMap()

	t.Run("successful", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, values)

		conversationID, err := mockRepo.GetConversationIDByExactInfo(ctx, mockDB.DB, lessonID, participants, conversationType)
		assert.NoError(t, err)
		assert.NotNil(t, conversationID)
	})

	t.Run("failed", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		conversationID, err := mockRepo.GetConversationIDByExactInfo(ctx, mockDB.DB, lessonID, participants, conversationType)
		assert.True(t, errors.Is(err, domain.ErrNoConversationFound))
		assert.Empty(t, conversationID)
	})
}

func TestLiveLessonConversationRepo_UpsertConversation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conversation := domain.LiveLessonConversation{
		ConversationID:   "conversation-id1",
		LessonID:         "lesson-id1",
		ParticipantList:  []string{"user-id1", "user-id2"},
		ConversationType: domain.LiveLessonConversationTypePrivate,
	}
	mockDTO := &LiveLessonConversation{}
	fields, _ := mockDTO.FieldMap()

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(len(fields))...)

	t.Run("insert failed", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := mockRepo.UpsertConversation(ctx, mockDB.DB, conversation)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.UpsertConversation(ctx, mockDB.DB, conversation)
		assert.Equal(t, err, domain.ErrNoConversationCreated)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert successful", func(t *testing.T) {
		mockRepo, mockDB := LiveLessonConversationRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.UpsertConversation(ctx, mockDB.DB, conversation)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
