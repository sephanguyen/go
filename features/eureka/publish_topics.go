package eureka

import (
	"context"
	"math/rand"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userPublicSomeMissingTopics(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	topicIDs := make([]string, 0)
	topicIDs = append(topicIDs, stepState.TopicIDs...)

	numberOfMissingTopicIDs := rand.Intn(5) + 1
	// Generate non-existed ids
	for i := 0; i < numberOfMissingTopicIDs; i++ {
		topicIDs = append(topicIDs, idutil.ULIDNow())
	}

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).Publish(ctx, &epb.PublishTopicsRequest{
		TopicIds: topicIDs,
	})

	return StepStateToContext(ctx, stepState), nil
}
