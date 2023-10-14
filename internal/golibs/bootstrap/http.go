package bootstrap

import (
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPServicer should be implemented to run a HTTP server.
//
// By default, it uses the following middlewares:
//   - ginzap package using bootstrap.Resource.Logger() for logging
//   - auto recovery from panic
type HTTPServicer[T any] interface {
	// SetupHTTP should be used to bind the HTTP handlers or middlewares to the gin.Engine.
	//
	// You should NOT start the HTTP server. The bootstrap package will automatically do that.
	SetupHTTP(T, *gin.Engine, *Resources) error
}

func (b *bootstrapper[T]) setupHTTPService(s HTTPServicer[T], c *T, rsc *Resources) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	l := rsc.Logger().WithOptions(zap.WithCaller(false))
	r.Use(ginzap.Ginzap(l, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(l, true))
	if err := s.SetupHTTP(*c, r, rsc); err != nil {
		return nil, err
	}
	return r, nil
}
