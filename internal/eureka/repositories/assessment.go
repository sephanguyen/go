package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type AssessmentRepo struct{}

func (r *AssessmentRepo) GetAssessmentByCourseAndLearningMaterial(ctx context.Context, db database.QueryExecer, courseIDs, learningMaterialIDs pgtype.TextArray) ([]*entities.Assessment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssessmentRepo.GetLatestByIdentity")
	defer span.End()

	var assessment entities.Assessment
	fields, _ := assessment.FieldMap()

	stmt := fmt.Sprintf(`
    SELECT %s
    FROM %s
    WHERE deleted_at IS NULL
    AND learning_material_id = ANY($1::_TEXT)
    AND course_id = ANY($2::_TEXT);`, strings.Join(fields, ", "), assessment.TableName())

	assessments := &entities.Assessments{}

	err := database.Select(ctx, db, stmt, &learningMaterialIDs, &courseIDs).ScanAll(assessments)

	if err != nil {
		return nil, err
	}

	return *assessments, nil
}
