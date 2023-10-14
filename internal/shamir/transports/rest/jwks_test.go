package rest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockJWKSetGetter struct {
	set jwk.Set
}

func (m *mockJWKSetGetter) GetJWKSet() jwk.Set {
	return m.set
}

func TestServer_GetJWKSet(t *testing.T) {
	t.Parallel()
	keys := make([]jwk.Key, 0, 3)
	for i := 0; i < 3; i++ {
		rk, err := rsa.GenerateKey(rand.Reader, 128)
		if err != nil {
			t.Fatal("failed to generate rsa key", err)
		}

		pk := jwk.NewRSAPrivateKey()
		if err := pk.FromRaw(rk); err != nil {
			t.Fatal("failed to generate jwk", err)
		}

		keys = append(keys, pk)
	}
	setJwk := jwk.NewSet()
	for i := range keys {
		setJwk.Add(keys[i])
	}
	m := &mockJWKSetGetter{
		set: setJwk,
	}

	gin.SetMode(gin.TestMode)
	resp := httptest.NewRecorder()
	ctx, r := gin.CreateTestContext(resp)
	SetupGinEngine(r, m, zap.NewNop(), "shamir_http_add")

	ctx.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	r.ServeHTTP(resp, ctx.Request)
	assert.Equal(t, http.StatusOK, resp.Code)

	output := map[string][]struct{}{}
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		t.Error(err)
	}

	if len(output["keys"]) != 3 {
		t.Error("unexpected number of keys returned", len(output["keys"]))
	}
}
