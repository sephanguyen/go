package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentEnrollmentStatusHistoryRepo_Upsert(t *testing.T) {
	t.Parallel()
	studentEnrollmentStatusHistoryRepo := &StudentEnrollmentStatusHistoryRepo{}
	db := &mock_database.Ext{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entity.StudentEnrollmentStatusHistory{
				StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}

				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: &entity.StudentEnrollmentStatusHistory{
				StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: errors.New("db.Exec: closed pool"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name: "no rows affected",
			req: &entity.StudentEnrollmentStatusHistory{
				StudentID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: errors.New("cannot upsert StudentEnrollmentStatusHistory"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentEnrollmentStatusHistoryRepo.Upsert(ctx, db, testCase.req.(*entity.StudentEnrollmentStatusHistory))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentEnrollmentStatusHistoryRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	studentEnrollmentStatusHistoryRepo := &StudentEnrollmentStatusHistoryRepo{}
	studentIDs := database.TextArray([]string{"student-1", "student-2"})
	mockDB := &mock_database.Ext{}

	testCases := []TestCase{
		{
			name:        "error cannot delete student_enrollment_status_history",
			expectedErr: fmt.Errorf("cannot delete student_enrollment_status_history"),
			setup: func(ctx context.Context) {
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(nil, fmt.Errorf("cannot delete student_enrollment_status_history"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &studentIDs).Once().Return(cmdTag, nil)
				mockDB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentEnrollmentStatusHistoryRepo.SoftDelete(ctx, mockDB, studentIDs)
		assert.Equal(t, testCase.expectedErr, err)
	}
}
