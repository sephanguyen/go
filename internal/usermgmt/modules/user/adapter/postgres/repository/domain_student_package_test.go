package repository

import (
	"context"
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

func DomainStudentPackageRepoWithSqlMock() (*DomainStudentPackageRepo, *testutil.MockDB) {
	r := &DomainStudentPackageRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainStudentPackageRepo_GetByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()

	_, enrollmentStatus := NewStudentPackage(entity.DefaultDomainStudentPackage{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray([]string{studentID})).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, []string{studentID})
		assert.Nil(t, err)
		assert.NotNil(t, studentPackages)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray([]string{studentID})).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, []string{studentID})
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray([]string{studentID})).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDs(ctx, mockDB.DB, []string{studentID})
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})
}

func TestDomainStudentPackageRepo_GetByStudentCourseAndLocationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()
	courseID := uuid.NewString()
	locationIDs := []string{uuid.NewString()}
	_, enrollmentStatus := NewStudentPackage(entity.DefaultDomainStudentPackage{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentCourseAndLocationIDs(ctx, mockDB.DB, studentID, courseID, locationIDs)
		assert.Nil(t, err)
		assert.NotNil(t, studentPackages)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		studentPackages, err := repo.GetByStudentCourseAndLocationIDs(ctx, mockDB.DB, studentID, courseID, locationIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.TextArray(locationIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentCourseAndLocationIDs(ctx, mockDB.DB, studentID, courseID, locationIDs)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})
}

func TestDomainStudentPackageRepo_GetByStudentIDAndCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()
	courseID := uuid.NewString()
	_, enrollmentStatus := NewStudentPackage(entity.DefaultDomainStudentPackage{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDAndCourseID(ctx, mockDB.DB, studentID, courseID)
		assert.Nil(t, err)
		assert.NotNil(t, studentPackages)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		studentPackages, err := repo.GetByStudentIDAndCourseID(ctx, mockDB.DB, studentID, courseID)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainStudentPackageRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		studentPackages, err := repo.GetByStudentIDAndCourseID(ctx, mockDB.DB, studentID, courseID)
		assert.NotNil(t, err)
		assert.Nil(t, studentPackages)
	})
}
