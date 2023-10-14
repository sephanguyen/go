package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LOProgressionRepo struct{}

func (r *LOProgressionRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.LOProgression) error {
	ctx, span := interceptors.StartSpan(ctx, "LOProgression.Upsert")
	defer span.End()

	fieldNames := database.GetFieldNames(e)
	scanFields := database.GetScanFields(e, fieldNames)

	stmt := `
    INSERT INTO %s (%s) VALUES (%s)
    ON CONFLICT (student_id, study_plan_id, learning_material_id) WHERE (deleted_at IS NULL) DO UPDATE SET
        shuffled_quiz_set_id = EXCLUDED.shuffled_quiz_set_id,
		quiz_external_ids = EXCLUDED.quiz_external_ids,
		last_index = EXCLUDED.last_index,
        updated_at = EXCLUDED.updated_at;
	`

	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(stmt, e.TableName(), strings.Join(fieldNames, ","), placeHolders)

	cmdTag, err := db.Exec(ctx, query, scanFields...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no row affected")
	}

	return nil
}

func (r *LOProgressionRepo) DeleteByStudyPlanIdentity(ctx context.Context, db database.QueryExecer, args StudyPlanItemIdentity) (int64, error) {
	ctx, span := interceptors.StartSpan(ctx, "LOProgressionRepo.DeleteByStudyPlanIdentity")
	defer span.End()
	e := entities.LOProgression{}
	query := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE learning_material_id = $1::TEXT AND student_id = $2::TEXT AND study_plan_id = $3::TEXT AND deleted_at IS NULL`, e.TableName())
	cmdTag, err := db.Exec(ctx, query, args.LearningMaterialID, args.StudentID, args.StudyPlanID)
	if err != nil {
		return 0, fmt.Errorf("db.Exec: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

func (r *LOProgressionRepo) GetByStudyPlanItemIdentity(ctx context.Context, db database.QueryExecer, arg StudyPlanItemIdentity, from pgtype.Int8, to pgtype.Int8) (*entities.LOProgression, error) {
	ctx, span := interceptors.StartSpan(ctx, "LOProgressionRepo.GetByStudyPlanItemIdentity")
	defer span.End()

	loProgression := &entities.LOProgression{}
	fieldNames := database.GetFieldNames(loProgression)
	selectFields := golibs.Replace(fieldNames, []string{"quiz_external_ids"}, []string{"quiz_external_ids[$4:$5]"})

	stmt := fmt.Sprintf(`
	SELECT %s
	FROM %s
	WHERE 
		deleted_at IS NULL
		AND study_plan_id = $1
		AND learning_material_id = $2
		AND student_id = $3 
	`, strings.Join(selectFields, ", "), loProgression.TableName())

	if err := database.Select(ctx, db, stmt, arg.StudyPlanID, arg.LearningMaterialID, arg.StudentID, from.Get(), to.Get()).ScanOne(loProgression); err != nil {
		return nil, err
	}

	return loProgression, nil
}
