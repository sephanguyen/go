package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrganizationRepoWithSqlMock() (*OrganizationRepo, *testutil.MockDB) {
	repo := &OrganizationRepo{}
	return repo, testutil.NewMockDB()
}

func TestOrganizationRepo_FindByStatuses(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := OrganizationRepoWithSqlMock()
	billItem := &entities.Organization{}
	_, fieldMap := billItem.FieldMap()
	// scanFields := database.GetScanFields(billItem, fields)

	scanFields := []interface{}{}
	for range fieldMap {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.GetOrganizations(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.GetOrganizations(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.GetOrganizations(ctx, mockDB.DB)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestOrganizationRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := OrganizationRepoWithSqlMock()
	mockE := &entities.Organization{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		org, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, org)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		org, err := repo.FindByID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("db.QueryRowEx : err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, org)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		org, err := repo.FindByID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("db.QueryRowEx : err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, org)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
