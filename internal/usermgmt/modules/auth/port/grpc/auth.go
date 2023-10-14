package grpc

import (
	"context"

	apb "github.com/manabie-com/backend/pkg/manabuf/auth/v1"
)

type AuthService struct {
	apb.UnimplementedAuthServiceServer

	DomainAuthService interface {
		ExchangeSalesforceToken(ctx context.Context) (string, error)
	}
}

func (s *AuthService) ExchangeSalesforceToken(ctx context.Context, _ *apb.ExchangeSalesforceTokenRequest) (*apb.ExchangeSalesforceTokenResponse, error) {
	token, err := s.DomainAuthService.ExchangeSalesforceToken(ctx)
	if err != nil {
		return nil, err
	}

	return &apb.ExchangeSalesforceTokenResponse{Token: token}, nil
}
