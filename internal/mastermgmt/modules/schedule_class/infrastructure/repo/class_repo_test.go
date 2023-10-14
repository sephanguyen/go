package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ClassRepoWithSqlMock() (*ClassRepo, *testutil.MockDB) {
	classRepo := &ClassRepo{}
	return classRepo, testutil.NewMockDB()
}

func TestClassRepo_GetMapClassByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	classRepo, mockDB := ClassRepoWithSqlMock()

	ids := []string{"01", "02"}

	t.Run("success", func(t *testing.T) {
		rc := &Class{}
		fields, value := rc.FieldMap()

		rc.Name.Set("class_name")
		rc.ClassID.Set("class_id")

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &ids)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			value,
		})
		resp, err := classRepo.GetMapClassByIDs(ctx, mockDB.DB, ids)
		expectedMapClass := make(map[string]*Class)
		expectedMapClass["class_id"] = &Class{
			ClassID: database.Text("class_id"),
			Name:    database.Text("class_name"),
		}
		require.NoError(t, err)
		require.Equal(t, expectedMapClass, resp)
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &ids)

		resp, err := classRepo.GetMapClassByIDs(ctx, mockDB.DB, ids)
		require.True(t, errors.Is(err, puddle.ErrClosedPool))
		require.Nil(t, resp)
	})
}
