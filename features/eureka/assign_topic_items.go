package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userTryToAssignTopicItemsWithInvalidRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &epb.AssignTopicItemsRequest{
		TopicId: stepState.Topics[0].Id,
	}

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).AssignTopicItems(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userTryToAssignTopicItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &epb.AssignTopicItemsRequest{
		TopicId: stepState.TopicIDs[0],
	}

	for _, lo := range stepState.LearningObjectives {
		req.Items = append(req.Items, &epb.AssignTopicItemsRequest_Item{
			ItemId: &epb.AssignTopicItemsRequest_Item_LoId{
				LoId: lo.Info.Id,
			},
			DisplayOrder: lo.Info.DisplayOrder,
		})
	}

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).AssignTopicItems(s.signedCtx(ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfValidLearningObjectives(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TopicID = idutil.ULIDNow()
	req := s.generateLOsReq(ctx)
	los := req.LearningObjectives
	for i := 0; i < len(los); i++ {
		los[i].TopicId = stepState.TopicIDs[i%len(stepState.Topics)]
	}
	stepState.LearningObjectives = los
	stepState.Request = &epb.UpsertLOsRequest{
		LearningObjectives: los,
	}
	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
