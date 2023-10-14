package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func SchedulerRepoWithSqlMock() (*SchedulerRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	schedulerRepo := &SchedulerRepo{}
	return schedulerRepo, mockDB
}

func TestSchedulerRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schedulerRepo, mockDB := SchedulerRepoWithSqlMock()
	scheduler := &dto.CreateSchedulerParams{
		SchedulerID: "scheduler-id",
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 1, 2),
		Frequency:   "weekly",
	}
	dto, err := NewScheduler(map[string]interface{}{
		"scheduler_id": scheduler.SchedulerID,
		"start_date":   scheduler.StartDate,
		"end_date":     scheduler.EndDate,
		"frequency":    scheduler.Frequency,
	})
	_, values := dto.FieldMap()
	dto.PreInsert()
	require.NoError(t, err)
	args := append([]interface{}{
		mock.Anything,
		mock.AnythingOfType("string")},
		values[0], values[1], values[2], values[3], mock.Anything, mock.Anything, mock.Anything)
	t.Run("error", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), pgx.ErrTxClosed, args...)
		schedulerID, err := schedulerRepo.Create(ctx, mockDB.DB, scheduler)
		require.ErrorIs(t, err, pgx.ErrTxClosed)
		require.Empty(t, schedulerID)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		schedulerID, err := schedulerRepo.Create(ctx, mockDB.DB, scheduler)
		require.NoError(t, err)
		require.Equal(t, scheduler.SchedulerID, schedulerID)
	})
}

func TestSchedulerRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schedulerRepo, mockDB := SchedulerRepoWithSqlMock()
	e := &Scheduler{}
	fields, value := e.FieldMap()
	schedulerID := idutil.ULIDNow()
	t.Run("error no row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, value)
		gotClass, err := schedulerRepo.GetByID(ctx, mockDB.DB, schedulerID)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Nil(t, gotClass)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, fields, value)
		gotClass, err := schedulerRepo.GetByID(ctx, mockDB.DB, schedulerID)
		assert.NoError(t, err)
		assert.NotNil(t, gotClass)
	})
}

func TestSchedulerRepo_CreateMany(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	schedulerRepo, mockDB := SchedulerRepoWithSqlMock()
	params := []*dto.CreateSchedulerParamWithIdentity{
		{
			ID: "lesson_id_01",
			CreateSchedulerParam: dto.CreateSchedulerParams{
				SchedulerID: "scheduler_id_01",
				StartDate:   now,
				EndDate:     now.Add(1 * time.Hour),
			},
		},
		{
			ID: "lesson_id_02",
			CreateSchedulerParam: dto.CreateSchedulerParams{
				SchedulerID: "scheduler_id_02",
				StartDate:   now,
				EndDate:     now.Add(1 * time.Hour),
			},
		},
		{
			ID: "lesson_id_03",
			CreateSchedulerParam: dto.CreateSchedulerParams{
				SchedulerID: "scheduler_id_03",
				StartDate:   now,
				EndDate:     now.Add(1 * time.Hour),
			},
		},
	}
	row := &mock_database.Row{}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Times(3).Return(row)
		row.On("Scan", mock.Anything, mock.Anything).Times(3).Return(nil)
		batchResults.On("Close").Once().Return(nil)
		resp, err := schedulerRepo.CreateMany(ctx, mockDB.DB, params)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Times(3).Return(row)
		row.On("Scan", mock.Anything, mock.Anything).Times(3).Return(fmt.Errorf("error"))
		batchResults.On("Close").Once().Return(nil)
		resp, err := schedulerRepo.CreateMany(ctx, mockDB.DB, params)
		require.Error(t, err)
		require.Empty(t, resp)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
