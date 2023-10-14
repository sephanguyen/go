package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LocationRepoWithSqlMock() (*LocationRepo, *testutil.MockDB) {
	r := &LocationRepo{}
	return r, testutil.NewMockDB()
}

func mockLocations() []*domain.Location {
	now := time.Now()
	location1 := &domain.Location{
		LocationID:              "location-1",
		PartnerInternalID:       "partner-1",
		PartnerInternalParentID: "partner-2",
		Name:                    "lesson name",
		LocationType:            "center",
		CreatedAt:               now,
		UpdatedAt:               now,
		ParentLocationID:        "",
		DeletedAt:               &now,
	}
	location2 := &domain.Location{
		LocationID:              "location-2",
		PartnerInternalID:       "partner-2",
		PartnerInternalParentID: "partner-2",
		Name:                    "lesson name",
		LocationType:            "center",
		CreatedAt:               now,
		UpdatedAt:               now,
		ParentLocationID:        "",
		DeletedAt:               &now,
	}
	locations := []*domain.Location{location1, location2}
	return locations
}

func TestLocationRepo_UpsertLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	e := &Location{}
	selectFields, values := e.FieldMap()

	t.Run("successfully", func(t *testing.T) {
		r, mockDB := LocationRepoWithSqlMock()
		locations := mockLocations()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		e := &Location{}
		_ = e.LocationID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			values,
		})

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Return(batchResults)
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Return(nil)

		errs := r.UpsertLocations(ctx, mockDB.DB, locations)
		require.Empty(t, errs)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		r, mockDB := LocationRepoWithSqlMock()
		locations := mockLocations()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		e := &Location{}
		_ = e.LocationID.Set("id")

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			values,
		})

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		errs := r.UpsertLocations(ctx, mockDB.DB, locations)
		require.NotEmpty(t, errs)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLocationRepo_GetLocationsByPartnerInternalIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		locations, err := r.GetLocationsByPartnerInternalIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &Location{}
		fields, values := e.FieldMap()
		_ = e.PartnerInternalID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationsByPartnerInternalIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"partner_internal_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLocationRepo_GetLocationByPartnerInternalID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	partnerInternalID := "id"
	e := &Location{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerInternalID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		location, err := r.GetLocationByPartnerInternalID(ctx, mockDB.DB, partnerInternalID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, location)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerInternalID)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationByPartnerInternalID(ctx, mockDB.DB, partnerInternalID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLocationRepo_GetLocationByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	partnerInternalID := "id"
	e := &Location{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerInternalID)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		location, err := r.GetLocationByID(ctx, mockDB.DB, partnerInternalID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, location)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &partnerInternalID)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationByID(ctx, mockDB.DB, partnerInternalID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLocationRepo_GetLocationsByLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		locations, err := r.GetLocationsByLocationIDs(ctx, mockDB.DB, ids, true)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select without deleted", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := new(Location)
		fields, values := e.FieldMap()
		_ = e.PartnerInternalID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationsByLocationIDs(ctx, mockDB.DB, ids, false)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":  {HasNullTest: true},
			"location_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})

	t.Run("success with select with deleted", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &Location{}
		fields, values := e.FieldMap()
		_ = e.PartnerInternalID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationsByLocationIDs(ctx, mockDB.DB, ids, true)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"location_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLocationRepo_RetrieveLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationRepoWithSqlMock()
	e := &Location{}
	fields, value := e.FieldMap()
	filter := domain.FilterLocation{
		IncludeIsArchived: false,
		UserID:            "user-1",
	}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		gotLocations, err := r.RetrieveLocations(ctx, mockDB.DB, filter)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		gotLocations, err := r.RetrieveLocations(ctx, mockDB.DB, filter)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_GetLocationByLocationTypeName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	typeName := "center"
	e := &Location{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &typeName)
		gotLocations, err := r.GetLocationByLocationTypeName(ctx, mockDB.DB, typeName)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &typeName)
		mockDB.MockScanFields(nil, selectFields, value)
		gotLocations, err := r.GetLocationByLocationTypeName(ctx, mockDB.DB, typeName)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_GetLocationByLocationTypeID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	typeID := "type-id"
	e := &Location{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &typeID)
		gotLocations, err := r.GetLocationByLocationTypeID(ctx, mockDB.DB, typeID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &typeID)
		mockDB.MockScanFields(nil, selectFields, value)
		gotLocations, err := r.GetLocationByLocationTypeID(ctx, mockDB.DB, typeID)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_UpdateAccessPath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationRepoWithSqlMock()
	ids := []string{"id", "id-1"}
	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &ids})
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.UpdateAccessPath(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), &ids})
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := r.UpdateAccessPath(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "locations")
		mockDB.RawStmt.AssertUpdatedFields(t, "access_path")
	})
}

func TestLocationRepo_RetrieveLowestLevelLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationRepoWithSqlMock()
	e := &Location{}
	fields, value := e.FieldMap()
	params := &GetLowestLevelLocationsParams{
		Name:        "center",
		Limit:       10,
		LocationIDs: []string{"location-1", "location-2"},
	}
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, params.Name, params.LocationIDs)
		gotLocations, err := r.RetrieveLowestLevelLocations(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.Name, params.LocationIDs)
		mockDB.MockScanFields(nil, fields, value)
		gotLocations, err := r.RetrieveLowestLevelLocations(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
	t.Run("success without locationIDs", func(t *testing.T) {
		params := &GetLowestLevelLocationsParams{
			Name:  "center",
			Limit: 10,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.Name)
		mockDB.MockScanFields(nil, fields, value)
		gotLocations, err := r.RetrieveLowestLevelLocations(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_GetLocationOrg(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := LocationRepoWithSqlMock()
	location := &Location{}
	fields, value := location.FieldMap()
	resourcePath := fmt.Sprint(constant.ManabieSchool)
	locationTypeOrg := domain.DefaultLocationType
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, locationTypeOrg, resourcePath)

		location, err := repo.GetLocationOrg(ctx, mockDB.DB, resourcePath)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, location)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationTypeOrg, resourcePath)
		mockDB.MockScanFields(nil, fields, value)
		locationOrg, err := repo.GetLocationOrg(ctx, mockDB.DB, resourcePath)
		assert.NoError(t, err)
		assert.NotNil(t, locationOrg)
	})
}

func TestLocationRepo_GetLowestLevelLocationsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationRepoWithSqlMock()
	params := &GetLowestLevelLocationsParams{
		Name:        "brand",
		Limit:       20,
		LocationIDs: []string{"loc-1", "loc-2"},
	}
	location := new(Location)
	fields, value := location.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, params.Name, params.LocationIDs)
		locs, err := r.RetrieveLowestLevelLocations(ctx, mockDB.DB, params)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locs)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.Name, params.LocationIDs)
		mockDB.MockScanFields(nil, fields, value)
		locs, err := r.GetLowestLevelLocationsV2(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, locs)
	})
	t.Run("success without locationIDs", func(t *testing.T) {
		params := &GetLowestLevelLocationsParams{
			Name:  "brand",
			Limit: 10,
		}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, params.Name)
		mockDB.MockScanFields(nil, fields, value)
		locs, err := r.GetLowestLevelLocationsV2(ctx, mockDB.DB, params)
		assert.NoError(t, err)
		assert.NotNil(t, locs)
	})
}

func TestLocationRepo_GetAllRawLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationRepoWithSqlMock()
	e := &Location{}
	fields, value := e.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		gotLocations, err := r.GetAllRawLocations(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		gotLocations, err := r.GetAllRawLocations(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_GetLocationByLocationTypeIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	typeIDs := []string{"type-id-01", "type-id-02"}
	e := &Location{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &typeIDs)
		gotLocations, err := r.GetLocationByLocationTypeIDs(ctx, mockDB.DB, typeIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &typeIDs)
		mockDB.MockScanFields(nil, selectFields, value)
		gotLocations, err := r.GetLocationByLocationTypeIDs(ctx, mockDB.DB, typeIDs)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
	})
}

func TestLocationRepo_GetChildLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationRepoWithSqlMock()
	id := "id"

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &id)

		locations, err := r.GetChildLocations(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &id)

		e := new(Location)
		fields, values := e.FieldMap()
		_ = e.PartnerInternalID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetChildLocations(ctx, mockDB.DB, id)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestLocationRepo_GetRootLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	l, mockDB := LocationRepoWithSqlMock()
	t.Run("err", func(t *testing.T) {
		args := []interface{}{mock.Anything, mock.Anything}

		fields := []string{"assigned_slot"}
		var locationID string
		values := []interface{}{&locationID}
		mockDB.MockQueryRowArgs(t, args...)
		mockDB.MockRowScanFields(errors.New("error"), fields, values)

		_, err := l.GetRootLocation(ctx, mockDB.DB)
		require.Error(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})

	t.Run("success", func(t *testing.T) {
		args := []interface{}{mock.Anything, mock.Anything}

		fields := []string{"assigned_slot"}
		var locationID string
		values := []interface{}{&locationID}
		mockDB.MockQueryRowArgs(t, args...)
		mockDB.MockRowScanFields(nil, fields, values)

		_, err := l.GetRootLocation(ctx, mockDB.DB)
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
			mockDB.Row,
		)
	})
}
