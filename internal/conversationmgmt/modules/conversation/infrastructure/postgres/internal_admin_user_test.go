package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternalAdminUserRepo_GetOne(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &InternalAdminUserRepo{}

	entInDB := &dto.InternalAdminUserPgDTO{}
	database.AllRandomEntity(entInDB)
	t.Run("success", func(t *testing.T) {

		fields, vals := entInDB.FieldMap()
		mockDB.MockRowScanFields(nil, fields, vals)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything)
		ret, err := r.GetOne(ctx, db)
		assert.Nil(t, err)
		assert.Equal(t, entInDB.ToInternalAdminUserDomain(), ret)
	})
	t.Run("err query row", func(t *testing.T) {
		fields, vals := entInDB.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, vals)
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything)
		ret, err := r.GetOne(ctx, db)
		assert.Nil(t, ret)
		assert.Equal(t, fmt.Errorf("db.QueryRow: [%v]", pgx.ErrNoRows).Error(), err.Error())
	})
}
