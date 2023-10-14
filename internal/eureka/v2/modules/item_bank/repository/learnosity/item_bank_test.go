package learnosity

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mock_learnosity "github.com/manabie-com/backend/mock/golibs/learnosity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestItemBankRepo_GetActivities(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	endpoint := learnosity.EndpointDataAPIGetActivities
	security := learnosity.Security{
		ConsumerKey:    "consumer_key",
		Domain:         "domain",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: "secret",
	}

	t.Run("get all activities successfully", func(t *testing.T) {
		// Arrange
		mockHTTP := &mock_learnosity.HTTP{}
		mockDataAPI := &mock_learnosity.DataAPI{}
		repo := NewItemBankRepo(mockHTTP, mockDataAPI)
		lmID := "LM_1"
		item1IDs := []any{"ITEM_1", "ITEM_2"}
		item2IDs := []any{
			map[string]any{
				"id":        "ID3",
				"reference": lmID,
			},
			map[string]any{
				"id":        "ID4",
				"reference": lmID,
			}}
		request := learnosity.Request{
			"references": []string{lmID},
			"status":     "published",
		}
		expectedActivities := []domain.Activity{
			{
				Reference: lmID,
				Data: domain.ActivityData{
					Items: item1IDs,
					Config: domain.Config{
						Regions: "ASIA",
					},
					RenderingType: "R1",
				},
				Tags: domain.Tags{Tenant: []string{"manabie"}},
			},
			{
				Reference: lmID,
				Data: domain.ActivityData{
					Items: item2IDs,
					Config: domain.Config{
						Regions: "ASIA",
					},
					RenderingType: "R2",
				},
				Tags: domain.Tags{Tenant: []string{"manabie"}},
			},
		}
		learnosityDataArr := sliceutils.Map(expectedActivities, func(s domain.Activity) map[string]any {
			return map[string]any{
				"reference": s.Reference,
				"data": map[string]any{
					"items":          s.Data.Items,
					"rendering_type": s.Data.RenderingType,
					"config": map[string]any{
						"regions": s.Data.Config.Regions,
					},
				},
				"tags": map[string]any{
					"tenant": s.Tags.Tenant,
				},
			}
		})
		dataRaw, _ := json.Marshal(learnosityDataArr)
		mockDataAPI.On("RequestIterator", mock.Anything, mockHTTP, endpoint, security, request).
			Once().
			Return([]learnosity.Result{
				{
					Meta: map[string]any{"records": float64(len(learnosityDataArr))},
					Data: dataRaw,
				},
			}, nil)

		// Act
		actualActivities, err := repo.GetActivities(ctx, security, request)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedActivities, actualActivities)
		mock.AssertExpectationsForObjects(t, mockDataAPI)
	})

	t.Run("get all activities failed", func(t *testing.T) {
		// Arrange
		mockHTTP := &mock_learnosity.HTTP{}
		mockDataAPI := &mock_learnosity.DataAPI{}
		repo := NewItemBankRepo(mockHTTP, mockDataAPI)
		lmID := "LM_1"
		request := learnosity.Request{
			"references": []string{lmID},
			"status":     "published",
		}
		apiErr := fmt.Errorf("%s", "some err")
		expectedErr := errors.NewLearnosityError("ItemBankRepo.GetActivities", apiErr)
		mockDataAPI.On("RequestIterator", mock.Anything, mockHTTP, endpoint, security, request).
			Once().
			Return(nil, apiErr)

		// Act
		statuses, err := repo.GetActivities(ctx, security, request)

		// Assert
		assert.Nil(t, statuses)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDataAPI)
	})
}
