package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func AcademicClosedDayRepoWithSqlMock() (*AcademicClosedDayRepo, *testutil.MockDB) {
	academicClosedDayRepo := &AcademicClosedDayRepo{}
	return academicClosedDayRepo, testutil.NewMockDB()
}

func TestAcademicClosedDayRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	academicClosedDayRepo, mockDB := AcademicClosedDayRepoWithSqlMock()
	academicClosedDays := []*domain.AcademicClosedDay{
		{
			AcademicClosedDayID: "academic_closed_day_id_1",
			Date:                now,
			AcademicYearID:      "academic_year_id",
			AcademicWeekID:      "academic_week_id",
			LocationID:          "location_id",
			UpdatedAt:           now,
			CreatedAt:           now,
		},
		{
			AcademicClosedDayID: "academic_closed_day_id_2",
			Date:                now.Add(24 * time.Hour),
			AcademicYearID:      "academic_year_id",
			AcademicWeekID:      "academic_week_id",
			LocationID:          "location_id",
			UpdatedAt:           now,
			CreatedAt:           now,
		},
	}
	t.Run("success", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := academicClosedDayRepo.Insert(ctx, mockDB.DB, academicClosedDays)
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
		err := academicClosedDayRepo.Insert(ctx, mockDB.DB, academicClosedDays)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestAcademicClosedDayRepo_GetLessonsWithSchedulerNull(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	academicClosedDayRepo, mockDB := AcademicClosedDayRepoWithSqlMock()
	acd := &AcademicClosedDay{}
	fields, values := acd.FieldMap()
	weekIDs := []string{}

	t.Run("err select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		resp, err := academicClosedDayRepo.GetAcademicClosedDayByWeeks(ctx, mockDB.DB, weekIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &weekIDs)
		mockDB.MockScanFields(nil, fields, values)

		resp, err := academicClosedDayRepo.GetAcademicClosedDayByWeeks(ctx, mockDB.DB, weekIDs)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
