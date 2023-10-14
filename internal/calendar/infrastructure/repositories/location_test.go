package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LocationRepoWithSqlMock() (*LocationRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()

	mockRepo := &LocationRepo{}
	return mockRepo, mockDB
}

func TestLocationRepo_GetLocationByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockLocation := &Location{}
	fields, values := mockLocation.FieldMap()

	sampleID := "sample"

	t.Run("fetch location by ID failed", func(t *testing.T) {
		mockLocationRepo, mockDB := LocationRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, sampleID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		location, err := mockLocationRepo.GetLocationByID(ctx, mockDB.DB, sampleID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, location)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("fetch location by ID successful", func(t *testing.T) {
		mockLocationRepo, mockDB := LocationRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, sampleID)
		mockDB.MockRowScanFields(nil, fields, values)

		location, err := mockLocationRepo.GetLocationByID(ctx, mockDB.DB, sampleID)

		assert.Nil(t, err)
		assert.NotNil(t, location)
		assert.IsType(t, &dto.Location{}, location)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}
