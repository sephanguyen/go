package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func studentSubmissionGradeRepoWithMock() (*StudentSubmissionGradeRepo, *testutil.MockDB) {
	r := &StudentSubmissionGradeRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentSubmissionGradeRepo_RetrieveByIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := studentSubmissionGradeRepoWithMock()
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})
	testCases := []TestCase{
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudentSubmissionGrade{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudentSubmissionGrade{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		e := &entities.StudentSubmissionGrade{}
		ids := testCase.req.(pgtype.TextArray)
		_, err := r.RetrieveByIDs(ctx, mockDB.DB, ids)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_submission_grade_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}

func TestStudentSubmissionGradeRepo_FindBySubmissionIDs(t *testing.T) {
	t.Parallel()
	r, mockDB := studentSubmissionGradeRepoWithMock()
	validReq := database.TextArray([]string{"valid-req-1", "valid-req-2"})
	testCases := []TestCase{
		{
			name:        "error no rows",
			req:         validReq,
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				e := &entities.StudentSubmissionGrade{}
				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudentSubmissionGrade{}
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, &validReq)
				fields, values := e.FieldMap()

				mockDB.MockScanArray(nil, fields, [][]interface{}{
					values,
				})
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		ids := testCase.req.(pgtype.TextArray)
		_, err := r.FindBySubmissionIDs(ctx, mockDB.DB, ids)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}

	return
}
