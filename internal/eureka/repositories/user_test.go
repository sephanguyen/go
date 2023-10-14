package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRepo_GetCountryByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &UserRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	userID := database.Text("user-id-mock")
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything).Once().Return(puddle.ErrClosedPool)
		_, err := r.GetCountryByUserID(ctx, db, userID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything).Once().Return(nil)
		_, err := r.GetCountryByUserID(ctx, db, userID)
		assert.NoError(t, err)
	})
}
