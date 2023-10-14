package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/domain"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

type LearnosityItemBankRepo interface {
	GetActivities(ctx context.Context, security learnosity.Security, request learnosity.Request) ([]domain.Activity, error)
}
