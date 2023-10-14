package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userUpdateLearningObjectiveName(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomLO := stepState.LearningObjectives[rand.Intn(len(stepState.LearningObjectives)-1)] //nolint:gosec
	stepState.LOName = fmt.Sprintf("new-name-%s", idutil.ULIDNow())
	stepState.LoID = randomLO.Info.Id

	stepState.Response, stepState.ResponseErr = pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpdateLearningObjectiveName(s.signedCtx(ctx), &pb.UpdateLearningObjectiveNameRequest{
		LoId:                     stepState.LoID,
		NewLearningObjectiveName: stepState.LOName,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateLearningObjectiveNameCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.LearningObjective{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lo_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	var rawLMEnt entities.LearningObjective
	if err := database.Select(ctx, s.DB, query, database.Text(stepState.LoID)).ScanOne(&rawLMEnt); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get learning objective by ID, err: %w", err)
	}
	if rawLMEnt.Name.String != stepState.LOName {
		return StepStateToContext(ctx, stepState), fmt.Errorf("name get incorrect, expected %s, get %s", stepState.LOName, rawLMEnt.Name.String)
	}
	return StepStateToContext(ctx, stepState), nil
}
