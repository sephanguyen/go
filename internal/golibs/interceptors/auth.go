package interceptors

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth intercepts request and verifies
type Auth struct {
	verifiers       []*TokenVerifier
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

	verifiers := make([]*TokenVerifier, len(issuers))
	for i, issuer := range issuers {
		v, err := NewTokenVerifier(issuer.Issuer, issuer.Audience, issuer.JWKSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("err init tokenVerifier: %w", err)
		}
		verifiers[i] = v
	}

	a := &Auth{
		verifiers:       verifiers,
		groupDecider:    groupDecider,
		skipAuthMethods: make(map[string]bool),
	}

	for _, method := range skipAuthMethods {
		a.skipAuthMethods[method] = true
	}

	return a, nil
}

var (
	sttNoDeciderProvided = status.Error(codes.PermissionDenied, "auth: no AccessControlDecider provided")
	sttDeniedAll         = status.Error(codes.PermissionDenied, "auth: denied all access")
	sttNotAllowed        = status.Error(codes.PermissionDenied, "auth: not allowed")
)

// UnaryServerInterceptor check jwt token in metadata
func (a *Auth) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := StartSpan(ctx, "Auth.UnaryServerInterceptor")
	defer span.End()

	if a.skipAuthMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	claims, err := a.verify(ctx)
	if err != nil {
		return nil, err
	}
	ctx = ContextWithUserID(ctx, claims.Subject)
	ctx = ContextWithJWTClaims(ctx, claims)

	ctxzap.AddFields(ctx, zap.String("userID", claims.Subject))
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		rid := md.Get("x-request-id")
		if len(rid) > 0 && len(rid[0]) > 0 {
			ctxzap.AddFields(ctx, zap.String("x-request-id", rid[0]))
		}
	}

	group, err := a.groupDecider.Check(ctx, claims.Subject, info.FullMethod)
	if err != nil {
		return nil, err
	}

	ctx = ContextWithUserGroup(ctx, group)

	return handler(ctx, req)
}

func (a *Auth) verify(ctx context.Context) (*CustomClaims, error) {
	ctx, span := StartSpan(ctx, "Auth.verify")
	defer span.End()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot check gRPC metadata")
	}

	s := md.Get("token")
	if len(s) == 0 || len(s[0]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	var verifyErrs error
	for _, verifier := range a.verifiers {
		c, err := verifier.Verify(ctx, s[0])
		if err == nil {
			return c, nil
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

	s := grpc_middleware.WrapServerStream(ss)
	s.WrappedContext = ContextWithUserID(ss.Context(), claims.Subject)
	s.WrappedContext = ContextWithJWTClaims(s.WrappedContext, claims)

	group, err := a.groupDecider.Check(ss.Context(), claims.Subject, info.FullMethod)
	if err != nil {
		return err
	}

	s.WrappedContext = ContextWithUserGroup(s.WrappedContext, group)

	ctxzap.AddFields(s.WrappedContext, zap.String("userID", claims.Subject))
	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok {
		rid := md.Get("x-request-id")
		if len(rid) > 0 && len(rid[0]) > 0 {
			ctxzap.AddFields(s.WrappedContext, zap.String("x-request-id", rid[0]))
		}
	}

	return handler(srv, s)
}

// TokenVerifier keeps data for auth interceptor
type TokenVerifier struct {
	issuer       string
	aud          string
	jwkURL       string
	httpClient   *http.Client
	keySet       *jose.JSONWebKeySet
	requestGroup singleflight.Group
}

// NewTokenVerifier creates new TokenVerifier
func NewTokenVerifier(issuer, aud, jwkURL string) (*TokenVerifier, error) {
	t := &TokenVerifier{
		issuer:       issuer,
		aud:          aud,
		jwkURL:       jwkURL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		requestGroup: singleflight.Group{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.fetchAllKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("err t.fetchAllKey: %w", err)
	}

	return t, nil
}

// fetch all jwk keyset on single flight
func (t *TokenVerifier) fetchAllKeySF(ctx context.Context) error {
	ctx, span := StartSpan(ctx, "TokenVerifier.fetchAllKeySF")
	defer span.End()

	_, err, _ := t.requestGroup.Do("get-jwk", func() (interface{}, error) {
		err := t.fetchAllKey(ctx)
		return nil, err
	})
	if err != nil {
		return fmt.Errorf("err getJwt: %w", err)
	}

	return nil
}

func (t *TokenVerifier) fetchAllKey(ctx context.Context) error {
	ctx, span := StartSpan(ctx, "TokenVerifier.fetchAllKey")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.jwkURL, nil)
	if err != nil {
		return fmt.Errorf("err make req: %w", err)
	}

	r, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("err get url: %w", err)
	}
	defer r.Body.Close()

	var result jose.JSONWebKeySet
	err = json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("err parse to JSONWebKeySet: %w", err)
	}

	t.keySet = &result

	return nil
}

// Verify parses and checks idToken validate
func (t *TokenVerifier) Verify(ctx context.Context, idToken string) (*CustomClaims, error) {
	ctx, span := StartSpan(ctx, "TokenVerifier.Verify")
	defer span.End()

	token, err := jwt.ParseSigned(idToken)
	if err != nil {
		return nil, fmt.Errorf("err parse idToken: %w", err)
	}

	claims := &CustomClaims{}
	err = token.Claims(t.keySet, claims)
	if err != nil {
		err = t.fetchAllKeySF(ctx)
		if err != nil {
			return nil, fmt.Errorf("error when re-fetch token: %w, issuer: %s", err, t.issuer)
		}

		// retry
		err = token.Claims(t.keySet, claims)
		if err != nil {
			return nil, fmt.Errorf("err parse claims: %w, issuer: %s", err, t.issuer)
		}
	}

	err = claims.Validate(jwt.Expected{
		Issuer:   t.issuer,
		Audience: []string{t.aud},
		Time:     time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("claims.Validate: %w, issuer: %s", err, t.issuer)
	}

	claims.JwkURL = t.jwkURL

	return claims, nil
}

// GroupDecider checks user's group
type GroupDecider struct {
	GroupFetcher  func(context.Context, string) (string, error)
	AllowedGroups map[string][]string
}

// Check checks if user allowed to call a method
func (g *GroupDecider) Check(ctx context.Context, userID, fullMethod string) (group string, err error) {
	ctx, span := StartSpan(ctx, "GroupDecider.Check")
	defer span.End()

	allowedGroups, ok := g.AllowedGroups[fullMethod]
	if !ok {
		return "", sttNoDeciderProvided
	}

	group, _ = g.GroupFetcher(ctx, userID)

	if gr, ok := constant.MapRoleWithLegacyUserGroup[group]; ok {
		group = gr
	}

	if allowedGroups == nil {
		// allowed all
		return group, nil
	}

	if len(allowedGroups) == 0 {
		return "", sttDeniedAll
	}

	for _, g := range allowedGroups {
		if group == g {
			return group, nil
		}
	}
	return "", sttNotAllowed
}

type (
	UserIDKey    int
	UserGroupKey int
	UserRolesKey int
	JwtClaims    int
)

// UserIDFromContext returns user ID from context, empty when not found
func UserIDFromContext(ctx context.Context) string {
	v := ctx.Value(UserIDKey(0))
	s, _ := v.(string)
	return s
}

// ContextWithUserID puts userID to ctx
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey(0), userID)
}

// UserGroupFromContext returns user group from ctx
func UserGroupFromContext(ctx context.Context) string {
	v := ctx.Value(UserGroupKey(0))
	s, _ := v.(string)
	return s
}

// ContextWithUserGroup puts group to ctx
func ContextWithUserGroup(ctx context.Context, group string) context.Context {
	return context.WithValue(ctx, UserGroupKey(0), group)
}

// ContextWithUserRoles puts roles to ctx
func ContextWithUserRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, UserRolesKey(0), roles)
}

// UserRolesFromContext returns user roles from ctx
func UserRolesFromContext(ctx context.Context) []string {
	v := ctx.Value(UserRolesKey(0))
	s, _ := v.([]string)
	return s
}

// JWTClaimsFromContext returns JWT claims from context, nil if not found
func JWTClaimsFromContext(ctx context.Context) *CustomClaims {
	v := ctx.Value(JwtClaims(0))
	s, _ := v.(*CustomClaims)
	return s
}

func GetUserInfoFromContext(ctx context.Context) (userGroup, userID string, schoolIDs []string) {
	userGroup = UserGroupFromContext(ctx)
	userID = UserIDFromContext(ctx)
	claims := JWTClaimsFromContext(ctx)
	if claims != nil && claims.Manabie != nil {
		schoolIDs = claims.Manabie.SchoolIDs
	}

	return constant.MapRoleWithLegacyUserGroup[userGroup], userID, schoolIDs
}

var (
	ErrFailedToParseJWT     = errors.New("failed to parse jwt")
	ErrManabieClaimsIsEmpty = errors.New("manabie claims is nil")
	ErrResourcePathIsEmpty  = errors.New("resource path is empty")
)

// JWTClaimsFromContextV2 returns JWT claims from context, nil and error if not found
func JWTClaimsFromContextV2(ctx context.Context) (*CustomClaims, error) {
	jwtClaims := ctx.Value(JwtClaims(0))
	customClaim, ok := jwtClaims.(*CustomClaims)
	if !ok {
		return nil, ErrFailedToParseJWT
	}
	return customClaim, nil
}

func OrganizationFromContext(ctx context.Context) (*Organization, error) {
	customClaims, err := JWTClaimsFromContextV2(ctx)
	if err != nil {
		return nil, err
	}

	if customClaims.Manabie == nil {
		return nil, ErrManabieClaimsIsEmpty
	}

	resourcePath := customClaims.Manabie.ResourcePath
	if resourcePath == "" {
		return nil, ErrResourcePathIsEmpty
	}

	schoolID, err := strconv.ParseInt(resourcePath, 10, 32)
	if err != nil {
		return nil, err
	}

	org := &Organization{
		organizationID: resourcePath,
		schoolID:       int32(schoolID),
	}
	return org, nil
}

// ResourcePathFromContext
func ResourcePathFromContext(ctx context.Context) (string, error) {
	claim := JWTClaimsFromContext(ctx)
	if claim == nil || claim.Manabie == nil {
		return "", fmt.Errorf("ctx has no resource path")
	}
	return claim.Manabie.ResourcePath, nil
}

// ContextWithJWTClaims puts claims to context
func ContextWithJWTClaims(ctx context.Context, claims *CustomClaims) context.Context {
	return context.WithValue(ctx, JwtClaims(0), claims)
}

func PrivateKeyFromString(data string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, fmt.Errorf("pem.Decode return nil block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to call x509.ParsePKCS1PrivateKey: %v", err)
	}
	return privateKey, nil
}
