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
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFlashcardProgressionRepo_Create(t *testing.T) {
	t.Parallel()
	db := &mock_database.QueryExecer{}
	flashcardProgressionRepo := &FlashcardProgressionRepo{}
	flashcardProgression := &entities.FlashcardProgression{
		OriginalQuizSetID:     database.Text("original-quiz-set-id"),
		StudySetID:            database.Text("study-set-id"),
		OriginalStudySetID:    database.Text("original-study-set_id"),
		StudentID:             database.Text("student-id"),
		StudyPlanItemID:       database.Text("study-plan-item-id"),
		LoID:                  database.Text("lo-id"),
		QuizExternalIDs:       database.TextArray([]string{"quiz-1", "quiz-2"}),
		StudyingIndex:         database.Int4(1),
		SkippedQuestionIDs:    database.TextArray([]string{"quiz-3"}),
		RememberedQuestionIDs: database.TextArray([]string{"quiz-4"}),
		CreatedAt:             database.Timestamptz(time.Now()),
		UpdatedAt:             database.Timestamptz(time.Now()),
		CompletedAt:           database.Timestamptz(time.Now()),
		DeletedAt:             database.Timestamptz(time.Now()),
	}
	testCases := []TestCase{
		{
			name:        "happy case",
			req:         flashcardProgression,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
		{
			name:        "error TxClosed",
			req:         flashcardProgression,
			expectedErr: pgx.ErrTxClosed,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, pgx.ErrTxClosed)
			},
		},
		{
			name:        "error no rows affected",
			req:         flashcardProgression,
			expectedErr: fmt.Errorf("can not create flashcard progression"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			_, err := flashcardProgressionRepo.Create(ctx, db, testCase.req.(*entities.FlashcardProgression))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestFlashcardProgressionRepo_Upsert(t *testing.T) {
}

func TestFlashcardProgressionRepo_GetLastFlashcardProgression(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text("study-plan-item-id"), database.Text("lo-id"), database.Text("student-id"), database.Bool(true))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetLastFlashcardProgression(ctx, mockDB.DB, database.Text("study-plan-item-id"), database.Text("lo-id"), database.Text("student-id"), database.Bool(true))
		assert.True(t, errors.Is(err, nil))
	})

	t.Run("err TxClosed", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.AnythingOfType("string"), database.Text("study-plan-item-id"), database.Text("lo-id"), database.Text("student-id"), database.Bool(true))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetLastFlashcardProgression(ctx, mockDB.DB, database.Text("study-plan-item-id"), database.Text("lo-id"), database.Text("student-id"), database.Bool(true))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})
}

func TestFlashcardProgressionRepo_GetByStudySetID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text("study-set-id"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetByStudySetID(ctx, mockDB.DB, database.Text("study-set-id"))
		assert.True(t, errors.Is(err, nil))
	})

	t.Run("err TxClosed", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.AnythingOfType("string"), database.Text("study-set-id"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetByStudySetID(ctx, mockDB.DB, database.Text("study-set-id"))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})
}

func TestFlashcardProgressionRepo_GetByStudySetIDAndStudentID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), database.Text("study-set-id"), database.Text("student-id"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetByStudySetIDAndStudentID(ctx, mockDB.DB, database.Text("student-id"), database.Text("study-set-id"))
		assert.True(t, errors.Is(err, nil))
	})

	t.Run("err TxClosed", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.AnythingOfType("string"), database.Text("study-set-id"), database.Text("student-id"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.GetByStudySetIDAndStudentID(ctx, mockDB.DB, database.Text("student-id"), database.Text("study-set-id"))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})
}

func TestFlashcardProgressionRepo_Get(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()

	t.Run("happy case with paging", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, nil, mock.Anything,
			mock.AnythingOfType("string"), database.Text("study-set-id"), database.Text("student-id"), database.Text("lo-id"), database.Text("study-plan-item-id"),
			database.Int8(1).Get(), database.Int8(2).Get())
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.Get(ctx, mockDB.DB, &GetFlashcardProgressionArgs{
			StudySetID:      database.Text("study-set-id"),
			StudentID:       database.Text("student-id"),
			LoID:            database.Text("lo-id"),
			StudyPlanItemID: database.Text("study-plan-item-id"),
			From:            database.Int8(1),
			To:              database.Int8(2),
		})
		assert.True(t, errors.Is(err, nil))
	})

	t.Run("err TxClosed with paging", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"), database.Text("study-set-id"), database.Text("student-id"), database.Text("lo-id"), database.Text("study-plan-item-id"),
			database.Int8(1).Get(), database.Int8(2).Get())
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.Get(ctx, mockDB.DB, &GetFlashcardProgressionArgs{
			StudySetID:      database.Text("study-set-id"),
			StudentID:       database.Text("student-id"),
			LoID:            database.Text("lo-id"),
			StudyPlanItemID: database.Text("study-plan-item-id"),
			From:            database.Int8(1),
			To:              database.Int8(2),
		})
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})

	t.Run("err TxClosed without paging", func(t *testing.T) {
		e := &entities.FlashcardProgression{}
		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything,
			mock.AnythingOfType("string"), database.Text("study-set-id"), database.Text("student-id"), database.Text("lo-id"), database.Text("study-plan-item-id"))
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})
		_, err := repo.Get(ctx, mockDB.DB, &GetFlashcardProgressionArgs{
			StudySetID:      database.Text("study-set-id"),
			StudentID:       database.Text("student-id"),
			LoID:            database.Text("lo-id"),
			StudyPlanItemID: database.Text("study-plan-item-id"),
			From:            pgtype.Int8{Status: pgtype.Null},
			To:              pgtype.Int8{Status: pgtype.Null},
		})
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
	})
}

func TestFlashcardProgressionRepo_UpdateCompletedAt(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()
	studySetID := database.Text("study-set-id")

	t.Run("err update", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studySetID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.UpdateCompletedAt(ctx, mockDB.DB, studySetID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studySetID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := repo.UpdateCompletedAt(ctx, mockDB.DB, studySetID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "flashcard_progressions")
		mockDB.RawStmt.AssertUpdatedFields(t, "completed_at", "updated_at")
	})
}

func TestFlashcardProgressionRepo_DeleteByStudySetID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	repo := &FlashcardProgressionRepo{}
	ctx := context.Background()
	studySetID := database.Text("study-set-id")

	t.Run("err delete", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studySetID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.DeleteByStudySetID(ctx, mockDB.DB, studySetID)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, &studySetID)
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
		err := repo.DeleteByStudySetID(ctx, mockDB.DB, studySetID)
		assert.Nil(t, err)
		mockDB.RawStmt.AssertUpdatedTable(t, "flashcard_progressions")
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
	})
}
