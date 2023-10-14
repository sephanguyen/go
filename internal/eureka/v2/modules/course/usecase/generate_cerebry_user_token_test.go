package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/cerebry"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_external "github.com/manabie-com/backend/mock/eureka/v2/modules/course/repository/external"

	"github.com/stretchr/testify/assert"
)

func TestCerebryUsecase_GenerateUserToken(t *testing.T) {
	t.Parallel()
	cfg := cerebry.Config{
		BaseURL:        "https://manabie.com",
		PermanentToken: "ABC",
	}
	parentCtx := context.Background()
	userID := idutil.ULIDNow()
	ctx := context.WithValue(parentCtx, interceptors.UserIDKey(0), userID)

	t.Run("return error when context does not contain user id", func(t *testing.T) {
		// arrange
		cerebryRepo := &mock_external.MockCerebryRepo{}
		sut := CerebryUsecase{CerebryConfig: cfg, CerebryRepo: cerebryRepo}
		expectedErr := errors.New("CerebryUsecase.GenerateUserToken: Context does not contain user id", nil)

		// act
		tok, err := sut.GenerateUserToken(parentCtx)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("return error when cerebry repo return an error", func(t *testing.T) {
		// arrange
		cerebryRepo := &mock_external.MockCerebryRepo{}
		sut := CerebryUsecase{CerebryConfig: cfg, CerebryRepo: cerebryRepo}
		rootErr := fmt.Errorf("some err")
		cerebryRepo.On("GetUserToken", ctx, userID).Return("", rootErr)
		expectedErr := errors.New("CerebryUsecase.GetUserToken", rootErr)

		// act
		tok, err := sut.GenerateUserToken(ctx)

		// assert
		assert.Empty(t, tok)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns token from the cerebry repo", func(t *testing.T) {
		// arrange
		cerebryRepo := &mock_external.MockCerebryRepo{}
		sut := CerebryUsecase{CerebryConfig: cfg, CerebryRepo: cerebryRepo}
		cerebryRepo.On("GetUserToken", ctx, userID).Return("TOKEN_IS_THE_SECRET_STUFF", nil)

		// act
		tok, err := sut.GenerateUserToken(ctx)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, "TOKEN_IS_THE_SECRET_STUFF", tok)
	})
}
