package eureka

import (
	"context"
	"math/rand"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userUpsertValidChapters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	numberOfChapter := rand.Intn(10) + 5

	reqChapters := s.generateChapters(ctx, numberOfChapter, nil)
	stepState.Response, stepState.ResponseErr = epb.NewChapterModifierServiceClient(s.Conn).UpsertChapters(ctx, &epb.UpsertChaptersRequest{
		Chapters: reqChapters,
		BookId:   stepState.BookID,
	})
	return StepStateToContext(ctx, stepState), nil
}
