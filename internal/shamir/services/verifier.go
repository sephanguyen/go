package services

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/square/go-jose/v3/jwt"
	"go.uber.org/multierr"
)

const (
	featureToggleUserAuthExchangeAPlus = "User_Auth_ExchangeAPlus"
	featureToggleUserAuthManabieRole   = "User_Auth_ManabieRole"

	ManabieRole = "MANABIE"
)

// tokenVerifier verifies token using JWKS
type tokenVerifier interface {
	Verify(ctx context.Context, idToken string) (*interceptors.CustomClaims, error)
}

// TokenVerifier serving jwks endpoint
type TokenVerifier struct {
	unleashClient unleashclient.ClientInstance
	environment   string
	set           jwk.Set
	privateKeys   map[string]*rsa.PrivateKey
	verifies      []tokenVerifier
	vendor        string
}

var (
	ErrWrongDivision = fmt.Errorf("verifier: unexpected JPREP student_division")
)

// NewTokenVerifier creates new verifier
func NewTokenVerifier(
	unleashClient unleashclient.ClientInstance,
	environment string,
	vendor string,
	privateKeys map[string]*rsa.PrivateKey,
	primaryKeyID string,
	issuers []configs.TokenIssuerConfig) (*TokenVerifier, error) {

	keys, err := generateJWKS(privateKeys, primaryKeyID)
	if err != nil {
		return nil, fmt.Errorf("generateJWKS: %w", err)
	}

	jwkSet := jwk.NewSet()
	for i := range keys {
		jwkSet.Add(keys[i])
	}

	verifiers := make([]tokenVerifier, 0, len(issuers))
	for _, i := range issuers {
		v, err := interceptors.NewTokenVerifier(i.Issuer, i.Audience, i.JWKSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("creating verifier error: %w", err)
		}

		verifiers = append(verifiers, v)
	}

	return &TokenVerifier{
		unleashClient: unleashClient,
		environment:   environment,
		set:           jwkSet,
		verifies:      verifiers,
		vendor:        vendor,
		privateKeys:   privateKeys,
	}, nil
}

// GetJWKSet returns JWKS json
func (c *TokenVerifier) GetJWKSet() jwk.Set {
	return c.set
}

// ExchangeVerifiedToken signs a new token and return if provided is valid
func (c *TokenVerifier) ExchangeVerifiedToken(claims *interceptors.CustomClaims, newTokenInfo *interceptors.TokenInfo) (string, error) {
	/*claims, err := c.Verify(ctx, originalToken)
	if err != nil {
		return "", err
	}*/

	if newTokenInfo.DefaultRole == entities.UserGroupStudent {
		if err := c.vendorSpecificVerification(claims); err != nil {
			return "", err
		}
	}

	token := generateNewToken(newTokenInfo, claims)
	hasuraClaimsWithManabieRole, err := c.unleashClient.IsFeatureEnabled(featureToggleUserAuthManabieRole, c.environment)
	if err != nil {
		hasuraClaimsWithManabieRole = false
	}

	if hasuraClaimsWithManabieRole {
		token.Hasura = generateHasuraClaimsWithManabieRole(token)
	}

	return c.signNewToken(token)
}

func (c *TokenVerifier) vendorSpecificVerification(claims *interceptors.CustomClaims) error {
	exchangeAPlusToken, err := c.unleashClient.IsFeatureEnabled(featureToggleUserAuthExchangeAPlus, c.environment)
	if err != nil {
		exchangeAPlusToken = false
	}

	switch c.vendor {
	case "jprep":
		studentDivision := strings.ToLower(claims.JPREPClaims.StudentDivision)
		if studentDivision == "kids" {
			return nil
		}
		if exchangeAPlusToken && studentDivision == "a_plus" {
			return nil
		}
		return ErrWrongDivision
	}

	return nil
}

// Verify returns valid claims if any available key match. If not, final error return is combined of all failed attempts
func (c *TokenVerifier) Verify(ctx context.Context, originalToken string) (*interceptors.CustomClaims, error) {
	possibleErrs := make([]error, 0, len(c.verifies))
	for _, v := range c.verifies {
		claims, err := v.Verify(ctx, originalToken)
		if err == nil {
			return claims, nil
		}

		possibleErrs = append(possibleErrs, err)
	}

	return nil, multierr.Combine(possibleErrs...)
}

type customClaims struct {
	jwt.Claims
	Hasura       *interceptors.HasuraClaims  `json:"https://hasura.io/jwt/claims,omitempty"`
	Manabie      *interceptors.ManabieClaims `json:"manabie,omitempty"`
	ResourcePath string                      `json:"resource_path"`
	UserGroup    string                      `json:"user_group"`
}

func generateNewToken(newTokenInfo *interceptors.TokenInfo, originalClaims *interceptors.CustomClaims) *customClaims {
	now := timeutil.Now()
	return &customClaims{
		Claims: jwt.Claims{
			ID:        idutil.ULID(now),
			Issuer:    "manabie",
			Subject:   originalClaims.Subject,
			Audience:  jwt.Audience([]string{newTokenInfo.Applicant}),
			Expiry:    jwt.NewNumericDate(originalClaims.Expiry.Time().Add(5 * time.Second)),
			NotBefore: jwt.NewNumericDate(originalClaims.NotBefore.Time()),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Manabie: &interceptors.ManabieClaims{
			UserID:       newTokenInfo.UserID,
			SchoolIDs:    golibs.ToArrayStringFromArrayInt64(newTokenInfo.SchoolIds),
			DefaultRole:  newTokenInfo.DefaultRole,
			AllowedRoles: newTokenInfo.AllowedRoles,
			UserGroup:    newTokenInfo.UserGroup,
			ResourcePath: newTokenInfo.ResourcePath,
		},
		Hasura: &interceptors.HasuraClaims{
			UserID:       newTokenInfo.UserID,
			SchoolIDs:    golibs.ToArrayStringPostgres(newTokenInfo.SchoolIds),
			DefaultRole:  newTokenInfo.DefaultRole,
			AllowedRoles: newTokenInfo.AllowedRoles,
			UserGroup:    newTokenInfo.UserGroup,
			ResourcePath: newTokenInfo.ResourcePath,
		},
		ResourcePath: newTokenInfo.ResourcePath,
		UserGroup:    strings.Join(newTokenInfo.AllowedRoles, ","),
	}
}

func generateHasuraClaimsWithManabieRole(tokenInfo *customClaims) *interceptors.HasuraClaims {
	return &interceptors.HasuraClaims{
		AllowedRoles: []string{ManabieRole},
		DefaultRole:  ManabieRole,
		UserGroup:    ManabieRole,
		UserID:       tokenInfo.Hasura.UserID,
		SchoolIDs:    tokenInfo.Hasura.SchoolIDs,
		ResourcePath: tokenInfo.Hasura.ResourcePath,
	}
}

func (c *TokenVerifier) signNewToken(claims *customClaims) (string, error) {
	key, valid := c.set.Get(0)
	if !valid {
		return "", fmt.Errorf("signNewToken: error when get key")
	}
	header := jws.NewHeaders()
	err := header.Set(jwk.KeyIDKey, key.KeyID())
	if err != nil {
		return "", fmt.Errorf("failed to set header: %w", err)
	}
	b, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	k, err := jws.Sign(b, jwa.RS256, c.privateKeys[key.KeyID()], jws.WithHeaders(header))
	if err != nil {
		return "", err
	}

	return string(k), nil
}

func generateJWK(id string, privateKey *rsa.PrivateKey) (jwk.Key, error) {
	key, err := jwk.New(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	err = multierr.Combine(
		key.Set(jwk.KeyIDKey, id),
		key.Set(jwk.KeyUsageKey, "sig"),
		key.Set(jwk.AlgorithmKey, jwa.RS256),
	)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func generateJWKS(privateKeys map[string]*rsa.PrivateKey, primaryKeyID string) ([]jwk.Key, error) {
	keys := make([]jwk.Key, 0, len(privateKeys))
	primaryKey, ok := privateKeys[primaryKeyID]
	// Add primary key to top
	// First key will be use to sign token
	// if not found, use first key
	if ok {
		key, err := generateJWK(primaryKeyID, primaryKey)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}
	for id, privateKey := range privateKeys {
		if id == primaryKeyID {
			//ignore primary key as it is already added to on top of keys
			continue
		}
		jwk, err := generateJWK(id, privateKey)
		if err != nil {
			return nil, fmt.Errorf("generateJWK: %w", err)
		}

		keys = append(keys, jwk)
	}

	return keys, nil
}
