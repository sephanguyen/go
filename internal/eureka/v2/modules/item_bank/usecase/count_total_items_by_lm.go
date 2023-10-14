package usecase

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/helper"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

func (a *ActivityUsecase) CountTotalLearnosityItemByLM(ctx context.Context, learningMaterialID string) (uint32, error) {
	var count uint32
	dataRequest := learnosity.Request{
		"references": []string{learningMaterialID},
		"status":     []string{"published"},
	}
	now := time.Now()
	security := helper.NewLearnositySecurity(ctx, a.LearnosityConfig, "localhost", now)
	activities, err := a.ItemBankRepo.GetActivities(ctx, security, dataRequest)
	if err != nil {
		return count, errors.New("ActivityUsecase.CountTotalLearnosityItemByLM", err)
	}
	for _, v := range activities {
		count += uint32(len(v.Data.Items))
	}

	return count, nil
}
