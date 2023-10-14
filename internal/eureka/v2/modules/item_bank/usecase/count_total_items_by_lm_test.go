package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	mock_item_bank_learnosity_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/item_bank/repository/learnosity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivityUsecase_CountItemByLM(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	lmID := idutil.ULIDNow()
	dataRequest := learnosity.Request{
		"references": []string{lmID},
		"status":     []string{"published"},
	}
	t.Run("return count of total items successfully when no error occurs", func(t *testing.T) {
		// Arrange
		mockItemBankRepo := &mock_item_bank_learnosity_repo.MockItemBankRepo{}
		usecase := &ActivityUsecase{ItemBankRepo: mockItemBankRepo}
		repoActivities := []domain.Activity{
			{
				Reference: lmID,
				Data: domain.ActivityData{
					Items:         []any{"Item_1", "Item_2"},
					Config:        domain.Config{Regions: "ASIA"},
					RenderingType: "R1",
				},
				Tags: domain.Tags{Tenant: []string{"manabie"}},
			},
			{
				Reference: lmID,
				Data: domain.ActivityData{
					Items:         []any{"Item_3", "Item_4", "Item_5"},
					Config:        domain.Config{Regions: "ASIA"},
					RenderingType: "",
				},
				Tags: domain.Tags{Tenant: []string{"manabie"}},
			},
			{
				Reference: lmID,
				Data: domain.ActivityData{
					Items:         []any{},
					Config:        domain.Config{Regions: "ASIA"},
					RenderingType: "",
				},
				Tags: domain.Tags{Tenant: []string{"manabie"}},
			}}
		mockItemBankRepo.On("GetActivities", mock.Anything, mock.Anything, dataRequest).
			Once().
			Return(repoActivities, nil)

		// Act
		count, err := usecase.CountTotalLearnosityItemByLM(ctx, lmID)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, uint32(5), count)
		mock.AssertExpectationsForObjects(t, mockItemBankRepo)
	})

	t.Run("return count of 0 when error occurs", func(t *testing.T) {
		// Arrange
		mockItemBankRepo := &mock_item_bank_learnosity_repo.MockItemBankRepo{}
		usecase := &ActivityUsecase{ItemBankRepo: mockItemBankRepo}
		rootErr := fmt.Errorf("%s", "some roots")
		repoLayerErr := errors.New("ItemBankRepo.GetActivities", rootErr)
		expectedErr := errors.New("ActivityUsecase.CountTotalLearnosityItemByLM", repoLayerErr)
		mockItemBankRepo.On("GetActivities", mock.Anything, mock.Anything, dataRequest).
			Once().
			Return(nil, repoLayerErr)

		// Act
		count, err := usecase.CountTotalLearnosityItemByLM(ctx, lmID)

		// Assert
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint32(0), count)
		mock.AssertExpectationsForObjects(t, mockItemBankRepo)
	})
}
