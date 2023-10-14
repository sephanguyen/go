package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocationRepo_GetLowestGrantedLocationsByUserIDAndPermissions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &LocationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	userID := "user-id"
	permission := []string{"permission-1", "permission-2"}
	locationIDsTextArrayReq := []pgtype.Text{database.Text("location-1"), database.Text("location-2")}
	locationAccessPathsTextArrayReq := []pgtype.Text{database.Text("location-access-path-mock-1"), database.Text("location-access-path-mock-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, userID, permission)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		_, _, err := r.GetLowestGrantedLocationsByUserIDAndPermissions(ctx, db, userID, permission)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, userID, permission)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		locationIDs, mapLocationAccessPath, err := r.GetLowestGrantedLocationsByUserIDAndPermissions(ctx, db, userID, permission)
		assert.NoError(t, err)
		for i, locationID := range locationIDs {
			assert.Equal(t, mapLocationAccessPath[locationID], locationAccessPathsTextArrayReq[i].String)
		}
	})
}

func TestLocationRepo_GetGrantedLocationsByUserIDAndPermissions(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &LocationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	userID := "user-id"
	permission := []string{"permission-1", "permission-2"}
	locationIDsTextArrayReq := []pgtype.Text{database.Text("location-1"), database.Text("location-2")}
	locationAccessPathsTextArrayReq := []pgtype.Text{database.Text("location-access-path-mock-1"), database.Text("location-access-path-mock-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, userID, permission)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		_, _, err := r.GetGrantedLocationsByUserIDAndPermissions(ctx, db, userID, permission)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, userID, permission)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		locationIDs, mapLocationAccessPath, err := r.GetGrantedLocationsByUserIDAndPermissions(ctx, db, userID, permission)
		assert.NoError(t, err)
		for i, locationID := range locationIDs {
			assert.Equal(t, mapLocationAccessPath[locationID], locationAccessPathsTextArrayReq[i].String)
		}
	})
}

func TestLocationRepo_GetLocationAccessPathsByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &LocationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	locationIDs := []string{"location-1", "location-2"}
	locationIDsTextArrayReq := []pgtype.Text{database.Text("location-1"), database.Text("location-2")}
	locationAccessPathsTextArrayReq := []pgtype.Text{database.Text("location-access-path-mock-1"), database.Text("location-access-path-mock-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, locationIDs)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		_, err := r.GetLocationAccessPathsByIDs(ctx, db, locationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIDs)
		mockDB.MockScanArray(nil, []string{"location_id", "access_path"}, [][]interface{}{{&locationIDsTextArrayReq[0], &locationAccessPathsTextArrayReq[0]}, {&locationIDsTextArrayReq[1], &locationAccessPathsTextArrayReq[1]}})

		mapLocationAccessPath, err := r.GetLocationAccessPathsByIDs(ctx, db, locationIDs)
		assert.NoError(t, err)
		for i, locationID := range locationIDs {
			assert.Equal(t, mapLocationAccessPath[locationID], locationAccessPathsTextArrayReq[i].String)
		}
	})
}

func TestLocationRepo_GetLowestLocationIDsByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &LocationRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	locationIDs := []string{"location-1", "location-2"}
	locationIDsTextArrayReq := []pgtype.Text{database.Text("location-1"), database.Text("location-2")}

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, locationIDs)

		_, err := r.GetLowestLocationIDsByIDs(ctx, db, locationIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIDs)
		mockDB.MockScanArray(nil, []string{"location_id"}, [][]interface{}{{&locationIDsTextArrayReq[0]}, {&locationIDsTextArrayReq[1]}})

		locIDs, err := r.GetLowestLocationIDsByIDs(ctx, db, locationIDs)
		assert.NoError(t, err)
		assert.Equal(t, locIDs, locationIDs)
	})
}
