package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

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

	locationIDs := []string{"location-1", "location-2"}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			locationIDs,
		)

		// fields, values := e.FieldMap()
		scannedValues := database.TextArray([]string{"location-accesspath"})
		mockDB.MockRowScanFields(pgx.ErrNoRows, []string{""}, []interface{}{&scannedValues})

		results, err := r.FindAccessPaths(ctx, mockDB.DB, locationIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, []string{}, results)
	})

	t.Run("scan field row success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.AnythingOfType("string"),
			locationIDs,
		)

		// fields, values := e.FieldMap()
		scannedValues := database.TextArray([]string{"location-accesspath"})
		mockDB.MockRowScanFields(nil, []string{""}, []interface{}{&scannedValues})

		results, err := r.FindAccessPaths(ctx, mockDB.DB, locationIDs)
		assert.NoError(t, err)
		assert.Equal(t, []string{"location-accesspath"}, results)
	})
}
