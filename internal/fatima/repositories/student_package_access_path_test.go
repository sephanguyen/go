package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentPackageAccessPathRepoWithSqlMock() (*StudentPackageAccessPathRepo, *testutil.MockDB) {
	r := &StudentPackageAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestStudentPackageAccessPathRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	r := &StudentPackageAccessPathRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentPackageAccessPath{
				{
					StudentPackageID: database.Text("student-package-id-1"),
					CourseID:         database.Text("course-id-1"),
					StudentID:        database.Text("student-id-1"),
					LocationID:       database.Text("location-id-1"),
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
			req: []*entities.StudentPackageAccessPath{
				{
					StudentPackageID: database.Text("student-package-id-1"),
					CourseID:         database.Text("course-id-1"),
					StudentID:        database.Text("student-id-1"),
					LocationID:       database.Text("location-id-1"),
				},
				{
					StudentPackageID: database.Text("student-package-id-2"),
					CourseID:         database.Text("course-id-2"),
					StudentID:        database.Text("student-id-2"),
					LocationID:       database.Text("location-id-2"),
				},
				{
					StudentPackageID: database.Text("student-package-id-3"),
					CourseID:         database.Text("course-id-3"),
					StudentID:        database.Text("student-id-3"),
					LocationID:       database.Text("location-id-3"),
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
		err := r.BulkUpsert(ctx, db, testCase.req.([]*entities.StudentPackageAccessPath))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageAccessPathRepo_DeleteByStudentPackageIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	r := &StudentPackageAccessPathRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"student-package-id-1", "student-package-id-1"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "err exec",
			req:         database.TextArray([]string{"student-package-id-1", "student-package-id-1"}),
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.DeleteByStudentPackageIDs(ctx, db, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageAccessPathRepo_DeleteByStudentIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	r := &StudentPackageAccessPathRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.TextArray([]string{"student-id-1", "student-id-1"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "err exec",
			req:         database.TextArray([]string{"student-id-1", "student-id-1"}),
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.DeleteByStudentIDs(ctx, db, testCase.req.(pgtype.TextArray))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageAccessPathRepo_GetByCourseIDAndLocationIDs(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	r := &StudentPackageAccessPathRepo{}
	locationIDs := []string{constants.ManabieOrgLocation}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         database.Text("course_id_1"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				e := &entities.StudentPackageAccessPath{}
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text("course_id_1"), database.TextArray(locationIDs)).Once().Return(rows, nil)
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
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text("course_id_2"), database.TextArray(locationIDs)).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := r.GetByCourseIDAndLocationIDs(ctx, db, testCase.req.(pgtype.Text), database.TextArray(locationIDs))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
