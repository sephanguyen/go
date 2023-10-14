package eureka

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userDeleteSomeTopicsWithRole(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).DeleteTopics(ctx, &epb.DeleteTopicsRequest{
		TopicIds: stepState.TopicIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}
