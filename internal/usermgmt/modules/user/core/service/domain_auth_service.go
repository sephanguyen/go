package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	sppb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"firebase.google.com/go/v4/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DomainAuthService struct {
	pb.AuthServiceServer
	DB database.Ext
	// FirebaseAuthClient internal_auth_tenant.TenantClient
	TenantManager  internal_auth_tenant.TenantManager
	FirebaseClient *auth.Client

	ShamirClient       spb.TokenReaderServiceClient
	EmailServiceClient sppb.EmailModifierServiceClient

	ExternalConfigurationRepo interface {
		GetConfigurationByKeys(ctx context.Context, db database.QueryExecer, keys []string) ([]entity.DomainConfiguration, error)
	}
}

type (
	WhitelistIP struct {
		IPv4 []string `json:"ipv4"`
		IPv6 []string `json:"ipv6"`
	}

	EmailTemplate struct {
		Subject string
		Body    string
	}
)

func (s *DomainAuthService) ExchangeCustomToken(ctx context.Context, token string) (string, error) {
	resp, err := s.ShamirClient.VerifyTokenV2(ctx, &spb.VerifyTokenRequest{OriginalToken: token})
	if err != nil {
		return "", fmt.Errorf("ShamirClient.VerifyToken: %w", err)
	}

	tenantClient, err := s.TenantManager.TenantClient(ctx, resp.GetTenantId())
	if err != nil {
		return "", fmt.Errorf("tenantManager.TenantClient: %w", err)
	}
	customToken, err := tenantClient.CustomToken(ctx, resp.UserId)
	if err != nil {
		return "", fmt.Errorf("tenantClient.CustomToken: %w", err)
	}

	return customToken, nil
}

func (s *DomainAuthService) ValidateIP(ctx context.Context, userIP string) (bool, error) {
	zLogger := ctxzap.Extract(ctx)
	externalConfigurations, err := s.ExternalConfigurationRepo.GetConfigurationByKeys(ctx, s.DB, []string{constant.KeyIPRestrictionFeatureConfig, constant.KeyIPRestrictionWhitelistConfig})
	if err != nil {
		zLogger.Error(
			"cannot get configurations",
			zap.Error(err),
			zap.String("Repo", "ExternalConfigurationRepo.GetConfigurationByKeys"),
		)
		return true, err
	}
	var featureConfig, whitelistConfig entity.DomainConfiguration
	for _, config := range externalConfigurations {
		if config.ConfigKey().String() == constant.KeyIPRestrictionFeatureConfig {
			featureConfig = config
		}
		if config.ConfigKey().String() == constant.KeyIPRestrictionWhitelistConfig {
			whitelistConfig = config
		}
	}
	if featureConfig == nil || whitelistConfig == nil {
		zLogger.Error("missing configurations")
		return true, nil
	}
	if featureConfig.ConfigValue().String() != constant.ConfigValueOn {
		return true, nil
	}
	var whitelistIPs WhitelistIP
	err = json.Unmarshal([]byte(whitelistConfig.ConfigValue().String()), &whitelistIPs)
	if err != nil {
		zLogger.Error(
			"cannot Unmarshal whitelist IP JSON string",
			zap.Error(err),
		)
		return true, err
	}
	if len(whitelistIPs.IPv4) == 0 && len(whitelistIPs.IPv6) == 0 {
		return true, nil
	}
	if !golibs.InArrayString(userIP, append(whitelistIPs.IPv6, whitelistIPs.IPv4...)) {
		zLogger.Error("user IP is not allowed",
			zap.String("user IP:", userIP),
		)
		return false, nil
	}
	return true, nil
}

func (s *DomainAuthService) GetAuthInfo(ctx context.Context, username, domainName string) (*entity.Organization, *entity.AuthUser, error) {
	resp, err := s.ShamirClient.GetAuthInfo(ctx, &spb.GetAuthInfoRequest{Username: username, DomainName: domainName})
	if err != nil {
		return nil, nil, err
	}

	org := &entity.Organization{
		OrganizationID: database.Text(resp.GetOrganizationId()),
		TenantID:       database.Text(resp.GetTenantId()),
	}
	user := &entity.AuthUser{
		Email:      database.Text(resp.GetEmail()),
		LoginEmail: database.Text(resp.GetLoginEmail()),
		UserID:     database.Text(resp.GetUserId()),
	}

	// feature username is enabled only if the response of getAuthInfo of Shamir service has login_email
	enableUsername := resp.LoginEmail != ""

	// if feature username is enabled, use LoginEmail as email
	if !enableUsername {
		user.LoginEmail = user.Email
	}

	return org, user, nil
}

func (s *DomainAuthService) ResetPassword(ctx context.Context, username, domainName, langCode string) error {
	logger := ctxzap.Extract(ctx)

	org, user, err := s.GetAuthInfo(ctx, username, domainName)
	if err != nil {
		return err
	}

	tenantClient, err := s.TenantManager.TenantClient(ctx, org.TenantID.String)
	if err != nil {
		logger.Error("tenantManager.TenantClient",
			zap.String("TenantID", org.TenantID.String),
			zap.Error(err),
		)
		return status.Error(codes.Internal, errors.Wrap(err, "TenantManager.TenantClient").Error())
	}

	passwordResetLink, err := tenantClient.PasswordResetLink(ctx, user.LoginEmail.String, langCode)
	if err != nil {
		logger.Error("tenantClient.PasswordResetLink",
			zap.String("LoginEmail", user.LoginEmail.String),
			zap.Error(err),
		)
		return err
	}

	err = s.sendResetPasswordEmail(ctx, org.OrganizationID.String, user.Email.String, passwordResetLink, langCode)
	if err != nil {
		logger.Error("sendResetPasswordEmail", zap.Error(err))
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *DomainAuthService) sendResetPasswordEmail(ctx context.Context, organizationID, email, passwordResetLink, langCode string) error {
	logger := ctxzap.Extract(ctx)

	resetPasswordEmail := GenerateResetPasswordEmail(email, passwordResetLink, langCode)
	resp, err := s.EmailServiceClient.SendEmail(ctx, &sppb.SendEmailRequest{
		Subject:        resetPasswordEmail.Subject,
		Content:        &sppb.SendEmailRequest_EmailContent{HTML: resetPasswordEmail.Body},
		Recipients:     []string{email},
		OrganizationId: organizationID,
	})
	if err != nil {
		return errors.Wrap(err, "EmailServiceClient.SendEmail")
	}
	logger.Debug("EmailServiceClient.SendEmail", zap.Any("resp", resp))

	return nil
}
