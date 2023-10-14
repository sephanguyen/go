package repo

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func LocationTypeRepoWithSqlMock() (*LocationTypeRepo, *testutil.MockDB) {
	r := &LocationTypeRepo{}
	return r, testutil.NewMockDB()
}

func TestLocationTypeRepo_UpsertLocationTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	r, mockDB := LocationTypeRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		locationType1 := &domain.LocationType{
			LocationTypeID: "location-1",
			DisplayName:    "partner-1",
			Name:           "lesson name 1",
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      &now,
		}
		locationType2 := &domain.LocationType{
			LocationTypeID: "location-2",
			DisplayName:    "partner-2",
			Name:           "lesson name 2",
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      &now,
		}
		locationTypes := make(map[int]*domain.LocationType)
		locationTypes[1] = locationType1
		locationTypes[2] = locationType2
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		errs := r.UpsertLocationTypes(ctx, mockDB.DB, locationTypes)
		require.Equal(t, len(errs), 0)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		locationType1 := &domain.LocationType{
			LocationTypeID: "location-1",
			DisplayName:    "partner-1",
			Name:           "lesson name 1",
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      &now,
		}
		locationType2 := &domain.LocationType{
			LocationTypeID: "location-2",
			DisplayName:    "partner-2",
			Name:           "lesson name 2",
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      &now,
		}
		locationTypes := make(map[int]*domain.LocationType)
		locationTypes[1] = locationType1
		locationTypes[2] = locationType2
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		errs := r.UpsertLocationTypes(ctx, mockDB.DB, locationTypes)
		require.Equal(t, len(errs), 1)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLocationTypeRepo_Import(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	t.Run("successfully", func(t *testing.T) {
		l := getRandomLocTypes()
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, l)
		require.Nil(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		l := getRandomLocTypes()

		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		err := r.Import(ctx, mockDB.DB, l)
		require.Error(t, err)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLocationTypeRepo_GetLocationTypeByName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	name := "name"
	e := &LocationType{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &name)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		locationType, err := r.GetLocationTypeByName(ctx, mockDB.DB, name, true)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locationType)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &name)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationTypeByName(ctx, mockDB.DB, name, true)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectField(t, selectFields...)
		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
	})
}

func TestLocationTypeRepo_GetLocationByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	id := "id"
	e := &LocationType{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &id)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		locationType, err := r.GetLocationTypeByID(ctx, mockDB.DB, id)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locationType)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &id)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationTypeByID(ctx, mockDB.DB, id)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectField(t, selectFields...)
		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
	})
}

func TestLocationTypeRepo_GetLocationTypeByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	ids := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		locations, err := r.GetLocationTypeByIDs(ctx, mockDB.DB, ids, true)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select with delete", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &LocationType{}
		fields, values := e.FieldMap()
		_ = e.LocationTypeID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationTypeByIDs(ctx, mockDB.DB, ids, false)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectField(t, fields...)

		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":       {HasNullTest: true},
			"location_type_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})

	t.Run("success with select without delete", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)

		e := &LocationType{}
		fields, values := e.FieldMap()
		_ = e.LocationTypeID.Set("id")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationTypeByIDs(ctx, mockDB.DB, ids, true)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectField(t, fields...)

		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"location_type_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLocationTypeRepo_GetLocationTypeByNames(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	names := database.TextArray([]string{"name", "name-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &names)

		locations, err := r.GetLocationTypeByNames(ctx, mockDB.DB, names)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locations)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &names)

		e := &LocationType{}
		fields, values := e.FieldMap()
		_ = e.Name.Set("name")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetLocationTypeByNames(ctx, mockDB.DB, names)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectField(t, fields...)

		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"name":       {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLocationRepo_RetrieveLocationTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	e := &LocationType{}
	fields, value := e.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		gotLocationTypes, err := r.RetrieveLocationTypes(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocationTypes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		gotLocationTypes, err := r.RetrieveLocationTypes(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocationTypes)
	})
}

func TestLocationTypeRepo_GetLocationTypeByParentName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	parentName := "name"
	e := &LocationType{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &parentName)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		locationType, err := r.GetLocationTypeByParentName(ctx, mockDB.DB, parentName)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locationType)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &parentName)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationTypeByParentName(ctx, mockDB.DB, parentName)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectField(t, selectFields...)
		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
	})
}

func TestLocationTypeRepo_GetLocationTypeByNameAndParent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LocationTypeRepoWithSqlMock()
	name := "name"
	parentName := "parent_name"
	e := &LocationType{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &name, &parentName)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		locationType, err := r.GetLocationTypeByNameAndParent(ctx, mockDB.DB, name, parentName)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, locationType)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &name, &parentName)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetLocationTypeByNameAndParent(ctx, mockDB.DB, name, parentName)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectField(t, selectFields...)
		mockDB.RawStmt.AssertFromClause(t, e.TableName(), "")
	})
}

func TestLocationTypeRepo_GetAllLocationTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	e := &LocationType{}
	fields, value := e.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		gotLocationTypes, err := r.GetAllLocationTypes(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, gotLocationTypes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		gotLocationTypes, err := r.GetAllLocationTypes(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, gotLocationTypes)
	})
}

func TestLocationTypeRepo_GetAllLocationTypesV2(t *testing.T) {
	t.Parallel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	e := &LocationType{}
	fields, value := e.FieldMap()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		actualLocTypes, err := r.GetAllLocationTypesV2(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, actualLocTypes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		actualLocTypes, err := r.GetAllLocationTypesV2(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, actualLocTypes)
	})
}

func TestLocationRepo_RetrieveLocationTypesV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	e := &LocationType{}
	fields, value := e.FieldMap()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)
		actualLocTypes, err := r.RetrieveLocationTypesV2(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, actualLocTypes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanFields(nil, fields, value)
		actualLocTypes, err := r.RetrieveLocationTypesV2(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.NotNil(t, actualLocTypes)
	})
}

func TestLocationTypeRepo_UpdateLevels(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	t.Run("error", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), puddle.ErrNotAvailable, mock.Anything, mock.AnythingOfType("string"))
		err := r.UpdateLevels(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.AnythingOfType("string"))
		err := r.UpdateLevels(ctx, mockDB.DB)
		assert.Nil(t, err)
	})
}

func getRandomLocTypes() []*domain.LocationType {
	now := time.Now()
	l1 := &domain.LocationType{
		LocationTypeID: idutil.ULIDNow(),
		Name:           "some name" + idutil.ULIDNow(),
		DisplayName:    "display" + idutil.ULIDNow(),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsArchived:     randBool(),
		DeletedAt:      nil,
	}
	l2 := &domain.LocationType{
		LocationTypeID: idutil.ULIDNow(),
		Name:           "some name" + idutil.ULIDNow(),
		DisplayName:    "display" + idutil.ULIDNow(),
		CreatedAt:      now,
		UpdatedAt:      now,
		IsArchived:     randBool(),
		DeletedAt:      nil,
	}
	lt := []*domain.LocationType{
		l1, l2,
	}
	return lt
}

func TestLocationTypeRepo_GetLocationTypesByLevel(t *testing.T) {
	t.Parallel()
	r, mockDB := LocationTypeRepoWithSqlMock()
	e := &LocationType{}
	fields, value := e.FieldMap()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	level := "2"
	defer cancel()
	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, level)
		actualLocTypes, err := r.GetLocationTypesByLevel(ctx, mockDB.DB, level)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, actualLocTypes)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, level)
		mockDB.MockScanFields(nil, fields, value)
		actualLocTypes, err := r.GetLocationTypesByLevel(ctx, mockDB.DB, level)
		assert.NoError(t, err)
		assert.NotNil(t, actualLocTypes)
	})
}

func randBool() bool {
	rand.Seed(time.Now().UnixNano())
	return (rand.Intn(2) == 1)
}
