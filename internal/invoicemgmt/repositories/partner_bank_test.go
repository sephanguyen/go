package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PartnerBankRepoWithSqlMock() (*PartnerBankRepo, *testutil.MockDB) {
	repo := &PartnerBankRepo{}
	return repo, testutil.NewMockDB()
}

func TestPartnerBankRepo_RetrievePartnerBankByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.PartnerBank{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - RetrievePartnerBankByID", func(t *testing.T) {
		repo, mockDB := PartnerBankRepoWithSqlMock()

		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.RetrievePartnerBankByID(ctx, mockDB.DB, mock.Anything)

		assert.Nil(t, err)
		assert.NotNil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test case - RetrievePartnerBankByID - ErrTxClosed", func(t *testing.T) {
		repo, mockDB := PartnerBankRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.RetrievePartnerBankByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - RetrievePartnerBankByID no rows affected", func(t *testing.T) {
		repo, mockDB := PartnerBankRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.RetrievePartnerBankByID(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPartnerBankRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := PartnerBankRepoWithSqlMock()

	partnerBank := &entities.PartnerBank{}
	_, partnerBankValues := partnerBank.FieldMap()
	argsPartnerBank := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(partnerBankValues))...,
	)
	internalErr := errors.New(" internal server error")
	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("Exec", argsPartnerBank...).Return(cmtTag, nil).Once()
			},
		},
		{
			name:      "error case fail to insert partner bank internal server error",
			expectErr: fmt.Errorf("err upsert partner bank: %w", internalErr),
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsPartnerBank...).Once().Return(cmtTag, internalErr)
			},
		},
		{
			name:      "error case row no rows affected",
			expectErr: fmt.Errorf("err upsert partner bank: %d RowsAffected", 0),
			setup: func() {
				cmtTag := pgconn.CommandTag(`0`)
				mockDB.DB.On("Exec", argsPartnerBank...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := repo.Upsert(ctx, mockDB.DB, partnerBank)
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}

func TestPartnerBankRepo_FindOne(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PartnerBankRepoWithSqlMock()
	mockE := &entities.PartnerBank{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindOne(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
