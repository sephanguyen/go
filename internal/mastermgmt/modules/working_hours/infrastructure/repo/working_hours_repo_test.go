package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func WorkingHoursRepoWithSqlMock() (*WorkingHoursRepo, *testutil.MockDB) {
	workingHoursRepo := &WorkingHoursRepo{}
	return workingHoursRepo, testutil.NewMockDB()
}

func TestWorkingHoursRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	workingHoursRepo, mockDB := WorkingHoursRepoWithSqlMock()
	locationIDs := []string{"location-id-1,location-id-2"}
	workingHours := []*domain.WorkingHours{
		{
			WorkingHoursID: "working_hours_id_1",
			Day:            "Monday",
			OpeningTime:    "08:00",
			ClosingTime:    "17:00",
			LocationID:     "location_id",
			UpdatedAt:      now,
			CreatedAt:      now,
		},
		{
			WorkingHoursID: "working_hours_id_2",
			Day:            "Tuesday",
			OpeningTime:    "09:00",
			ClosingTime:    "18:00",
			LocationID:     "location_id",
			UpdatedAt:      now,
			CreatedAt:      now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := workingHoursRepo.Upsert(ctx, mockDB.DB, workingHours, locationIDs)
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
		err := workingHoursRepo.Upsert(ctx, mockDB.DB, workingHours, locationIDs)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
