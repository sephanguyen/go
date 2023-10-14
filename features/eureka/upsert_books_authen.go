package eureka

import (
	"context"
	"math/rand"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userUpsertBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	numberOfBooks := rand.Intn(20) + 1
	reqBooks := s.generateBooks(numberOfBooks, nil)
	stepState.Response, stepState.ResponseErr = epb.NewBookModifierServiceClient(s.Conn).UpsertBooks(ctx, &epb.UpsertBooksRequest{
		Books: reqBooks,
	})
	return StepStateToContext(ctx, stepState), nil
}
