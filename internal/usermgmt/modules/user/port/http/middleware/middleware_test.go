package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	unitTestEndpoint = "/unit-test"
)

type mockTokenReaderService struct {
	verifyTokenFn     func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	exchangeTokenFn   func(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error)
	verifyTokenV2Fn   func(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error)
	verifySignatureFn func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error)
	getAuthInfo       func(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error)
	exchangeSalesforceToken func(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error)
}

func (mockTokenReaderService mockTokenReaderService) VerifyToken(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
	return mockTokenReaderService.verifyTokenFn(ctx, in)
}

func (mockTokenReaderService mockTokenReaderService) ExchangeToken(ctx context.Context, in *spb.ExchangeTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeTokenResponse, error) {
	return mockTokenReaderService.exchangeTokenFn(ctx, in)
}

func (mockTokenReaderService mockTokenReaderService) VerifyTokenV2(ctx context.Context, in *spb.VerifyTokenRequest, opts ...grpc.CallOption) (*spb.VerifyTokenResponse, error) {
	return mockTokenReaderService.verifyTokenV2Fn(ctx, in)
}

func (mockTokenReaderService mockTokenReaderService) VerifySignature(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
	return mockTokenReaderService.verifySignatureFn(ctx, in)
}

func (mockTokenReaderService mockTokenReaderService) GetAuthInfo(ctx context.Context, in *spb.GetAuthInfoRequest, opts ...grpc.CallOption) (*spb.GetAuthInfoResponse, error) {
	return mockTokenReaderService.getAuthInfo(ctx, in)
}

func (mockTokenReaderService mockTokenReaderService) ExchangeSalesforceToken(ctx context.Context, in *spb.ExchangeSalesforceTokenRequest, opts ...grpc.CallOption) (*spb.ExchangeSalesforceTokenResponse, error) {
	return mockTokenReaderService.exchangeSalesforceToken(ctx, in)
}

func TestMiddleware_VerifySignature(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	req, _ := http.NewRequest(http.MethodGet, unitTestEndpoint, buf)
	allowedGroups := map[string][]string{
		unitTestEndpoint: {constant.RoleOpenAPI},
	}

	groupDecider := &interceptors.GroupDecider{
		AllowedGroups: allowedGroups,
	}

	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		groupDecider.GroupFetcher = func(ctx context.Context, userID string) ([]string, error) {
			return []string{constant.RoleOpenAPI}, nil
		}
		c, engine := gin.CreateTestContext(w)
		c.Request = req
		engine.Use(VerifySignature(zap.NewNop(), groupDecider, mockTokenReaderService{
			verifySignatureFn: func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
				return &spb.VerifySignatureResponse{}, nil
			},
		}))
		engine.GET(unitTestEndpoint, func(ctx *gin.Context) {
			return
		})
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("error case: INVALID_SIGNATURE", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()

		c, engine := gin.CreateTestContext(w)
		c.Request = req
		groupDecider.GroupFetcher = func(ctx context.Context, userID string) ([]string, error) {
			return []string{constant.RoleOpenAPI}, nil
		}

		engine.Use(VerifySignature(zap.NewNop(), groupDecider, mockTokenReaderService{
			verifySignatureFn: func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
				return nil, status.Error(codes.PermissionDenied, errorx.ErrShamirInvalidSignature.Error())
			},
		}))
		engine.GET(unitTestEndpoint, func(ctx *gin.Context) {
			return
		})
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error case: INVALID_PUBLIC_KEY", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()

		c, engine := gin.CreateTestContext(w)
		c.Request = req
		groupDecider.GroupFetcher = func(ctx context.Context, userID string) ([]string, error) {
			return []string{constant.RoleOpenAPI}, nil
		}

		engine.Use(VerifySignature(zap.NewNop(), groupDecider, mockTokenReaderService{
			verifySignatureFn: func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
				return nil, status.Error(codes.PermissionDenied, errorx.ErrShamirInvalidPublicKey.Error())
			},
		}))
		engine.GET(unitTestEndpoint, func(ctx *gin.Context) {
			return
		})
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("error case: INTERNAL_ERROR", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()

		c, engine := gin.CreateTestContext(w)
		c.Request = req
		groupDecider.GroupFetcher = func(ctx context.Context, userID string) ([]string, error) {
			return []string{constant.RoleOpenAPI}, nil
		}

		engine.Use(VerifySignature(zap.NewNop(), groupDecider, mockTokenReaderService{
			verifySignatureFn: func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
				return nil, status.Error(codes.Internal, "")
			},
		}))
		engine.GET(unitTestEndpoint, func(ctx *gin.Context) {
			return
		})
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

}

func TestMiddleware_VerifyAuthorization(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	req, _ := http.NewRequest(http.MethodGet, unitTestEndpoint, buf)
	allowedGroups := map[string][]string{
		unitTestEndpoint: {constant.RoleOpenAPI},
	}

	groupDecider := &interceptors.GroupDecider{
		AllowedGroups: allowedGroups,
	}
	t.Run("error case: PERMISSION_DENIED", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()

		c, engine := gin.CreateTestContext(w)
		c.Request = req
		groupDecider.GroupFetcher = func(ctx context.Context, userID string) ([]string, error) {
			return nil, nil
		}

		engine.Use(VerifySignature(zap.NewNop(), groupDecider, mockTokenReaderService{
			verifySignatureFn: func(ctx context.Context, in *spb.VerifySignatureRequest, opts ...grpc.CallOption) (*spb.VerifySignatureResponse, error) {
				return &spb.VerifySignatureResponse{}, nil
			},
		}))
		engine.GET(unitTestEndpoint, func(ctx *gin.Context) {
			return
		})
		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
