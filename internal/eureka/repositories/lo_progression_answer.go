package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	dbeureka "github.com/manabie-com/backend/internal/eureka/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LOProgressionAnswerRepo struct{}

func (r *LOProgressionAnswerRepo) DeleteByStudyPlanIdentity(ctx context.Context, db database.QueryExecer, args StudyPlanItemIdentity) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "LOProgressionAnswerRepo.DeleteByStudyPlanIdentity")
	defer span.End()
	e := entities.LOProgressionAnswer{}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, e.TableName())
	cmdTag, err := db.Exec(ctx, query, args.LearningMaterialID, args.StudentID, args.StudyPlanID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *LOProgressionAnswerRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.LOProgressionAnswer) error {
	ctx, span := interceptors.StartSpan(ctx, "LOProgressionAnswerRepo.BulkUpsert")
	defer span.End()

	stmt := `
    INSERT INTO %s (%s) VALUES %s
    ON CONFLICT (progression_id, quiz_external_id) DO UPDATE SET
        student_text_answer = EXCLUDED.student_text_answer,
        student_index_answer = EXCLUDED.student_index_answer,
		submitted_keys_answer = EXCLUDED.submitted_keys_answer,
        updated_at = EXCLUDED.updated_at;
	`

	if err := dbeureka.BulkUpsert(ctx, db, stmt, items); err != nil {
		return fmt.Errorf("database.BulkUpsert error: %s", err.Error())
	}

	return nil
}

func (r *LOProgressionAnswerRepo) ListByProgressionAndExternalIDs(ctx context.Context, db database.QueryExecer, progressionID pgtype.Text, externalIDs pgtype.TextArray) (entities.LOProgressionAnswers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LOProgressionAnswerRepo.ListByProgressionAndExternalIDs")
	defer span.End()

	e := &entities.LOProgressionAnswer{}
	lpAnswers := entities.LOProgressionAnswers{}
	fieldNames := database.GetFieldNames(e)

	stmt := fmt.Sprintf(`
	SELECT %s 
	FROM %s 
	WHERE deleted_at IS NULL AND progression_id = $1 AND quiz_external_id = ANY($2)
	`, strings.Join(fieldNames, ", "), e.TableName())

	if err := database.Select(ctx, db, stmt, progressionID, externalIDs).ScanAll(&lpAnswers); err != nil {
		return nil, err
	}

	return lpAnswers, nil
}
