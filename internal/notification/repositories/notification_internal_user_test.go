package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationInternalUserRepo_GetByOrgID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &NotificationInternalUserRepo{}

	ent := &entities.NotificationInternalUser{}
	database.AllRandomEntity(ent)
	recourcePath := "recource_path_1"
	t.Run("success", func(t *testing.T) {
		fields, vals1 := ent.FieldMap()
		mockDB.MockScanFields(nil, fields, vals1)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(recourcePath))
		result, err := r.GetByOrgID(ctx, db, recourcePath)
		assert.Nil(t, err)
		assert.Equal(t, ent, result)
	})
	t.Run("error query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(recourcePath))
		questions, err := r.GetByOrgID(ctx, db, recourcePath)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
	t.Run("error scan", func(t *testing.T) {
		fields, vals := ent.FieldMap()
		mockDB.MockScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(recourcePath))
		questions, err := r.GetByOrgID(ctx, db, recourcePath)
		assert.Nil(t, questions)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}
