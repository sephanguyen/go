package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func StudentPackageRepoWithSqlMock() (*StudentPackageRepo, *testutil.MockDB) {
	r := &StudentPackageRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentPackageRepo_CurrentPackage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentPackageRepoWithSqlMock()

	userID := ksuid.New().String()
	pgUserID := database.Text(userID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgUserID,
		)

		studentPackages, err := r.CurrentPackage(ctx, mockDB.DB, pgUserID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentPackages)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgUserID,
		)

		e := &entities.StudentPackage{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		users, err := r.CurrentPackage(ctx, mockDB.DB, pgUserID)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.StudentPackage{e}, users)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"now":        {HasNullTest: false, BetweenExpr: &testutil.BetweenExpr{Field: "now", Args: []string{"start_at", "end_at"}}},
			"is_active":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{Type: "bool", Value: true}},
		})
	})
}

func TestStudentPackageRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentPackageRepoWithSqlMock()

	t.Run("err insert", func(t *testing.T) {
		e := &entities.StudentPackage{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag(""), puddle.ErrClosedPool, args...)

		err := r.Insert(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.StudentPackage{}
		_, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Insert(ctx, mockDB.DB, e)
		assert.Equal(t, fmt.Errorf("cannot create student_packages"), err)
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.StudentPackage{}
		fields, values := e.FieldMap()

		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Insert(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertInsertedTable(t, e.TableName())
		mockDB.RawStmt.AssertInsertedFields(t, fields...)
	})
}

func TestStudentPackageRepo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentPackageRepoWithSqlMock()

	t.Run("err update", func(t *testing.T) {
		e := &entities.StudentPackage{}
		_, values := e.FieldMap()

		// move primaryField to the last
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, append(values[1:], values[0])...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.Update(ctx, mockDB.DB, e)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("no rows affected", func(t *testing.T) {
		e := &entities.StudentPackage{}
		_, values := e.FieldMap()

		// move primaryField to the last
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, append(values[1:], values[0])...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("0"), nil, args...)

		err := r.Update(ctx, mockDB.DB, e)
		assert.Equal(t, fmt.Errorf("cannot update student_packages"), err)
	})

	t.Run("success", func(t *testing.T) {
		e := &entities.StudentPackage{}
		fields, values := e.FieldMap()

		// move primaryField to the last
		args := append([]interface{}{ctx, mock.AnythingOfType("string")}, append(values[1:], values[0])...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.Update(ctx, mockDB.DB, e)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, e.TableName())
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, fields[1:]...)
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_package_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 10}},
		})
	})
}

func TestStudentPackageRepo_Get(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentPackageRepoWithSqlMock()

	studentPackageID := ksuid.New().String()
	pgStudentPackageID := database.Text(studentPackageID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgStudentPackageID,
		)

		studentPackages, err := r.Get(ctx, mockDB.DB, pgStudentPackageID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, studentPackages)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgStudentPackageID,
		)

		e := &entities.StudentPackage{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		sPackage, err := r.Get(ctx, mockDB.DB, pgStudentPackageID)
		assert.Nil(t, err)
		assert.Equal(t, e, sPackage)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_package_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestCourseStudentRepo_BulkInsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudentRepo := &StudentPackageRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentPackage{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entities.StudentPackage{
				{
					ID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
				{
					ID: pgtype.Text{String: "3", Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, pgx.ErrTxClosed)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := courseStudentRepo.BulkInsert(ctx, db, testCase.req.([]*entities.StudentPackage))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestCourseStudentRepo_GetByStudentIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudentRepo := &StudentPackageRepo{}
	testCases := []struct {
		name        string
		req         pgtype.TextArray
		expectedErr error
		setup       func(context.Context)
	}{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"student_id_1", "student_id_2"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudentPackage{}
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray([]string{"student_id_1", "student_id_2"})).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", database.GetScanFields(e, database.GetFieldNames(e))...).Once().Return(nil)
				rows.On("Scan", database.GetScanFields(e, database.GetFieldNames(e))...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudentRepo.GetByStudentIDs(ctx, db, testCase.req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageRepo_GetByCourseIDAndLocationIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	courseStudentRepo := &StudentPackageRepo{}
	locationIDs := []string{constants.ManabieOrgLocation}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.Text("course_id_1"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudentPackage{}
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(locationIDs)).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", database.GetScanFields(e, database.GetFieldNames(e))...).Once().Return(nil)
				rows.On("Scan", database.GetScanFields(e, database.GetFieldNames(e))...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "error select no rows",
			req:         database.Text("course_id_2"),
			expectedErr: fmt.Errorf("%w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), database.TextArray(locationIDs)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := courseStudentRepo.GetByCourseIDAndLocationIDs(ctx, db, testCase.req.(pgtype.Text), database.TextArray(locationIDs))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentPackageRepo := &StudentPackageRepo{}
	studentId := database.Text("student_id_1")

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studentId,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentId).Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentPackageRepo.SoftDelete(ctx, db, testCase.req.(pgtype.Text))
		assert.Nil(t, err)
	}
}

func TestStudentPackageRepo_SoftDeleteByIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	studentPackageRepo := &StudentPackageRepo{}
	studentIds := database.TextArray([]string{"student_id_1", "student_id_2"})

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         studentIds,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentIds).Once().Return(nil, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentPackageRepo.SoftDeleteByIDs(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Nil(t, err)
	}
}

func TestStudentPackageRepo_GetByStudentPackageIDAndStudentIDAndCourseID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentPackageRepoWithSqlMock()

	studentPackageID := ksuid.New().String()
	studentID := ksuid.New().String()
	courseID := ksuid.New().String()

	pgStudentPackageID := database.Text(studentPackageID)
	pgStudentID := database.Text(studentID)
	pgCourseID := database.Text(courseID)

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			&pgStudentPackageID,
			&pgStudentID,
		)

		sp, err := r.GetByStudentPackageIDAndStudentIDAndCourseID(ctx, mockDB.DB, pgStudentPackageID, pgStudentID, pgCourseID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, sp)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"),
			&pgStudentPackageID,
			&pgStudentID,
		)

		e := &entities.StudentPackage{}
		fields, values := e.FieldMap()

		_ = e.ID.Set(ksuid.New().String())
		mockDB.MockScanFields(nil, fields, values)

		sp, err := r.GetByStudentPackageIDAndStudentIDAndCourseID(ctx, mockDB.DB, pgStudentPackageID, pgStudentID, pgCourseID)
		assert.Nil(t, err)
		assert.Equal(t, e, sp)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}
