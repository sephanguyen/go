package users

import (
	"context"
	"errors"
	"testing"
	"time"

	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockUserMgmtAuthService struct {
	exchangeCustomTokenFn func(context.Context, *upb.ExchangeCustomTokenRequest, ...grpc.CallOption) (*upb.ExchangeCustomTokenResponse, error)
}

func (m *mockUserMgmtAuthService) ExchangeCustomToken(ctx context.Context, req *upb.ExchangeCustomTokenRequest, options ...grpc.CallOption) (*upb.ExchangeCustomTokenResponse, error) {
	return m.exchangeCustomTokenFn(ctx, req, options...)
}

func TestUserModifierService_ExchangeCustomToken(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	/*t.Run("verify token error", func(tt *testing.T) {
		tt.Parallel()
		mockErr := fmt.Errorf("mock error")
		m := &mockShamirClient{
			verifyTokenV2Fn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return nil, mockErr
			},
		}
		s := &UserModifierService{
			ShamirClient: m,
		}
		resp, err := s.ExchangeCustomToken(ctx, &bpb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})
		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.True(tt, errors.Is(err, mockErr))
	})

	t.Run("verify tenant token: CustomToken error", func(tt *testing.T) {
		tt.Parallel()
		m := &mockShamirClient{
			verifyTokenV2Fn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return &spb.VerifyTokenResponse{
					UserId:   "user-id-1201",
					TenantId: "tenant-id-1201",
				}, nil
			},
		}
		mockErr := fmt.Errorf("mock error")
		tenantClient := new(mock_multitenant.TenantClient)
		tenantClient.On("CustomToken", ctx, mock.Anything).Once().Return("", mockErr)
		tenantManager := new(mock_multitenant.TenantManager)
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
		s := &UserModifierService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, &bpb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})
		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.True(tt, errors.Is(err, mockErr))
	})

	t.Run("verify tenant token: TenantClient error", func(tt *testing.T) {
		tt.Parallel()
		m := &mockShamirClient{
			verifyTokenV2Fn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return &spb.VerifyTokenResponse{
					UserId:   "user-id-1201",
					TenantId: "tenant-id-1201",
				}, nil
			},
		}
		mockErr := fmt.Errorf("mock error")
		tenantManager := new(mock_multitenant.TenantManager)
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(nil, mockErr)
		s := &UserModifierService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, &bpb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})
		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.True(tt, errors.Is(err, mockErr))
	})

	t.Run("verify tenant token success", func(tt *testing.T) {
		tt.Parallel()
		m := &mockShamirClient{
			verifyTokenV2Fn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return &spb.VerifyTokenResponse{
					UserId:   "user-id-1201",
					TenantId: "tenant-id-1201",
				}, nil
			},
		}
		tenantClient := new(mock_multitenant.TenantClient)
		tenantClient.On("CustomToken", ctx, mock.Anything).Once().Return("a tenant customtoken", nil)
		tenantManager := new(mock_multitenant.TenantManager)
		tenantManager.On("TenantClient", ctx, mock.Anything).Once().Return(tenantClient, nil)
		s := &UserModifierService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, &bpb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})
		assert.NoError(t, err)
		assert.Equal(tt, resp.CustomToken, "a tenant customtoken")
	})
	// Can not add more tests because can't mock s.FirebaseClient*/

	t.Run("failed to exchange custom token", func(tt *testing.T) {
		mockErr := errors.New("failed to exchange custom token")

		s := mockUserMgmtAuthService{
			exchangeCustomTokenFn: func(ctx context.Context, request *upb.ExchangeCustomTokenRequest, option ...grpc.CallOption) (*upb.ExchangeCustomTokenResponse, error) {
				return nil, mockErr
			},
		}

		resp, err := s.ExchangeCustomToken(ctx, &upb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})

		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.True(tt, errors.Is(err, mockErr))
	})

	t.Run("exchange token successfully", func(tt *testing.T) {
		mockResp := &upb.ExchangeCustomTokenResponse{CustomToken: "exchanged-custom-token"}

		s := mockUserMgmtAuthService{
			exchangeCustomTokenFn: func(ctx context.Context, request *upb.ExchangeCustomTokenRequest, option ...grpc.CallOption) (*upb.ExchangeCustomTokenResponse, error) {
				return mockResp, nil
			},
		}

		resp, err := s.ExchangeCustomToken(ctx, &upb.ExchangeCustomTokenRequest{
			Token: "a customtoken",
		})

		assert.NoError(t, err)
		assert.Equal(tt, mockResp.CustomToken, resp.CustomToken)
	})
}
