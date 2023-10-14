package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TimeSlotRepoWithSqlMock() (*TimeSlotRepo, *testutil.MockDB) {
	timeSlotRepo := &TimeSlotRepo{}
	return timeSlotRepo, testutil.NewMockDB()
}

func TestTimeSlotRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	timeSlotRepo, mockDB := TimeSlotRepoWithSqlMock()
	locationIDs := []string{"location-id-1,location-id-2"}
	timeSlot := []*domain.TimeSlot{
		{
			TimeSlotID:         "time_slot_01",
			TimeSlotInternalID: "1",
			StartTime:          "11:00",
			EndTime:            "13:00",
			LocationID:         "location_id",
			UpdatedAt:          now,
			CreatedAt:          now,
		},
		{
			TimeSlotID:         "time_slot_02",
			TimeSlotInternalID: "2",
			StartTime:          "14:00",
			EndTime:            "16:00",
			LocationID:         "location_id",
			UpdatedAt:          now,
			CreatedAt:          now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := timeSlotRepo.Upsert(ctx, mockDB.DB, timeSlot, locationIDs)
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)
		err := timeSlotRepo.Upsert(ctx, mockDB.DB, timeSlot, locationIDs)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
