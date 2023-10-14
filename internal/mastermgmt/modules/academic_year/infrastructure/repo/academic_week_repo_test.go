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

func AcademicWeekRepoWithSqlMock() (*AcademicWeekRepo, *testutil.MockDB) {
	academicWeekRepo := &AcademicWeekRepo{}
	return academicWeekRepo, testutil.NewMockDB()
}

func TestAcademicWeekRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	academicWeekRepo, mockDB := AcademicWeekRepoWithSqlMock()
	weeks := []*domain.AcademicWeek{
		{
			AcademicWeekID: "academic_week_id_1",
			WeekOrder:      1,
			Name:           "Week 1",
			StartDate:      now,
			EndDate:        now.Add(24 * 7 * time.Hour),
			Period:         "Term 1",
			AcademicYearID: "academic_year_id",
			LocationID:     "location_id",
			UpdatedAt:      now,
			CreatedAt:      now,
		},
		{
			AcademicWeekID: "academic_week_id_2",
			WeekOrder:      1,
			Name:           "Week 2",
			StartDate:      now.Add(24 * 8 * time.Hour),
			EndDate:        now.Add(24 * 7 * time.Hour * 3),
			Period:         "Term 1",
			AcademicYearID: "academic_year_id",
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
		err := academicWeekRepo.Insert(ctx, mockDB.DB, weeks)
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
		err := academicWeekRepo.Insert(ctx, mockDB.DB, weeks)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLessonRepo_GetLessonsWithSchedulerNull(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	academicWeekRepo, mockDB := AcademicWeekRepoWithSqlMock()
	weekDTO := &AcademicWeek{}
	fields, values := weekDTO.FieldMap()
	academicYearID := "academic_year_id"
	locationIDs := []string{"location_id_01", "location_id_02"}

	t.Run("err select", func(t *testing.T) {
		mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		resp, err := academicWeekRepo.GetAcademicWeeksByYearAndLocationIDs(ctx, mockDB.DB, academicYearID, locationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, resp)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &academicYearID, &locationIDs)
		mockDB.MockScanFields(nil, fields, values)

		resp, err := academicWeekRepo.GetAcademicWeeksByYearAndLocationIDs(ctx, mockDB.DB, academicYearID, locationIDs)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
