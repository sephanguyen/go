package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/shamir/services"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/square/go-jose/v3/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements gRPC
type Service struct {
	spb.UnimplementedTokenReaderServiceServer

	UnleashClient unleashclient.ClientInstance

	Verifier interface {
		ExchangeVerifiedToken(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error)
		Verify(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error)
	}

	DB                database.Ext
	AuthDB            database.Ext
	SalesforceService interface {
		GetAccessToken(orgID, userID string) (string, error)
	}
	DefaultOrganizationAuthValues string

	UserRepoV2 interface {
		GetByAuthInfo(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error)
		GetByAuthInfoV2(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.AuthUser, error)
		GetByUsername(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error)
		GetByEmail(ctx context.Context, db database.QueryExecer, username string, organizationID string) (*entity.AuthUser, error)
	}
	DomainAPIKeypairRepo interface {
		GetByPublicKey(ctx context.Context, db database.QueryExecer, publicKey string) (entity.DomainAPIKeypair, error)
	}
	OrganizationRepo interface {
		GetByDomainName(ctx context.Context, db database.QueryExecer, domainName string) (*entity.Organization, error)
		// 	GetByID(ctx context.Context, db database.QueryExecer, orgID string) ()
	}
	OrganizationRepoV2 interface {
		GetSalesforceClientIDByOrganizationID(ctx context.Context, db database.QueryExecer, organizationID string) (string, error)
	}
	Env            string
	FeatureManager interface {
		IsEnableUsernameStudentParentStaff(ctx context.Context, org valueobj.HasOrganizationID) bool
		IsEnableDecouplingUserAndAuthDB(org valueobj.HasOrganizationID) bool
	}
}

func (s *Service) exchangeVerifiedToken(ctx context.Context, newTokenInfo *interceptors.TokenInfo, verifiedIDToken *interceptors.CustomClaims) (string, *interceptors.TokenInfo, error) {
	// Only verified tokens by Identity Platform can have tenant field
	if verifiedIDToken.JwkURL != auth.FirebaseAndIdentityJwkURL {
		if verifiedIDToken.FirebaseClaims != nil {
			verifiedIDToken.FirebaseClaims.Identity.Tenant = ""
		}
	}

	isEnablingAuthDBConnection, err := s.UnleashClient.IsFeatureEnabledOnOrganization(unleash.FeatureDecouplingUserAndAuthDB, s.Env, verifiedIDToken.OrganizationID().String())
	if err != nil {
		isEnablingAuthDBConnection = false
	}

	userID := verifiedIDToken.Claims.Subject
	projectID := verifiedIDToken.GetProjectID()
	tenantID := verifiedIDToken.GetTenantID()

	user := &entity.LegacyUser{}
	if isEnablingAuthDBConnection && s.AuthDB != nil {
		authUser, err := s.UserRepoV2.GetByAuthInfoV2(ctx, s.AuthDB, s.DefaultOrganizationAuthValues, userID, projectID, tenantID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get user V2: %w", err)
		}
		user.UserID = authUser.UserID
		user.ResourcePath = authUser.ResourcePath
		user.Group = authUser.UserGroup
	} else {
		user, err = s.UserRepoV2.GetByAuthInfo(ctx, s.DB, s.DefaultOrganizationAuthValues, userID, projectID, tenantID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	if user.DeactivatedAt.Status == pgtype.Present {
		return "", nil, errorx.ErrDeactivatedUser
	}

	schoolID, err := strconv.Atoi(user.ResourcePath.String)
	if err != nil {
		return "", nil, err
	}
	newTokenInfo.DefaultRole = user.Group.String
	newTokenInfo.AllowedRoles = []string{user.Group.String}
	newTokenInfo.SchoolIds = []int64{int64(schoolID)}
	newTokenInfo.UserGroup = user.Group.String
	newTokenInfo.ResourcePath = user.ResourcePath.String

	newToken, err := s.Verifier.ExchangeVerifiedToken(verifiedIDToken, newTokenInfo)
	if err != nil {
		if errors.Is(err, services.ErrWrongDivision) {
			return "", nil, status.Error(codes.FailedPrecondition, err.Error())
		}

		return "", nil, err
	}
	return newToken, newTokenInfo, nil
}

// ExchangeToken implements gRPC method
func (s *Service) ExchangeToken(ctx context.Context, req *spb.ExchangeTokenRequest) (*spb.ExchangeTokenResponse, error) {
	newTokenInfo := &interceptors.TokenInfo{
		Applicant: req.NewTokenInfo.Applicant,
		UserID:    req.NewTokenInfo.UserId,
	}

	verifiedIDToken, err := s.Verifier.Verify(ctx, req.OriginalToken)
	if err != nil {
		return nil, err
	}
	token, _, err := s.exchangeVerifiedToken(ctx, newTokenInfo, verifiedIDToken)
	if err != nil {
		return nil, err
	}
	return &spb.ExchangeTokenResponse{
		NewToken: token,
	}, nil
}

// VerifyToken implements gRPC method
func (s *Service) VerifyToken(ctx context.Context, req *spb.VerifyTokenRequest) (*spb.VerifyTokenResponse, error) {
	claims, err := s.Verifier.Verify(ctx, req.OriginalToken)
	if err != nil {
		return nil, err
	}

	return &spb.VerifyTokenResponse{
		UserId:   claims.Subject,
		TenantId: claims.GetTenantID(),
	}, nil
}

func (s *Service) VerifyTokenV2(ctx context.Context, req *spb.VerifyTokenRequest) (*spb.VerifyTokenResponse, error) {
	claims, err := s.Verifier.Verify(ctx, req.OriginalToken)
	if err != nil {
		return nil, err
	}

	userID := claims.Claims.Subject
	projectID := claims.GetProjectID()
	tenantID := claims.GetTenantID()

	isEnablingAuthDBConnection, err := s.UnleashClient.IsFeatureEnabledOnOrganization(unleash.FeatureDecouplingUserAndAuthDB, s.Env, claims.OrganizationID().String())
	if err != nil {
		isEnablingAuthDBConnection = false
	}

	user := &entity.LegacyUser{}
	if isEnablingAuthDBConnection && s.AuthDB != nil {
		authUser, err := s.UserRepoV2.GetByAuthInfoV2(ctx, s.AuthDB, s.DefaultOrganizationAuthValues, userID, projectID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user V2: %w", err)
		}
		user.UserID = authUser.UserID
		user.ResourcePath = authUser.ResourcePath
		user.Group = authUser.UserGroup
	} else {
		user, err = s.UserRepoV2.GetByAuthInfo(ctx, s.DB, s.DefaultOrganizationAuthValues, userID, projectID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}
	if user.Group.String != entities.UserGroupStudent && user.Group.String != entities.UserGroupParent {
		return nil, errors.New("not belong to student or parent group")
	}

	return &spb.VerifyTokenResponse{
		UserId:   userID,
		TenantId: tenantID,
	}, nil
}

func (s *Service) ExchangeSalesforceToken(ctx context.Context, req *spb.ExchangeSalesforceTokenRequest) (*spb.ExchangeSalesforceTokenResponse, error) {
	// salesforceClientID, err := s.OrganizationRepoV2.GetSalesforceClientIDByOrganizationID(ctx, s.DB, req.OrganizationId)
	// if err != nil {
	// 	return nil, err
	// }

	token, err := s.SalesforceService.GetAccessToken(req.OrganizationId, req.UserId)
	if err != nil {
		return nil, err
	}
	return &spb.ExchangeSalesforceTokenResponse{
		Token: token,
	}, nil
}

func fakeVerifiedClaimAndNewTokenInfo(env string, req *spb.GenerateFakeTokenRequest) (verifiedClaim *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo, err error) {
	// jwt applicant found in deployments/helm/backend/bob/configs/.../bob.config.yaml
	var jwtApplicant string

	switch env {
	case "local":
		jwtApplicant = "manabie-local"
	case "stag":
		jwtApplicant = "manabie-stag"
	case "uat":
		jwtApplicant = "manabie-stag"
	case "prod":
		mapSchoolIDWithJWTApplicantInProd := map[string]string{
			fmt.Sprint(constants.AICSchool):       "prod-aic",
			fmt.Sprint(constants.GASchool):        "prod-ga",
			fmt.Sprint(constants.RenseikaiSchool): "prod-renseikai",
			fmt.Sprint(constants.SynersiaSchool):  "prod-synersia",
		}

		jwtApplicant = mapSchoolIDWithJWTApplicantInProd[req.GetSchoolId()]
		if jwtApplicant == "" {
			jwtApplicant = "prod-tokyo"
		}
	default:
		return nil, nil, fmt.Errorf("api not allowed for env %s", env)
	}

	newTokenInfo = &interceptors.TokenInfo{
		Applicant: jwtApplicant,
		UserID:    req.GetUserId(),
	}

	verifiedClaim = fakeVerifiedCustomClaim(req.GetUserId(), req.GetTenantId(), req.GetProjectId())
	return
}

func (s *Service) GenerateFakeToken(ctx context.Context, req *spb.GenerateFakeTokenRequest) (*spb.GenerateFakeTokenResponse, error) {
	verifiedClaim, newTokenInfo, err := fakeVerifiedClaimAndNewTokenInfo(s.Env, req)
	if err != nil {
		return nil, err
	}
	token, newTokenInfo, err := s.exchangeVerifiedToken(ctx, newTokenInfo, verifiedClaim)
	if err != nil {
		return nil, err
	}
	if newTokenInfo.ResourcePath != req.SchoolId {
		return nil, fmt.Errorf("sanity check failed, resource path of id (%s) is different from resource path from request (%s)",
			newTokenInfo.ResourcePath, req.SchoolId)
	}

	return &spb.GenerateFakeTokenResponse{
		Token: token,
	}, nil
}

func (s *Service) VerifySignature(ctx context.Context, req *spb.VerifySignatureRequest) (*spb.VerifySignatureResponse, error) {
	isEnablingAuthDBConnection, err := s.UnleashClient.IsFeatureEnabled(unleash.FeatureDecouplingUserAndAuthDB, s.Env)
	if err != nil {
		isEnablingAuthDBConnection = false
	}

	db := s.DB
	if isEnablingAuthDBConnection && s.AuthDB != nil {
		db = s.AuthDB
	}

	apiKeypair, err := s.DomainAPIKeypairRepo.GetByPublicKey(ctx, db, req.PublicKey)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, status.Error(codes.PermissionDenied, errorx.ErrShamirInvalidPublicKey.Error())
		default:
			return nil, status.Error(codes.Internal, errors.Wrap(err, "s.DomainAPIKeypairRepo.GetByPublicKey").Error())
		}
	}

	mac := hmac.New(sha256.New, []byte(apiKeypair.PrivateKey().String()))
	_, err = mac.Write(req.Body)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "mac.Write").Error())
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	if signature != req.Signature {
		return nil, status.Error(codes.PermissionDenied, errorx.ErrShamirInvalidSignature.Error())
	}

	return &spb.VerifySignatureResponse{
		UserId:         apiKeypair.UserID().String(),
		OrganizationId: apiKeypair.OrganizationID().String(),
	}, nil
}

func (s *Service) GetAuthInfo(ctx context.Context, request *spb.GetAuthInfoRequest) (*spb.GetAuthInfoResponse, error) {
	org, user, err := s.getAuthInfoByUsernameDomainName(ctx, request.GetUsername(), request.GetDomainName())
	if err != nil {
		return nil, err
	}
	return &spb.GetAuthInfoResponse{
		TenantId:       org.TenantID.String,
		OrganizationId: org.OrganizationID.String,
		LoginEmail:     user.LoginEmail.String,
		Email:          user.Email.String,
		UserId:         user.UserID.String,
	}, nil
}

func (s *Service) getAuthInfoByUsernameDomainName(ctx context.Context, username string, domainName string) (*entity.Organization, *entity.AuthUser, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	org, err := s.OrganizationRepo.GetByDomainName(ctx, s.DB, domainName)
	if err != nil {
		zapLogger.Errorw("get organization by domain name failed",
			"domainName", domainName,
			"error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, status.Error(codes.NotFound, errorx.ErrOrganizationNotFound.Error())
		}
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	schoolID, err := strconv.ParseInt(org.OrganizationID.String, 10, 64)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	organization := interceptors.NewOrganization(org.OrganizationID.String, int32(schoolID))
	enableAuthDB := s.FeatureManager.IsEnableDecouplingUserAndAuthDB(organization)
	enableUsername := s.FeatureManager.IsEnableUsernameStudentParentStaff(ctx, organization)

	db := s.DB
	if enableAuthDB && s.AuthDB != nil {
		db = s.AuthDB
	}

	var user *entity.AuthUser
	if enableUsername {
		user, err = s.UserRepoV2.GetByUsername(ctx, db, username, org.OrganizationID.String)
	} else {
		user, err = s.UserRepoV2.GetByEmail(ctx, db, username, org.OrganizationID.String)
	}

	if err != nil {
		zapLogger.Errorw("get user by username failed",
			"username", username,
			"org_id", org.OrganizationID.String,
			"err", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, status.Error(codes.NotFound, errorx.ErrUsernameNotFound.Error())
		}
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	// if username feature was disabled, we should return empty string for login email
	if !enableUsername {
		user.LoginEmail.String = ""
	}

	return org, user, nil
}

func fakeVerifiedCustomClaim(userID string, tenantID string, projectID string) *interceptors.CustomClaims {
	now := time.Now()
	exp := now.Add(2 * time.Hour)
	jwt := jwt.Claims{
		Issuer:   fmt.Sprintf("https://securetoken.google.com/%s", projectID),
		Subject:  userID,
		Expiry:   jwt.NewNumericDate(exp),
		IssuedAt: jwt.NewNumericDate(now),
		ID:       idutil.ULIDNow(),
	}
	return &interceptors.CustomClaims{
		Claims: jwt,
		FirebaseClaims: &interceptors.FirebaseClaims{
			Identity: interceptors.FirebaseIdentity{
				Tenant: tenantID,
			},
		},
	}
}
