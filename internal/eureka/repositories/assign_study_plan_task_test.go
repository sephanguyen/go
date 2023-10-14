package repositories

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAssignStudyPlanTask(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignStudyPlanTaskRepo := &AssignStudyPlanTaskRepo{}
	testCases := []TestCase{
		{
			name: "error scan",
			req: &entities.AssignStudyPlanTask{
				ID:           database.Text("id"),
				StudyPlanIDs: database.TextArray([]string{"study-plan-id-1"}),
				Status:       database.Text("status"),
				CourseID:     database.Text("course-id"),
			},
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				row := &mock_database.Row{}
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
		{
			name: "happy case",
			req: &entities.AssignStudyPlanTask{
				ID:           database.Text("id"),
				StudyPlanIDs: database.TextArray([]string{"study-plan-id-1"}),
				Status:       database.Text("status"),
				CourseID:     database.Text("course-id"),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				row := &mock_database.Row{}
				db.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(row)
				row.On("Scan", mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := assignStudyPlanTaskRepo.Create(ctx, db, testCase.req.(*entities.AssignStudyPlanTask))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUpdateStatus(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignStudyPlanTaskRepo := &AssignStudyPlanTaskRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "happy case",
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignStudyPlanTaskRepo.UpdateStatus(ctx, db, database.Text("id"), database.Text("status"))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUpdateDetailError(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	assignStudyPlanTaskRepo := &AssignStudyPlanTaskRepo{}
	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "happy case",
			expectedErr: pgx.ErrNoRows,
			setup: func(ctx context.Context) {
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := assignStudyPlanTaskRepo.UpdateDetailError(ctx, db, &AssignStudyPlanTaskDetailErrorArgs{
			ID:          database.Text("id"),
			Status:      database.Text("status"),
			ErrorDetail: database.Text("error_detail"),
		})
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
