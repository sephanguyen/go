package jerry

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.uber.org/zap"
)

var (
	cacheStatusTag      = tag.MustNewKey("cache_status")
	cacheHitMissCounter = stats.Int64("hasura/cache", "cache hit counter, with status hit/miss", stats.UnitDimensionless)
)

// CacheCounterView should be export if caching is being use
var CacheCounterView = &view.View{
	Name:        "hasura/cache",
	Description: "cache hits counter, by call",
	TagKeys:     []tag.Key{HasuraQueryNameKey, HasuraProxyService, cacheStatusTag},
	Measure:     cacheHitMissCounter,
	Aggregation: view.Count(),
}

func (s *server) addCache(k CacheKey, c cachedItem) {
	s.cacheSimulator.Add(k, &c)
}

var cacheTTL = 60 * time.Second

func (s *server) checkCache(k CacheKey) (*cachedItem, bool) {
	item, has := s.cacheSimulator.Get(k)
	if has {
		if item.createdAt.Before(time.Now().Add(-cacheTTL)) {
			s.logger.Warn("cache hit but timed out", zap.Reflect("cachekey", k))
			s.cacheSimulator.Remove(k)
			return nil, false
		}
		return item, true
	}
	return nil, false
}

type HasuraSessionVar struct {
	DefaultRole  string `json:"x-hasura-default-role"` // note that http header "X-Hasura-Role" will override this value
	UserID       string `json:"x-hasura-user-id"`
	UserGroup    string `json:"x-hasura-user-group"`
	ResourcePath string `json:"x-hasura-resource-path"`
}

type TokenCont struct {
	HasuraVars struct {
		DefaultRole  string `json:"x-hasura-default-role"` // note that http header "X-Hasura-Role" will override this value
		UserID       string `json:"x-hasura-user-id"`
		UserGroup    string `json:"x-hasura-user-group"`
		ResourcePath string `json:"x-hasura-resource-path"`
	} `json:"https://hasura.io/jwt/claims"`
}

func (t TokenCont) IsValid() bool {
	return t.HasuraVars.DefaultRole != "" &&
		t.HasuraVars.UserID != "" &&
		t.HasuraVars.UserGroup != "" &&
		t.HasuraVars.ResourcePath != ""
}

var errNoAuthHeader = errors.New("authorization header is missing")

// extractSessionVarFromReq returns a HasuraSessionVar from a GraphQL request.
func (s *server) extractSessionVarFromReq(req *http.Request) (sessionVar *HasuraSessionVar, err error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errNoAuthHeader
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		s.logger.Error("invalid Authorization header (lacks \"Bearer \" prefix)", zap.String("header", authHeader))
		return nil, fmt.Errorf("invalid auth header")
	}

	tok := authHeader[7:]
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		s.logger.Error("invalid jwt token (does not have 3 parts)", zap.Int("part_count", len(parts)), zap.String("token", tok))
		return nil, fmt.Errorf("invalid jwt token")
	}
	pl, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		s.logger.Error("failed to decode base64 string", zap.String("string", parts[1]))
		return nil, fmt.Errorf("failed to decode base64 string: %s", err)
	}
	var cont TokenCont
	if err := json.Unmarshal(pl, &cont); err != nil {
		s.logger.Error("failed to unmarshal json", zap.ByteString("jsonstring", pl))
		return nil, fmt.Errorf("failed to unmarshal json: %s", err)
	}
	if !cont.IsValid() {
		s.logger.Error("invalid hasura claims in jwt token", zap.Reflect("claims", cont))
		return nil, fmt.Errorf("invalid hasura claims in jwt token")
	}

	return &HasuraSessionVar{
		DefaultRole:  cont.HasuraVars.DefaultRole,
		UserID:       cont.HasuraVars.UserID,
		UserGroup:    cont.HasuraVars.UserGroup,
		ResourcePath: cont.HasuraVars.ResourcePath,
	}, nil
}
