package interceptors

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat/go-jwx/jws"
	"github.com/segmentio/ksuid"
	"github.com/square/go-jose/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var privateKeys = []string{
	`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCsP8I2tfnNnL1K+LNZzgkLIs977aKffd7uj6L5cGG464F/8/X+
rkGepW+dUvhqfAR4k3yP543rUXzxOfl7ZEPawsrWMPmxvFqfVp26Z4rJniAeHqzi
NagZ19Db7aKYxHOdXwa32/N1VudAhAyKx3QEN9GfPjx/Ehx67E8cW+Qs0QIDAQAB
AoGALSSoqd4XkiO6GKQFnUu6YwjEiB5HuLUscCmE9QrXEbfnQLmXhx/0YrfJANp1
8LKAGXnN84kkUMASlsYy9HvarFS2Drm+NYGhGlyAJ9ba5VEkZRjEjBLX1yTqGqa2
L1Kvfgq/FrJ4K3e4OZgvr1M+a63HYvuR9SN6BW2wvifh3oECQQDMxD02cVsyxB16
jOhlcGgE+l7scvWn6OwY7NWT0lP+7ISQJyA2fkgHvKo98shhFGudOh3WEVSuX+fu
bzlInMZVAkEA11izNqxbyUffBM6UYLxcaTBil41vuPie8vpymuduA6rbXjzuyOND
QBZqoJTwBhBwelo07rZsPwY/pmHqk+8wjQJBAILe0TigjmcNzMFsmYNrqi+0TULV
3oeoaG0twPsvLBv70mXHe+EYKLU5MZ1SNBtHz9e3MyaEARlJlpRfZb8w49kCQQCd
fEjT2wjlEqKOqWTpudb7Nl9j5hOmemwD1hSqJEXYeMMlD/qw/0LXQ42HEmTWin83
e3Dqgo53KOKzkzgyJ+KhAkAaZ42lHWLITJrTEUBhhnZtYFAGuAO6G/TT8Q3xRpZJ
+QBq+Nyd8kvBPbhff7q8JA/XA+Zg1UvhcgqFWCXIcKvz
-----END RSA PRIVATE KEY-----
`,
	`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQD8jhnqBthMkuVAKe6BhBI7F+NUxfbL1vkudGeBWFCsvhRPIZDG
C7+yfTqvRfq1bqSULnwhoOv0kdExF0gMBlEtstsLFDrxPOG2Arbh0a80UgQSYae7
Y7swBPICMiOshhsv6HjQW4j78DK2KxvEUGJWdscCDmJ221Vwy1oYbkueqQIDAQAB
AoGACIM9ql66sbIN2iDPsjviZW2DsxrNG8fONFumFX0Fkx0BED3AZHyG5JxF+xxv
u+fT0k7SzktfSKoVlAMF4ang2IlXjRODLhrs7WnsuCAyS8DrcO75oKpks7h+Svon
UomoV3BLRdeEgpHQK755K9IExQ4JMn5Ss9/AoBq2u3WuMVECQQD83X1YpijJmOwU
jXpwWNc7QpMuBMf/lyh0TO2MisZrBuoq+sGQF9qEoyKdRN7k5hihpuUffc1RGlfI
5UyQVsjlAkEA/6+gnYk9BeF97IYcD9FxsTbLSvsvMBBQqy6WRoh6clumM9lc7bRW
HMk5jJMRFpX14zwY0RY9eml4WO3D0qO2dQJALp/YYOQiWSmtPgzoKpb2+KJJ6Q5Q
ojwI5YjZtDRSaaGYw9wAnMUJMbOyyjTMtIAIwqW3UZsspGurDAvbljGqUQJBAODQ
c9EQkm9RTX9ii9n8tpKzMxzCr0L7lXJujAOIjOFwZDrCrEr1faHh7JAF38iUIpei
h8+QVo4DnqXSqZPUDuECQFO90GbqCJ9C4UrH/py3/y04B0XUSWcv5ugsGuC5TLT9
dL/0SceBflGz1GKToafryX6q6pdKa97dexrsk1NTNpo=
-----END RSA PRIVATE KEY-----
`,
}

func TestAuthUnaryInterceptor(t *testing.T) {
	t.Parallel()
	type testCase struct {
		mockGroupFetcher func(ctx context.Context, userID string) (string, error)
		userID           string
		deciders         map[string][]string
		expectedErrResp  map[string]error
	}

	testCases := map[string]testCase{
		"admin user access admin allowed method": {
			expectedErrResp: map[string]error{
				"Service.Test1": nil,
			},
			userID: "admin-id",
			deciders: map[string][]string{
				"Service.Test1": {"ADMIN"},
			},
			mockGroupFetcher: func(context.Context, string) (string, error) {
				return "ADMIN", nil
			},
		},
		"admin user access allow all method": {
			expectedErrResp: map[string]error{
				"Service.Test1": nil,
			},
			userID: "admin-id",
			deciders: map[string][]string{
				"Service.Test1": nil,
			},
			mockGroupFetcher: func(context.Context, string) (string, error) {
				return "ADMIN", nil
			},
		},
		"admin user access denied method": {
			expectedErrResp: map[string]error{
				"Service.Test1": sttDeniedAll,
			},
			userID: "admin-id",
			deciders: map[string][]string{
				"Service.Test1": {},
			},
			mockGroupFetcher: func(context.Context, string) (string, error) {
				return "ADMIN", nil
			},
		},
		"admin access student only method": {
			expectedErrResp: map[string]error{
				"Service.Test1": sttNotAllowed,
			},
			userID: "admin-id",
			deciders: map[string][]string{
				"Service.Test1": {"STUDENT"},
			},
			mockGroupFetcher: func(context.Context, string) (string, error) {
				return "ADMIN", nil
			},
		},
		"admin access mixed cases": {
			expectedErrResp: map[string]error{
				"Service.AdminOnly":   nil,
				"Service.StudentOnly": sttNotAllowed,
				"Service.DeniedAll":   sttDeniedAll,
				"Service.AllowedNone": nil,
			},
			userID: "admin-id",
			deciders: map[string][]string{
				"Service.AdminOnly":   {"ADMIN"},
				"Service.StudentOnly": {"STUDENT"},
				"Service.DeniedAll":   {},
				"Service.AllowedNone": nil,
			},
			mockGroupFetcher: func(context.Context, string) (string, error) {
				return "ADMIN", nil
			},
		},
	}

	for name, testCase := range testCases {
		name := name
		testCase := testCase
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()
			fa := Auth{
				groupDecider: &GroupDecider{
					GroupFetcher:  testCase.mockGroupFetcher,
					AllowedGroups: testCase.deciders,
				},
			}

			for path, expectedErr := range testCase.expectedErrResp {

				_, err := fa.groupDecider.Check(context.Background(), testCase.userID, path)
				assert.Equal(tt, expectedErr, err, "case %s: unexpected error returned for path: %s", name, path)
			}
		})
	}
}

func TestNewAuth(t *testing.T) {
	t.Parallel()
	t.Run("missing issuer", func(t *testing.T) {
		t.Parallel()
		a, err := NewAuth(nil, nil, nil)
		assert.Nil(t, a)
		assert.EqualError(t, err, "no issuer provided")
	})

	t.Run("err fetch jwk", func(t *testing.T) {
		t.Parallel()
		randomserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
		}))
		defer randomserver.Close()
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: randomserver.URL + "/jwkset",
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
		}
		a, err := NewAuth(nil, nil, iss)
		assert.Nil(t, a)
		assert.EqualError(t, err, "err init tokenVerifier: err t.fetchAllKey: err parse to JSONWebKeySet: EOF")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tokenGenerator := newTokenGenerator()
		server := httptest.NewServer(http.HandlerFunc(tokenGenerator.ServeJWKSet()))
		defer server.Close()

		skipMethod := []string{"/IgnoreEndpoint"}
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
		}
		a, err := NewAuth(skipMethod, nil, iss)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(a.verifiers))
		assert.Equal(t, map[string]bool{"/IgnoreEndpoint": true}, a.skipAuthMethods)

		groupDecider := &GroupDecider{
			GroupFetcher: func(context.Context, string) (string, error) { panic("not implemented") },
			AllowedGroups: map[string][]string{
				"login":       nil,
				"get-profile": {"admin"},
			},
		}

		iss = []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-2",
				Audience:     "aud-2",
			},
		}
		a, err = NewAuth(skipMethod, groupDecider, iss)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(a.verifiers))
		assert.Equal(t, map[string]bool{"/IgnoreEndpoint": true}, a.skipAuthMethods)
		assert.Equal(t, groupDecider, a.groupDecider)
	})
}

func TestAuthVerify(t *testing.T) {
	t.Parallel()
	tokenGenerator := newTokenGenerator()
	server := httptest.NewServer(http.HandlerFunc(tokenGenerator.ServeJWKSet()))
	defer server.Close()

	t.Run("empty metadata", func(t *testing.T) {
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
		}
		a, err := NewAuth(nil, nil, iss)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(a.verifiers))

		userID, err := a.verify(context.Background())
		assert.Empty(t, userID)
		assert.Equal(t, status.Error(codes.Internal, "cannot check gRPC metadata"), err)
	})

	t.Run("missing token", func(t *testing.T) {
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
		}
		a, err := NewAuth(nil, nil, iss)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(a.verifiers))

		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})

		userID, err := a.verify(ctx)
		assert.Empty(t, userID)
		assert.Equal(t, status.Error(codes.Unauthenticated, "missing token"), err)
	})

	t.Run("invalid token", func(t *testing.T) {
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "issuer-1",
				Audience:     "aud-1",
			},
		}
		a, err := NewAuth(nil, nil, iss)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(a.verifiers))

		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"token": {"invalid token"}})

		userID, err := a.verify(ctx)
		assert.Empty(t, userID)
		assert.Equal(t, status.Error(codes.Unauthenticated, "Unauthenticated"), err, "must return only generic message when unauthenthicated")
	})

	t.Run("valid token", func(t *testing.T) {
		iss := []configs.TokenIssuerConfig{
			{
				JWKSEndpoint: server.URL,
				Issuer:       "http://firebase:8080/fake_aud",
				Audience:     "fake_aud",
			},
		}
		a, err := NewAuth(nil, nil, iss)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(a.verifiers))

		token := tokenGenerator.FakeToken(&TokenValue{UserID: ksuid.New().String()})
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"token": {token}})

		userID, err := a.verify(ctx)
		assert.NotEmpty(t, userID)
		assert.Nil(t, err)
	})
}

func TestTokenVerifierVerify(t *testing.T) {
	t.Parallel()
	tokenGenerator := newTokenGenerator()
	server := httptest.NewServer(http.HandlerFunc(tokenGenerator.ServeJWKSet()))
	defer server.Close()

	t.Run("err refetch new key after verify failed", func(t *testing.T) {
		randomserver := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
		}))
		defer randomserver.Close()

		a := &TokenVerifier{
			jwkURL:     randomserver.URL + "/jwkset",
			issuer:     "http://fake_firebase:8080/fake_aud",
			aud:        "fake_aud",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImNhM2YxOGFhOGI2ZWZlZjY4ZTNmMWUxMDQyZWM3YjNiN2FkYWMzYzIifQ.ewogICAgImlzcyI6ICJodHRwOi8vZmFrZV9maXJlYmFzZTo0MDQwMS9mYWtlX2F1ZCIsCiAgICAiYXVkIjogImZha2VfYXVkIiwKICAgICJhdXRoX3RpbWUiOiAxNTk2NTMzMjI1LAogICAgInVzZXJfaWQiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJzdWIiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJpYXQiOiAxNTk2NTMzMjg1LAogICAgImV4cCI6IDE1OTY1MzY4ODUsCiAgICAicGhvbmVfbnVtYmVyIjogIis4NDE1OTY1MzMyODUiLAogICAgImZpcmViYXNlIjogewogICAgICAgICJpZGVudGl0aWVzIjogewogICAgICAgICAgICAicGhvbmUiOiBbCiAgICAgICAgICAgICAgICAiKzg0MTU5NjUzMzI4NSIKICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgInNpZ25faW5fcHJvdmlkZXIiOiAicGhvbmUiCiAgICB9Cn0.K-1bm07_Qj66dOFRnSZVCHJ9NYizUvRoD2A8gBMkBfvPv4BvfY1NO_q1ON5umT4yq34tkQ_Ka4MW07m9S399NXWvbSvbUq9hX5SkJZ3CUTLSBscqlJb4yQ2LHq9zyuS9xfnSwo1alw9MLojt-miwBr7QFPtkzVm560AmKXGAkjI"

		claim, err := a.Verify(context.Background(), token)
		assert.Empty(t, claim)
		assert.Error(t, err)
	})

	t.Run("err parse token claims", func(t *testing.T) {
		a := &TokenVerifier{
			jwkURL:     server.URL,
			issuer:     "issuer-1",
			aud:        "aud-1",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImNhM2YxOGFhOGI2ZWZlZjY4ZTNmMWUxMDQyZWM3YjNiN2FkYWMzYzIifQ.ew0gICAgImlzcyI6ICJodHRwOi8vZmFrZV9maXJlYmFzZTo0MDQwMS9mYWtlX2F1ZCIsCiAgICAiYXVkIjogImZha2VfYXVkIiwKICAgICJhdXRoX3RpbWUiOiAxNTk2NTMzMjI1LAogICAgInVzZXJfaWQiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJzdWIiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJpYXQiOiAxNTk2NTMzMjg1LAogICAgImV4cCI6IDE1OTY1MzY4ODUsCiAgICAicGhvbmVfbnVtYmVyIjogIis4NDE1OTY1MzMyODUiLAogICAgImZpcmViYXNlIjogewogICAgICAgICJpZGVudGl0aWVzIjogewogICAgICAgICAgICAicGhvbmUiOiBbCiAgICAgICAgICAgICAgICAiKzg0MTU5NjUzMzI4NSIKICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgInNpZ25faW5fcHJvdmlkZXIiOiAicGhvbmUiCiAgICB9Cn0.K-1bm07_Qj66dOFRnSZVCHJ9NYizUvRoD2A8gBMkBfvPv4BvfY1NO_q1ON5umT4yq34tkQ_Ka4MW07m9S399NXWvbSvbUq9hX5SkJZ3CUTLSBscqlJb4yQ2LHq9zyuS9xfnSwo1alw9MLojt-miwBr7QFPtkzVm560AmKXGAkjI"

		claim, err := a.Verify(context.Background(), token)
		assert.Empty(t, claim)
		assert.Equal(t, "err parse claims: square/go-jose: error in cryptographic primitive, issuer: issuer-1", err.Error())
	})

	t.Run("valid token, invalid issuer", func(t *testing.T) {
		a := &TokenVerifier{
			jwkURL:     server.URL,
			issuer:     "issuer-1",
			aud:        "aud-1",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		token := tokenGenerator.FakeToken(&TokenValue{UserID: ksuid.New().String()})
		claim, err := a.Verify(context.Background(), token)
		assert.Empty(t, claim)
		assert.Equal(t, "claims.Validate: square/go-jose/jwt: validation failed, invalid issuer claim (iss), issuer: issuer-1", err.Error())
	})

	t.Run("valid token, invalid aud", func(t *testing.T) {
		a := &TokenVerifier{
			jwkURL:     server.URL,
			issuer:     "http://firebase:8080/fake_aud",
			aud:        "aud-1",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		token := tokenGenerator.FakeToken(&TokenValue{UserID: ksuid.New().String()})
		claim, err := a.Verify(context.Background(), token)
		assert.Empty(t, claim)
		assert.Equal(t, "claims.Validate: square/go-jose/jwt: validation failed, invalid audience claim (aud), issuer: http://firebase:8080/fake_aud", err.Error())
	})

	t.Run("valid token, expired", func(t *testing.T) {
		a := &TokenVerifier{
			jwkURL:     server.URL,
			issuer:     "http://fake_firebase:40401/fake_aud",
			aud:        "fake_aud",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImNhM2YxOGFhOGI2ZWZlZjY4ZTNmMWUxMDQyZWM3YjNiN2FkYWMzYzIifQ.ewogICAgImlzcyI6ICJodHRwOi8vZmFrZV9maXJlYmFzZTo0MDQwMS9mYWtlX2F1ZCIsCiAgICAiYXVkIjogImZha2VfYXVkIiwKICAgICJhdXRoX3RpbWUiOiAxNTk2NTMzMjI1LAogICAgInVzZXJfaWQiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJzdWIiOiAiZWNkNTlhMDAtZjJkZC00NzcxLTgwNDEtNjA4ODg5NmQwMmM2IiwKICAgICJpYXQiOiAxNTk2NTMzMjg1LAogICAgImV4cCI6IDE1OTY1MzY4ODUsCiAgICAicGhvbmVfbnVtYmVyIjogIis4NDE1OTY1MzMyODUiLAogICAgImZpcmViYXNlIjogewogICAgICAgICJpZGVudGl0aWVzIjogewogICAgICAgICAgICAicGhvbmUiOiBbCiAgICAgICAgICAgICAgICAiKzg0MTU5NjUzMzI4NSIKICAgICAgICAgICAgXQogICAgICAgIH0sCiAgICAgICAgInNpZ25faW5fcHJvdmlkZXIiOiAicGhvbmUiCiAgICB9Cn0.K-1bm07_Qj66dOFRnSZVCHJ9NYizUvRoD2A8gBMkBfvPv4BvfY1NO_q1ON5umT4yq34tkQ_Ka4MW07m9S399NXWvbSvbUq9hX5SkJZ3CUTLSBscqlJb4yQ2LHq9zyuS9xfnSwo1alw9MLojt-miwBr7QFPtkzVm560AmKXGAkjI"

		claim, err := a.Verify(context.Background(), token)
		assert.Empty(t, claim)
		assert.Equal(t, "claims.Validate: square/go-jose/jwt: validation failed, token is expired (exp), issuer: http://fake_firebase:40401/fake_aud", err.Error())
	})

	t.Run("success with firebase token", func(t *testing.T) {
		a := &TokenVerifier{
			jwkURL:     server.URL,
			issuer:     "http://firebase:8080/fake_aud",
			aud:        "fake_aud",
			keySet:     &jose.JSONWebKeySet{},
			httpClient: &http.Client{},
		}

		userID := ksuid.New().String()
		token := tokenGenerator.FakeToken(&TokenValue{UserID: userID})
		claim, err := a.Verify(context.Background(), token)
		assert.Nil(t, err)
		assert.NotEmpty(t, claim)
		assert.Equal(t, userID, claim.Subject)
		assert.Equal(t, a.jwkURL, claim.JwkURL)
	})
}

func loadKeys() map[string]*rsa.PrivateKey {
	keys := make(map[string]*rsa.PrivateKey, len(privateKeys))
	for _, key := range privateKeys {
		block, _ := pem.Decode([]byte(key))
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}

		h := sha1.New()
		h.Write(block.Bytes)
		keys[fmt.Sprintf("%x", h.Sum(nil))] = privateKey
	}

	return keys
}

type publicKeyServer struct {
	certsMap map[string]string
}

func newPublicKeyServer(keysMap map[string]*rsa.PrivateKey, certNotBefore, certNotAfter string) *publicKeyServer {
	s := &publicKeyServer{}
	s.certsMap = make(map[string]string, len(keysMap))
	for keyID, privateKey := range keysMap {

		notBefore, err := time.Parse("Jan 2 15:04:05 2006", certNotBefore)

		notAfter, err := time.Parse("Jan 2 15:04:05 2006", certNotAfter)

		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Fatalf("failed to generate serial number: %s", err)
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"Acme Co"},
			},
			NotBefore: notBefore,
			NotAfter:  notAfter,

			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			log.Fatalf("Failed to create certificate: %s", err)
		}

		s.certsMap[keyID] = string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: derBytes,
		}))
	}
	return s
}

type TokenGenerator struct {
	tmplFirebase *template.Template
	tmplCognito  *template.Template
	keysMap      map[string]*rsa.PrivateKey
}

func newTokenGenerator() *TokenGenerator {
	tmplFirebase := `{
    "iss": "{{ or .IssuerPrefix "http://firebase:8080" }}/{{ or .Audience "fake_aud" }}",
    "aud": "{{ or .Audience "fake_aud" }}",
    "auth_time": {{ .AuthTime }},
    "user_id": "{{ .UserID }}",
    "sub": "{{ .UserID }}",
    "iat": {{ .IssueAt }},
    "exp": {{ .Expiration }},
    "phone_number": "+84{{ .PhoneNumber }}",
    "firebase": {
        "identities": {
            "phone": [
                "+84{{ .PhoneNumber }}"
            ]
        },
        "sign_in_provider": "phone"
    }
}
`

	tmplCognito := `{
    "iss": "{{ or .IssuerPrefix "http://fake_cognito:40401" }}/{{ or .Audience "fake_aud" }}",
    "client_id": "{{ or .Audience "fake_cognito_aud" }}",
    "auth_time": {{ .AuthTime }},
    "username": "{{ .UserID }}",
    "sub": "{{ .UserID }}",
    "iat": {{ .IssueAt }},
    "exp": {{ .Expiration }},
   	"version": 2,
    "origin_jti": "{{ .UserID }}",
    "event_id": "{{ .UserID }}",
	"token_use": "access",
	"scope": "openid",
    "jti": "{{ .UserID }}"
}`

	s := &TokenGenerator{
		tmplFirebase: template.Must(template.New("firebase").Parse(tmplFirebase)),
		tmplCognito:  template.Must(template.New("cognitor").Parse(tmplCognito)),
		keysMap:      loadKeys(),
	}

	return s
}

func (s *TokenGenerator) selectKey(id string) (string, *rsa.PrivateKey) {
	key, ok := s.keysMap[id]
	if ok {
		return id, key
	}

	for id, key := range s.keysMap {
		return id, key
	}

	return "", nil
}

type TokenValue struct {
	IssuerPrefix string
	Audience     string
	AuthTime     string
	UserID       string
	IssueAt      string
	Expiration   string
	PhoneNumber  string
	IsCognito    bool
}

func (s *TokenGenerator) FakeToken(tokenValue *TokenValue) string {
	now := time.Now()
	if tokenValue.AuthTime == "" {
		tokenValue.AuthTime = strconv.FormatInt(now.Add(-1*time.Minute).Unix(), 10)
	}

	if tokenValue.IssueAt == "" {
		tokenValue.IssueAt = strconv.FormatInt(now.Unix(), 10)
	}

	if tokenValue.Expiration == "" {
		tokenValue.Expiration = strconv.FormatInt(now.Add(1*time.Hour).Unix(), 10)
	}

	if tokenValue.UserID == "" {
		tokenValue.UserID = strconv.FormatInt(now.UnixNano(), 10)
	}

	if tokenValue.PhoneNumber == "" {
		tokenValue.PhoneNumber = strconv.FormatInt(now.Unix(), 10)
	}

	var buff bytes.Buffer
	if tokenValue.IsCognito {
		s.tmplCognito.Execute(&buff, &tokenValue)
	} else {
		s.tmplFirebase.Execute(&buff, &tokenValue)
	}

	kid, key := s.selectKey("")
	header := &jws.StandardHeaders{}
	header.Set("kid", kid)
	k, err := jws.Sign(buff.Bytes(), "RS256", key, jws.WithHeaders(header))
	if err != nil {
		log.Panic(err)
	}

	return string(k)
}

func (s *TokenGenerator) ServeJWKSet() func(w http.ResponseWriter, req *http.Request) {
	set := jwk.NewSet()

	keysMap := loadKeys()
	for keyID, privateKey := range keysMap {
		set.Add(convertJWK(privateKey, keyID))
	}

	header := &jws.StandardHeaders{}
	firstKey, valid := set.Get(0)
	if !valid {
		log.Panic(fmt.Errorf("signNewToken: error when get key"))
	}

	header.Set(jwk.KeyIDKey, firstKey)

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&set)
	}
}

func convertJWK(privateKey *rsa.PrivateKey, id string) jwk.Key {
	key, err := jwk.New(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	key.Set(jwk.KeyIDKey, id)
	key.Set(jwk.KeyUsageKey, "sig")
	key.Set(jwk.AlgorithmKey, jwa.RS256)

	return key
}

func TestPrivateKeyFromString(t *testing.T) {
	privateKeyString := privateKeys[0]
	privateKey, err := PrivateKeyFromString(privateKeyString)
	assert.NoError(t, err)
	assert.NotNil(t, privateKey)
}

func TestJWTClaimsFromContextV2(t *testing.T) {
	testCases := []struct {
		name           string
		input          context.Context
		expectedOutput *CustomClaims
		expectedError  error
	}{
		{
			name:           "get claims from ctx successfully",
			input:          context.WithValue(context.Background(), JwtClaims(0), &CustomClaims{}),
			expectedOutput: &CustomClaims{},
			expectedError:  nil,
		},
		{
			name:           "can not get claims because ctx store incorrect value type",
			input:          context.WithValue(context.Background(), JwtClaims(0), time.Time{}),
			expectedOutput: nil,
			expectedError:  ErrFailedToParseJWT,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claims, err := JWTClaimsFromContextV2(testCase.input)
			assert.Equal(t, testCase.expectedOutput, claims)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestOrganizationFromContext(t *testing.T) {
	testCases := []struct {
		name           string
		input          context.Context
		expectedOutput *Organization
		expectedError  error
	}{
		{
			name:           "can not get claims because ctx store incorrect value type",
			input:          context.WithValue(context.Background(), JwtClaims(0), time.Time{}),
			expectedOutput: nil,
			expectedError:  ErrFailedToParseJWT,
		},
		{
			name:           "don't allow manabie claims if it's empty",
			input:          context.WithValue(context.Background(), JwtClaims(0), &CustomClaims{}),
			expectedOutput: nil,
			expectedError:  ErrManabieClaimsIsEmpty,
		},
		{
			name: "don't allow resource path claims if it's empty",
			input: context.WithValue(
				context.Background(),
				JwtClaims(0),
				&CustomClaims{
					Manabie: &ManabieClaims{},
				},
			),
			expectedOutput: nil,
			expectedError:  ErrResourcePathIsEmpty,
		},
		{
			name: "get organization successfully",
			input: context.WithValue(
				context.Background(),
				JwtClaims(0),
				&CustomClaims{
					Manabie: &ManabieClaims{
						ResourcePath: "1",
					},
				},
			),
			expectedOutput: &Organization{
				organizationID: "1",
				schoolID:       1,
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claims, err := OrganizationFromContext(testCase.input)
			assert.Equal(t, testCase.expectedOutput, claims)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
