package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type FlashcardProgressionRepo struct{}

type GetFlashcardProgressionArgs struct {
	StudySetID      pgtype.Text
	StudentID       pgtype.Text
	LoID            pgtype.Text
	StudyPlanItemID pgtype.Text
	From            pgtype.Int8
	To              pgtype.Int8
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

func (r *FlashcardProgressionRepo) GetLastFlashcardProgression(
	ctx context.Context, db database.QueryExecer,
	studyPlanItemID, loID, studentID pgtype.Text, isCompleted pgtype.Bool,
) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	fieldNames := database.GetFieldNames(flashcardProgression)
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND study_plan_item_id = $1
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

func (r *FlashcardProgressionRepo) Get(ctx context.Context, db database.QueryExecer, args *GetFlashcardProgressionArgs) (*entities.FlashcardProgression, error) {
	flashcardProgression := &entities.FlashcardProgression{}
	originalFieldNames := database.GetFieldNames(flashcardProgression)
	fieldNames := strings.Join(originalFieldNames, ",")
	selectArgs := []interface{}{args.StudySetID, args.StudentID, args.LoID, args.StudyPlanItemID}

	originalStmt := `SELECT %s 
	FROM %s 
	WHERE deleted_at IS NULL 
	AND study_set_id = $1
	AND ($2::TEXT IS NULL OR student_id = $2)
	AND ($3::TEXT IS NULL OR lo_id = $3)
	AND ($4::TEXT IS NULL OR study_plan_item_id = $4);`

	if args.From.Status == pgtype.Present && args.To.Status == pgtype.Present {
		// quiz_external_ids[$5:$6], $5, $6 are from and to. It helps to filter quiz_external_ids by from and to
		fieldNames = strings.Replace(fieldNames, "quiz_external_ids", "quiz_external_ids[$5:$6]", 1)

		selectArgs = append(selectArgs, args.From.Get(), args.To.Get())
	}

	stmt := fmt.Sprintf(originalStmt, fieldNames, flashcardProgression.TableName())
	if err := database.Select(ctx, db, stmt, selectArgs...).ScanOne(flashcardProgression); err != nil {
		return nil, err
	}
	return flashcardProgression, nil
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
	AND study_set_id = $1 
	AND student_id = $2;`, strings.Join(fieldNames, ","), flashcardProgression.TableName())

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

func (r *FlashcardProgressionRepo) DeleteByStudySetID(ctx context.Context, db database.QueryExecer, studySetID pgtype.Text) error {
	flashcardProgression := &entities.FlashcardProgression{}

	stmt := fmt.Sprintf(`UPDATE %s 
		SET deleted_at = NOW()
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
