package eureka

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userSomePublicTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topicIDs := make([]string, 0)
	topicIDs = append(topicIDs, stepState.TopicIDs...)

	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).Publish(ctx, &epb.PublishTopicsRequest{
		TopicIds: topicIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}
