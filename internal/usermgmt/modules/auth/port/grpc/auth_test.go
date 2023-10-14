package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	apb "github.com/manabie-com/backend/pkg/manabuf/auth/v1"
	"github.com/stretchr/testify/assert"
)

type DomainAuthService interface {
	ExchangeSalesforceToken(ctx context.Context) (string, error)
}

type mockDomainAuthService struct {
	exchangeSalesforceToken func(ctx context.Context) (string, error)
}

func (d *mockDomainAuthService) ExchangeSalesforceToken(ctx context.Context) (string, error) {
	return d.exchangeSalesforceToken(ctx)
}

func TestAuthService_ExchangeSalesforceToken(t *testing.T) {
	t.Parallel()

	t.Run("exchange salesforce token success", func(tt *testing.T) {
		tt.Parallel()

		s := &AuthService{
			DomainAuthService: &mockDomainAuthService{
				exchangeSalesforceToken: func(ctx context.Context) (string, error) {
					return "salesforce access token", nil
				},
			},
		}

		claim := &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprint(constants.ManabieSchool),
			},
		}
		ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)
		ctx = interceptors.ContextWithUserID(ctx, "user-id")

		resp, err := s.ExchangeSalesforceToken(ctx, &apb.ExchangeSalesforceTokenRequest{})
		assert.NoError(tt, err)
		assert.Equal(tt, "salesforce access token", resp.Token)
	})

	t.Run("exchange salesforce token failed", func(tt *testing.T) {
		tt.Parallel()

		expectedError := fmt.Errorf("request access token failed")
		s := &AuthService{
			DomainAuthService: &mockDomainAuthService{
				exchangeSalesforceToken: func(ctx context.Context) (string, error) {
					return "", fmt.Errorf("request access token failed")
				},
			},
		}

		claim := &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprint(constants.ManabieSchool),
			},
		}
		ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)
		ctx = interceptors.ContextWithUserID(ctx, "user-id")

		resp, err := s.ExchangeSalesforceToken(ctx, &apb.ExchangeSalesforceTokenRequest{})
		assert.Equal(tt, expectedError, err)
		assert.Nil(tt, resp)
	})
}
