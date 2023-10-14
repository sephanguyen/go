package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/item_bank/usecase"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestItemBankService_GetTotalItemsByLM(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("returns count when no error occur", func(t *testing.T) {
		// Arrange
		lmID := idutil.ULIDNow()
		expectedCount := uint32(200000)
		activityUsecase := &mock_usecase.MockActivityUsecase{}
		itemBankService := &ItemBankService{ActivityGetter: activityUsecase}
		activityUsecase.On("CountTotalLearnosityItemByLM", ctx, lmID).Once().Return(expectedCount, nil)

		// Act
		resp, err := itemBankService.GetTotalItemsByLM(ctx, &epb.GetTotalItemsByLMRequest{LearningMaterialId: lmID})

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, resp.GetTotalItems(), expectedCount)
		mock.AssertExpectationsForObjects(t, activityUsecase)
	})

	t.Run("returns error when error occur", func(t *testing.T) {
		// Arrange
		lmID := idutil.ULIDNow()
		expectedCount := uint32(0)
		activityUsecase := &mock_usecase.MockActivityUsecase{}
		itemBankService := &ItemBankService{ActivityGetter: activityUsecase}
		usecaseErr := errors.New("Test", fmt.Errorf("%s", "some thing"))
		expectedErr := errors.NewGrpcError(usecaseErr, transport.GrpcErrorMap)
		activityUsecase.On("CountTotalLearnosityItemByLM", ctx, lmID).Once().Return(expectedCount, usecaseErr)

		// Act
		_, err := itemBankService.GetTotalItemsByLM(ctx, &epb.GetTotalItemsByLMRequest{LearningMaterialId: lmID})

		// Assert
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, activityUsecase)
	})
}
