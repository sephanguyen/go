package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainEnrollmentStatusHistoryRepoWithSqlMock() (*DomainEnrollmentStatusHistoryRepo, *testutil.MockDB) {
	r := &DomainEnrollmentStatusHistoryRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainEnrollmentStatusHistoryRepo_create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, enrollmentStatus)
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := repo.Create(ctx, mockDB.DB, enrollmentStatus)
		assert.Equal(t, err, err)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, enrollmentStatus, enrollmentStatus)
		assert.Nil(t, err)
	})
	t.Run("create fail", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := repo.Update(ctx, mockDB.DB, enrollmentStatus, enrollmentStatus)
		assert.Equal(t, err, err)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetByStudentIDAndLocationID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()
	locationID := uuid.NewString()

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID, false)
		assert.Nil(t, err)
		assert.NotNil(t, grantedRoles)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		grantedRoles, err := repo.GetByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID, false)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByStudentIDAndLocationID(ctx, mockDB.DB, studentID, locationID, false)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetLatestEnrollmentStudentOfLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()
	locationID := uuid.NewString()

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetLatestEnrollmentStudentOfLocation(ctx, mockDB.DB, studentID, locationID)
		assert.Nil(t, err)
		assert.NotNil(t, grantedRoles)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		grantedRoles, err := repo.GetLatestEnrollmentStudentOfLocation(ctx, mockDB.DB, studentID, locationID)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID), database.Text(locationID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetLatestEnrollmentStudentOfLocation(ctx, mockDB.DB, studentID, locationID)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_SoftDeleteEnrollments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
	enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
			database.Text(enrollmentStatus.UserID().String()),
			database.Text(enrollmentStatus.LocationID().String()),
			database.Timestamptz(enrollmentStatus.StartDate().Time()),
			database.Text(enrollmentStatus.EnrollmentStatus().String()),
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDeleteEnrollments(ctx, mockDB.DB, enrollmentStatus)
		assert.True(t, errors.Is(err, err))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
			database.Text(enrollmentStatus.UserID().String()),
			database.Text(enrollmentStatus.LocationID().String()),
			database.Timestamptz(enrollmentStatus.StartDate().Time()),
			database.Text(enrollmentStatus.EnrollmentStatus().String()),
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteEnrollments(ctx, mockDB.DB, enrollmentStatus)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.DB.On("Exec", args...).Return(pgconn.CommandTag("1"), nil)
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestDomainEnrollmentStatusHistoryRepo_DeactivateEnrollmentStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
	enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})
	dateReq := time.Time{}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
			database.TimestamptzNull(dateReq.Add(-1*time.Second)),
			database.Text(enrollmentStatus.UserID().String()),
			database.Text(enrollmentStatus.LocationID().String()),
			database.Text(enrollmentStatus.EnrollmentStatus().String()),
			database.Timestamptz(enrollmentStatus.StartDate().Time()),
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.DeactivateEnrollmentStatus(ctx, mockDB.DB, enrollmentStatus, dateReq.Add(-1*time.Second))
		assert.True(t, errors.Is(err, err))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")},
			database.TimestamptzNull(dateReq.Add(-1*time.Second)),
			database.Text(enrollmentStatus.UserID().String()),
			database.Text(enrollmentStatus.LocationID().String()),
			database.Text(enrollmentStatus.EnrollmentStatus().String()),
			database.Timestamptz(enrollmentStatus.StartDate().Time()),
		)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.DeactivateEnrollmentStatus(ctx, mockDB.DB, enrollmentStatus, dateReq.Add(-1*time.Second))
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.DB.On("Exec", args...).Return(pgconn.CommandTag("1"), nil)
		mockDB.RawStmt.AssertUpdatedFields(t, "end_date")
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := uuid.NewString()

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentID(ctx, mockDB.DB, studentID, false)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})

	t.Run("happy case: get current enrollment history", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentID(ctx, mockDB.DB, studentID, false)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		enrollmentStatusHistories, err := repo.GetByStudentID(ctx, mockDB.DB, studentID, false)
		assert.NotNil(t, err)
		assert.Nil(t, enrollmentStatusHistories)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(studentID)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentID(ctx, mockDB.DB, studentID, false)
		assert.NotNil(t, err)
		assert.Nil(t, enrollmentStatusHistories)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	enrollmentStatusHistory := entity.DefaultDomainEnrollmentStatusHistory{}
	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On(
			"Query", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentIDLocationIDEnrollmentStatusStartDateAndEndDate(ctx, mockDB.DB, enrollmentStatusHistory)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentIDs := []string{"student-id"}

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})

	t.Run("happy case: get current enrollment history", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		enrollmentStatusHistories, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, enrollmentStatusHistories)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(studentIDs)).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.NotNil(t, err)
		assert.Nil(t, enrollmentStatusHistories)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_BulkInsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.BulkInsert(ctx, mockDB.DB, entity.DomainEnrollmentStatusHistories{
			enrollmentStatus,
		})
		assert.Nil(t, err)
	})
	t.Run("BulkInsert fail", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{})

		_, enrollmentStatusValues := enrollmentStatus.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(enrollmentStatusValues))...)
		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := repo.BulkInsert(ctx, mockDB.DB, entity.DomainEnrollmentStatusHistories{
			enrollmentStatus,
		})
		assert.Equal(t, err, err)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_UpdateStudentStatusBasedEnrollmentStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	expectedQuery := `
		WITH inactive_students as (
			SELECT sesh1.student_id, MAX(sesh1.start_date) AS latest_start_date 
			FROM student_enrollment_status_history sesh1
			WHERE NOT EXISTS 
				(
					SELECT 1
					FROM student_enrollment_status_history sesh2
					WHERE NOT (sesh2.enrollment_status = ANY($1::text[]))
					AND sesh2.student_id = sesh1.student_id 
					AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
					AND start_date < CLOCK_TIMESTAMP()
					AND deleted_at IS NULL
				)
			AND (end_date > CLOCK_TIMESTAMP() OR end_date IS NULL)
			AND start_date < CLOCK_TIMESTAMP()
			AND deleted_at IS NULL 
			AND ((ARRAY_LENGTH('{student-id}'::text[], 1) IS NULL) or (sesh1.student_id = ANY('{student-id}')))
			GROUP BY sesh1.student_id
		),
		upsert_students AS (
			SELECT student_enrollment_status_history.student_id, COALESCE(inactive_students.latest_start_date,NULL) as deactivation_date
			FROM student_enrollment_status_history left join inactive_students on student_enrollment_status_history.student_id = inactive_students.student_id
			WHERE student_enrollment_status_history.deleted_at IS NULL
			AND ((ARRAY_LENGTH('{student-id}'::text[], 1) IS NULL) or (student_enrollment_status_history.student_id = ANY('{student-id}')))
			GROUP BY student_enrollment_status_history.student_id,inactive_students.latest_start_date
		) UPDATE users SET deactivated_at = upsert_students.deactivation_date FROM upsert_students
		WHERE users.user_id = upsert_students.student_id 
		AND users.deactivated_at IS DISTINCT FROM upsert_students.deactivation_date`
	t.Run("the query is called correctly", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()

		mockDB.DB.On(
			"Exec", mock.Anything, expectedQuery, database.TextArray(
				[]string{"withdrawn", "non-potential", "graduated"},
			),
		).Once().Return(pgconn.CommandTag("1"), nil)

		err := repo.UpdateStudentStatusBasedEnrollmentStatus(ctx, mockDB.DB, []string{"student-id"}, []string{"withdrawn", "non-potential", "graduated"})
		assert.Nil(t, err)
	})
	t.Run("the query is throw error correctly", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On(
			"Exec", mock.Anything, expectedQuery, database.TextArray(
				[]string{"withdrawn", "non-potential", "graduated"},
			),
		).Once().Return(nil, errors.New("query error"))

		err := repo.UpdateStudentStatusBasedEnrollmentStatus(ctx, mockDB.DB, []string{"student-id"}, []string{"withdrawn", "non-potential", "graduated"})
		assert.Error(t, err)
	})
}

func TestDomainEnrollmentStatusHistoryRepo_GetSameStartDateEnrollmentStatusHistory(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, enrollmentStatus := NewEnrollmentStatusHistory(entity.DefaultDomainEnrollmentStatusHistory{}).FieldMap()
	argsEnrollmentStatus := append([]interface{}{}, genSliceMock(len(enrollmentStatus))...)

	enrollmentStatusHistory := entity.DefaultDomainEnrollmentStatusHistory{}
	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainEnrollmentStatusHistoryRepoWithSqlMock()
		mockDB.DB.On(
			"Query", mock.Anything, mock.AnythingOfType("string"),
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsEnrollmentStatus...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		enrollmentStatusHistories, err := repo.GetSameStartDateEnrollmentStatusHistory(ctx, mockDB.DB, enrollmentStatusHistory)
		assert.Nil(t, err)
		assert.NotNil(t, enrollmentStatusHistories)
	})
}
