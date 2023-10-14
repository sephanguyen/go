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

func TestStudentRepo_GetCountryByStudent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r := &StudentRepo{}
	mockDB := testutil.NewMockDB()
	db := mockDB.DB
	row := mockDB.Row

	studentID := database.Text("student-id-mock")
	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &studentID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything).Once().Return(puddle.ErrClosedPool)
		_, err := r.GetCountryByStudent(ctx, db, studentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &studentID)
		db.On("QueryRow").Once().Return(row, nil)
		row.On("Scan", mock.Anything).Once().Return(nil)
		_, err := r.GetCountryByStudent(ctx, db, studentID)
		assert.NoError(t, err)
	})
}
