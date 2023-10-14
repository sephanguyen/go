package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/external"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/cerebry"
)

type CerebryUsecase struct {
	CerebryConfig cerebry.Config
	CerebryRepo   repository.CerebryRepo
}

func NewCerebryUsecase(config cerebry.Config) *CerebryUsecase {
	return &CerebryUsecase{
		CerebryConfig: config,
		CerebryRepo:   external.NewCerebryRepo(config),
	}
}

type CerebryTokenGenerator interface {
	GenerateUserToken(ctx context.Context) (string, error)
}
