package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
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

func TestUserRepo_FindUserWithEmailByEmailReference(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.User{}
	_, fieldMap := mockE.FieldMap()

	studentReferenceID := "student-test"

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()

		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		_, err := repo.FindUserWithEmailByEmailReference(ctx, mockDB.DB, studentReferenceID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("find user with email failed", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		_, err := repo.FindUserWithEmailByEmailReference(ctx, mockDB.DB, studentReferenceID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UserRepo FindUserWithEmailByEmailReference: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("No rows affected FindUserWithEmailByEmailReference", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		userID, err := repo.FindUserWithEmailByEmailReference(ctx, mockDB.DB, studentReferenceID)
		assert.Nil(t, err)
		assert.Empty(t, userID)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestUserRepo_FindByExternalID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	externalID := "id"
	_, userValues := (&entities.User{}).FieldMap()
	argsStudent := append([]interface{}{}, genSliceMock(len(userValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &externalID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(nil)
		students, err := repo.FindByUserExternalID(ctx, mockDB.DB, externalID)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := UserRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &externalID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsStudent...).Once().Return(puddle.ErrClosedPool)
		user, err := repo.FindByUserExternalID(ctx, mockDB.DB, externalID)
		assert.Nil(t, user)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
