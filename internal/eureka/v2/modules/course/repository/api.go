package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
)

type ExternalCourseRepo interface {
	Upsert(ctx context.Context, courses []domain.Course) ([]domain.Course, error)
}

type CerebryRepo interface {
	GetUserToken(ctx context.Context, userID string) (string, error)
}
