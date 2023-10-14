package learning_material

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) ourSystemMustUpdateLearningMaterialNameCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	e := &entities.LearningMaterial{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	var rawLMEnt entities.LearningMaterial
	if err := database.Select(ctx, s.EurekaDB, query, database.Text(stepState.LearningMaterialID)).ScanOne(&rawLMEnt); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to get learning material by ID, err: %w", err)
	}
	if rawLMEnt.Name.String != stepState.LMName {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("name get incorrect, expected %s, get %s", stepState.LMName, rawLMEnt.Name.String)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateLMName(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	randomLMId := stepState.LearningMaterialIDs[rand.Intn(len(stepState.LearningMaterialIDs)-1)]

	stepState.LMName = fmt.Sprintf("lm-name-%s", idutil.ULIDNow())
	stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).UpdateLearningMaterialName(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.UpdateLearningMaterialNameRequest{
		LearningMaterialId:      randomLMId,
		NewLearningMaterialName: stepState.LMName,
	})
	stepState.LearningMaterialID = randomLMId
	return utils.StepStateToContext(ctx, stepState), nil
}
