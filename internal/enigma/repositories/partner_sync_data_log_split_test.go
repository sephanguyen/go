package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func PartnerSyncDataLogSplitRepoWithSqlMock() (*PartnerSyncDataLogSplitRepo, *testutil.MockDB) {
	r := &PartnerSyncDataLogSplitRepo{}
	return r, testutil.NewMockDB()
}

func TestPartnerSyncDataLogSplitRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PartnerSyncDataLogSplitRepoWithSqlMock()

	t.Run("err create", func(t *testing.T) {
		partnerLog := &entities.PartnerSyncDataLogSplit{}
		_, values := partnerLog.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, partnerLog)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.PartnerSyncDataLogSplit{}
		fields, values := e.FieldMap()

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestPartnerSyncDataLogSplitRepo_UpdateLogStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PartnerSyncDataLogSplitRepoWithSqlMock()

	logID := database.Text("mock-log-id")
	t.Run("err update", func(t *testing.T) {
		partnerLog := &entities.PartnerSyncDataLogSplit{
			PartnerSyncDataLogSplitID: logID,
			Status:                    database.Text(string(entities.StatusProcessing)),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.Text(string(entities.StatusProcessing)), logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := repo.UpdateLogStatus(ctx, mockDB.DB, partnerLog)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("no row affected", func(t *testing.T) {
		partnerLog := &entities.PartnerSyncDataLogSplit{
			PartnerSyncDataLogSplitID: logID,
			Status:                    database.Text(string(entities.StatusProcessing)),
		}
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.Text(string(entities.StatusProcessing)), logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := repo.UpdateLogStatus(ctx, mockDB.DB, partnerLog)
		assert.Equal(t, err, fmt.Errorf("no rows affected"))
	})

	t.Run("success", func(t *testing.T) {
		partnerLog := &entities.PartnerSyncDataLogSplit{
			PartnerSyncDataLogSplitID: logID,
			Status:                    database.Text(string(entities.StatusSuccess)),
		}
		fields := []string{"status", "updated_at"}

		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.Text(string(entities.StatusSuccess)), logID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.UpdateLogStatus(ctx, mockDB.DB, partnerLog)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, partnerLog.TableName())
		mockDB.RawStmt.AssertUpdatedFields(t, fields...)
	})
}

func TestPartnerSyncDataLogSplitRepo_GetLogsBySignature(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PartnerSyncDataLogSplitRepoWithSqlMock()

	signature := database.Text("mock-signature")
	t.Run("err", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &signature)
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, args...)

		logs, err := r.GetLogsBySignature(ctx, mockDB.DB, signature)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, logs)
	})

	t.Run("success", func(t *testing.T) {
		id1 := database.Text(idutil.ULIDNow())
		id2 := database.Text(idutil.ULIDNow())
		ids := []pgtype.Text{id1, id2}
		status := database.Text(string(entities.StatusProcessing))
		updatedAt := pgtype.Timestamp{}
		updatedAt.Set(time.Now())
		log := entities.PartnerSyncDataLogSplit{}

		fields, _ := log.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			{
				&id1, &status, &updatedAt,
			},
			{
				&id2, &status, &updatedAt,
			},
		})
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &signature)
		mockDB.MockQueryArgs(t, nil, args...)

		logs, err := r.GetLogsBySignature(ctx, mockDB.DB, signature)
		assert.NoError(t, err)
		assert.Equal(t, len(logs), len(ids))
	})
}

func TestPartnerSyncDataLogSplitRepo_GetLogsReportByDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := PartnerSyncDataLogSplitRepoWithSqlMock()
	now := time.Now()
	from, to := pgtype.Date{Time: now}, pgtype.Date{Time: now}
	t.Run("err", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &from, &to)
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, args...)

		logs, err := r.GetLogsReportByDate(ctx, mockDB.DB, from, to)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, logs)
	})

	t.Run("success", func(t *testing.T) {
		total := pgtype.Int8{Int: 1}
		status := database.Text(string(entities.StatusProcessing))
		createdAt := pgtype.Date{}
		createdAt.Set(time.Now())

		mockDB.MockScanArray(nil, []string{"total", "status", "created_at"}, [][]interface{}{
			{
				&total, &status, &createdAt,
			},
			{
				&total, &status, &createdAt,
			},
		})
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &from, &to)
		mockDB.MockQueryArgs(t, nil, args...)

		reports, err := r.GetLogsReportByDate(ctx, mockDB.DB, from, to)
		assert.NoError(t, err)
		assert.Equal(t, len(reports), 2)
	})
}

func TestPartnerSyncDataLogSplitRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	p, mockDB := PartnerSyncDataLogSplitRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		logs := []*entities.PartnerSyncDataLogSplit{
			{
				PartnerSyncDataLogSplitID: database.Text("1"),
				Status:                    database.Text(string(entities.StatusProcessing)),
				RetryTimes:                pgtype.Int4{Int: 2},
				CreatedAt:                 pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
		}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := p.UpdateLogsStatusAndRetryTime(ctx, mockDB.DB, logs)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		logs := []*entities.PartnerSyncDataLogSplit{
			{
				PartnerSyncDataLogSplitID: database.Text("1"),
				Status:                    database.Text(string(entities.StatusProcessing)),
				RetryTimes:                pgtype.Int4{Int: 2},
				CreatedAt:                 pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
			},
		}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := p.UpdateLogsStatusAndRetryTime(ctx, mockDB.DB, logs)
		require.NotEqual(t, err, 1)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
