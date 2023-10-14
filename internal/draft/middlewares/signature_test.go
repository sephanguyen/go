package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func TestEnsureAuthRedirectsToLogin(t *testing.T) {
	t.Run("Signature match", func(t *testing.T) {
		w := httptest.NewRecorder()
		logStr := `{
		"tables": "courses",
		"service": "bob",
		"school_id": "100000",
		"per_batch": 100,
		"before_at": "2023-02-02T13:46:20+07:00",
		"after_at": "2023-02-02T13:45:49+07:00"
	}`
		_, engine := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodPost, "/draft-http/v1/data_clean/payload", strings.NewReader(logStr))
		req.Header.Set("Manabie-Signature", "970a7dd554e917498ca9fcddb88df25ba0a81612bf9e41537d32f121f040cdbb")
		engine.Use(VerifySignature(HeaderKey, "123"))
		engine.ServeHTTP(w, req)

		assert.Equal(t, 404, w.Result().StatusCode)
	})

	t.Run("Signature not match", func(t *testing.T) {
		w := httptest.NewRecorder()
		logStr := `{
		"tables": "courses",
		"service": "bob",
		"school_id": "100000",
		"per_batch": 100,
		"before_at": "2023-02-02T13:46:20+07:00",
		"after_at": "2023-02-02T13:45:49+07:00"
	}`
		_, engine := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodPost, "/draft-http/v1/data_clean/payload", strings.NewReader(logStr))
		req.Header.Set("Manabie-Signature", "example")
		engine.Use(VerifySignature(HeaderKey, "123"))
		engine.ServeHTTP(w, req)

		assert.Equal(t, 400, w.Result().StatusCode)
	})
}
