package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LocationRepoSqlMock() (*LocationRepository, *testutil.MockDB) {
	r := &LocationRepository{}
	return r, testutil.NewMockDB()
}

func TestLocationRepo_GetLocationByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	LocationRepo, mockDB := LocationRepoSqlMock()

	args := []interface{}{mock.Anything, mock.Anything, &[]string{"l1", "l2"}}

	t.Run("success", func(t *testing.T) {
		u := &domain.Location{}
		fields, values := u.FieldMap()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		locations, err := LocationRepo.GetLocationByID(ctx, mockDB.DB, []string{"l1", "l2"})
		assert.Nil(t, err)
		assert.NotNil(t, locations)
	})
}
