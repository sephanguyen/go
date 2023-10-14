package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func TestStudentPackageClassRepo_BulkUpsert(t *testing.T) {
	t.Parallel()
	db := new(mock_database.Tx)
	repo := &StudentPackageClassRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentPackageClass{
				{
					StudentPackageID: database.Text("student-package-id-1"),
					CourseID:         database.Text("course-id-1"),
					StudentID:        database.Text("student-id-1"),
					LocationID:       database.Text("location-id-1"),
					ClassID:          database.Text("class-id-1"),
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
			req: []*entities.StudentPackageClass{
				{
					StudentPackageID: database.Text("student-package-id-1"),
					CourseID:         database.Text("course-id-1"),
					StudentID:        database.Text("student-id-1"),
					LocationID:       database.Text("location-id-1"),
					ClassID:          database.Text("class-id-1"),
				},
				{
					StudentPackageID: database.Text("student-package-id-2"),
					CourseID:         database.Text("course-id-2"),
					StudentID:        database.Text("student-id-2"),
					LocationID:       database.Text("location-id-2"),
					ClassID:          database.Text("class-id-2"),
				},
				{
					StudentPackageID: database.Text("student-package-id-3"),
					CourseID:         database.Text("course-id-3"),
					StudentID:        database.Text("student-id-3"),
					LocationID:       database.Text("location-id-3"),
					ClassID:          database.Text("class-id-3"),
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
		err := repo.BulkUpsert(ctx, db, testCase.req.([]*entities.StudentPackageClass))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPackageClassRepo_DeleteByStudentPackageIDs(t *testing.T) {
	db := &mock_database.QueryExecer{}
	r := &StudentPackageClassRepo{}
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

func TestStudentPackageClassRepo_DeleteByStudentPackageIDAndCourseID(t *testing.T) {
	db := &mock_database.QueryExecer{}
	r := &StudentPackageClassRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "err exec",
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				db.On("Exec", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := r.DeleteByStudentPackageIDAndCourseID(ctx, db, "student_package_id", "course_id")
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
