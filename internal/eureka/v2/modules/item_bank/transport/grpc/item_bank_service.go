package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/item_bank/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ItemBankService struct {
	ActivityGetter usecase.ActivityGetter
}

func (i *ItemBankService) GetItemsByLM(_ context.Context, _ *epb.GetItemsByLMRequest) (*epb.GetItemsByLMResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetItemsByLM not implemented")
}

func NewItemBankService(activityUsecase *usecase.ActivityUsecase) *ItemBankService {
	return &ItemBankService{
		ActivityGetter: activityUsecase,
	}
}

func (i *ItemBankService) GetTotalItemsByLM(ctx context.Context, req *epb.GetTotalItemsByLMRequest) (*epb.GetTotalItemsByLMResponse, error) {
	count, err := i.ActivityGetter.CountTotalLearnosityItemByLM(ctx, req.LearningMaterialId)

	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &epb.GetTotalItemsByLMResponse{TotalItems: count}, nil
}
