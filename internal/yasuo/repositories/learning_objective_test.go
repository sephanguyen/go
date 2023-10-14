package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func LearningObjectiveRepoWithSqlMock() (*LearningObjectiveRepo, *testutil.MockDB) {
	r := &LearningObjectiveRepo{}
	return r, testutil.NewMockDB()
}

func TestLearningObjectiveRepo_FindSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, mockDB := LearningObjectiveRepoWithSqlMock()
	ids := []string{"id", "id-1"}

	pgIDs := database.TextArray([]string{"id", "id-1"})

	t.Run("err select", func(t *testing.T) {
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, &pgIDs)

		schoolIDs, err := r.FindSchoolIDs(ctx, mockDB.DB, ids)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
		assert.Nil(t, schoolIDs)
	})

	t.Run("success with select all fields", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &pgIDs)

		e := &repositories_bob.EnSchoolID{}
		fields, values := e.FieldMap()
		e.SchoolID = 1

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		schoolIDs, err := r.FindSchoolIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.Equal(t, []int32{e.SchoolID}, schoolIDs)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)

		mockDB.RawStmt.AssertSelectedTable(t, "learning_objectives", "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"lo_id":      {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestLearningObjectiveRepo_SoftDeleteWithLoIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	LoRepo := &LearningObjectiveRepo{}

	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"lo_id-1", "lo_id-2"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "unexpected error",
			req:         database.TextArray([]string{"lo_id-1", "lo_id-2"}),
			expectedErr: fmt.Errorf("err db.Exec: %w", fmt.Errorf("unexpected error")),
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("unexpected error"))
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)

		_, err := LoRepo.SoftDeleteWithLoIDs(ctx, db, testCase.req.(pgtype.TextArray))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
