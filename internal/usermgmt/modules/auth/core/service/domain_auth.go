package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
)

type DomainAuthService struct {
	ShamirClient spb.TokenReaderServiceClient
}

func (s *DomainAuthService) ExchangeSalesforceToken(ctx context.Context) (string, error) {
	userID := interceptors.UserIDFromContext(ctx)
	orgID := golibs.ResourcePathFromCtx(ctx)

	token, err := s.ShamirClient.ExchangeSalesforceToken(ctx, &spb.ExchangeSalesforceTokenRequest{
		UserId:         userID,
		OrganizationId: orgID,
	})
	if err != nil {
		return "", fmt.Errorf("s.ShamirClient.ExchangeSalesforceToken: %s", err.Error())
	}

	return token.Token, nil
}
