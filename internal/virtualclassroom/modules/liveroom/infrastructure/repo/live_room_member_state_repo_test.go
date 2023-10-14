package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LiveRoomMemberStateRepoWithSqlMock() (*LiveRoomMemberStateRepo, *testutil.MockDB) {
	r := &LiveRoomMemberStateRepo{}
	return r, testutil.NewMockDB()
}

func TestLiveRoomMemberStateRepo_GetLiveRoomMemberStatesWithParams(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dto := &LiveRoomMemberState{}
	fields, values := dto.FieldMap()
	params := &domain.SearchLiveRoomMemberStateParams{
		ChannelID: "channel-id1",
		UserIDs:   []string{"user-id1", "user-id2"},
		StateType: "state-type",
	}

	t.Run("failed to select", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)

		liveRoomMemberStates, err := mockRepo.GetLiveRoomMemberStatesWithParams(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, liveRoomMemberStates)
	})

	t.Run("success with select", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		liveRoomMemberStates, err := mockRepo.GetLiveRoomMemberStatesWithParams(ctx, mockDB.DB, params)
		assert.Nil(t, err)
		assert.NotNil(t, liveRoomMemberStates)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestLiveRoomMemberStateRepo_BulkUpsertLiveRoomMembersState(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	userIDs := []string{"user-id1", "user-id2"}
	stateType := vc_domain.LearnerStateTypeAnnotation
	state := &vc_domain.StateValue{
		BoolValue:        true,
		StringArrayValue: []string{},
	}

	t.Run("bulk upsert failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		err := mockRepo.BulkUpsertLiveRoomMembersState(ctx, mockDB.DB, channelID, userIDs, stateType, state)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("bulk upsert successful", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(userIDs); i++ {
			batchResults.On("Exec").Return(cmdTag, nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		err := mockRepo.BulkUpsertLiveRoomMembersState(ctx, mockDB.DB, channelID, userIDs, stateType, state)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}

func TestLiveRoomMemberStateRepo_UpdateAllLiveRoomMembersState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	stateType := vc_domain.LearnerStateTypeHandsUp
	state := &vc_domain.StateValue{
		BoolValue:        false,
		StringArrayValue: []string{},
	}

	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		genSliceMock(5)...)

	t.Run("upsert failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, pgx.ErrTxClosed)

		err := mockRepo.UpdateAllLiveRoomMembersState(ctx, mockDB.DB, channelID, stateType, state)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("upsert successful", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Once().Return(cmdTag, nil)

		err := mockRepo.UpdateAllLiveRoomMembersState(ctx, mockDB.DB, channelID, stateType, state)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestLiveRoomMemberStateRepo_CreateLiveRoomMemberState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channelID := "channel-id1"
	userID := "user-id1"
	stateType := vc_domain.LearnerStateTypeChat
	state := &vc_domain.StateValue{
		BoolValue: true,
	}

	dto := &LiveRoomMemberState{}
	fields := database.GetFieldNamesExcepts(dto, []string{"deleted_at"})
	args := append([]interface{}{mock.Anything, mock.Anything}, genSliceMock(len(fields))...)

	t.Run("failed", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)

		err := mockRepo.CreateLiveRoomMemberState(ctx, mockDB.DB, channelID, userID, stateType, state)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("successful", func(t *testing.T) {
		mockRepo, mockDB := LiveRoomMemberStateRepoWithSqlMock()
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := mockRepo.CreateLiveRoomMemberState(ctx, mockDB.DB, channelID, userID, stateType, state)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

}
