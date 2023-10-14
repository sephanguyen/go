package auth

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/Nerzal/gocloak/v10"
	"google.golang.org/api/identitytoolkit/v3"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type IdentityService interface {
	VerifyEmailPassword(ctx context.Context, email string, password string) (string, error)
}

// KeyCloakClient is implemented by gocloak.GoCloak
type KeyCloakClient interface {
	GetToken(ctx context.Context, realm string, opts gocloak.TokenOptions) (*gocloak.JWT, error)
}

type IdentityServiceImpl struct {
	cl       *identitytoolkit.Service
	keycloak KeyCloakClient
	realm    string
	clientID string
}

// for example with this config (check helm config):
// issuers:
//   - issuer: https://ji-sso.jprep.jp/auth/realms/jprep
//     audience: manabie
//     jwks_endpoint: https://ji-sso.jprep.jp/auth/realms/jprep/protocol/openid-connect/certs
//
// Path:  https://ji-sso.jprep.jp
// Realm: jprep
// ClientID: manabie
type KeyCloakOpts = configs.KeyCloakAuthConfig

func NewKeyCloakClient(opts KeyCloakOpts) (*IdentityServiceImpl, error) {
	client := gocloak.NewClient(opts.Path)
	return &IdentityServiceImpl{keycloak: client, realm: opts.Realm, clientID: opts.ClientID}, nil
}

func NewIdentityService(ctx context.Context, opts ...option.ClientOption) (*IdentityServiceImpl, error) {
	identitytoolkitService, err := identitytoolkit.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &IdentityServiceImpl{cl: identitytoolkitService}, nil
}

func (s *IdentityServiceImpl) VerifyEmailPassword(ctx context.Context, email, password string) (string, error) {
	if s.keycloak != nil {
		token, err := s.keycloak.GetToken(ctx, s.realm, gocloak.TokenOptions{
			ClientID:  &s.clientID,
			GrantType: gocloak.StringP("password"),
			Scope:     gocloak.StringP("openid"),
			Username:  &email,
			Password:  &password,
		})
		if err != nil {
			return "", err
		}
		return token.IDToken, nil
	}
	call := s.cl.Relyingparty.VerifyPassword(&identitytoolkit.IdentitytoolkitRelyingpartyVerifyPasswordRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	})
	res, err := call.Context(ctx).Do()
	if err != nil {
		return "", err
	}
	return res.IdToken, nil
}

func ContextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0", "token", token)
}

// BobUserModifierServiceClient is implemented by bob's bob.UserModifierServiceClient.
type BobUserModifierServiceClient interface {
	UpdateUserProfile(ctx context.Context, in *bpb.UpdateUserProfileRequest, opts ...grpc.CallOption) (*bpb.UpdateUserProfileResponse, error)
	UpdateUserDeviceToken(ctx context.Context, in *bpb.UpdateUserDeviceTokenRequest, opts ...grpc.CallOption) (*bpb.UpdateUserDeviceTokenResponse, error)
	ExchangeToken(ctx context.Context, in *bpb.ExchangeTokenRequest, opts ...grpc.CallOption) (*bpb.ExchangeTokenResponse, error)
	Register(ctx context.Context, in *bpb.RegisterRequest, opts ...grpc.CallOption) (*bpb.RegisterResponse, error)
	ExchangeCustomToken(ctx context.Context, in *bpb.ExchangeCustomTokenRequest, opts ...grpc.CallOption) (*bpb.ExchangeCustomTokenResponse, error)
	UpdateUserLastLoginDate(ctx context.Context, in *bpb.UpdateUserLastLoginDateRequest, opts ...grpc.CallOption) (*bpb.UpdateUserLastLoginDateResponse, error)
}

func ContextWithTokenFromEmailPassword(ctx context.Context, identitySvc IdentityService, userSvc BobUserModifierServiceClient, email, password string) (context.Context, error) {
	token, err := identitySvc.VerifyEmailPassword(ctx, email, password)
	if err != nil {
		return ctx, err
	}
	tokenRes, err := userSvc.ExchangeToken(
		ctx, &bpb.ExchangeTokenRequest{Token: token})
	if err != nil {
		return ctx, err
	}
	return ContextWithToken(ctx, tokenRes.Token), nil
}

func InjectFakeJwtToken(ctx context.Context, resourcePath string) context.Context {
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole:  entities.UserGroupSchoolAdmin,
			UserGroup:    entities.UserGroupSchoolAdmin,
			ResourcePath: resourcePath,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claim)
}
