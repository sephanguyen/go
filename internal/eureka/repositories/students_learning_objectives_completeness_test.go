package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentsLearningObjectivesCompletenessRepoWithSQLMock() (*StudentsLearningObjectivesCompletenessRepo, *testutil.MockDB) {
	r := &StudentsLearningObjectivesCompletenessRepo{}
	return r, testutil.NewMockDB()
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

// TODO:
func TestUpsertLOCompleteness(t *testing.T) {

}
