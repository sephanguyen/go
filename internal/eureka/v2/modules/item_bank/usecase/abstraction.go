package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/repository"
	item_bank_learnosity_repo "github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/repository/learnosity"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

type ActivityUsecase struct {
	DB               database.Ext
	LearnosityConfig configurations.LearnosityConfig
	HTTP             learnosity.HTTP
	DataAPI          learnosity.DataAPI

	ItemBankRepo repository.LearnosityItemBankRepo
}

func NewActivityUsecase(db database.Ext, learnosityConfig configurations.LearnosityConfig,
	http learnosity.HTTP, api learnosity.DataAPI) *ActivityUsecase {
	return &ActivityUsecase{
		DB:               db,
		LearnosityConfig: learnosityConfig,
		HTTP:             http,
		DataAPI:          api,
		ItemBankRepo:     item_bank_learnosity_repo.NewItemBankRepo(http, api),
	}
}

type ActivityGetter interface {
	CountTotalLearnosityItemByLM(ctx context.Context, learningMaterialID string) (uint32, error)
}
