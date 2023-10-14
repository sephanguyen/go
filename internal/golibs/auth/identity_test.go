package auth

import (
	"context"
	"fmt"
	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_auth "github.com/manabie-com/backend/mock/golibs/auth"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/Nerzal/gocloak/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func Test_KeyCloakIdentity(t *testing.T) {
	keycloakMock := &mock_auth.KeyCloakClient{}
	realm := idutil.ULIDNow()
	clientID := idutil.ULIDNow()
	identitySvc := &IdentityServiceImpl{
		realm:    realm,
		clientID: clientID,
		keycloak: keycloakMock,
	}
	mockusersvc := &mock_auth.BobUserModifierServiceClient{}
	email, password := "", ""
	keycloakToken := idutil.ULIDNow()
	manabietoken := idutil.ULIDNow()
	keycloakMock.On("GetToken", mock.Anything, realm, gocloak.TokenOptions{
		ClientID:  &identitySvc.clientID,
		GrantType: gocloak.StringP("password"),
		Scope:     gocloak.StringP("openid"),
		Username:  &email,
		Password:  &password,
	}).Once().Return(&gocloak.JWT{
		IDToken: keycloakToken,
	}, nil)
	mockusersvc.On("ExchangeToken", mock.Anything, &bpb.ExchangeTokenRequest{Token: keycloakToken}).Once().Return(&bpb.ExchangeTokenResponse{
		Token: manabietoken,
	}, nil)

	ctx, err := ContextWithTokenFromEmailPassword(context.Background(), identitySvc, mockusersvc, email, password)
	assert.NoError(t, err)
	md, exist := metadata.FromOutgoingContext(ctx)
	assert.True(t, exist)
	assert.NoError(t, err)
	assert.Equal(t, manabietoken, md["token"][0])
}

func Test_ContextWithTokenFromEmailPassword(t *testing.T) {
	identityMock := &mock_auth.IdentityService{}
	mockusersvc := &mock_auth.BobUserModifierServiceClient{}
	email, password := "", ""
	firebasetoken := idutil.ULIDNow()
	manabietoken := idutil.ULIDNow()

	identityMock.On("VerifyEmailPassword", mock.Anything, email, password).Once().Return(firebasetoken, nil)
	mockusersvc.On("ExchangeToken", mock.Anything, &bpb.ExchangeTokenRequest{Token: firebasetoken}).Once().Return(&bpb.ExchangeTokenResponse{
		Token: manabietoken,
	}, nil)

	ctx, err := ContextWithTokenFromEmailPassword(context.Background(), identityMock, mockusersvc, email, password)
	assert.NoError(t, err)
	md, exist := metadata.FromOutgoingContext(ctx)
	assert.True(t, exist)
	assert.NoError(t, err)
	assert.Equal(t, manabietoken, md["token"][0])
}

func Test_InjectFakeJwtToken(t *testing.T) {
	ctx := InjectFakeJwtToken(context.Background(), "manabie")
	claim := interceptors.JWTClaimsFromContext(ctx)
	assert.Equal(t, claim.Manabie.UserGroup, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())
	assert.Equal(t, claim.Manabie.ResourcePath, "manabie")
}

func TestHashedPassword(t *testing.T) {
	key, err := crypt.DecodeBase64("mAaX5DSYQLUj3XD60McZ3n6m/AdZxEpfiLYqIFtYf2jlNIVaJ6Esu1sWe5HrsyLO1sTD/pygrtoFsQaFhfuRDg==")
	assert.NoError(t, err)

	salt, err := crypt.DecodeBase64("Bw==")
	assert.NoError(t, err)

	hashConfig := &gcp.HashConfig{
		HashAlgorithm: "SCRYPT",
		HashSignerKey: gcp.Base64EncodedStr{
			Value:        "",
			DecodedBytes: key,
		},
		HashSaltSeparator: gcp.Base64EncodedStr{
			Value:        "",
			DecodedBytes: salt,
		},
		HashRounds:     8,
		HashMemoryCost: 14,
	}

	now := time.Now()
	_, err = HashedPassword(hashConfig, []byte("password"), []byte("salt"))
	assert.NoError(t, err)

	fmt.Println(time.Now().Sub(now))
}
