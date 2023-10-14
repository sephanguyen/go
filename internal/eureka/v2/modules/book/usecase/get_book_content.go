package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (b *BookUsecase) GetPublishedBookContent(ctx context.Context, bookID string) (book domain.Book, err error) {
	book, err = b.BookRepo.GetPublishedBookContent(ctx, bookID)
	if err != nil {
		if errors.CheckErrType(errors.ErrNoRowsExisted, err) {
			return book, errors.NewEntityNotFoundError("BookUsecase.GetPublishedBookContent", err)
		}
		return book, errors.New("BookUsecase.GetPublishedBookContent", err)
	}

	book.RemoveUnpublishedContent()
	return book, nil
}
