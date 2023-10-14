package learning_objective

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

var (
	learningObjectiveUpdatedName        = "LearningObjective updated"
	learningObjectiveUpdatedVideoID     = "video_id updated"
	learningObjectiveUpdatedVideoScript = "video_script updated"
	learningObjectiveUpdatedStudyGuide  = "study_guide updated"
)

func (s *Suite) userUpdatesALearningObjective(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lo := &sspb.LearningObjectiveBase{
		Base: &sspb.LearningMaterialBase{
			LearningMaterialId: stepState.LearningMaterialIDs[0],
			Name:               learningObjectiveUpdatedName,
		},
		VideoId:     learningObjectiveUpdatedVideoID,
		VideoScript: learningObjectiveUpdatedVideoScript,
		StudyGuide:  learningObjectiveUpdatedStudyGuide,
	}

	stepState.Response, stepState.ResponseErr = sspb.NewLearningObjectiveClient(s.EurekaConn).UpdateLearningObjective(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpdateLearningObjectiveRequest{
		LearningObjective: lo,
	})
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemUpdatesTheLearningObjectiveCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	lo := &entities.LearningObjectiveV2{
		LearningMaterial: entities.LearningMaterial{
			ID: database.Text(stepState.LearningMaterialIDs[0]),
		},
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1", strings.Join(database.GetFieldNames(lo), ","), lo.TableName())
	if err := database.Select(ctx, s.EurekaDB, query, lo.LearningMaterial.ID).ScanOne(lo); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if lo.Name.String != learningObjectiveUpdatedName {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect LearningObjective name: expected %s, got %s", learningObjectiveUpdatedName, lo.Name.String)
	}
	if lo.Video.String != learningObjectiveUpdatedVideoID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect LearningObjective video: expected %s, got %s", learningObjectiveUpdatedVideoID, lo.Video.String)
	}
	if lo.VideoScript.String != learningObjectiveUpdatedVideoScript {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect LearningObjective video script: expected %s, got %s", learningObjectiveUpdatedVideoScript, lo.VideoScript.String)
	}
	if lo.StudyGuide.String != learningObjectiveUpdatedStudyGuide {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("Incorrect LearningObjective study guide: expected %s, got %s", learningObjectiveUpdatedStudyGuide, lo.StudyGuide.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
