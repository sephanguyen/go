package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userListBooksByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	books := stepState.Request.([]*entities.Book)
	bookIDs := make([]string, 0, len(books))
	for _, book := range books {
		bookIDs = append(bookIDs, book.ID.String)
	}

	stepState.Response, stepState.ResponseErr = epb.NewBookReaderServiceClient(s.Conn).ListBooks(ctx, &epb.ListBooksRequest{
		Filter: &cpb.CommonFilter{
			Ids: bookIDs,
		},
	})

	return StepStateToContext(ctx, stepState), nil
}
