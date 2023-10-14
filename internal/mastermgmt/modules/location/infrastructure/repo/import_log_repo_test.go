package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ImportLogRepoWithSqlMock() (*ImportLogRepo, *testutil.MockDB) {
	return &ImportLogRepo{}, testutil.NewMockDB()
}
func TestImportLogRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := ImportLogRepoWithSqlMock()
	t.Run("err insert", func(t *testing.T) {
		e := &domain.ImportLog{}
		dto, _ := ToImportLog(e)
		_, values := dto.FieldMap()

		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrNotAvailable, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrNotAvailable))
	})

	t.Run("success", func(t *testing.T) {
		e := &domain.ImportLog{}
		dto, _ := ToImportLog(e)
		fields, values := dto.FieldMap()

		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Create(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, dto.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}
