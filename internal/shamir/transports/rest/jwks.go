package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/oauth2-proxy/mockoidc"
	"go.uber.org/zap"
)

// JWKSetGetter implemented by *services.TokenVerifier
type JWKSetGetter interface {
	GetJWKSet() jwk.Set
}

// Server serving jwks endpoint
type Server struct {
	verifier JWKSetGetter
}

// SetupGinEngine starts HTTP REST API
func SetupGinEngine(r *gin.Engine, v JWKSetGetter, zapLogger *zap.Logger, shamirAdd string) error {
	c := &Server{
		verifier: v,
	}

	r.GET("/.well-known/jwks.json", c.GetJWKSet)
	// oidc endpoint
	m := newmockoidc()
	server := &http.Server{
		Addr: shamirAdd,
	}
	m.Server = server
	r.GET(mockoidc.DiscoveryEndpoint, WrapF(Discovery(m)))
	return nil
}

// GetJWKSet returns JWKS json
func (c *Server) GetJWKSet(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.verifier.GetJWKSet())
}

type discoveryResponse struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSUri               string `json:"jwks_uri"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`

	GrantTypesSupported               []string `json:"grant_types_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
}

func Discovery(m *mockoidc.MockOIDC) http.HandlerFunc {
	return func(rw http.ResponseWriter, _ *http.Request) {
		discovery := &discoveryResponse{
			JWKSUri: m.Addr() + "/.well-known/jwks.json",
		}

		resp, err := json.Marshal(discovery)
		if err != nil {
			internalServerError(rw, err.Error())
			return
		}
		jsonResponse(rw, resp)
	}
}

func WrapF(f http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("identifier here\n")
		f(c.Writer, c.Request)
	}
}

func newmockoidc() *mockoidc.MockOIDC {
	m, err := mockoidc.NewServer(nil)
	if err != nil {
		panic(err)
	}

	return m
}
