package repo

import (
	"context"
	"testing"
	"time"

	testing_util "github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrganizationRepoWithSqlMock() (*OrganizationRepo, *testing_util.MockDB) {
	o := &OrganizationRepo{}
	return o, testing_util.NewMockDB()
}
func TestOrganizationRepo_GetIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := OrganizationRepoWithSqlMock()

	t.Run("err select", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string")).Once().Return(nil, puddle.ErrClosedPool)
		results, err := r.GetIDs(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("successfully", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string")).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Close").Once().Return(nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", mock.Anything).Once().Return(nil)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)

		_, err := r.GetIDs(ctx, mockDB.DB)
		assert.Nil(t, err)
	})
}
