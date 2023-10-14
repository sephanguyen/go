package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func setupMockForSubmissionRepo() (*SubmissionRepo, *testutil.MockDB) {
	r := &SubmissionRepo{}
	return r, testutil.NewMockDB()
}

func TestSubmissionRepo_GetOneBySessionID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	repo := &SubmissionRepo{}
	key := "key"

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Text(key))
		expectedErr := errors.NewDBError("SubmissionRepo.GetOneBySessionID", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool))

		// act
		_, err := repo.GetOneBySessionID(ctx, mockDB.DB, key)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("not found return no rows exists", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(key))
		expectedErr := errors.NewNoRowsExistedError("SubmissionRepo.GetOneBySessionID", fmt.Errorf("err db.Query: %w", pgx.ErrNoRows))

		// act
		_, err := repo.GetOneBySessionID(ctx, mockDB.DB, key)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded", func(t *testing.T) {
		// arrange
		e := &dto.Submission{}
		fields, values := e.FieldMap()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(key))
		mockDB.MockScanFields(nil, fields, values)

		// act
		_, err := repo.GetOneBySessionID(ctx, mockDB.DB, key)

		// assert
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"session_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestSubmissionRepo_GetOneBySubmissionID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	repo := &SubmissionRepo{}
	key := "key"

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Text(key))
		expectedErr := errors.NewDBError("SubmissionRepo.GetOneBySubmissionID", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool))

		// act
		actual, err := repo.GetOneBySubmissionID(ctx, mockDB.DB, key)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("not found return no rows exists", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(key))
		expectedErr := errors.NewNoRowsExistedError("SubmissionRepo.GetOneBySubmissionID", fmt.Errorf("err db.Query: %w", pgx.ErrNoRows))

		// act
		actual, err := repo.GetOneBySubmissionID(ctx, mockDB.DB, key)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded", func(t *testing.T) {
		// arrange
		e := &dto.Submission{}
		fields, values := e.FieldMap()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(key))
		mockDB.MockScanFields(nil, fields, values)

		// act
		_, err := repo.GetOneBySubmissionID(ctx, mockDB.DB, key)

		// assert
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at": {HasNullTest: true},
			"id":         {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestSubmissionRepo_GetManyBySessionIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sessionIDs := []string{idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()}

	t.Run("query failed returns DB error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"),
			sessionIDs[0],
			sessionIDs[1],
			sessionIDs[2],
			sessionIDs[3],
			sessionIDs[4],
		)
		expectedErr := errors.NewDBError("SubmissionRepo.GetManyBySessionIDs", puddle.ErrClosedPool)

		// act
		actual, err := repo.GetManyBySessionIDs(ctx, mockDB.DB, sessionIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed returns Conversion error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		holder := dto.Submission{}
		_, val := holder.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"),
				sessionIDs[0],
				sessionIDs[1],
				sessionIDs[2],
				sessionIDs[3],
				sessionIDs[4]).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("SubmissionRepo.scanSubmissions", scanErr)

		// act
		actual, err := repo.GetManyBySessionIDs(ctx, mockDB.DB, sessionIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded returns nil err", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			sessionIDs[0],
			sessionIDs[1],
			sessionIDs[2],
			sessionIDs[3],
			sessionIDs[4])

		e := &dto.Submission{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyBySessionIDs(ctx, mockDB.DB, sessionIDs)

		// assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestSubmissionRepo_GetManyByAssessments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID, asmID := idutil.ULIDNow(), idutil.ULIDNow()

	t.Run("query failed returns DB error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"),
			database.Text(studentID),
			database.Text(asmID),
		)
		expectedErr := errors.NewDBError("SubmissionRepo.GetManyByAssessments", puddle.ErrClosedPool)

		// act
		actual, err := repo.GetManyByAssessments(ctx, mockDB.DB, studentID, asmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed returns Conversion error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		holder := dto.Submission{}
		_, val := holder.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"),
				database.Text(studentID),
				database.Text(asmID),
			).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("SubmissionRepo.scanSubmissions", scanErr)

		// act
		actual, err := repo.GetManyByAssessments(ctx, mockDB.DB, studentID, asmID)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded returns nil err", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			database.Text(studentID),
			database.Text(asmID))

		e := &dto.Submission{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByAssessments(ctx, mockDB.DB, studentID, asmID)

		// assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestSubmissionRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}

		now := time.Now()
		submissionDTO := dto.Submission{}
		database.AllNullEntity(&submissionDTO)
		_ = multierr.Combine(
			submissionDTO.ID.Set("id"),
			submissionDTO.SessionID.Set("session_id"),
			submissionDTO.AssessmentID.Set("assessment_id"),
			submissionDTO.StudentID.Set("student_id"),
			submissionDTO.GradingStatus.Set(domain.GradingStatusNotMarked),
			submissionDTO.MaxScore.Set(10),
			submissionDTO.GradedScore.Set(5),
			submissionDTO.CompletedAt.Set(now),
			submissionDTO.CreatedAt.Set(now),
			submissionDTO.UpdatedAt.Set(now),
		)
		_, values := submissionDTO.FieldMap()

		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// actual
		submission := domain.Submission{
			ID:            "id",
			SessionID:     "session_id",
			AssessmentID:  "assessment_id",
			StudentID:     "student_id",
			GradingStatus: domain.GradingStatusNotMarked,
			MaxScore:      10,
			GradedScore:   5,
			CompletedAt:   now,
		}

		err := repo.Insert(ctx, mockDB.DB, now, submission)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("unexpected case", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &SubmissionRepo{}

		now := time.Now()
		submissionDTO := dto.Submission{}
		database.AllNullEntity(&submissionDTO)
		_ = multierr.Combine(
			submissionDTO.ID.Set("id"),
			submissionDTO.SessionID.Set("session_id"),
			submissionDTO.AssessmentID.Set("assessment_id"),
			submissionDTO.StudentID.Set("student_id"),
			submissionDTO.GradingStatus.Set(domain.GradingStatusNotMarked),
			submissionDTO.MaxScore.Set(10),
			submissionDTO.GradedScore.Set(5),
			submissionDTO.CompletedAt.Set(now),
			submissionDTO.CreatedAt.Set(now),
			submissionDTO.UpdatedAt.Set(now),
		)
		_, values := submissionDTO.FieldMap()

		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)

		expectedErr := errors.NewDBError("database.Insert", pgx.ErrTxClosed)

		// actual
		submission := domain.Submission{
			ID:            "id",
			SessionID:     "session_id",
			AssessmentID:  "assessment_id",
			StudentID:     "student_id",
			GradingStatus: domain.GradingStatusNotMarked,
			MaxScore:      10,
			GradedScore:   5,
			CompletedAt:   now,
		}

		err := repo.Insert(ctx, mockDB.DB, now, submission)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestSubmissionRepo_UpdateAllocateMarkerSubmissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	submissionRepo, mockDB := setupMockForSubmissionRepo()

	testCases := []TestCase{
		{
			name: "Happy case",
			req: []domain.Submission{
				{
					ID:                "submission_1",
					AllocatedMarkerID: "marker_1",
				},
				{
					ID:                "submission_2",
					AllocatedMarkerID: "",
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))

				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
			expectedResp: nil,
		},
		{
			name: "Batch Exec error",
			req: []domain.Submission{
				{
					ID:                "submission_1",
					AllocatedMarkerID: "marker_1",
				},
				{
					ID:                "submission_2",
					AllocatedMarkerID: "",
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))

				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)

				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
			expectedErr: errors.NewDBError("SubmissionRepo.UpdateAllocateMarkerSubmissions Exec", puddle.ErrClosedPool),
		},
		{
			name: "Batch Exec with no RowsAffected",
			req: []domain.Submission{
				{
					ID:                "submission_1",
					AllocatedMarkerID: "marker_1",
				},
				{
					ID:                "submission_2",
					AllocatedMarkerID: "",
				},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))

				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)

				batchResults.On("Close").Once().Return(nil)
			},
			expectedErr: errors.NewNoRowsUpdatedError("SubmissionRepo.UpdateAllocateMarkerSubmissions", nil),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)

			err := submissionRepo.UpdateAllocateMarkerSubmissions(ctx, mockDB.DB, testCase.req.([]domain.Submission))

			if testCase.expectedErr != nil {
				assert.Equal(t, err, testCase.expectedErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
