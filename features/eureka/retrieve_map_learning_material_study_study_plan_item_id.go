package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userRetrieveLMStudyPlanItemID(ctx context.Context) (context.Context, error) {
	ctx = contextWithToken(s, ctx)
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanItemReaderServiceClient(s.Conn).RetrieveMappingLmIDToStudyPlanItemID(ctx, &epb.RetrieveMappingLmIDToStudyPlanItemIDRequest{
		StudyPlanId: stepState.StudyPlanID,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnLMAndStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	e := entities.StudyPlanItem{}
	query := fmt.Sprintf(`
	SELECT
	COALESCE(NULLIF(content_structure ->> 'lo_id', ''), content_structure->>'assignment_id', '') AS learning_material_id, study_plan_item_id 
		FROM %s spi
	WHERE study_plan_id  = $1
	AND deleted_at IS NULL`, e.TableName())

	rows, err := s.DB.Query(ctx, query, database.Text(stepState.StudyPlanID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	res := stepState.Response.(*epb.RetrieveMappingLmIDToStudyPlanItemIDResponse)
	result := []*repositories.FindLearningMaterialByStudyPlanID{}
	for rows.Next() {
		i := new(repositories.FindLearningMaterialByStudyPlanID)

		if err := rows.Scan(&i.LearningMaterialID, &i.StudyPlanItemID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("StudyPlanItemRepo.FindLearningMaterialByStudyPlanID.Err: %w", err)
		}
		result = append(result, i)
	}

	if len(result) != len(res.Pairs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("FindLearningMaterialByStudyPlanID return %v elements but response is %v", len(result), len(res.Pairs))
	}

	for _, v := range result {
		if res.Pairs[v.LearningMaterialID.String] != v.StudyPlanItemID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("LM ID %s should match with SP ID %s but got %s", v.LearningMaterialID.String, v.StudyPlanItemID.String, res.Pairs[v.LearningMaterialID.String])
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
