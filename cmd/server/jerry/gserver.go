package jerry

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/jerry/configurations"

	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yudai/gojsondiff"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

func init() {
	s := &server{}
	bootstrap.WithHTTP[configurations.Config](s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	logger *zap.Logger
	configurations.Config
	routingRules []RoutingRule

	cacheSimulator *lru.Cache[CacheKey, *cachedItem]
}

type CacheKey struct {
	Query        string
	DefaultRole  string
	UserID       string
	UserGroup    string
	ResourcePath string
	Variables    string
}
type cachedItem struct {
	createdAt time.Time
	result    []byte
	// to sanity check if we really impl cache
	// in Manabie usecase (at the time this is written), rls policy works based on resource_path and user_id
	// we sanity check to ensure cache of this school does not leak to other school's response
	defaultRole  string
	userID       string
	userGroup    string
	resourcePath string
}

type RoutingRule struct {
	exactMatch    string
	matchPrefix   string
	rewritePrefix string
	addr          string
	port          string
	proxy         *httputil.ReverseProxy
}

func (*server) ServerName() string {
	return "jerry"
}

func (s *server) WithPrometheusCollectors(*bootstrap.Resources) []prometheus.Collector {
	return nil
}

func (s *server) InitMetricsValue() {
}

// Provide us with your own opencensus, if any
func (s *server) WithOpencensusViews() []*view.View {
	return HasuraClientViews
}

func (s *server) SetupHTTP(c configurations.Config, e *gin.Engine, rsc *bootstrap.Resources) error {
	e.Any("/*proxyPath", s.proxy)
	return nil
}

func rewritePrefix(pathWithoutPrefix string, newPrefix string) string {
	return newPrefix + pathWithoutPrefix
}

func isRouteMatch(requestedPath string, r RoutingRule) bool {
	if r.exactMatch != "" {
		return requestedPath == r.exactMatch
	}
	return strings.HasPrefix(requestedPath, r.matchPrefix)
}

func (s *server) proxy(c *gin.Context) {
	for _, r := range s.routingRules {
		p := c.Param("proxyPath")
		isMatched := isRouteMatch(p, r)
		if !isMatched {
			continue
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("reading request body: %s", err))
			return
		}
		c.Request.Body.Close() //  must close
		var reqBody requestBody
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		err = json.Unmarshal(bodyBytes, &reqBody)
		// We still wants to proxy the request even if the payload does not follow syntax of a graphql request
		// TODO: we can recognize which path only serves graphql requests and which only serves normal http request of Hasura
		// so we can save time parsing payload that will eventually return errors
		if err != nil {
			s.logger.Error("failed to unmarshal request body",
				zap.Error(err), zap.String("endpoint", c.Request.RequestURI),
				zap.ByteString("request_body", bodyBytes),
			)
			r.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}

		queryName, isQuery, foundQueryName := extractQueryName(reqBody.Query)
		if foundQueryName {
			if isQuery {
				s.logger.Info("incoming hasura query", zap.String("query", queryName))
			} else {
				s.logger.Info("incoming hasura mutation", zap.String("query", queryName))
			}
			ctx, err := tag.New(c.Request.Context(),
				tag.Upsert(HasuraQueryNameKey, queryName),
				tag.Upsert(HasuraProxyService, r.addr),
			)
			if err != nil {
				_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to upsert opencensus tag %w", err))
				return
			}
			var (
				cacheAble = false
				cacheHit  = false
				cacheKey  CacheKey
				cacheRet  *cachedItem

				defaultRole, userID, userGroup, resourcePath string
			)

			if isQuery {
				sessionvars, err := s.extractSessionVarFromReq(c.Request)
				// TODO(@anhpngt) DRY out this part of the code later
				if errors.Is(err, errNoAuthHeader) { //nolint:gocritic
					cacheAble = true
					cacheKey = CacheKey{
						Query:        queryName,
						DefaultRole:  "",
						UserID:       "",
						UserGroup:    "",
						ResourcePath: "",
						Variables:    string(reqBody.Variable),
					}
					cachedItem, hasCache := s.checkCache(cacheKey)
					cacheHit = hasCache
					if !hasCache {
						s.logger.Debug("cache missed", zap.Reflect("query", reqBody))
						err = stats.RecordWithTags(ctx, []tag.Mutator{
							tag.Upsert(cacheStatusTag, "miss"),
						}, cacheHitMissCounter.M(1))
						if err != nil {
							s.logger.Error("stats.RecordWithTags", zap.Error(err))
						}
					} else {
						cacheRet = cachedItem
					}
				} else if err == nil {
					cacheAble = true
					defaultRole = sessionvars.DefaultRole
					userID = sessionvars.UserID
					userGroup = sessionvars.UserGroup
					resourcePath = sessionvars.ResourcePath
					cacheKey = CacheKey{
						Query:        queryName,
						DefaultRole:  sessionvars.DefaultRole,
						UserID:       sessionvars.UserID,
						UserGroup:    sessionvars.UserGroup,
						ResourcePath: sessionvars.ResourcePath,
						Variables:    string(reqBody.Variable),
					}
					cachedItem, hasCache := s.checkCache(cacheKey)
					cacheHit = hasCache
					if !hasCache {
						s.logger.Debug("cache missed", zap.Reflect("query", reqBody))
						err = stats.RecordWithTags(ctx, []tag.Mutator{
							tag.Upsert(cacheStatusTag, "miss"),
						}, cacheHitMissCounter.M(1))
						if err != nil {
							s.logger.Error("stats.RecordWithTags", zap.Error(err))
						}
					} else {
						cacheRet = cachedItem
					}
				} else {
					s.logger.Error("extractSessionVarFromReq", zap.Error(err))
				}
			}

			newReq := c.Request.WithContext(ctx)

			buff := &bytes.Buffer{}

			// We want to get result from Hasura to check if that response == content cached
			// TODO: if we really impl cache, we must have token validation, now we don't need it
			customWriter := &teeHTTPResponseWriter{
				inner:     c.Writer,
				teeWriter: io.MultiWriter(c.Writer, buff),
			}
			r.proxy.ServeHTTP(customWriter, newReq)
			var rawbs []byte
			enc := customWriter.Header().Get("Content-Encoding")
			if enc == "gzip" {
				gzreader, err := gzip.NewReader(buff)
				if err != nil {
					s.logger.Error("gzip.NewReader", zap.Error(err))
					return
				}
				decodedBs, err := io.ReadAll(gzreader)
				if err != nil {
					s.logger.Error("read from gzReader", zap.Error(err))
					return
				}
				rawbs = decodedBs
			} else {
				rawbs = buff.Bytes()
			}

			if cacheHit {
				diff, err := gojsondiff.New().Compare(cacheRet.result, rawbs)
				if err != nil {
					s.logger.Error("jsondiff", zap.Error(err))
					return
				}
				if diff.Modified() {
					err = stats.RecordWithTags(ctx, []tag.Mutator{
						tag.Upsert(cacheStatusTag, "hit-different"),
					}, cacheHitMissCounter.M(1))
					s.logger.Debug("cache INCORRECTLY hit", zap.Reflect("query", reqBody))
					cacheHit = false
				} else {
					err = stats.RecordWithTags(ctx, []tag.Mutator{
						tag.Upsert(cacheStatusTag, "hit-equal"),
					}, cacheHitMissCounter.M(1))
					s.logger.Debug("cache correctly hit", zap.Reflect("query", reqBody))
				}
				if err != nil {
					s.logger.Error("stats.RecordWithTags", zap.Error(err))
					return
				}
				// sanity check
				if cacheRet.defaultRole != defaultRole {
					s.logger.Error("cache item defaultRole != session's defaultRole", zap.String("cache_default_role", cacheRet.defaultRole), zap.String("session_default_role", defaultRole))
				}
				if cacheRet.userID != userID {
					s.logger.Error("cache item userID != session's userID", zap.String("cache_user_id", cacheRet.userID), zap.String("session_user_id", userID))
				}
				if cacheRet.userID != userID {
					s.logger.Error("cache item userGroup != session's userGroup", zap.String("cache_user_group", cacheRet.userGroup), zap.String("session_user_group", userGroup))
				}
				if cacheRet.resourcePath != resourcePath {
					s.logger.Error("cache item resourcePath != session's resourcePath", zap.String("cache_rp", cacheRet.resourcePath), zap.String("session_rp", resourcePath))
				}
			}

			if cacheAble && !cacheHit {
				s.addCache(cacheKey, cachedItem{
					createdAt:    time.Now(),
					defaultRole:  defaultRole,
					userID:       userID,
					userGroup:    userGroup,
					resourcePath: resourcePath,
					result:       rawbs,
				})
			}

			return
		} else {
			s.logger.Warn("failed to find query name in incoming request body",
				zap.String("request_uri", c.Request.RequestURI),
				zap.String("request_body", reqBody.Query),
			)
			r.proxy.ServeHTTP(c.Writer, c.Request)
			return
		}
	}

	s.logger.Error("no route match for hasura request",
		zap.String("method", c.Request.Method),
		zap.Reflect("header", c.Request.Header),
		zap.String("url", c.Request.RequestURI),
	)
	_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("no route match"))
}

type teeHTTPResponseWriter struct {
	inner     http.ResponseWriter
	teeWriter io.Writer
}

func (t *teeHTTPResponseWriter) Header() http.Header {
	return t.inner.Header()
}

func (t *teeHTTPResponseWriter) WriteHeader(statusCode int) {
	t.inner.WriteHeader(statusCode)
}

func (t *teeHTTPResponseWriter) Write(b []byte) (int, error) {
	return t.teeWriter.Write(b)
}

var queryNameRegexp = regexp.MustCompile("(query|mutation) ([^({ ]*)")

func extractQueryName(fullQuery string) (string, bool, bool) {
	matches := queryNameRegexp.FindStringSubmatch(fullQuery)
	if len(matches) != 0 {
		return matches[2], matches[1] == "query", true
	}
	return "", false, false
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	cacheSimulator, err := lru.New[CacheKey, *cachedItem](200)
	if err != nil {
		return err
	}
	s.cacheSimulator = cacheSimulator
	for idx := range c.HasuraRoutingRules {
		rule := c.HasuraRoutingRules[idx]
		remoteUrl, err := url.Parse(fmt.Sprintf("%s://%s:%s", "http", rule.ForwardHost, rule.ForwardPort))
		if err != nil {
			return err
		}
		for _, exact := range rule.MatchedExacts {
			proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
			proxy.Director = func(req *http.Request) {
				host := fmt.Sprintf("%s:%s", rule.ForwardHost, rule.ForwardPort)
				req.Host = host
				req.URL.Scheme = "http"
				req.URL.Host = host
				path := req.URL.Path
				if rule.RewriteUri != "" {
					path = rule.RewriteUri
				}
				req.URL.Path = path
			}
			proxy.Transport = &ochttp.Transport{
				Base: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}

			s.routingRules = append(s.routingRules, RoutingRule{
				exactMatch:    exact,
				rewritePrefix: rule.RewriteUri,
				addr:          rule.ForwardHost,
				port:          rule.ForwardPort,
				proxy:         proxy,
			})
		}
		for _, matchedPrefix := range rule.MatchedPrefixes {
			proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
			proxy.Director = func(req *http.Request) {
				host := fmt.Sprintf("%s:%s", rule.ForwardHost, rule.ForwardPort)
				req.Host = host
				req.URL.Scheme = "http"
				req.URL.Host = host
				path := req.URL.Path
				if rule.RewriteUri != "" {
					pathWithoutPrefix := path[len(matchedPrefix):]
					path = rewritePrefix(pathWithoutPrefix, rule.RewriteUri)
				}
				req.URL.Path = path
			}
			proxy.Transport = &ochttp.Transport{
				Base: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			s.routingRules = append(s.routingRules, RoutingRule{
				matchPrefix:   matchedPrefix,
				rewritePrefix: rule.RewriteUri,
				addr:          rule.ForwardHost,
				port:          rule.ForwardPort,
				proxy:         proxy,
			})
		}
	}
	s.logger = rsc.Logger()
	return nil
}

func (*server) GracefulShutdown(context.Context) {}

var (
	HasuraQueryNameKey = tag.MustNewKey("hasura_query")
	HasuraProxyService = tag.MustNewKey("hasura_proxy_service")
)

type hasuraRoundTrip struct {
	base http.RoundTripper
}
type hasuraQueryNameKey struct{}

type requestBody struct {
	Query    string          `json:"query"`
	Variable json.RawMessage `json:"variables"`
}

var (
	HasuraClientViews = []*view.View{
		ClientCompletedCount,
		ClientRoundtripLatencyDistribution,
		CacheCounterView,
	}
	ClientCompletedCount = &view.View{
		Name:        "opencensus.io/http/client/completed_count",
		Measure:     ochttp.ClientRoundtripLatency,
		Aggregation: view.Count(),
		Description: "Count of completed requests, by HTTP method and response status",
		TagKeys:     []tag.Key{ochttp.KeyClientStatus, HasuraQueryNameKey, HasuraProxyService},
	}
	ClientRoundtripLatencyDistribution = &view.View{
		Name:        "opencensus.io/http/client/roundtrip_latency",
		Measure:     ochttp.ClientRoundtripLatency,
		Aggregation: interceptors.MillisecondsDistribution,
		Description: "End-to-end latency, by HTTP method and response status",
		TagKeys:     []tag.Key{ochttp.KeyClientStatus, HasuraQueryNameKey, HasuraProxyService},
	}
)
