package users

import (
	"context"
	"fmt"
	"testing"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockShamirClient struct {
	verifyTokenFn     func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	exchangeTokenFn   func(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error)
	verifyTokenV2Fn   func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	verifySignatureFn func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error)
	getAuthInfo       func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error)
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
	t.Run("verify token error", func(tt *testing.T) {
		tt.Parallel()
		mockErr := fmt.Errorf("mock error")
		m := &mockShamirClient{
			verifyTokenFn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return nil, mockErr
			},
		}
		s := &UserModifierService{
			ShamirClient: m,
		}
		resp, err := s.ExchangeToken(context.Background(), &bpb.ExchangeTokenRequest{
			Token: "a token",
		})
		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.EqualError(tt, err, "s.ShamirClient.VerifyToken: mock error")
	})

	t.Run("exchange token error", func(tt *testing.T) {
		tt.Parallel()
		mockErr := fmt.Errorf("mock error")
		s := &UserModifierService{
			ShamirClient: &mockShamirClient{
				verifyTokenFn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
					return &spb.VerifyTokenResponse{
						UserId: "user-id",
					}, nil
				},
				exchangeTokenFn: func(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error) {
					return nil, mockErr
				},
			},
		}
		resp, err := s.ExchangeToken(context.Background(), &bpb.ExchangeTokenRequest{
			Token: "a token",
		})
		assert.Error(tt, err)
		assert.Nil(tt, resp)
		assert.EqualError(tt, err, "s.ShamirClient.ExchangeToken: mock error")
	})

	t.Run("exchange token success", func(tt *testing.T) {
		tt.Parallel()
		s := &UserModifierService{
			ApplicantID: "applicant-id-1",
			ShamirClient: &mockShamirClient{
				verifyTokenFn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
					return &spb.VerifyTokenResponse{
						UserId: "user-id-12",
					}, nil
				},
				exchangeTokenFn: func(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error) {
					assert.Equal(tt, "a token", in.OriginalToken)
					assert.Equal(tt, "applicant-id-1", in.NewTokenInfo.Applicant)
					assert.Equal(tt, "user-id-12", in.NewTokenInfo.UserId)

					return &spb.ExchangeTokenResponse{
						NewToken: "new token",
					}, nil
				},
			},
		}

		resp, err := s.ExchangeToken(context.Background(), &bpb.ExchangeTokenRequest{
			Token: "a token",
		})
		assert.NoError(tt, err)
		assert.Equal(tt, "new token", resp.Token)
	})
}
