package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"google.golang.org/grpc/metadata"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer

	DomainAuthService interface {
		ValidateIP(ctx context.Context, userIP string) (bool, error)
		ExchangeCustomToken(ctx context.Context, token string) (string, error)
		GetAuthInfo(ctx context.Context, username, domainName string) (*entity.Organization, *entity.AuthUser, error)
		ResetPassword(ctx context.Context, username, domainName, langCode string) error
	}
}

func (s *AuthService) ExchangeCustomToken(ctx context.Context, req *pb.ExchangeCustomTokenRequest) (*pb.ExchangeCustomTokenResponse, error) {
	customToken, err := s.DomainAuthService.ExchangeCustomToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	return &pb.ExchangeCustomTokenResponse{CustomToken: customToken}, nil
}

func (s *AuthService) ValidateUserIP(ctx context.Context, _ *pb.ValidateUserIPRequest) (*pb.ValidateUserIPResponse, error) {
	zLogger := ctxzap.Extract(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		zLogger.Error("cannot get metadata from context")
		return &pb.ValidateUserIPResponse{
			Allow: true,
		}, nil
	}

	userIP, ok := md["cf-connecting-ip"]
	if !ok {
		zLogger.Error("cannot get cf-connecting-ip from context")
		return &pb.ValidateUserIPResponse{
			Allow: true,
		}, nil
	}
	if len(userIP) == 0 {
		zLogger.Error("user IP is not sent",
			zap.Strings("user IPs:", userIP),
		)
		return &pb.ValidateUserIPResponse{
			Allow: true,
		}, nil
	}
	zLogger.Info("User IP Address", zap.Strings("IP list", userIP))
	allow, err := s.DomainAuthService.ValidateIP(ctx, userIP[0])
	if err != nil {
		zLogger.Error("can not validate user ip",
			zap.Error(err),
			zap.String("DomainAuthService", "ValidateIP"),
		)
		return &pb.ValidateUserIPResponse{
			Allow: true,
		}, nil
	}
	return &pb.ValidateUserIPResponse{
		Allow: allow,
	}, nil
}

func (s *AuthService) GetAuthInfo(ctx context.Context, request *pb.GetAuthInfoRequest) (*pb.GetAuthInfoResponse, error) {
	logger := ctxzap.Extract(ctx)

	org, user, err := s.DomainAuthService.GetAuthInfo(ctx, request.Username, request.DomainName)
	if err != nil {
		logger.Error("can not get auth info",
			zap.String("username", request.Username),
			zap.String("domain_name", request.DomainName),
		)
		return nil, err
	}

	return &pb.GetAuthInfoResponse{
		TenantId:   org.TenantID.String,
		LoginEmail: user.LoginEmail.String,
	}, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, request *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	logger := ctxzap.Extract(ctx)

	// use default language english if language code is invalid or empty
	if _, err := language.Parse(request.LanguageCode); err != nil {
		request.LanguageCode = constant.EnglishLanguageCode
	}

	err := s.DomainAuthService.ResetPassword(ctx, request.GetUsername(), request.GetDomainName(), request.GetLanguageCode())
	if err != nil {
		logger.Error("can not reset password",
			zap.String("username", request.GetUsername()),
			zap.String("domain_name", request.GetDomainName()),
			zap.String("language_code", request.GetLanguageCode()),
			zap.Error(err),
		)
		return nil, err
	}

	return &pb.ResetPasswordResponse{}, nil
}
