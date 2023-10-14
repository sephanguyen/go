package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSchoolHistoryRepo_GetCurrentSchoolInfoByStudentIDs(t *testing.T) {
	r := &SchoolHistoryRepo{}
	mockDB := testutil.NewMockDB()

	t.Run("Get error", func(t *testing.T) {
		studentIDs := database.TextArray([]string{"1", "2", "3"})
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything, studentIDs)

		_, err := r.GetCurrentSchoolInfoByStudentIDs(context.Background(), mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("Get success", func(t *testing.T) {
		studentIDs := database.TextArray([]string{"1"})
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, studentIDs)

		schoolInfo := entities.SchoolInfo{}
		fields, values := schoolInfo.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		_, err := r.GetCurrentSchoolInfoByStudentIDs(context.Background(), mockDB.DB, studentIDs)
		assert.NoError(t, err)
	})
}
