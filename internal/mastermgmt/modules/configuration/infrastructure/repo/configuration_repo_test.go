package repo

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ConfigRepoWithSqlMock() (*ConfigRepo, *testutil.MockDB) {
	r := &ConfigRepo{}
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

		e := &domain.InternalConfiguration{}
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
	e := &domain.InternalConfiguration{}
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
	searchOpts := domain.ConfigSearchArgs{
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

		e := &domain.InternalConfiguration{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		searchOpts.Keyword = "random-key"
		_, err := r.SearchWithKey(ctx, mockDB.DB, searchOpts)
		assert.Nil(t, err)
		expectedQuery := `SELECT configuration_id,config_key,config_value,config_value_type,last_editor,created_at,updated_at,deleted_at,resource_path FROM internal_configuration_value WHERE deleted_at IS NULL  AND config_key ILIKE '%random-key%' UNION SELECT configuration_id,config_key,config_value,config_value_type,last_editor,created_at,updated_at,deleted_at,resource_path FROM external_configuration_value WHERE deleted_at IS NULL  AND config_key ILIKE '%random-key%' ORDER BY config_key DESC LIMIT $1 OFFSET $2`
		mockDB.DB.AssertCalled(t, "Query", mock.Anything, expectedQuery, database.Int8(limit), database.Int8(offset))
	})
}
