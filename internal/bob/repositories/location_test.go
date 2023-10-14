package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LocationRepoWithSqlMock() (*LocationRepo, *testutil.MockDB) {
	r := &LocationRepo{}
	return r, testutil.NewMockDB()
}

func TestLocationRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()

	locationID := database.Text("location-id")

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&locationID,
		)

		e := &entities_bob.Location{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)

		results, err := r.FindByID(ctx, mockDB.DB, locationID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			&locationID,
		)

		e := &entities_bob.Location{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		results, err := r.FindByID(ctx, mockDB.DB, locationID)
		assert.Nil(t, err)
		assert.Equal(t, e, results)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
