package repository

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func SchoolRepoWithSqlMock() (*SchoolRepo, *testutil.MockDB) {
	repo := &SchoolRepo{}
	return repo, testutil.NewMockDB()
}

func TestSchoolRepo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	schoolID := database.Int4(math.MinInt32)
	_, schoolValues := (&entity.School{}).FieldMap()
	argsSchool := append([]interface{}{}, genSliceMock(len(schoolValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := SchoolRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &schoolID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsSchool...).Once().Return(nil)
		school, err := repo.Find(ctx, mockDB.DB, schoolID)
		assert.Nil(t, err)
		assert.NotNil(t, school)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := SchoolRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &schoolID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsSchool...).Once().Return(puddle.ErrClosedPool)
		school, err := repo.Find(ctx, mockDB.DB, schoolID)
		assert.Nil(t, school)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
