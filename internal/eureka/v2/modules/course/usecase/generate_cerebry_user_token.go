package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func (c *CerebryUsecase) GenerateUserToken(ctx context.Context) (tok string, err error) {
	requesterID := interceptors.UserIDFromContext(ctx)

	if requesterID == "" {
		return "", errors.New("CerebryUsecase.GenerateUserToken: Context does not contain user id", nil)
	}

	tok, err = c.CerebryRepo.GetUserToken(ctx, requesterID)
	if err != nil {
		return "", errors.New("CerebryUsecase.GetUserToken", err)
	}

	return tok, nil
}
