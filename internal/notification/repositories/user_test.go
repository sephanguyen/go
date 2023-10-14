package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRepo_FindUser(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &UserRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB

	userIDs := []string{"user-1", "user-2"}
	e1 := &entities.User{}
	database.AllRandomEntity(e1)
	e2 := &entities.User{}
	database.AllRandomEntity(e2)

	t.Run("err select", func(t *testing.T) {
		fields, values1 := e1.FieldMap()
		_, values2 := e2.FieldMap()
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.TextArray(userIDs))
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values1,
			values2,
		})
		_, _, err := r.FindUser(ctx, db, &FindUserFilter{
			UserIDs: database.TextArray(userIDs),
		})
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		fields, values1 := e1.FieldMap()
		_, values2 := e2.FieldMap()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.TextArray(userIDs))
		users, mapUsers, err := r.FindUser(ctx, db, &FindUserFilter{
			UserIDs: database.TextArray(userIDs),
		})
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values1,
			values2,
		})
		assert.NoError(t, err)
		assert.Equal(t, e1, mapUsers[e1.UserID.String])
		assert.Equal(t, e1, users[0])
		assert.Equal(t, e2, mapUsers[e2.UserID.String])
		assert.Equal(t, e2, users[1])
	})
}
