package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	eureka_db "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type FlashcardProgressionRepo struct{}

const bulkUpsertFlashCardProgression = `
INSERT INTO %s (%s) 
VALUES %s 
ON CONFLICT ON CONSTRAINT flashcard_progressions_pk DO UPDATE 
SET 
	original_quiz_set_id = excluded.original_quiz_set_id, 
	original_study_set_id = excluded.original_study_set_id, 
	student_id = excluded.student_id, 
	study_plan_item_id = excluded.study_plan_item_id, 
	lo_id = excluded.lo_id, 
	quiz_external_ids = excluded.quiz_external_ids, 
	studying_index = excluded.studying_index, 
	skipped_question_ids = excluded.skipped_question_ids, 
	remembered_question_ids = excluded.remembered_question_ids, 
	updated_at = excluded.updated_at, 
	completed_at = excluded.completed_at,
	deleted_at = excluded.deleted_at
`

func (r *FlashcardProgressionRepo) Upsert(ctx context.Context, db database.QueryExecer, flashcardProgressions []*entities.FlashcardProgression) error {
	ctx, span := interceptors.StartSpan(ctx, "FlashcardProgressionRepo.Upsert")
	defer span.End()

	now := time.Now()
	for _, flashcardProgression := range flashcardProgressions {
		err := multierr.Combine(
			flashcardProgression.CreatedAt.Set(now),
			flashcardProgression.UpdatedAt.Set(now),
		)
		if err != nil {
			return err
		}
	}

	err := eureka_db.BulkUpsert(ctx, db, bulkUpsertFlashCardProgression, flashcardProgressions)
	if err != nil {
		return fmt.Errorf("eureka_db.BulkUpsertFlashCardProgression error: %s", err.Error())
	}
	return nil
}

func (r *FlashcardProgressionRepo) GetLastFlashcardProgression(
	ctx context.Context, db database.QueryExecer,
	studyPlanItemID, loID, studentID pgtype.Text, isCompleted pgtype.Bool,
) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	fieldNames := database.GetFieldNames(flashcardProgression)
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND ($1::TEXT IS NULL OR study_plan_item_id = $1)
	AND lo_id = $2
	AND student_id = $3
	AND (($4 = TRUE AND completed_at IS NOT NULL) OR ($4 = FALSE AND completed_at IS NULL))
	ORDER BY updated_at DESC
	LIMIT 1`, strings.Join(fieldNames, ","), flashcardProgression.TableName())

	err := database.Select(ctx, db, stmt, studyPlanItemID, loID, studentID, isCompleted).ScanOne(flashcardProgression)
	if err != nil {
		return nil, err
	}

	return flashcardProgression, nil
}

func (r *FlashcardProgressionRepo) GetLastFlashcardProgressionV2(
	ctx context.Context, db database.QueryExecer,
	itemIdentity *StudyPlanItemIdentity, isCompleted pgtype.Bool,
) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	fieldNames := database.GetFieldNames(flashcardProgression)
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND study_plan_id = $1
	AND learning_material_id = $2
	AND student_id = $3
	AND (($4 = TRUE AND completed_at IS NOT NULL) OR ($4 = FALSE AND completed_at IS NULL))
	ORDER BY updated_at DESC
	LIMIT 1`, strings.Join(fieldNames, ","), flashcardProgression.TableName())

	err := database.Select(ctx, db, stmt, itemIdentity.StudyPlanID, itemIdentity.LearningMaterialID, itemIdentity.StudentID, isCompleted).ScanOne(flashcardProgression)
	if err != nil {
		return nil, err
	}

	return flashcardProgression, nil
}

type GetFlashcardProgressionArgs struct {
	StudySetID      pgtype.Text
	StudentID       pgtype.Text
	LoID            pgtype.Text
	StudyPlanItemID pgtype.Text
	LmID            pgtype.Text
	StudyPlanID     pgtype.Text
	From            pgtype.Int8
	To              pgtype.Int8
}

func (r *FlashcardProgressionRepo) Get(ctx context.Context, db database.QueryExecer, args *GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error) {
	// set default Null instead of Undefined
	if args.StudentID.Status == pgtype.Undefined {
		args.StudentID.Status = pgtype.Null
	}
	if args.LoID.Status == pgtype.Undefined {
		args.LoID.Status = pgtype.Null
	}
	if args.StudyPlanItemID.Status == pgtype.Undefined {
		args.StudyPlanItemID.Status = pgtype.Null
	}
	if args.LmID.Status == pgtype.Undefined {
		args.LmID.Status = pgtype.Null
	}
	if args.StudyPlanID.Status == pgtype.Undefined {
		args.StudyPlanID.Status = pgtype.Null
	}

	flashcardProgression := &entities.FlashcardProgression{}
	originalFieldNames := database.GetFieldNames(flashcardProgression)
	fieldNames := strings.Join(originalFieldNames, ",")
	selectArgs := []interface{}{args.StudySetID, args.StudentID, args.LoID, args.StudyPlanItemID, args.LmID, args.StudyPlanID}

	originalStmt := `SELECT %s 
	FROM %s 
	WHERE deleted_at IS NULL 
	AND study_set_id = $1
	AND ($2::TEXT IS NULL OR student_id = $2)
	AND ($3::TEXT IS NULL OR lo_id = $3)
	AND ($4::TEXT IS NULL OR study_plan_item_id = $4)
	AND ($5::TEXT IS NULL OR learning_material_id = $5)
	AND ($6::TEXT IS NULL OR study_plan_id = $6);`

	if args.From.Status == pgtype.Present && args.To.Status == pgtype.Present {
		// quiz_external_ids[$7:$8], $7, $8 are from and to. It helps to filter quiz_external_ids by from and to
		fieldNames = strings.Replace(fieldNames, "quiz_external_ids", "quiz_external_ids[$7:$8]", 1)

		selectArgs = append(selectArgs, args.From.Get(), args.To.Get())
	}

	stmt := fmt.Sprintf(originalStmt, fieldNames, flashcardProgression.TableName())
	if err := database.Select(ctx, db, stmt, selectArgs...).ScanOne(flashcardProgression); err != nil {
		return nil, err
	}
	return flashcardProgression, nil
}

func (r *FlashcardProgressionRepo) Create(ctx context.Context, db database.QueryExecer, flashcardProgression *entities.FlashcardProgression) (pgtype.Text, error) {
	cmd, err := database.Insert(ctx, flashcardProgression, db.Exec)
	if err != nil {
		return pgtype.Text{Status: pgtype.Null}, err
	}

	if cmd.RowsAffected() != 1 {
		return pgtype.Text{Status: pgtype.Null}, fmt.Errorf("can not create flashcard progression")
	}

	return flashcardProgression.StudySetID, nil
}

func (r *FlashcardProgressionRepo) DeleteByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error {
	flashcardProgression := &entities.FlashcardProgression{}

	stmt := fmt.Sprintf(`UPDATE %s 
		SET deleted_at = NOW()
		WHERE study_set_id = $1::TEXT AND deleted_at IS NULL`, flashcardProgression.TableName())
	cmd, err := db.Exec(ctx, stmt, &studySetID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any flashcard progression to update: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *FlashcardProgressionRepo) GetByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	fieldNames := database.GetFieldNames(flashcardProgression)

	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND study_set_id = $1;`, strings.Join(fieldNames, ","), flashcardProgression.TableName())

	if err := database.Select(ctx, db, stmt, studySetID).ScanOne(flashcardProgression); err != nil {
		return nil, err
	}
	return flashcardProgression, nil
}

func (r *FlashcardProgressionRepo) GetByStudySetIDAndStudentID(ctx context.Context, db database.QueryExecer, studentID, studySetID pgtype.Text) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	fieldNames := database.GetFieldNames(flashcardProgression)

	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND study_set_id = $1::TEXT 
	AND student_id = $2::TEXT;`, strings.Join(fieldNames, ","), flashcardProgression.TableName())

	if err := database.Select(ctx, db, stmt, studySetID, studentID).ScanOne(flashcardProgression); err != nil {
		return nil, err
	}
	return flashcardProgression, nil
}

func (r *FlashcardProgressionRepo) UpdateCompletedAt(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error {
	flashcardProgression := &entities.FlashcardProgression{}

	stmt := fmt.Sprintf(`UPDATE %s 
		SET completed_at = NOW(), updated_at = NOW()
		WHERE study_set_id = $1 AND deleted_at IS NULL`, flashcardProgression.TableName())
	cmd, err := db.Exec(ctx, stmt, &studySetID)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("not found any flashcard progression to update: %w", pgx.ErrNoRows)
	}

	return nil
}
