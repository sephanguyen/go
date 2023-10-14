package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	mock_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/mock"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_spb "github.com/manabie-com/backend/mock/usermgmt/external_services"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	sppb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestAuthService_ExchangeCustomToken(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("verify token error", func(tt *testing.T) {
		tt.Parallel()
		mockErr := fmt.Errorf("mock error")
		m := &mockShamirClient{
			verifyTokenV2Fn: func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
				return nil, mockErr
			},
		}
		s := &DomainAuthService{
			ShamirClient: m,
		}
		resp, err := s.ExchangeCustomToken(ctx, "a customtoken")
		assert.Error(tt, err)
		assert.Empty(tt, resp)
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
		s := &DomainAuthService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, "a customtoken")
		assert.Error(tt, err)
		assert.Empty(tt, resp)
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
		s := &DomainAuthService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, "a customtoken")
		assert.Error(tt, err)
		assert.Empty(tt, resp)
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
		s := &DomainAuthService{
			ShamirClient:  m,
			TenantManager: tenantManager,
		}
		resp, err := s.ExchangeCustomToken(ctx, "a customtoken")
		assert.NoError(t, err)
		assert.Equal(tt, resp, "a tenant customtoken")
	})
	// Can not add more tests because can't mock s.FirebaseClient
}

func TestService_ValidateUserIP(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockFeatureConfig := mock_usermgmt.Configuration{
		RandomConfiguration: mock_usermgmt.RandomConfiguration{
			ConfigKey:   field.NewString("user.authentication.ip_address_restriction"),
			ConfigValue: field.NewString("on"),
		},
	}
	mockWhitelistConfig := mock_usermgmt.Configuration{
		RandomConfiguration: mock_usermgmt.RandomConfiguration{
			ConfigKey:   field.NewString("user.authentication.allowed_ip_address"),
			ConfigValue: field.NewString(`{"ipv4":["183.80.142.65"],"ipv6":["2405:4802:9035:7ff0:8863:b99b:9d4e:a932"]}`),
		},
	}
	tests := []struct {
		name     string
		userIP   string
		wantResp bool
		setup    func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo)
	}{
		{
			name:     "user IP is in whitelist",
			userIP:   "2405:4802:9035:7ff0:8863:b99b:9d4e:a932",
			wantResp: true,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return([]entity.DomainConfiguration{mockFeatureConfig, mockWhitelistConfig}, nil)
			},
		},
		{
			name:     "user IP is not in whitelist",
			userIP:   "1535:3856:8352:3fd9:8125:bu1b:9d4e:0i13",
			wantResp: false,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return([]entity.DomainConfiguration{mockFeatureConfig, mockWhitelistConfig}, nil)
			},
		},
		{
			name:     "skip validate when missing configurations",
			userIP:   "2405:4802:9035:7ff0:8863:b99b:9d4e:a932",
			wantResp: true,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return([]entity.DomainConfiguration{mockFeatureConfig}, nil)
			},
		},
		{
			name:     "skip validate when feature is not on",
			userIP:   "2405:4802:9035:7ff0:8863:b99b:9d4e:a932",
			wantResp: true,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return([]entity.DomainConfiguration{mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigKey:   field.NewString("user.authentication.ip_address_restriction"),
						ConfigValue: field.NewString("off"),
					},
				}, mockWhitelistConfig}, nil)
			},
		},
		{
			name:     "skip validate when get external configuration failed",
			userIP:   "2405:4802:9035:7ff0:8863:b99b:9d4e:a932",
			wantResp: true,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return(nil, errors.New("query error"))
			},
		},
		{
			name:     "skip validate when whitelist is empty",
			userIP:   "2405:4802:9035:7ff0:8863:b99b:9d4e:a932",
			wantResp: true,
			setup: func(ctx context.Context, repo *mock_repositories.MockDomainConfigurationRepo) {
				repo.On("GetConfigurationByKeys", ctx, mock.Anything, []string{"user.authentication.ip_address_restriction", "user.authentication.allowed_ip_address"}).Return([]entity.DomainConfiguration{mockFeatureConfig, mock_usermgmt.Configuration{
					RandomConfiguration: mock_usermgmt.RandomConfiguration{
						ConfigKey:   field.NewString("user.authentication.allowed_ip_address"),
						ConfigValue: field.NewString(`{"ipv4":[],"ipv6":[]}`),
					},
				}}, nil)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mock_repositories.MockDomainConfigurationRepo{}
			if tc.setup != nil {
				tc.setup(ctx, mockRepo)
			}
			service := &DomainAuthService{
				DB:                        &mock_database.Tx{},
				ExternalConfigurationRepo: mockRepo,
			}
			resp, _ := service.ValidateIP(ctx, tc.userIP)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestDomainAuthService_GetAuthInfo(t *testing.T) {
	type args struct {
		ctx        context.Context
		username   string
		domainName string
	}
	tests := []struct {
		name             string
		args             args
		setup            func() *DomainAuthService
		wantOrganization *entity.Organization
		wantAuthUser     *entity.AuthUser
		wantErr          error
	}{
		{
			name: "get auth info when enabled username",
			args: args{},
			setup: func() *DomainAuthService {
				return &DomainAuthService{
					ShamirClient: &mockShamirClient{
						getAuthInfo: func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error) {
							return &spb.GetAuthInfoResponse{
								LoginEmail:     "LoginEmail",
								TenantId:       "TenantId",
								Email:          "Email",
								OrganizationId: "OrganizationId",
								UserId:         "UserId",
							}, nil
						},
					},
				}
			},
			wantOrganization: &entity.Organization{
				OrganizationID: database.Text("OrganizationId"),
				TenantID:       database.Text("TenantId"),
			},
			wantAuthUser: &entity.AuthUser{
				Email:      database.Text("Email"),
				LoginEmail: database.Text("LoginEmail"),
				UserID:     database.Text("UserId"),
			},
			wantErr: nil,
		},
		{
			name: "get auth info when disabled username",
			args: args{},
			setup: func() *DomainAuthService {
				return &DomainAuthService{
					ShamirClient: &mockShamirClient{
						getAuthInfo: func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error) {
							return &spb.GetAuthInfoResponse{
								TenantId:       "TenantId",
								Email:          "Email",
								OrganizationId: "OrganizationId",
								UserId:         "UserId",
							}, nil
						},
					},
				}
			},
			wantOrganization: &entity.Organization{
				OrganizationID: database.Text("OrganizationId"),
				TenantID:       database.Text("TenantId"),
			},
			wantAuthUser: &entity.AuthUser{
				Email:      database.Text("Email"),
				LoginEmail: database.Text("Email"),
				UserID:     database.Text("UserId"),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			org, user, err := s.GetAuthInfo(tt.args.ctx, tt.args.username, tt.args.domainName)
			assert.Equalf(t, tt.wantOrganization, org, "GetAuthInfo(%v, %v, %v)", tt.args.ctx, tt.args.username, tt.args.domainName)
			assert.Equalf(t, tt.wantAuthUser, user, "GetAuthInfo(%v, %v, %v)", tt.args.ctx, tt.args.username, tt.args.domainName)
			assert.Equalf(t, tt.wantErr, err, "GetAuthInfo(%v, %v, %v)", tt.args.ctx, tt.args.username, tt.args.domainName)
		})
	}
}

func TestDomainAuthService_sendResetPasswordEmail(t *testing.T) {
	type args struct {
		ctx               context.Context
		organizationID    string
		email             string
		passwordResetLink string
		langCode          string
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *DomainAuthService
		wantErr error
	}{
		{
			name: "send reset password link by email service",
			args: args{
				ctx:               context.Background(),
				organizationID:    "organizationID",
				email:             "email",
				passwordResetLink: "passwordResetLink",
				langCode:          "langCode",
			},
			setup: func() *DomainAuthService {
				emailService := new(mock_spb.MockemailModifierServiceClient)
				emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).
					Return(&sppb.SendEmailResponse{}, nil)
				return &DomainAuthService{EmailServiceClient: emailService}
			},
			wantErr: nil,
		},
		{
			name: "send reset password link by email service with error",
			args: args{
				ctx:               context.Background(),
				organizationID:    "organizationID",
				email:             "email",
				passwordResetLink: "passwordResetLink",
				langCode:          "langCode",
			},
			setup: func() *DomainAuthService {
				emailService := new(mock_spb.MockemailModifierServiceClient)
				emailService.On("SendEmail", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
				return &DomainAuthService{EmailServiceClient: emailService}
			},
			wantErr: errors.Wrap(assert.AnError, "EmailServiceClient.SendEmail"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			err := s.sendResetPasswordEmail(tt.args.ctx, tt.args.organizationID, tt.args.email, tt.args.passwordResetLink, tt.args.langCode)
			if err != nil || tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
		})
	}
}

func TestDomainAuthService_ResetPassword(t *testing.T) {
	type args struct {
		ctx        context.Context
		username   string
		domainName string
		langCode   string
	}
	tests := []struct {
		name    string
		args    args
		setup   func() *DomainAuthService
		wantErr error
	}{
		{
			name: "happy case",
			args: args{
				ctx:        context.Background(),
				username:   "username",
				domainName: "domainName",
				langCode:   constant.EnglishLanguageCode,
			},
			setup: func() *DomainAuthService {
				tenantClient := new(mock_multitenant.TenantClient)
				tenantManager := new(mock_multitenant.TenantManager)
				emailService := new(mock_spb.MockemailModifierServiceClient)
				samplePasswordResetLink := "https://firebase.net/resetlink"
				authInfoResp := &spb.GetAuthInfoResponse{
					LoginEmail:     "LoginEmail",
					TenantId:       "TenantId",
					Email:          "Email",
					OrganizationId: "OrganizationId",
					UserId:         "UserId",
				}

				tenantManager.On("TenantClient", mock.Anything, mock.Anything).Return(tenantClient, nil)
				tenantClient.On("PasswordResetLink", mock.Anything, authInfoResp.LoginEmail, constant.EnglishLanguageCode).Return(samplePasswordResetLink, nil)
				resetPasswordEmail := GenerateResetPasswordEmail(authInfoResp.Email, samplePasswordResetLink, constant.EnglishLanguageCode)
				emailService.On("SendEmail", mock.Anything, &sppb.SendEmailRequest{
					Subject:        resetPasswordEmail.Subject,
					Content:        &sppb.SendEmailRequest_EmailContent{HTML: resetPasswordEmail.Body},
					Recipients:     []string{authInfoResp.Email},
					OrganizationId: authInfoResp.OrganizationId,
				}, mock.Anything).Return(&sppb.SendEmailResponse{}, nil)

				return &DomainAuthService{
					ShamirClient: &mockShamirClient{
						getAuthInfo: func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error) {
							return authInfoResp, nil
						},
					},
					TenantManager:      tenantManager,
					EmailServiceClient: emailService,
				}
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			err := s.ResetPassword(tt.args.ctx, tt.args.username, tt.args.domainName, tt.args.langCode)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
