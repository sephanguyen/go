package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	pbv2 "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

type CerebryService struct {
	CerebryTokenGenerator usecase.CerebryTokenGenerator
}

func NewCerebryService(cerebryUsecase *usecase.CerebryUsecase) *CerebryService {
	return &CerebryService{
		CerebryTokenGenerator: cerebryUsecase,
	}
}

func (c *CerebryService) GetCerebryUserToken(ctx context.Context, _ *pbv2.GetCerebryUserTokenRequest) (res *pbv2.GetCerebryUserTokenResponse, err error) {
	tok, err := c.CerebryTokenGenerator.GenerateUserToken(ctx)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pbv2.GetCerebryUserTokenResponse{Token: tok}, nil
}
