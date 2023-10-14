package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/manabie-com/backend/internal/jerry/configurations"
	"github.com/manabie-com/backend/internal/jerry/services/hasura"

	"github.com/gin-gonic/gin"
	"go.opencensus.io/plugin/ochttp"
	"go.uber.org/zap"
)

func RegisterHasuraCacheService(ge *gin.Engine, l *zap.Logger, c configurations.HasuraCacheConfig) error {
	s, err := newHasuraCacheService(l, c)
	if err != nil {
		return err
	}

	ge.POST("/v1/graphql", s.handle)
	return nil
}

type hasuraCacheService struct {
	eavedropper *hasura.Eavesdropper
	proxy       *httputil.ReverseProxy
	cacher      *hasura.CacheStore

	l *zap.Logger
}

func newHasuraCacheService(l *zap.Logger, c configurations.HasuraCacheConfig) (*hasuraCacheService, error) {
	targetURL, err := url.Parse(c.HasuraURL())
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = &ochttp.Transport{
		Base: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
	}
	if err != nil {
		return nil, fmt.Errorf("failed to init reverse proxy: %s", err)
	}

	s := &hasuraCacheService{
		eavedropper: hasura.NewEavesdropper(c.HasuraHost, l),
		proxy:       proxy,
		cacher:      hasura.NewCacheStore(c.Name(), c.RedisAddress(), c.CacheTTL(), l),
		l:           l,
	}
	return s, nil
}

func (s *hasuraCacheService) handle(c *gin.Context) {
	hasuraRequest, err := s.eavedropper.EavesdropHTTPRequest(c.Request)
	if err != nil {
		s.l.Error("failed to intercept Hasura request", zap.Error(err))
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if !hasuraRequest.Cachable() {
		s.l.Warn("query is not cachable")
		return
	}

	hasuraLiveResponse, err := hasura.EavesdropHTTPResponse(s.proxy, c.Writer, c.Request)
	if err != nil {
		s.l.Error("failed to intercept Hasura response", zap.Error(err))
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err := s.cacher.UpsertAnalyze(c.Request.Context(), hasuraRequest, hasuraLiveResponse); err != nil {
		s.l.Error("failed to analyze/update cache", zap.Error(err))
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}
