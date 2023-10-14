package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LocationRepoWithSqlMock() (*LocationRepo, *testutil.MockDB) {
	locationRepo := &LocationRepo{}
	return locationRepo, testutil.NewMockDB()
}

func TestLocationRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	locationRepoWithSqlMock, mockDB := LocationRepoWithSqlMock()

	const locationID = "testID"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			locationID,
		)
		entities := &entities.Location{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		location, err := locationRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, locationID)
		assert.Nil(t, err)
		assert.NotNil(t, location)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			locationID,
		)
		entities := &entities.Location{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		location, err := locationRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, locationID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, location)

	})
}

func TestLocationRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	locationRepoWithSqlMock, mockDB := LocationRepoWithSqlMock()

	locationIDs := []string{"10", "20", "30"}
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			locationIDs,
		)
		e := &entities.Location{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		discount, err := locationRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, locationIDs)
		assert.Nil(t, err)
		assert.NotNil(t, discount)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			locationIDs,
		)
		e := &entities.Location{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		discount, err := locationRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, locationIDs)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, discount)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, locationIDs)
		discount, err := locationRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, locationIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discount)
	})
}

func TestLocationRepo_GetLowestGrantedLocationIDsByUserIDAndPermissions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &LocationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	params := GetGrantedLowestLevelLocationsParams{
		Limit:           30,
		UserID:          "user-id",
		PermissionNames: []string{"permission-1", "permission-2"},
	}
	locationIDsTextArrayReq := []pgtype.Text{database.Text("location-1"), database.Text("location-2")}
	locationAccessPathsTextArrayReq := []pgtype.Text{database.Text("location-access-path-mock-1"), database.Text("location-access-path-mock-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, params.UserID, params.PermissionNames, params.Name)

		_, err := r.GetLowestGrantedLocationIDsByUserIDAndPermissions(ctx, db, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.UserID, params.PermissionNames, params.Name)
		mockDB.MockScanArray(nil, []string{"location_id"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		locationIDs, err := r.GetLowestGrantedLocationIDsByUserIDAndPermissions(ctx, db, params)
		assert.NoError(t, err)
		assert.NotNil(t, locationIDs)
	})
}
