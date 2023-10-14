package repository

import (
	"context"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PrefectureRepoWithSqlMock() (*PrefectureRepo, *testutil.MockDB) {
	repo := &PrefectureRepo{}
	return repo, testutil.NewMockDB()
}

func TestPrefectureRepo_GetByPrefectureID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	prefectureCode := database.Text("01")
	_, prefectureValues := (&entity.Prefecture{}).FieldMap()
	argsPrefecture := append([]interface{}{}, genSliceMock(len(prefectureValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PrefectureRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &prefectureCode).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsPrefecture...).Once().Return(nil)
		prefecture, err := repo.GetByPrefectureID(ctx, mockDB.DB, prefectureCode)
		assert.Nil(t, err)
		assert.NotNil(t, prefecture)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := PrefectureRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &prefectureCode).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsPrefecture...).Once().Return(puddle.ErrClosedPool)
		prefecture, err := repo.GetByPrefectureID(ctx, mockDB.DB, prefectureCode)
		assert.Nil(t, prefecture)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPrefectureRepo_GetByPrefectureCode(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	prefectureCode := database.Text("01")
	_, prefectureValues := (&entity.Prefecture{}).FieldMap()
	argsPrefecture := append([]interface{}{}, genSliceMock(len(prefectureValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PrefectureRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &prefectureCode).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsPrefecture...).Once().Return(nil)
		prefecture, err := repo.GetByPrefectureCode(ctx, mockDB.DB, prefectureCode)
		assert.Nil(t, err)
		assert.NotNil(t, prefecture)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := PrefectureRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &prefectureCode).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsPrefecture...).Once().Return(puddle.ErrClosedPool)
		prefecture, err := repo.GetByPrefectureCode(ctx, mockDB.DB, prefectureCode)
		assert.Nil(t, prefecture)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPrefectureRepo_GetByPrefectureIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	prefectureCodes := database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow()})
	_, prefectureValues := (&entity.Prefecture{}).FieldMap()
	repo, mockDB := PrefectureRepoWithSqlMock()

	rows := mockDB.Rows
	t.Run("happy case", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", prefectureValues...).Once().Return(nil)
		rows.On("Next").Once().Return(false)

		schools, err := repo.GetByPrefectureIDs(ctx, mockDB.DB, prefectureCodes)
		assert.Nil(t, err)
		assert.NotNil(t, schools)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", prefectureValues...).Once().Return(pgx.ErrNoRows)

		_, err := repo.GetByPrefectureIDs(ctx, mockDB.DB, prefectureCodes)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
