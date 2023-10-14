package repository

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func GradeOrganizationRepoWithSqlMock() (*GradeOrganizationRepo, *testutil.MockDB) {
	r := &GradeOrganizationRepo{}
	return r, testutil.NewMockDB()
}

func TestGradeOrganizationRepo_GetByGradeIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gradeIDDs := []string{uuid.NewString()}
	_, domainGradeValues := NewGrade(entity.NullDomainGrade{}).FieldMap()
	argsDomainGrades := append([]interface{}{}, genSliceMock(len(domainGradeValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(gradeIDDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsDomainGrades...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByGradeIDs(ctx, mockDB.DB, gradeIDDs)
		assert.Nil(t, err)
		assert.NotNil(t, grantedRoles)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(gradeIDDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		grantedRoles, err := repo.GetByGradeIDs(ctx, mockDB.DB, gradeIDDs)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(gradeIDDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsDomainGrades...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByGradeIDs(ctx, mockDB.DB, gradeIDDs)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})
}

func TestGradeOrganizationRepo_GetByGradeValues(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gradeValues := []int32{rand.Int31()}
	_, domainGradeValues := NewGrade(entity.NullDomainGrade{}).FieldMap()
	argsDomainGrades := append([]interface{}{}, genSliceMock(len(domainGradeValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Int4Array(gradeValues)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsDomainGrades...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByGradeValues(ctx, mockDB.DB, gradeValues)
		assert.Nil(t, err)
		assert.NotNil(t, grantedRoles)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Int4Array(gradeValues)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		grantedRoles, err := repo.GetByGradeValues(ctx, mockDB.DB, gradeValues)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := GradeOrganizationRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Int4Array(gradeValues)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsDomainGrades...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByGradeValues(ctx, mockDB.DB, gradeValues)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})
}
