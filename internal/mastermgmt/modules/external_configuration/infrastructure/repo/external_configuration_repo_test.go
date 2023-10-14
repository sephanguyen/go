package repo

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ConfigRepoWithSqlMock() (*ExternalConfigRepo, *testutil.MockDB) {
	r := &ExternalConfigRepo{}
	return r, testutil.NewMockDB()
}

func TestConfigRepo_GetByMultipleKeys(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()
	keys := []string{"key", "key-1"}

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(keys))

		configs, err := r.GetByMultipleKeys(ctx, mockDB.DB, keys)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, configs)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys))

		e := &domain.ExternalConfiguration{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByMultipleKeys(ctx, mockDB.DB, keys)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"config_key": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestConfigRepo_GetByKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()
	key := "key_" + idutil.ULIDNow()
	e := &domain.ExternalConfiguration{}
	selectFields, value := e.FieldMap()
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, key)
		mockDB.MockRowScanFields(puddle.ErrClosedPool, selectFields, value)

		cfg, err := r.GetByKey(ctx, mockDB.DB, key)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, cfg)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, key)
		mockDB.MockRowScanFields(nil, selectFields, value)
		_, err := r.GetByKey(ctx, mockDB.DB, key)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestConfigRepo_SearchWithKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rand.Seed(time.Now().Unix())

	r, mockDB := ConfigRepoWithSqlMock()
	limit := rand.Int63n(200)
	offset := rand.Int63n(200)
	searchOpts := domain.ExternalConfigSearchArgs{
		Limit:  limit,
		Offset: offset,
	}

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Int8(limit), database.Int8(offset))

		searchOpts.Keyword = ""
		configs, err := r.SearchWithKey(ctx, mockDB.DB, searchOpts)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, configs)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Int8(limit), database.Int8(offset))

		e := &domain.ExternalConfiguration{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		searchOpts.Keyword = "random-key"
		_, err := r.SearchWithKey(ctx, mockDB.DB, searchOpts)
		assert.Nil(t, err)
		expectedQuery := `
		SELECT configuration_id,config_key,config_value,config_value_type,last_editor,created_at,updated_at,deleted_at,resource_path FROM external_configuration_value
		WHERE deleted_at IS NULL  AND config_key ILIKE '%random-key%'
		ORDER BY config_key DESC LIMIT $1 OFFSET $2`
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.DB.AssertCalled(t, "Query", mock.Anything, expectedQuery, database.Int8(limit), database.Int8(offset))
	})
}

func TestConfigRepo_CreateMultipleConfigs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rand.Seed(time.Now().Unix())

	r, mockDB := ConfigRepoWithSqlMock()
	time := time.Now()

	configs := []*domain.ExternalConfiguration{
		{
			ID:              idutil.ULIDNow(),
			ConfigKey:       "config-key",
			ConfigValue:     "true",
			ConfigValueType: "boolean",
			ResourcePath:    "1234",
			CreatedAt:       time,
			UpdatedAt:       time,
		},
		{
			ID:              idutil.ULIDNow(),
			ConfigKey:       "config-key-2",
			ConfigValue:     "[]",
			ConfigValueType: "json",
			ResourcePath:    "1234",
			CreatedAt:       time,
			UpdatedAt:       time,
		}}

	t.Run("successfully", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := r.CreateMultipleConfigs(ctx, mockDB.DB, configs)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		r.CreateMultipleConfigs(ctx, mockDB.DB, configs)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}

func TestLocationConfigRepo_GetByKeysAndLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()
	keys := []string{"key", "key-1"}
	locations := []string{"A", "B"}

	t.Run("select failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(keys), database.TextArray(locations))

		configs, err := r.GetByKeysAndLocations(ctx, mockDB.DB, keys, locations)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, configs)
	})

	t.Run("select succeeded", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys), database.TextArray(locations))

		e := &domain.LocationConfiguration{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByKeysAndLocations(ctx, mockDB.DB, keys, locations)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":  {HasNullTest: true},
			"config_key":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"location_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}

func TestLocationConfigRepo_GetByKeysAndLocationsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()
	keys := []string{"key", "key-1"}
	locations := []string{"A", "B"}

	t.Run("using V2 repo", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(keys), database.TextArray(locations))

		e := &domain.LocationConfigurationV2{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.GetByKeysAndLocationsV2(ctx, mockDB.DB, keys, locations)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":  {HasNullTest: true},
			"config_key":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"location_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}
