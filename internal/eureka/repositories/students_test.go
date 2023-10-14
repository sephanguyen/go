package repositories

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentRepo_FindStudentsByCourseID(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			req:         database.Text("course-id"),
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			req:         database.Text("course-id"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(pgtype.Text)
		_, err := topicRepo.FindStudentsByCourseID(ctx, mockDB.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentRepo_FindStudentsByClassIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			req:         database.TextArray([]string{"class-id"}),
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			req:         database.TextArray([]string{"class-id"}),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(pgtype.TextArray)
		_, err := topicRepo.FindStudentsByClassIDs(ctx, mockDB.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentRepo_FindClassesByCourseID(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			req:         database.Text("course-id"),
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			req:         database.Text("course-id"),
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.(pgtype.Text)
		_, err := topicRepo.FindClassesByCourseID(ctx, mockDB.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentRepo_FindStudentsByCourseLocation(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := topicRepo.FindStudentsByCourseLocation(ctx, mockDB.DB, database.Text("course-id"), database.TextArray([]string{"location-id"}))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentRepo_FindStudentsByLocation(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			expectedErr: fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := topicRepo.FindStudentsByLocation(ctx, mockDB.DB, database.TextArray([]string{"location-id"}))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentRepo_FilterOutDeletedStudentIDs(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	topicRepo := &StudentRepo{}
	studentIDs := []string{"student-id-0", "student-id-1"}
	testCases := []TestCase{
		{
			name:        "retrieve error",
			req:         studentIDs,
			expectedErr: fmt.Errorf("db.Query: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		{
			name:        "find success",
			req:         studentIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("Query").Once().Return(rows, nil)
				rows.On("Close").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", mock.Anything).Once().Return(nil)
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(false)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		req := testCase.req.([]string)
		_, err := topicRepo.FilterOutDeletedStudentIDs(ctx, mockDB.DB, req)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
