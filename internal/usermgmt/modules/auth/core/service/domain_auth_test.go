package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockShamirClient struct {
	verifyTokenFn           func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	exchangeTokenFn         func(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error)
	verifyTokenV2Fn         func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	verifySignatureFn       func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error)
	getAuthInfo             func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error)
	exchangeSalesforceToken func(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error)
}

func (m *mockShamirClient) VerifyToken(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
	return m.verifyTokenFn(ctx, in)
}

func (m *mockShamirClient) ExchangeToken(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error) {
	return m.exchangeTokenFn(ctx, in)
}

func (m *mockShamirClient) VerifyTokenV2(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
	return m.verifyTokenV2Fn(ctx, in)
}

func (m *mockShamirClient) VerifySignature(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
	return m.verifySignatureFn(ctx, in)
}

func (m *mockShamirClient) GetAuthInfo(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error) {
	return m.getAuthInfo(ctx, in)
}

func (m *mockShamirClient) ExchangeSalesforceToken(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error) {
	return m.exchangeSalesforceToken(ctx, in)
}

func TestUserModifierService_ExchangeToken(t *testing.T) {
	t.Parallel()

	t.Run("exchange salesforce token success", func(tt *testing.T) {
		tt.Parallel()

		s := &DomainAuthService{
			ShamirClient: &mockShamirClient{
				exchangeSalesforceToken: func(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error) {
					return &spb.ExchangeSalesforceTokenResponse{
						Token: "new token",
					}, nil
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

		resp, err := s.ExchangeSalesforceToken(ctx)
		assert.NoError(tt, err)
		assert.Equal(tt, "new token", resp)
	})

	t.Run("exchange salesforce token error", func(tt *testing.T) {
		tt.Parallel()
		mockErr := fmt.Errorf("exchange salesforce token error")
		s := &DomainAuthService{
			ShamirClient: &mockShamirClient{
				exchangeSalesforceToken: func(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error) {
					return nil, mockErr
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

		resp, err := s.ExchangeSalesforceToken(ctx)
		assert.Error(tt, err)
		assert.Empty(tt, resp)
		assert.EqualError(tt, err, "s.ShamirClient.ExchangeSalesforceToken: exchange salesforce token error")
	})

}
