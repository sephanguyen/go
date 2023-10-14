package interceptors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	glInterceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth intercepts request and verifies
type Auth struct {
	verifiers       []*glInterceptors.TokenVerifier
	skipAuthMethods map[string]bool
	groupDecider    *GroupDecider
}

// NewAuth returns error if no issuer provided
func NewAuth(
	skipAuthMethods []string,
	groupDecider *GroupDecider,
	issuers []configs.TokenIssuerConfig,
) (*Auth, error) {
	if len(issuers) == 0 {
		return nil, fmt.Errorf("no issuer provided")
	}

	verifiers := make([]*glInterceptors.TokenVerifier, len(issuers))
	for i, issuer := range issuers {
		verifier, err := glInterceptors.NewTokenVerifier(issuer.Issuer, issuer.Audience, issuer.JWKSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("err init tokenVerifier: %w", err)
		}
		verifiers[i] = verifier
	}

	auth := &Auth{
		verifiers:       verifiers,
		groupDecider:    groupDecider,
		skipAuthMethods: make(map[string]bool),
	}

	for _, method := range skipAuthMethods {
		auth.skipAuthMethods[method] = true
	}

	return auth, nil
}

var (
	sttNoDeciderProvided = status.Error(codes.PermissionDenied, "auth: no AccessControlDecider provided")
	sttDeniedAll         = status.Error(codes.PermissionDenied, "auth: denied all access")
	sttNotAllowed        = status.Error(codes.PermissionDenied, "auth: not allowed")
	sttDeactivatedUser   = status.Error(codes.PermissionDenied, errorx.ErrDeactivatedUser.Error())
)

// UnaryServerInterceptor check jwt token in metadata
func (a *Auth) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := glInterceptors.StartSpan(ctx, "Auth.UnaryServerInterceptor")
	defer span.End()

	if a.skipAuthMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	claims, err := a.verify(ctx)
	if err != nil {
		return nil, err
	}
	ctx = glInterceptors.ContextWithUserID(ctx, claims.Subject)
	ctx = glInterceptors.ContextWithJWTClaims(ctx, claims)

	ctxzap.AddFields(ctx, zap.String("userID", claims.Subject))
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		rid := md.Get("x-request-id")
		if len(rid) > 0 && len(rid[0]) > 0 {
			ctxzap.AddFields(ctx, zap.String("x-request-id", rid[0]))
		}
	}

	groups, err := a.groupDecider.Check(ctx, claims.Subject, info.FullMethod)
	if err != nil {
		return nil, err
	}

	if len(groups) > 0 {
		ctx = glInterceptors.ContextWithUserGroup(ctx, groups[0])
		ctx = glInterceptors.ContextWithUserRoles(ctx, groups)
	}

	return handler(ctx, req)
}

func (a *Auth) verify(ctx context.Context) (*glInterceptors.CustomClaims, error) {
	ctx, span := glInterceptors.StartSpan(ctx, "Auth.verify")
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot check gRPC metadata")
	}

	token := md.Get("token")
	if len(token) == 0 || len(token[0]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	var verifyErrs error
	for _, verifier := range a.verifiers {
		claim, err := verifier.Verify(ctx, token[0])
		if err == nil {
			return claim, nil
		}

		verifyErrs = multierr.Append(verifyErrs, err)
	}

	ctxzap.Extract(ctx).Error("Auth verifyErrs", zap.Error(verifyErrs))

	return nil, status.Error(codes.Unauthenticated, codes.Unauthenticated.String())
}

// StreamServerInterceptor check for jwt token
func (a *Auth) StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if a.skipAuthMethods[info.FullMethod] {
		return handler(srv, ss)
	}

	claims, err := a.verify(ss.Context())
	if err != nil {
		return err
	}

	wrapStream := grpc_middleware.WrapServerStream(ss)
	wrapStream.WrappedContext = glInterceptors.ContextWithUserID(ss.Context(), claims.Subject)
	wrapStream.WrappedContext = glInterceptors.ContextWithJWTClaims(wrapStream.WrappedContext, claims)

	groups, err := a.groupDecider.Check(ss.Context(), claims.Subject, info.FullMethod)
	if err != nil {
		return err
	}
	if len(groups) > 0 {
		wrapStream.WrappedContext = glInterceptors.ContextWithUserGroup(wrapStream.WrappedContext, groups[0])
	}

	ctxzap.AddFields(wrapStream.WrappedContext, zap.String("userID", claims.Subject))
	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok {
		rid := md.Get("x-request-id")
		if len(rid) > 0 && len(rid[0]) > 0 {
			ctxzap.AddFields(wrapStream.WrappedContext, zap.String("x-request-id", rid[0]))
		}
	}

	return handler(srv, wrapStream)
}

// GroupDecider checks user's groups
type GroupDecider struct {
	GroupFetcher  func(ctx context.Context, userID string) ([]string, error)
	AllowedGroups map[string][]string
}

// Check checks if user allowed to call a method
func (g *GroupDecider) Check(ctx context.Context, userID, fullMethod string) (groups []string, err error) {
	ctx, span := glInterceptors.StartSpan(ctx, "GroupDecider.Check")
	defer span.End()

	allowedGroups, ok := g.AllowedGroups[fullMethod]
	if !ok {
		return nil, sttNoDeciderProvided
	}
	groups, err = g.GroupFetcher(ctx, userID)
	if err == sttDeactivatedUser {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, sttDeniedAll
	}

	if allowedGroups == nil {
		// allowed all
		return groups, nil
	}

	if len(allowedGroups) == 0 {
		return nil, sttDeniedAll
	}

	for _, group := range groups {
		if golibs.InArrayString(group, allowedGroups) {
			return groups, nil
		}
	}
	return nil, sttNotAllowed
}

type UserRepoInterface interface {
	GetUserRoles(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (entity.Roles, error)
	GetUserGroupMembers(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupMember, error)
	Get(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.LegacyUser, error)
}

func RetrieveUserRoles(ctx context.Context, userRepo UserRepoInterface, db database.QueryExecer) ([]string, error) {
	userID := glInterceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	userGroupMembers, err := userRepo.GetUserGroupMembers(ctx, db, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("get user group member: %w", err).Error())
	}

	if len(userGroupMembers) == 0 {
		return nil, status.Error(codes.Internal, "user haven't been assigned user_group")
	}

	listUserRoles, err := getUserRoles(ctx, userRepo, db, userID)
	if err != nil {
		return nil, err
	}

	return listUserRoles, nil
}

func RetrieveUserRolesV2(ctx context.Context, userRepo UserRepoInterface, db database.QueryExecer) ([]string, error) {
	userID := glInterceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	user, err := userRepo.Get(ctx, db, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("get user: %w", err).Error())
	}

	if user.DeactivatedAt.Status == pgtype.Present {
		return nil, sttDeactivatedUser
	}

	userGroupMembers, err := userRepo.GetUserGroupMembers(ctx, db, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("get user group member: %w", err).Error())
	}

	if len(userGroupMembers) == 0 {
		return nil, status.Error(codes.Internal, "user haven't been assigned user_group")
	}

	listUserRoles, err := getUserRoles(ctx, userRepo, db, userID)
	if err != nil {
		return nil, err
	}

	return listUserRoles, nil
}

func getUserRoles(ctx context.Context, userRepo UserRepoInterface, db database.QueryExecer, userID string) ([]string, error) {
	userRoles, err := userRepo.GetUserRoles(ctx, db, database.Text(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("get user's roles: %w", err).Error())
	}

	if len(userRoles) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing grant role for user")
	}

	listUserRoles := []string{}
	for _, ur := range userRoles {
		role := ur.RoleName.String
		if golibs.InArrayString(role, constant.AllowListRoles) {
			listUserRoles = append(listUserRoles, role)
		}
	}

	return listUserRoles, nil
}

func RetrieveLegacyUserGroups(ctx context.Context, userRepo UserRepoInterface, db database.QueryExecer) ([]string, error) {
	listUserRoles, err := RetrieveUserRoles(ctx, userRepo, db)
	if err != nil {
		return nil, err
	}

	return constant.MapRoleToLegacyUserGroup[listUserRoles[0]], nil
}

// LoginInAuthPlatform login in auth platform, returns id token
func LoginInAuthPlatform(ctx context.Context, apiKey string, tenantID string, email string, password string) (string, error) {
	url := fmt.Sprintf("%s%s", constant.IdentityToolkitURL, apiKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId,omitempty"`
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		TenantID:          tenantID,
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to login firebase and failed to decode error")
	}

	if resp.StatusCode == http.StatusOK {
		type result struct {
			IDToken string `json:"idToken"`
		}

		r := &result{}
		if err := json.Unmarshal(data, &r); err != nil {
			return "", errors.Wrap(err, "failed to login and failed to decode error")
		}
		return r.IDToken, nil
	}

	return "", errors.New("failed to login firebase" + string(data))
}

func ExchangeToken(ctx context.Context, conn grpc.ClientConnInterface, applicant, userID, originalToken string) (string, error) {
	rsp, err := spb.NewTokenReaderServiceClient(conn).ExchangeToken(ctx, &spb.ExchangeTokenRequest{
		NewTokenInfo: &spb.ExchangeTokenRequest_TokenInfo{
			Applicant: applicant,
			UserId:    userID,
		},
		OriginalToken: originalToken,
	})
	if err != nil {
		return "", err
	}
	return rsp.NewToken, nil
}

func GRPCContext(ctx context.Context, key, val string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(key, val, "pkg", "com.manabie.liz", "version", "1.0.0"))
}
