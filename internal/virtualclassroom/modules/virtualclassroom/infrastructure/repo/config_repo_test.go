package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
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

func TestConfigRepo_GetConfigWithResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	configRepo, mockDB := ConfigRepoWithSqlMock()
	config := &Config{}
	fields, values := config.FieldMap()

	keys := []string{"specificCourseIDsForLesson"}
	group := "lesson"
	country := domain.CountryMaster
	resourcePath := "1"

	queryArgs := []interface{}{
		mock.Anything,                 // context
		mock.AnythingOfType("string"), // query string
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	}

	t.Run("successful", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, queryArgs...)
		mockDB.MockScanFields(nil, fields, values)

		configs, err := configRepo.GetConfigWithResourcePath(ctx, mockDB.DB, country, group, keys, resourcePath)
		assert.NoError(t, err)
		assert.NotNil(t, configs)
	})

	t.Run("failed", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, queryArgs...)
		mockDB.MockScanFields(nil, fields, values)

		configs, err := configRepo.GetConfigWithResourcePath(ctx, mockDB.DB, country, group, keys, resourcePath)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, configs)
	})
}
