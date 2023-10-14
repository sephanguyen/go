package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

type LearnositySessionRepo interface {
	GetSessionStatuses(ctx context.Context, security learnosity.Security, request learnosity.Request) ([]domain.Session, error)
	GetSessionResponses(ctx context.Context, security learnosity.Security, request learnosity.Request) (domain.Sessions, error)
}
