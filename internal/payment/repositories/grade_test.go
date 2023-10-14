package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func GradeRepoWithSqlMock() (*GradeRepo, *testutil.MockDB, *mock_database.Tx) {
	gradeRepo := &GradeRepo{}
	return gradeRepo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestGradeRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	gradeRepoWithSqlMock, mockDB, _ := GradeRepoWithSqlMock()
	var gradeID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			gradeID,
		)
		entities := &entities.Grade{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		grade, err := gradeRepoWithSqlMock.GetByID(ctx, mockDB.DB, gradeID)
		assert.Nil(t, err)
		assert.NotNil(t, grade)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			gradeID,
		)
		entities := &entities.Grade{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		grade, err := gradeRepoWithSqlMock.GetByID(ctx, mockDB.DB, gradeID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, grade)

	})
}

func TestGradeRepo_GetGradeNamesByGradeIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gradeRepoWithSqlMock, mockDB, _ := GradeRepoWithSqlMock()
	gradeIDs := []string{
		"grade1", "grade2",
	}
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, gradeIDs)
		entity := &entities.Grade{}
		fields, _ := entity.FieldMap()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		rows.On("Scan", fields).Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := gradeRepoWithSqlMock.GetGradeNamesByGradeIDs(ctx, mockDB.DB, gradeIDs)
		assert.Nil(t, err)

	})
}
