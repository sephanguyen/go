package eureka

import (
	"context"
	"math/rand"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userHasCreateSomeTopics(ctx context.Context, typeTopic string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	numberOfTopics := rand.Intn(10) + 2
	reqTopics := s.generateTopics(ctx, numberOfTopics, nil)

	stepState.Response, stepState.ResponseErr = epb.NewTopicModifierServiceClient(s.Conn).Upsert(ctx, &epb.UpsertTopicsRequest{
		Topics: reqTopics,
	})

	return StepStateToContext(ctx, stepState), nil
}
