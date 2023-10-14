package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ReserveClassRepoWithSqlMock() (*ReserveClassRepo, *testutil.MockDB) {
	reserveClassRepo := &ReserveClassRepo{}
	return reserveClassRepo, testutil.NewMockDB()
}

func TestReserveClassRepo_InsertOne(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()
	reserveClassRepo, mockDB := ReserveClassRepoWithSqlMock()
	reserveClass := &domain.ReserveClass{
		ReserveClassID:   "reserve_class_id_01",
		StudentID:        "student_id_01",
		StudentPackageID: "student_package_id_01",
		CourseID:         "course_id_01",
		ClassID:          "class_id_01",
		EffectiveDate:    now,
		UpdatedAt:        now,
		CreatedAt:        now,
	}
	t.Run("success", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte("1"))
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
		mockDB.DB.On("Close").Once().Return(nil)
		err := reserveClassRepo.InsertOne(ctx, mockDB.DB, reserveClass)
		require.NoError(t, err)
	})
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte("1"))
		mockDB.DB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, fmt.Errorf("exec error"))
		mockDB.DB.On("Close").Once().Return(nil)
		err := reserveClassRepo.InsertOne(ctx, mockDB.DB, reserveClass)
		require.Error(t, err, "exec error")
	})
}

func TestReserveClassRepo_DeleteOldReserveClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reserveClassRepo, mockDB := ReserveClassRepoWithSqlMock()

	studentPackageID := "student_package_id_01"
	studentID := "student_id_01"
	courseID := "course_id_01"
	fields := []string{"class_id", "effective_date"}

	t.Run("success", func(t *testing.T) {
		classID := pgtype.Text{String: "class_id_01"}
		effectiveDate := pgtype.Date{Time: time.Now()}
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentPackageID, studentID, courseID)
		mockDB.MockRowScanFields(nil, fields, []interface{}{
			&classID,
			&effectiveDate,
		})
		respClassID, respEffectiveDate, err := reserveClassRepo.DeleteOldReserveClass(ctx, mockDB.DB, studentPackageID, studentID, courseID)
		require.NoError(t, err)
		require.Equal(t, classID, respClassID)
		require.Equal(t, effectiveDate, respEffectiveDate)
	})

	t.Run("query no row", func(t *testing.T) {
		var classID pgtype.Text
		var effectiveDate pgtype.Date
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentPackageID, studentID, courseID)
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, []interface{}{
			&classID,
			&effectiveDate,
		})
		respClassID, respEffectiveDate, err := reserveClassRepo.DeleteOldReserveClass(ctx, mockDB.DB, studentPackageID, studentID, courseID)
		require.NoError(t, err)
		require.Equal(t, classID, respClassID)
		require.Equal(t, effectiveDate, respEffectiveDate)
	})

	t.Run("error", func(t *testing.T) {
		var classID pgtype.Text
		var effectiveDate pgtype.Date
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, studentPackageID, studentID, courseID)
		mockDB.MockRowScanFields(pgx.ErrInvalidLogLevel, fields, []interface{}{
			&classID,
			&effectiveDate,
		})
		respClassID, respEffectiveDate, err := reserveClassRepo.DeleteOldReserveClass(ctx, mockDB.DB, studentPackageID, studentID, courseID)
		require.Error(t, err)
		require.Equal(t, classID, respClassID)
		require.Equal(t, effectiveDate, respEffectiveDate)
	})
}

func TestReserveClassRepo_GetByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reserveClassRepo, mockDB := ReserveClassRepoWithSqlMock()

	studentID := "student_id_01"

	t.Run("success", func(t *testing.T) {
		rc := &ReserveClassDTO{}
		fields, value := rc.FieldMap()

		rc.ReserveClassID.Set("reserve_class_id_01")
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &studentID)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			value,
		})
		resp, err := reserveClassRepo.GetByStudentIDs(ctx, mockDB.DB, studentID)
		require.NoError(t, err)
		require.Equal(t, resp, []*domain.ReserveClass{rc.ToReserveClassDomain()})
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &studentID)

		resp, err := reserveClassRepo.GetByStudentIDs(ctx, mockDB.DB, studentID)
		require.True(t, errors.Is(err, puddle.ErrClosedPool))
		require.Nil(t, resp)
	})
}

func TestReserveClassRepo_GetByEffectiveDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	reserveClassRepo, mockDB := ReserveClassRepoWithSqlMock()

	date := "2023/07/31"

	t.Run("success", func(t *testing.T) {
		rc := &ReserveClassDTO{}
		fields, value := rc.FieldMap()

		rc.ReserveClassID.Set("reserve_class_id_01")
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			value,
		})
		resp, err := reserveClassRepo.GetByEffectiveDate(ctx, mockDB.DB, date)
		require.NoError(t, err)
		require.Equal(t, resp, []*domain.ReserveClass{rc.ToReserveClassDomain()})
	})

	t.Run("error", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything)

		resp, err := reserveClassRepo.GetByEffectiveDate(ctx, mockDB.DB, date)
		require.True(t, errors.Is(err, puddle.ErrClosedPool))
		require.Nil(t, resp)
	})
}

