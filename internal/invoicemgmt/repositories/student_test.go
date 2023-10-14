package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentRepoWithSqlMock() (*StudentRepo, *testutil.MockDB) {
	repo := &StudentRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	studentID := "id"
	_, studentValues := (&entities.Student{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(studentValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(nil)
		students, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := StudentRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(puddle.ErrClosedPool)
		student, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.Nil(t, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
