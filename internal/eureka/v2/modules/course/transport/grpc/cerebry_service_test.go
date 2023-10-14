package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/course/usecase"
	pbv2 "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/stretchr/testify/assert"
)

func TestCerebryService_GetCerebryUserToken(t *testing.T) {
	t.Parallel()
	req := &pbv2.GetCerebryUserTokenRequest{}

	t.Run("Return error when there is an error occurred in usecase", func(t *testing.T) {
		// arrange
		ctx := context.Background()
		cerebryUsecase := &mock_usecase.MockCerebryUsecase{}
		sut := CerebryService{CerebryTokenGenerator: cerebryUsecase}
		rootErr := errors.New("Some err", nil)
		cerebryUsecase.On("GenerateUserToken", ctx).Return("", rootErr)
		expectedErr := errors.NewGrpcError(rootErr, transport.GrpcErrorMap)

		// act
		res, err := sut.GetCerebryUserToken(ctx, req)

		// assert
		assert.Nil(t, res)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Return token successfully", func(t *testing.T) {
		// arrange
		ctx := context.Background()
		cerebryUsecase := &mock_usecase.MockCerebryUsecase{}
		sut := CerebryService{CerebryTokenGenerator: cerebryUsecase}
		cerebryUsecase.On("GenerateUserToken", ctx).Return("KIEN_RANG_DEP", nil)

		// act
		res, err := sut.GetCerebryUserToken(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, "KIEN_RANG_DEP", res.Token)
	})
}
