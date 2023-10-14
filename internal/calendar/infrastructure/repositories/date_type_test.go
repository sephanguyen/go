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

func DateTypeRepoWithSqlMock() (*DateTypeRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()

	mockRepo := &DateTypeRepo{}
	return mockRepo, mockDB
}

func TestDateTypeRepo_GetDateTypeByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDateType := &DateType{}
	fields, values := mockDateType.FieldMap()

	sampleID := "sample"

	t.Run("get date type by ID failed", func(t *testing.T) {
		mockDateTypeRepo, mockDB := DateTypeRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, sampleID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		dateType, err := mockDateTypeRepo.GetDateTypeByID(ctx, mockDB.DB, sampleID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, dateType)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("get date type by ID successful", func(t *testing.T) {
		mockDateTypeRepo, mockDB := DateTypeRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, sampleID)
		mockDB.MockRowScanFields(nil, fields, values)

		dateType, err := mockDateTypeRepo.GetDateTypeByID(ctx, mockDB.DB, sampleID)

		assert.Nil(t, err)
		assert.NotNil(t, dateType)
		assert.IsType(t, &dto.DateType{}, dateType)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestDateTypeRepo_GetDateTypeByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDateType := &DateType{}
	fields, values := mockDateType.FieldMap()

	sampleID := []string{"regular", "closed"}

	t.Run("fetch date type by IDs failed", func(t *testing.T) {
		mockDateTypeRepo, mockDB := DateTypeRepoWithSqlMock()

		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, sampleID)
		mockDB.MockScanFields(pgx.ErrNoRows, fields, values)

		dateType, err := mockDateTypeRepo.GetDateTypeByIDs(ctx, mockDB.DB, sampleID)

		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, dateType)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("fetch date type by IDs successful", func(t *testing.T) {
		mockDateTypeRepo, mockDB := DateTypeRepoWithSqlMock()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, sampleID)
		mockDB.MockScanFields(nil, fields, values)

		dateType, err := mockDateTypeRepo.GetDateTypeByIDs(ctx, mockDB.DB, sampleID)

		assert.Nil(t, err)
		assert.NotNil(t, dateType)
		assert.IsType(t, []*dto.DateType{}, dateType)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
