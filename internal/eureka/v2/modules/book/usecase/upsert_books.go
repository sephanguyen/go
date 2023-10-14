package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (b *BookUsecase) UpsertBooks(ctx context.Context, books []domain.Book) error {
	if err := b.BookRepo.Upsert(ctx, books); err != nil {
		return errors.New("BookUsecase.UpsertBooks", err)
	}

	return nil
}
