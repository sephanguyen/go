package learnosity

import (
	"context"
	"encoding/json"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/repository"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	learnosity_entity "github.com/manabie-com/backend/internal/golibs/learnosity/entity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

type ItemBankRepo struct {
	http learnosity.HTTP
	api  learnosity.DataAPI
}

func NewItemBankRepo(http learnosity.HTTP, api learnosity.DataAPI) repository.LearnosityItemBankRepo {
	return &ItemBankRepo{http: http, api: api}
}

func (i *ItemBankRepo) GetActivities(ctx context.Context, security learnosity.Security, request learnosity.Request) (rs []domain.Activity, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ItemBankRepo.GetActivities")
	defer span.End()
	endpoint := learnosity.EndpointDataAPIGetActivities

	results, err := i.api.RequestIterator(ctx, i.http, endpoint, security, request)
	if err != nil {
		return rs, errors.NewLearnosityError("ItemBankRepo.GetActivities", err)
	}
	for _, r := range results {
		records := r.Meta.Records()
		ssr := make([]learnosity_entity.Activity, records)
		_ = json.Unmarshal(r.Data, &ssr)
		activities := sliceutils.Map(ssr, toActivityDomain)
		rs = append(rs, activities...)
	}
	return rs, nil
}

func toActivityDomain(s learnosity_entity.Activity) domain.Activity {
	return domain.Activity{
		Reference: s.Reference,
		Data: domain.ActivityData{
			Items:         s.Data.Items,
			Config:        domain.Config{Regions: s.Data.Config.Regions},
			RenderingType: s.Data.RenderingType,
		},
		Tags: domain.Tags{Tenant: s.Tags.Tenant},
	}
}
