package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
)

type CourseBookRepo interface {
	Upsert(ctx context.Context, courseBooks []*dto.CourseBookDto) error
	RetrieveAssociatedBook(ctx context.Context, bookID string) ([]*dto.CourseBookDto, error)
}

type CourseRepo interface {
	RetrieveByIDs(ctx context.Context, ids []string) ([]*dto.CourseDto, error)
}
