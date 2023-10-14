package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveRoomPollRepoWithSqlMock() (*LiveRoomPollRepo, *testutil.MockDB) {
	r := &LiveRoomPollRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveRoomPollRepo_CreateLiveRoomPoll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	liveRoomPoll := &domain.LiveRoomPoll{
		ChannelID: "channel-id1",
		StudentAnswers: domain.StudentAnswersList{
			{
				UserID:    "user-id1",
				Answers:   []string{"A", "B", "C"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				UserID:    "user-id2",
				Answers:   []string{"A", "B", "C"},
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				UserID:    "user-id3",
				Answers:   []string{"A", "B", "C"},
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		StoppedAt: &now,
		EndedAt:   &now,
		Options: &vc_domain.CurrentPollingOptions{
			{
				Answer:    "A",
				IsCorrect: true,
				Content:   "hello",
			},
			{
				Answer:    "B",
				IsCorrect: false,
				Content:   "hello",
			},
			{
				Answer:    "C",
				IsCorrect: false,
				Content:   "hello",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockLiveRoomPoll := &LiveRoomPoll{}
	fields := database.GetFieldNamesExcepts(mockLiveRoomPoll, []string{"deleted_at"})

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(len(fields))...)

	t.Run("insert failed", func(t *testing.T) {
		liveRoomPollRepo, mockDB := LiveRoomPollRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := liveRoomPollRepo.CreateLiveRoomPoll(ctx, mockDB.DB, liveRoomPoll)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert successful", func(t *testing.T) {
		liveRoomPollRepo, mockDB := LiveRoomPollRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := liveRoomPollRepo.CreateLiveRoomPoll(ctx, mockDB.DB, liveRoomPoll)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
