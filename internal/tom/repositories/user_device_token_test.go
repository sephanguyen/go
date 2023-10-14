package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserDeviceTokenRepo_BulkUpdateResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	r := &UserDeviceTokenRepo{}

	offsetID := pgtype.Text{}
	offsetID.Set(nil)
	userIDs := []string{"user-1", "user-2"}
	resourcePath := "manabie"

	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, nil, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "user_device_tokens" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(userIDs))
		err := r.BulkUpdateResourcePath(ctx, db, userIDs, resourcePath)
		assert.NoError(t, err)
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockExecArgs(t, pgconn.CommandTag{}, pgx.ErrTxClosed, mock.Anything, mock.MatchedBy(func(execString string) bool {
			stmt := testutil.ParseSQL(t, execString)
			return stmt.MustGetUpdatedTable() == "user_device_tokens" && cmp.Equal(stmt.MustGetUpdatedFields(), []string{"resource_path"})
		}), database.Text(resourcePath), database.TextArray(userIDs))
		err := r.BulkUpdateResourcePath(ctx, db, userIDs, resourcePath)
		assert.ErrorIs(t, err, pgx.ErrTxClosed)
	})
}
