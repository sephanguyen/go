package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserRepoWithSqlMock() (*UserRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	userRepo := &UserRepo{}
	return userRepo, mockDB
}

func TestUserRepo_GetStaffsByLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userRepo, mockDB := UserRepoWithSqlMock()
	e := &User{}
	fields, values := e.FieldMap()
	location := "location-1"
	status := []string{"status-1", "status-2"}
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, location, status)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationAndWorkingStatus(ctx, mockDB.DB, location, status, false)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})
	t.Run("success with using user basic info", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, location, status)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationAndWorkingStatus(ctx, mockDB.DB, location, status, true)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, location, status)

		staff, err := userRepo.GetStaffsByLocationAndWorkingStatus(ctx, mockDB.DB, location, status, false)

		assert.Error(t, err)
		assert.Empty(t, staff)
	})
}

func TestUserRepo_GetStaffsByLocationIDsAndNameOrEmail(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userRepo, mockDB := UserRepoWithSqlMock()
	e := &User{}
	fields, values := e.FieldMap()
	locationIds := []string{"location-1", "location-2"}
	teacherIds := []string{"teacher1"}
	keyword := "name"
	t.Run("success with full filter", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIds, teacherIds, 10)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, teacherIds, keyword, 10)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})

	t.Run("success without filter", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIds)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, []string{}, "", 0)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})

	t.Run("success with filter teacherIDs", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIds, teacherIds)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, teacherIds, "", 0)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})

	t.Run("success with filter keyword", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIds)
		mockDB.MockScanFields(nil, fields, values)

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, []string{}, keyword, 0)

		assert.NoError(t, err)
		assert.NotEmpty(t, staff)
	})

	t.Run("error when resp return more items than limit", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, locationIds, teacherIds, 1)
		mockDB.MockScanArray(nil, fields, [][]interface{}{values, values})

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, teacherIds, keyword, 1)

		assert.Error(t, err)
		assert.Empty(t, staff)
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, locationIds, teacherIds, 10)

		staff, err := userRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, mockDB.DB, locationIds, teacherIds, keyword, 10)

		assert.Error(t, err)
		assert.Empty(t, staff)
	})
}

func TestUserRepo_GetStudentCurrentGradeByUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockUserRepo, mockDB := UserRepoWithSqlMock()
	userIDs := []string{"test-user-id-1", "test-user-id-2", "test-user-id-3"}
	fields := []string{
		"user_id",
		"student_grade",
	}
	var (
		userID       pgtype.Text
		studentGrade pgtype.Text
	)
	values := make([]interface{}, 0, 2)
	values = append(values, &userID)
	values = append(values, &studentGrade)

	t.Run("failed to get student grades", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), userIDs)

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs, false)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentGradeMap)
	})

	t.Run("successfully fetched student grades", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), userIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs, false)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})

	t.Run("successfully fetched student grades using user basic info", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), userIDs)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		studentGradeMap, err := mockUserRepo.GetStudentCurrentGradeByUserIDs(ctx, mockDB.DB, userIDs, true)
		assert.Nil(t, err)
		assert.NotNil(t, studentGradeMap)
	})
}
