package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TaggedUserRepoWithSqlMock() (*TaggedUserRepo, *testutil.MockDB) {
	r := &TaggedUserRepo{}
	return r, testutil.NewMockDB()
}

func TestChapterRepo_FindByTagIDsAndUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := TaggedUserRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, pgIDs, pgIDs)

		schoolIDs, err := r.FindByTagIDsAndUserIDs(ctx, mockDB.DB, database.TextArray(ids), database.TextArray(ids))
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, pgIDs, pgIDs)

		e := &EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		_, err := r.FindByTagIDsAndUserIDs(ctx, mockDB.DB, database.TextArray(ids), database.TextArray(ids))
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedTable(t, "tagged_user", "")

	})
}
