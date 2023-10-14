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
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFeedbackSessionRepo_GetOneBySubmissionID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	repo := &FeedbackSessionRepo{}
	key := "key"

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Text(key))
		expectedErr := errors.NewDBError("FeedbackSessionRepo.GetOneBySubmissionID",
			fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool))

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
		expectedErr := errors.NewNoRowsExistedError("FeedbackSessionRepo.GetOneBySubmissionID",
			fmt.Errorf("err db.Query: %w", pgx.ErrNoRows))

		// act
		actual, err := repo.GetOneBySubmissionID(ctx, mockDB.DB, key)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded", func(t *testing.T) {
		// arrange
		e := &dto.FeedbackSession{}
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
			"deleted_at":    {HasNullTest: true},
			"submission_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
		})
	})
}

func TestFeedbackSessionRepo_GetManyBySubmissionIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	submissionIDs := []string{idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()}

	t.Run("query failed returns DB error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"),
			submissionIDs[0],
			submissionIDs[1],
			submissionIDs[2],
			submissionIDs[3],
			submissionIDs[4],
		)
		expectedErr := errors.NewDBError("FeedbackSessionRepo.GetManyBySubmissionIDs", puddle.ErrClosedPool)

		// act
		actual, err := repo.GetManyBySubmissionIDs(ctx, mockDB.DB, submissionIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed returns Conversion error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}
		holder := dto.FeedbackSession{}
		_, val := holder.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"),
				submissionIDs[0],
				submissionIDs[1],
				submissionIDs[2],
				submissionIDs[3],
				submissionIDs[4]).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("FeedbackSessionRepo.scanFeedbackSessions", scanErr)

		// act
		actual, err := repo.GetManyBySubmissionIDs(ctx, mockDB.DB, submissionIDs)

		// assert
		assert.Nil(t, actual)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("return empty slice when there is no row", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.AnythingOfType("string"),
			submissionIDs[0],
			submissionIDs[1],
			submissionIDs[2],
			submissionIDs[3],
			submissionIDs[4],
		)

		// act
		actual, err := repo.GetManyBySubmissionIDs(ctx, mockDB.DB, submissionIDs)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, []domain.FeedbackSession{}, actual)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded returns nil err", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.AnythingOfType("string"),
			submissionIDs[0],
			submissionIDs[1],
			submissionIDs[2],
			submissionIDs[3],
			submissionIDs[4])

		e := &dto.FeedbackSession{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyBySubmissionIDs(ctx, mockDB.DB, submissionIDs)

		// assert
		assert.Nil(t, err)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
	})
}

func TestFeedbackSessionRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case: create successful", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}

		now := time.Now()
		feedback := domain.FeedbackSession{
			ID:           uuid.New().String(),
			SubmissionID: idutil.ULIDNow(),
			CreatedBy:    "SOME ONE",
			CreatedAt:    now,
		}
		feedbackDto := dto.FeedbackSession{
			ID:           database.Text(feedback.ID),
			SubmissionID: database.Text(feedback.SubmissionID),
			CreatedBy:    database.Text(feedback.CreatedBy),
			BaseEntity: dto.BaseEntity{
				CreatedAt: database.Timestamptz(feedback.CreatedAt),
				UpdatedAt: database.Timestamptz(feedback.CreatedAt),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
		}

		_, values := feedbackDto.FieldMap()

		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		// act
		err := repo.Insert(ctx, mockDB.DB, feedback)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("return error when an error occurred", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}

		now := time.Now()
		feedback := domain.FeedbackSession{
			ID:           uuid.New().String(),
			SubmissionID: idutil.ULIDNow(),
			CreatedBy:    "SOME ONE",
			CreatedAt:    now,
		}
		feedbackDto := dto.FeedbackSession{
			ID:           database.Text(feedback.ID),
			SubmissionID: database.Text(feedback.SubmissionID),
			CreatedBy:    database.Text(feedback.CreatedBy),
			BaseEntity: dto.BaseEntity{
				CreatedAt: database.Timestamptz(feedback.CreatedAt),
				UpdatedAt: database.Timestamptz(feedback.CreatedAt),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
		}

		_, values := feedbackDto.FieldMap()
		args := append([]interface{}{mock.Anything, mock.Anything}, values...)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), pgx.ErrTxClosed, args...)
		expectedErr := errors.NewDBError("FeedbackSessionRepo.Insert", pgx.ErrTxClosed)

		// act
		err := repo.Insert(ctx, mockDB.DB, feedback)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("return error when feedback id is not uuid", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &FeedbackSessionRepo{}

		now := time.Now()
		feedback := domain.FeedbackSession{
			ID:           "SOMETHING_ID",
			SubmissionID: idutil.ULIDNow(),
			CreatedBy:    "SOME ONE",
			CreatedAt:    now,
		}

		expectedErr := errors.NewValidationError("Feedback session id must be UUID", nil)

		// act
		err := repo.Insert(ctx, mockDB.DB, feedback)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
