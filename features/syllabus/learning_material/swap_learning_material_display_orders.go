package learning_material

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
)

func (s *Suite) swapLMDisplayOrder(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	//TODO: add the rest type.
	lmTypes := []string{"assignment", "learning_objective", "flashcard"}
	var (
		randomN1, randomN2 int
	)
	randomN1 = rand.Intn(len(lmTypes))
	if randomN1-1 >= 0 {
		randomN2 = randomN1 - 1
	} else if randomN1+1 < len(lmTypes) {
		randomN2 = randomN1 + 1
	}
	req := &sspb.SwapDisplayOrderRequest{
		FirstLearningMaterialId:  stepState.MapLearningMaterial[lmTypes[randomN1]].LearningMaterialId,
		SecondLearningMaterialId: stepState.MapLearningMaterial[lmTypes[randomN2]].LearningMaterialId,
	}
	s.StepState.Request = req
	stepState.MapLearningMaterial[lmTypes[randomN1]].ParamPosition = 1
	stepState.MapLearningMaterial[lmTypes[randomN2]].ParamPosition = 2
	lmIDs := []string{req.FirstLearningMaterialId, req.SecondLearningMaterialId}
	// save the existed created LMs
	e := &entities.LearningMaterial{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	rawlmsEnt := &entities.LearningMaterials{}
	if err := database.Select(ctx, s.EurekaDB, query, database.TextArray(lmIDs)).ScanAll(rawlmsEnt); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	lmEnts := rawlmsEnt.Get()
	for _, lm := range lmEnts {
		if lm.ID.String == stepState.MapLearningMaterial[lmTypes[randomN1]].LearningMaterialId {
			stepState.MapLearningMaterial[lmTypes[randomN1]].DisplayOrder = int32(lm.DisplayOrder.Int)
		}
		if lm.ID.String == stepState.MapLearningMaterial[lmTypes[randomN2]].LearningMaterialId {
			stepState.MapLearningMaterial[lmTypes[randomN2]].DisplayOrder = int32(lm.DisplayOrder.Int)
		}
	}
	stepState.Response, stepState.ResponseErr = sspb.NewLearningMaterialClient(s.EurekaConn).SwapDisplayOrder(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		req,
	)

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) displayOrdersOfFlashcardsSwapped(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	req := s.StepState.Request.(*sspb.SwapDisplayOrderRequest)
	lmIDs := []string{req.FirstLearningMaterialId, req.SecondLearningMaterialId}

	// save the existed created LMs
	e := &entities.LearningMaterial{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE learning_material_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	rawlmsEnt := &entities.LearningMaterials{}
	if err := database.Select(ctx, s.EurekaDB, query, database.TextArray(lmIDs)).ScanAll(rawlmsEnt); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	lmEnts := rawlmsEnt.Get()

	for _, val := range stepState.MapLearningMaterial {
		if val.ParamPosition > 0 {
			for _, lm := range lmEnts {
				// the same learningmaterialID shouldn't have the same display_order
				if lm.ID.String == val.LearningMaterialId {
					if lm.DisplayOrder.Int == int16(val.DisplayOrder) {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("[1]learning material display order didn't swap")
					}
				}
				// this compare just ensure 100%
				if lm.ID.String != val.LearningMaterialId {
					if lm.DisplayOrder.Int != int16(val.DisplayOrder) {
						return utils.StepStateToContext(ctx, stepState), fmt.Errorf("[2]learning material display order didn't swap")
					}
				}
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
