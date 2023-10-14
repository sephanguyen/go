package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSchoolHistoryRepo_SoftDeleteByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := schoolHistoryRepoWithSqlMock()

	studentIDs := database.TextArray([]string{"studentID-1", "studentID-2"})

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := r.SoftDeleteByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentIDs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := r.SoftDeleteByStudentIDs(ctx, mockDB.DB, studentIDs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "school_history")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestSchoolHistoryRepo_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	schoolHistoryRepo := &SchoolHistoryRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.SchoolHistory{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					SchoolID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			name: "happy case: upsert multiple user_access_paths",
			req: []*entity.SchoolHistory{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					SchoolID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					SchoolID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			req: []*entity.SchoolHistory{
				{
					StudentID:    pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					SchoolID:     pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := schoolHistoryRepo.Upsert(ctx, db, testCase.req.([]*entity.SchoolHistory))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func schoolHistoryRepoWithSqlMock() (*SchoolHistoryRepo, *testutil.MockDB) {
	repo := &SchoolHistoryRepo{}
	return repo, testutil.NewMockDB()
}

func TestSchoolHistoryRepo_GetByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := pgtype.Text{}
	studentID.Set(uuid.NewString())
	_, schoolHistoryValues := (&entity.SchoolHistory{}).FieldMap()
	argsSchoolHistories := append([]interface{}{}, genSliceMock(len(schoolHistoryValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistorys, err := repo.GetByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, schoolHistorys)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		schoolHistorys, err := repo.GetByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistorys)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistorys, err := repo.GetByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistorys)
	})
}

func TestSchoolHistoryRepo_GetCurrentSchoolByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := pgtype.Text{}
	studentID.Set(uuid.NewString())
	_, schoolHistoryValues := (&entity.SchoolHistory{}).FieldMap()
	argsSchoolHistories := append([]interface{}{}, genSliceMock(len(schoolHistoryValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistorys, err := repo.GetByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.NotNil(t, schoolHistorys)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		schoolHistorys, err := repo.GetCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistorys)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &studentID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistorys, err := repo.GetCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistorys)
	})
}

func TestSchoolHistoryRepo_GetSchoolHistoryByGradeIDAndStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	gradeID := pgtype.Text{}
	gradeID.Set(uuid.NewString())
	studentID := pgtype.Text{}
	studentID.Set(uuid.NewString())
	isCurrent := database.Bool(false)
	_, schoolHistoryValues := (&entity.SchoolHistory{}).FieldMap()
	argsSchoolHistories := append([]interface{}{}, genSliceMock(len(schoolHistoryValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &gradeID, &studentID, &isCurrent).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistories, err := repo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, mockDB.DB, gradeID, studentID, isCurrent)
		assert.Nil(t, err)
		assert.NotNil(t, schoolHistories)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &gradeID, &studentID, &isCurrent).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		schoolHistories, err := repo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, mockDB.DB, gradeID, studentID, isCurrent)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistories)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := schoolHistoryRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &gradeID, &studentID, &isCurrent).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsSchoolHistories...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		schoolHistories, err := repo.GetSchoolHistoriesByGradeIDAndStudentID(ctx, mockDB.DB, gradeID, studentID, isCurrent)
		assert.NotNil(t, err)
		assert.Nil(t, schoolHistories)
	})
}

func TestSchoolHistoryRepo_SetCurrentSchoolByStudentIDAndSchoolID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolHistoryRepo, mockDB := schoolHistoryRepoWithSqlMock()

	schoolID := database.Text("school-ID")
	studentID := database.Text("student-ID")

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &schoolID, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := schoolHistoryRepo.SetCurrentSchoolByStudentIDAndSchoolID(ctx, mockDB.DB, schoolID, studentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &schoolID, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := schoolHistoryRepo.SetCurrentSchoolByStudentIDAndSchoolID(ctx, mockDB.DB, schoolID, studentID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "school_history")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "is_current")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"school_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestSchoolHistoryRepo_RemoveCurrentSchoolByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolHistoryRepo, mockDB := schoolHistoryRepoWithSqlMock()

	studentID := database.Text("student-ID")

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := schoolHistoryRepo.RemoveCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := schoolHistoryRepo.RemoveCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "school_history")
		// move primaryField to the last
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestSchoolHistoryRepo_UnsetCurrentSchoolByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolHistoryRepo, mockDB := schoolHistoryRepoWithSqlMock()

	studentID := database.Text("student-ID")

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := schoolHistoryRepo.UnsetCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studentID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := schoolHistoryRepo.UnsetCurrentSchoolByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "school_history")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "is_current")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestSchoolHistoryRepo_SetCurrentSchool(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	schoolHistoryRepo, mockDB := schoolHistoryRepoWithSqlMock()

	orgs := database.Text("manabie-school")

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &orgs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := schoolHistoryRepo.SetCurrentSchool(ctx, mockDB.DB, orgs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &orgs)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := schoolHistoryRepo.SetCurrentSchool(ctx, mockDB.DB, orgs)
		assert.Nil(t, err)

		mockDB.RawStmt.AssertUpdatedTable(t, "school_history")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "is_current")
	})
}
