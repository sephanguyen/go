package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type mockDomainAuthService struct {
	exchangeCustomToken func(ctx context.Context) (string, error)
	validateIP          func(ctx context.Context) (bool, error)
	getAuthInfo         func(ctx context.Context, username, domainName string) (*entity.Organization, *entity.AuthUser, error)
	resetPassword       func(ctx context.Context, username, domainName, langCode string) error
}

func (m *mockDomainAuthService) ValidateIP(ctx context.Context, userIP string) (bool, error) {
	return m.validateIP(ctx)
}

func (m *mockDomainAuthService) ExchangeCustomToken(ctx context.Context, token string) (string, error) {
	return m.exchangeCustomToken(ctx)
}

func (m *mockDomainAuthService) GetAuthInfo(ctx context.Context, username, domainName string) (*entity.Organization, *entity.AuthUser, error) {
	return m.getAuthInfo(ctx, username, domainName)
}

func (m *mockDomainAuthService) ResetPassword(ctx context.Context, username, domainName, langCode string) error {
	return m.resetPassword(ctx, username, domainName, langCode)
}

func TestAuthService_ValidateUserIP(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tests := []struct {
		name          string
		md            metadata.MD
		domainService *mockDomainAuthService
		want          bool
	}{
		{
			name: "user IP is in whitelist",
			md:   metadata.MD{"cf-connecting-ip": []string{"2405:4802:9035:7ff0:8863:b99b:9d4e:a932"}},
			domainService: &mockDomainAuthService{
				validateIP: func(ctx context.Context) (bool, error) {
					return true, nil
				},
			},
			want: true,
		},
		{
			name: "user IP is not in whitelist",
			md:   metadata.MD{"cf-connecting-ip": []string{"2405:4802:9035:7ff0:8863:b99b:9d4e:a932"}},
			domainService: &mockDomainAuthService{
				validateIP: func(ctx context.Context) (bool, error) {
					return false, nil
				},
			},
			want: false,
		},
		{
			name: "skip validate when user IP is not in metadata",
			md:   metadata.MD{"cf-connecting-ip": []string{}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{
				DomainAuthService: tt.domainService,
			}
			ctx := metadata.NewIncomingContext(ctx, tt.md)
			resp, _ := service.ValidateUserIP(ctx, &pb.ValidateUserIPRequest{})
			assert.Equal(t, tt.want, resp.Allow)
		})
	}
}

func TestAuthService_ResetPassword(t *testing.T) {
	type args struct {
		ctx     context.Context
		request *pb.ResetPasswordRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *AuthService
		want    *pb.ResetPasswordResponse
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx: context.Background(),
				request: &pb.ResetPasswordRequest{
					Username:     "Username",
					DomainName:   "DomainName",
					LanguageCode: "en",
				},
			},
			setup: func() *AuthService {
				return &AuthService{
					DomainAuthService: &mockDomainAuthService{
						resetPassword: func(ctx context.Context, username, domainName, langCode string) error {
							return nil
						},
					},
				}
			},
			want:    &pb.ResetPasswordResponse{},
			wantErr: nil,
		},
		{
			name: "happy case with invalid language code",
			args: args{
				ctx: context.Background(),
				request: &pb.ResetPasswordRequest{
					Username:     "Username",
					DomainName:   "DomainName",
					LanguageCode: "invalid",
				},
			},
			setup: func() *AuthService {
				return &AuthService{
					DomainAuthService: &mockDomainAuthService{
						resetPassword: func(ctx context.Context, username, domainName, langCode string) error {
							return nil
						},
					},
				}
			},
			want:    &pb.ResetPasswordResponse{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ResetPassword(tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.want, got, "ResetPassword(%v, %v)", tt.args.ctx, tt.args.request)
			assert.Equalf(t, tt.wantErr, err, "ResetPassword(%v, %v)", tt.args.ctx, tt.args.request)
		})
	}
}
