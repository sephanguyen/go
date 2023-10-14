package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserRepoWithSqlMock() (*UserRepo, *testutil.MockDB) {
	repo := &UserRepo{}
	return repo, testutil.NewMockDB()
}

func TestUserRepo_FindByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userID := "id"
	_, userValues := (&entities.User{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(userValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(nil)
		students, err := repo.FindByID(ctx, mockDB.DB, userID)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(puddle.ErrClosedPool)
		user, err := repo.FindByID(ctx, mockDB.DB, userID)
		assert.Nil(t, user)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
