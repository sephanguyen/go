package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsertLOCompleteness_Batch(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentsLearningObjectivesCompletenessRepo := &StudentsLearningObjectivesCompletenessRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentsLearningObjectivesCompleteness{
				{
					StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			req: []*entities.StudentsLearningObjectivesCompleteness{
				{
					StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
				{
					StudentID: pgtype.Text{String: "2", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentsLearningObjectivesCompletenessRepo.UpsertLOCompleteness(ctx, db, testCase.req.([]*entities.StudentsLearningObjectivesCompleteness))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
	return
}

//  test repo with mock
func StudentsLearningObjectivesCompletenessRepoWithSQLMock() (*StudentsLearningObjectivesCompletenessRepo, *testutil.MockDB) {
	r := &StudentsLearningObjectivesCompletenessRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentsLearningObjectivesCompleteness_UpdateHighestQuizScore(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentsLearningObjectivesCompletenessRepoWithSQLMock()
	loID := database.Text("loID")
	studentID := database.Text("studentID")
	newScore := database.Float4(50)
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.MockExecArgs(t, cmdTag, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
			studentID,
			newScore,
			mock.Anything,
			mock.Anything,
		)

		err := r.UpsertHighestQuizScore(ctx, mockDB.DB, loID, studentID, newScore)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.MockExecArgs(t, cmdTag, nil, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
			studentID,
			newScore,
			mock.Anything,
			mock.Anything,
		)

		err := r.UpsertHighestQuizScore(ctx, mockDB.DB, loID, studentID, newScore)
		assert.True(t, errors.Is(err, nil))
	})
}

func TestStudentsLearningObjectivesCompleteness_UpsertLOFirstQuizCompleteness(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentsLearningObjectivesCompletenessRepoWithSQLMock()
	loID := database.Text("loID")
	studentID := database.Text("studentID")
	firstScore := database.Float4(50)
	t.Run("error", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.MockExecArgs(t, cmdTag, puddle.ErrClosedPool, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
			studentID,
			firstScore,
			mock.Anything,
			mock.Anything,
		)

		err := r.UpsertFirstQuizCompleteness(ctx, mockDB.DB, loID, studentID, firstScore)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})
	t.Run("success", func(t *testing.T) {
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.MockExecArgs(t, cmdTag, nil, mock.Anything,
			mock.AnythingOfType("string"),
			loID,
			studentID,
			firstScore,
			mock.Anything,
			mock.Anything,
		)

		err := r.UpsertFirstQuizCompleteness(ctx, mockDB.DB, loID, studentID, firstScore)
		assert.True(t, errors.Is(err, nil))
	})
}

func TestStudentsLearningObjectivesCompletenessRepo_CountTotalLOsFinished(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := StudentsLearningObjectivesCompletenessRepoWithSQLMock()
	userID := database.Text("mock-student-id")
	fromTimestamptz := database.Timestamptz(time.Now())
	toTimestamptz := database.Timestamptz(time.Now().Add(time.Hour))
	testcases := []struct {
		name        string
		ctx         context.Context
		studentID   pgtype.Text
		from        *pgtype.Timestamptz
		to          *pgtype.Timestamptz
		expectedErr error
		setup       func(context.Context)
	}{
		{
			name:        "from and to is nil",
			ctx:         ctx,
			studentID:   userID,
			from:        nil,
			to:          nil,
			expectedErr: nil,
			setup: func(c context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
				mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "from and to is nil but error",
			ctx:         ctx,
			studentID:   userID,
			from:        nil,
			to:          nil,
			expectedErr: fmt.Errorf("row.Scan: %w", puddle.ErrClosedPool),
			setup: func(c context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
				mockDB.Row.On("Scan", mock.Anything).Once().Return(puddle.ErrClosedPool)
			},
		},
		{
			name:        "from and to existed",
			ctx:         ctx,
			studentID:   userID,
			from:        &fromTimestamptz,
			to:          &toTimestamptz,
			expectedErr: nil,
			setup: func(c context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID, &fromTimestamptz, &toTimestamptz)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
				mockDB.Row.On("Scan", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "error in case `from` and `to` existed",
			ctx:         ctx,
			studentID:   userID,
			from:        &fromTimestamptz,
			to:          &toTimestamptz,
			expectedErr: fmt.Errorf("row.Scan: %w", puddle.ErrClosedPool),
			setup: func(c context.Context) {
				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, &userID, &fromTimestamptz, &toTimestamptz)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row, nil)
				mockDB.Row.On("Scan", mock.Anything).Once().Return(puddle.ErrClosedPool)
			},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.setup(testcase.ctx)
			_, err := r.CountTotalLOsFinished(testcase.ctx, mockDB.DB, testcase.studentID, testcase.from, testcase.to)
			assert.Equal(t, testcase.expectedErr, err)
		})
	}
}

func TestStudentsLearningObjectivesCompletenessRepo_Find(t *testing.T) {
	r, mockDB := StudentsLearningObjectivesCompletenessRepoWithSQLMock()
	userID := database.Text("mock-student-id")
	pgtypeLOIds := database.TextArray([]string{"loIDs"})
	testCases := []TestCase{
		{
			name: "err select",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything,
					mock.AnythingOfType("string"),
					&userID,
					&pgtypeLOIds,
				)
			},
			expectedErr: fmt.Errorf("r.DB.QueryEx: %w", puddle.ErrClosedPool),
		},
		{
			name: "success with find",
			setup: func(ctx context.Context) {
				mockDB.MockQueryArgs(t, nil, mock.Anything,
					mock.AnythingOfType("string"),
					&userID,
					&pgtypeLOIds,
				)
				lo := &entities.StudentsLearningObjectivesCompleteness{}
				fields := database.GetFieldNames(lo)
				p := new(entities.StudentsLearningObjectivesCompleteness)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", database.GetScanFields(p, fields)...).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.Find(ctx, mockDB.DB, userID, pgtypeLOIds)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
