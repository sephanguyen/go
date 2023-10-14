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

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
)

func TestAssessmentSessionRepo_GetLatestByIdentity(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	repo := &AssessmentSessionRepo{}

	type Request struct {
		AssessmentID string
		UserID       string
	}
	request := Request{
		AssessmentID: "assessment_id",
		UserID:       "user_id",
	}

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name:    "happy case",
			Ctx:     ctx,
			Request: request,
			Setup: func(ctx context.Context) {
				assessmentSessionDto := dto.AssessmentSession{
					SessionID:    database.Text("session_id"),
					AssessmentID: database.Text("assessment_id"),
					UserID:       database.Text("user_id"),
					Status:       database.Text("COMPLETED"),
				}
				fields, values := assessmentSessionDto.FieldMap()

				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
					database.Text(request.AssessmentID), database.Text(request.UserID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Session{
				ID:           "session_id",
				AssessmentID: "assessment_id",
				UserID:       "user_id",
				Status:       domain.SessionStatusCompleted,
			},
			ExpectedError: nil,
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: request,
			Setup: func(ctx context.Context) {
				assessmentSessionDto := dto.AssessmentSession{
					SessionID:    database.Text("session_id"),
					AssessmentID: database.Text("assessment_id"),
					UserID:       database.Text("user_id"),
					Status:       database.Text("COMPLETED"),
				}
				fields, values := assessmentSessionDto.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything,
					database.Text(request.AssessmentID), database.Text(request.UserID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Session{},
			ExpectedError:    errors.NewNoRowsExistedError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrNoRows)),
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: request,
			Setup: func(ctx context.Context) {
				assessmentSessionDto := dto.AssessmentSession{}
				fields, values := assessmentSessionDto.FieldMap()

				mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything,
					database.Text(request.AssessmentID), database.Text(request.UserID))
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
			},
			ExpectedResponse: domain.Session{},
			ExpectedError:    errors.NewDBError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			res, err := repo.GetLatestByIdentity(tc.Ctx, mockDB.DB, tc.Request.(Request).AssessmentID, tc.Request.(Request).UserID)
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, tc.ExpectedResponse, res)
			}
		})
	}
}

func TestAssessmentSessionRepo_Insert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	repo := &AssessmentSessionRepo{}

	now := time.Now()
	req := domain.Session{
		AssessmentID: "assessment_id",
		UserID:       "user_id",
	}

	testCases := []struct {
		Name          string
		Ctx           context.Context
		Request       any
		Setup         func(ctx context.Context)
		ExpectedError error
	}{
		{
			Name:    "happy case",
			Ctx:     ctx,
			Request: req,
			Setup: func(ctx context.Context) {
				assessmentSessionDto := dto.AssessmentSession{}
				database.AllNullEntity(&assessmentSessionDto)
				_ = multierr.Combine(
					assessmentSessionDto.SessionID.Set(req.ID),
					assessmentSessionDto.AssessmentID.Set(req.AssessmentID),
					assessmentSessionDto.UserID.Set(req.UserID),
					assessmentSessionDto.Status.Set(""),
					assessmentSessionDto.CreatedAt.Set(now),
					assessmentSessionDto.UpdatedAt.Set(now),
				)
				_, values := assessmentSessionDto.FieldMap()

				args := append([]interface{}{mock.Anything, mock.Anything}, values...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
			ExpectedError: nil,
		},
		{
			Name:    "unexpected error",
			Ctx:     ctx,
			Request: req,
			Setup: func(ctx context.Context) {
				assessmentSessionDto := dto.AssessmentSession{}
				database.AllNullEntity(&assessmentSessionDto)
				_ = multierr.Combine(
					assessmentSessionDto.SessionID.Set(req.ID),
					assessmentSessionDto.AssessmentID.Set(req.AssessmentID),
					assessmentSessionDto.UserID.Set(req.UserID),
					assessmentSessionDto.Status.Set(""),
					assessmentSessionDto.CreatedAt.Set(now),
					assessmentSessionDto.UpdatedAt.Set(now),
				)
				_, values := assessmentSessionDto.FieldMap()

				args := append([]interface{}{mock.Anything, mock.Anything}, values...)
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("error execute query"), args...)
			},
			ExpectedError: errors.NewDBError("database.Insert", fmt.Errorf("error execute query")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			err := repo.Insert(tc.Ctx, mockDB.DB, now, tc.Request.(domain.Session))
			if err != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			}
		})
	}
}

func TestAssessmentSessionRepo_UpdateStatus(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentSessionRepo{}

		now := time.Now()
		session := domain.Session{
			ID:     "session_id",
			Status: domain.SessionStatusCompleted,
		}
		assessmentDto := dto.AssessmentSession{}
		err := assessmentDto.FromEntity(now, session)
		assert.Nil(t, err)

		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, mock.Anything, mock.Anything, &assessmentDto.Status, &assessmentDto.UpdatedAt, &assessmentDto.SessionID)

		// actual
		err = repo.UpdateStatus(ctx, mockDB.DB, now, session)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("unexpected error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentSessionRepo{}

		now := time.Now()
		session := domain.Session{
			ID:     "session_id",
			Status: domain.SessionStatusCompleted,
		}
		assessmentDto := dto.AssessmentSession{}
		err := assessmentDto.FromEntity(now, session)
		assert.Nil(t, err)

		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), fmt.Errorf("unexpected error"), mock.Anything, mock.Anything, &assessmentDto.Status, &assessmentDto.UpdatedAt, &assessmentDto.SessionID)
		expectedErr := errors.NewDBError("database.UpdateFields", fmt.Errorf("unexpected error"))

		// actual
		err = repo.UpdateStatus(ctx, mockDB.DB, now, session)

		// assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestAssessmentSessionRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentSessionRepo{}

		assessmentSessionDto := dto.AssessmentSession{
			SessionID:    database.Text("session_id"),
			AssessmentID: database.Text("assessment_id"),
			UserID:       database.Text("user_id"),
			Status:       database.Text("COMPLETED"),
		}
		fields, values := assessmentSessionDto.FieldMap()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text("session_id"))
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		expectedResult := domain.Session{
			ID:           "session_id",
			AssessmentID: "assessment_id",
			UserID:       "user_id",
			Status:       domain.SessionStatusCompleted,
		}

		// actual
		result, err := repo.GetByID(ctx, mockDB.DB, "session_id")

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("unexpected error", func(t *testing.T) {
		// arrange
		mockDB := testutil.NewMockDB()
		repo := &AssessmentSessionRepo{}

		assessmentSessionDto := dto.AssessmentSession{
			SessionID:    database.Text("session_id"),
			AssessmentID: database.Text("assessment_id"),
			UserID:       database.Text("user_id"),
			Status:       database.Text("COMPLETED"),
		}
		fields, values := assessmentSessionDto.FieldMap()

		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything, database.Text("session_id"))
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		expectedResult := domain.Session{}
		expectedErr := errors.NewDBError("database.Select", fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed))

		// actual
		result, err := repo.GetByID(ctx, mockDB.DB, "session_id")

		// assert
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, expectedResult, result)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestAssessmentSessionRepo_GetManyByAssessments(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockDB := testutil.NewMockDB()
	repo := &AssessmentSessionRepo{}
	asmID := idutil.ULIDNow()
	userID := idutil.ULIDNow()

	t.Run("query failed return DB error", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.Anything, database.Text(asmID), database.Text(userID))
		expectedErr := errors.NewDBError("AssessmentSessionRepo.GetManyByAssessments", puddle.ErrClosedPool)

		// act
		c, err := repo.GetManyByAssessments(ctx, mockDB.DB, asmID, userID)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("scan failed return conversion error", func(t *testing.T) {
		// arrange
		asm := &dto.AssessmentSession{}
		_, val := asm.FieldMap()
		mockDB.DB.
			On("Query", mock.Anything, mock.AnythingOfType("string"), database.Text(asmID), database.Text(userID)).
			Once().
			Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Close").Once().Return(nil)
		scanErr := fmt.Errorf("%s", "some err")
		mockDB.Rows.On("Scan", val...).Once().Return(scanErr)
		expectedErr := errors.NewConversionError("AssessmentSessionRepo.scanAssessmentSessions", scanErr)

		// act
		c, err := repo.GetManyByAssessments(ctx, mockDB.DB, asmID, userID)

		// assert
		assert.Nil(t, c)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("return empty slice when there is no row", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, database.Text(asmID), database.Text(userID))

		// act
		actual, err := repo.GetManyByAssessments(ctx, mockDB.DB, asmID, userID)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, []domain.Session{}, actual)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("select succeeded return all sessions", func(t *testing.T) {
		// arrange
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, database.Text(asmID), database.Text(userID))
		e := &dto.AssessmentSession{
			Status: database.Text("NONE"),
		}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})

		// act
		_, err := repo.GetManyByAssessments(ctx, mockDB.DB, asmID, userID)

		// assert
		assert.Nil(t, err)

		mockDB.RawStmt.AssertSelectedFields(t, fields...)
		mockDB.RawStmt.AssertSelectedTable(t, e.TableName(), "")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"deleted_at":    {HasNullTest: true},
			"assessment_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"user_id":       {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
		})
	})
}
