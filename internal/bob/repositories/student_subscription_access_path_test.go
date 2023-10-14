package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func StudentSubscriptionAccessPathRepoWithSqlMock() (*StudentSubscriptionAccessPathRepo, *testutil.MockDB) {
	r := &StudentSubscriptionAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentSubscriptionAccessPathRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := StudentSubscriptionAccessPathRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		sub1 := &entities_bob.StudentSubscriptionAccessPath{
			LocationID:            database.Text("location-1"),
			StudentSubscriptionID: database.Text("sub-1"),
			CreatedAt:             database.Timestamptz(time.Now()),
			UpdatedAt:             database.Timestamptz(time.Now()),
		}
		sub2 := &entities_bob.StudentSubscriptionAccessPath{
			LocationID:            database.Text("location-2"),
			StudentSubscriptionID: database.Text("sub-2"),
			CreatedAt:             database.Timestamptz(time.Now()),
			UpdatedAt:             database.Timestamptz(time.Now()),
		}
		caps := []*entities_bob.StudentSubscriptionAccessPath{sub1, sub2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Upsert(ctx, mockDB.DB, caps)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		sub1 := &entities_bob.StudentSubscriptionAccessPath{
			LocationID:            database.Text("location-1"),
			StudentSubscriptionID: database.Text("sub-1"),
			CreatedAt:             database.Timestamptz(time.Now()),
			UpdatedAt:             database.Timestamptz(time.Now()),
		}
		sub2 := &entities_bob.StudentSubscriptionAccessPath{
			LocationID:            database.Text("location-2"),
			StudentSubscriptionID: database.Text("sub-2"),
			CreatedAt:             database.Timestamptz(time.Now()),
			UpdatedAt:             database.Timestamptz(time.Now()),
		}
		caps := []*entities_bob.StudentSubscriptionAccessPath{sub1, sub2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		r.Upsert(ctx, mockDB.DB, caps)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
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

		e := &entities_bob.StudentSubscriptionAccessPath{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		result, err := r.FindStudentSubscriptionIDsByLocationIDs(ctx, mockDB.DB, locationIds)
		assert.Nil(t, err)
		assert.Equal(t, "", result[0])
	})
}
