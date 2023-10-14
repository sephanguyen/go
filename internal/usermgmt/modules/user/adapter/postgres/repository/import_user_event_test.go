package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ImportUserEventWithSqlMock() (*ImportUserEventRepo, *testutil.MockDB) {
	r := &ImportUserEventRepo{}
	return r, testutil.NewMockDB()
}

func TestImportUserEvent_Upsert(t *testing.T) {
	t.Parallel()

	t.Run("failed to Upsert import_user_event", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := ImportUserEventWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Once().Return(mockDB.Row)
		batchResults.On("Close").Once().Return(nil)

		usrEmail := &entity.ImportUserEvent{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()
		importUserEvents := []*entity.ImportUserEvent{
			{
				UserID: database.Text(idutil.ULIDNow()),
			},
			{
				UserID: database.Text(idutil.ULIDNow()),
			},
		}

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		_, err := r.Upsert(ctx, mockDB.DB, importUserEvents)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("Upsert import_user_event successfully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, mockDB := ImportUserEventWithSqlMock()

		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("QueryRow").Twice().Return(mockDB.Row)
		batchResults.On("Close").Once().Return(nil)

		usrEmail := &entity.ImportUserEvent{}
		database.AllNullEntity(usrEmail)
		fields, values := usrEmail.FieldMap()
		importUserEvents := []*entity.ImportUserEvent{
			{
				UserID: database.Text(idutil.ULIDNow()),
			},
			{
				UserID: database.Text(idutil.ULIDNow()),
			},
		}

		mockDB.MockRowScanFields(nil, fields, values)
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := r.Upsert(ctx, mockDB.DB, importUserEvents)

		assert.Nil(t, err)
	})
}

func TestImportUserEvent_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ImportUserEventWithSqlMock()
	ids := database.Int8Array([]int64{ksuid.Max.Time().Unix(), ksuid.Max.Time().Unix(), ksuid.Max.Time().Unix()})

	mockE := &entity.ImportUserEvent{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("err select", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), ids).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		users, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, users)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		latestRecords, err := r.GetByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []*entity.ImportUserEvent{mockE}, latestRecords)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

}
