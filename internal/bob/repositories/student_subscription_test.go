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
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func StudentSubscriptionRepoWithSqlMock() (*StudentSubscriptionRepo, *testutil.MockDB) {
	r := &StudentSubscriptionRepo{}
	return r, testutil.NewMockDB()
}
func TestStudentSubscriptionRepo_BulkInsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	studentSubscriptionRepo := &StudentSubscriptionRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentSubscription{
				{
					StudentSubscriptionID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				db.On("SoftDelete", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "insert duplicate student subscription case",
			req: []*entities.StudentSubscription{
				{
					StudentSubscriptionID: pgtype.Text{String: "unit_test_student_subscription_id_1", Status: pgtype.Present},
					SubscriptionID:        pgtype.Text{String: "unit_test_subscription_id_2", Status: pgtype.Present},
					StartAt:               pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
					EndAt:                 pgtype.Timestamptz{Time: time.Date(2021, 12, 12, 0, 0, 0, 0, time.UTC)},
				},
				{
					StudentSubscriptionID: pgtype.Text{String: "unit_test_student_subscription_id_1", Status: pgtype.Present},
					SubscriptionID:        pgtype.Text{String: "unit_test_subscription_id_2", Status: pgtype.Present},
					StartAt:               pgtype.Timestamptz{Time: time.Date(2020, 12, 12, 0, 0, 0, 0, time.UTC)},
					EndAt:                 pgtype.Timestamptz{Time: time.Date(2030, 12, 12, 0, 0, 0, 0, time.UTC)},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				db.On("SoftDelete", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := studentSubscriptionRepo.BulkUpsert(ctx, db, testCase.req.([]*entities.StudentSubscription))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentSubscriptionRepo_DeleteByCourseIDAndStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentSubscriptionDTO := &entities.StudentSubscription{
		StudentID: database.Text("test-id"),
		CourseID:  database.Text("test-id"),
	}

	_, fieldMap := studentSubscriptionDTO.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentSubscriptionRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.DeleteByCourseIDAndStudentID(ctx, mockDB.DB, studentSubscriptionDTO.StudentID, studentSubscriptionDTO.CourseID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("err while deleting student_subscription", func(t *testing.T) {
		repo, mockDB := StudentSubscriptionRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.DeleteByCourseIDAndStudentID(ctx, mockDB.DB, studentSubscriptionDTO.StudentID, studentSubscriptionDTO.CourseID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err db.Exec: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}
