package repo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentSubscriptionAccessPathRepoWithSqlMock() (*StudentSubscriptionAccessPathRepo, *testutil.MockDB) {
	r := &StudentSubscriptionAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentSubscriptionAccessPathRepo_FindStudentSubscriptionIDsByLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
	locationIds := []string{"location-id-1", "location-id-2"}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.AnythingOfType("string"),
			&locationIds,
		)
		_, err := r.FindStudentSubscriptionIDsByLocationIDs(ctx, mockDB.DB, locationIds)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t,
			nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			&locationIds,
		)

		e := &StudentSubscriptionAccessPath{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		result, err := r.FindStudentSubscriptionIDsByLocationIDs(ctx, mockDB.DB, locationIds)
		assert.Nil(t, err)
		assert.Equal(t, "", result[0])
	})
}

func TestStudentSubscriptionAccessPathRepo_FindLocationsByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		subs, err := r.FindLocationsByStudentSubscriptionIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, subs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &StudentSubscriptionAccessPath{}
		fields, values := e.FieldMap()
		_ = e.StudentSubscriptionID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		subLocations, err := r.FindLocationsByStudentSubscriptionIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, map[string][]string{
			e.StudentSubscriptionID.String: {e.LocationID.String},
		}, subLocations)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":              {HasNullTest: true},
			"student_subscription_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestStudentSubscriptionAccessPathRepo_BulkUpsertStudentSubscriptionAccessPath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sampleData := domain.StudentSubscriptionAccessPaths{
		{
			SubscriptionID: "id-1",
			LocationID:     "loc-1",
		},
		{
			SubscriptionID: "id-2",
			LocationID:     "loc-2",
		},
		{
			SubscriptionID: "id-3",
			LocationID:     "loc-3",
		},
	}

	t.Run("err bulk upsert", func(t *testing.T) {
		r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, errors.New("error")).Once()
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscriptionAccessPath(ctx, mockDB.DB, sampleData)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("no rows affected after upsert", func(t *testing.T) {
		r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		batchResults.On("Exec").Return(cmdTag, nil).Once()
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscriptionAccessPath(ctx, mockDB.DB, sampleData)

		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("bulk upsert successful", func(t *testing.T) {
		r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).
			Once().
			Return(batchResults, nil)
		for i := 0; i < len(sampleData); i++ {
			batchResults.On("Exec").Return(cmdTag, nil).Once()
		}
		batchResults.On("Close").Once().Return(nil)

		err := r.BulkUpsertStudentSubscriptionAccessPath(ctx, mockDB.DB, sampleData)

		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}

func TestStudentSubscriptionAccessPathRepo_DeleteByStudentSubscriptionIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sampleData := []string{
		"sample-id-1",
		"sample-id-2",
		"sample-id-3",
	}

	t.Run("err delete", func(t *testing.T) {
		r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, sampleData).Return(cmdTag, errors.New("error")).Once()

		err := r.DeleteByStudentSubscriptionIDs(ctx, mockDB.DB, sampleData)
		assert.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("delete successful", func(t *testing.T) {
		r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, sampleData).Return(cmdTag, nil).Once()

		err := r.DeleteByStudentSubscriptionIDs(ctx, mockDB.DB, sampleData)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
