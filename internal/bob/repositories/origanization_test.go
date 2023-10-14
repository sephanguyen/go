package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrganizationRepoWithSqlMock() (*OrganizationRepo, *testutil.MockDB) {
	r := &OrganizationRepo{}
	return r, testutil.NewMockDB()
}

func TestOrganizationRepo_GetOrganizations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := OrganizationRepoWithSqlMock()

	t.Run("err select", func(t *testing.T) {
		mockDB.DB.On("QueryRow", mock.Anything,
			mock.AnythingOfType("string"),
		).Return(mockDB.Row, nil).Once()
		mockDB.Row.On("Scan", mock.Anything).Once().Return(puddle.ErrClosedPool)

		results, err := r.GetOrganizations(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, results)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		orIDs := []string{"orID-1", "orID-2"}
		mockDB.DB.On("QueryRow", mock.Anything,
			mock.AnythingOfType("string"),
		).Return(mockDB.Row, nil).Once()
		mockDB.Row.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			result := args[0].(*pgtype.TextArray)
			*result = database.TextArray(orIDs)
		}).Once().Return(nil)

		results, err := r.GetOrganizations(ctx, mockDB.DB)
		assert.NoError(t, err)
		assert.Equal(t, orIDs[0], results[0])
		assert.Equal(t, orIDs[1], results[1])
	})
}
