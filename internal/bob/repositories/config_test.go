package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ConfigRepoWithSqlMock() (*ConfigRepo, *testutil.MockDB) {
	r := &ConfigRepo{}
	return r, testutil.NewMockDB()
}

func TestConfigTes_Retrieve(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()

	country := database.Text("country")
	group := database.Text("group")
	keys := database.TextArray([]string{"key-1", "key-2"})

	config := &entities.Config{}
	fields, _ := config.FieldMap()
	scanFields := database.GetScanFields(config, fields)

	t.Run("err select", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(pgx.ErrNoRows)

		_, err := r.Retrieve(ctx, mockDB.DB, country, group, keys)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("scan field row success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		_, err := r.Retrieve(ctx, mockDB.DB, country, group, keys)
		assert.Nil(t, err)
	})
}

func TestConfigTes_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()

	country := database.Text("country")
	group := database.Text("group")

	config := &entities.Config{}
	fields, _ := config.FieldMap()
	scanFields := database.GetScanFields(config, fields)

	t.Run("err select", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(pgx.ErrNoRows)

		_, err := r.Find(ctx, mockDB.DB, country, group)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("scan field row success", func(t *testing.T) {
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		_, err := r.Find(ctx, mockDB.DB, country, group)
		assert.Nil(t, err)
	})
}

func TestConfigRepo_RetrieveWithResourcePath(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ConfigRepoWithSqlMock()
	country := database.Text("COUNTRY_MASTER")
	group := database.Text("lesson")
	keys := database.TextArray([]string{"specificCourseIDsForLesson"})
	resourcePath := database.Text("1")
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrNotAvailable, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		configs, err := r.RetrieveWithResourcePath(ctx, mockDB.DB, country, group, keys, resourcePath)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
		assert.Nil(t, configs)
	})

	t.Run("success with select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		e := &entities.Config{}
		fields, values := e.FieldMap()
		_ = e.Country.Set("COUNTRY_MASTER")

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.RetrieveWithResourcePath(ctx, mockDB.DB, country, group, keys, resourcePath)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestConfigRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c, mockDB := ConfigRepoWithSqlMock()
	config1 := &entities.Config{
		Key:       database.Text("key-1"),
		Group:     database.Text("group-1"),
		Country:   database.Text("country-1"),
		Value:     database.Text("value-1"),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}
	config2 := &entities.Config{
		Key:       database.Text("key-2"),
		Group:     database.Text("group-2"),
		Country:   database.Text("country-2"),
		Value:     database.Text("value-2"),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}
	t.Run("successfully", func(t *testing.T) {
		configs := []*entities.Config{config1, config2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		err := c.Upsert(ctx, mockDB.DB, configs)
		require.Equal(t, err, nil)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
	t.Run("error", func(t *testing.T) {
		configs := []*entities.Config{config1, config2}
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
		batchResults.On("Close").Once().Return(nil)

		c.Upsert(ctx, mockDB.DB, configs)

		mock.AssertExpectationsForObjects(
			t,
			mockDB.DB,
		)
	})
}
