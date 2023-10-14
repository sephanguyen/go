package eureka

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userTryToAssignTopicItemsWithRole(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

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

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).AssignTopicItems(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}
