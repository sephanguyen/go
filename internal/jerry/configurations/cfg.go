package configurations

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
)

type Config struct {
	Common             configs.CommonConfig
	HasuraRoutingRules []RoutingRule `yaml:"hasura_routing_rules"`
}
type RoutingRule struct {
	MatchedExacts   []string `yaml:"matched_exacts"`
	MatchedPrefixes []string `yaml:"matched_prefixes"`
	RewriteUri      string   `yaml:"rewrite_uri"`
	ForwardHost     string   `yaml:"forward_host"`
	ForwardPort     string   `yaml:"forward_port"`
}

type Config2 struct {
	Common            configs.CommonConfig
	HasuraCacheConfig HasuraCacheConfig `yaml:"hasura_cache_config"`
}

type HasuraCacheConfig struct {
	// HasuraHost is the host value to Hasura server (e.g. bob-hasura). Required.
	// Use HasuraURL to retrieve the full Hasura URL.
	HasuraHost string `yaml:"hasura_host"`

	// HasuraPort is the port value to Hasura server. Default is "8080".
	// Use HasuraURL to retrieve the full Hasura URL.
	HasuraPort *string `yaml:"hasura_port"`

	// RedisURL is the URL to the Redis server to store cache.
	// Default is ":6379".
	RedisAddr *string `yaml:"redis_addr"`

	// TTLInSeconds is the time-to-live for cache in Redis.
	// It must be positive. Default is 60.
	//
	// Use CacheTTL to retrieve this value.
	TTLInSeconds *int `yaml:"ttl_in_seconds"`
}

// Name returns the name for metric identitication.
func (c HasuraCacheConfig) Name() string {
	return c.HasuraHost
}

var defaultHasuraPort = "8080"

// HasuraURL returns the URL to upstream Hasura server.
func (c HasuraCacheConfig) HasuraURL() string {
	if c.HasuraHost == "" {
		panic("HasuraCacheConfig.HasuraHost is empty")
	}
	port := c.HasuraPort
	if port == nil {
		port = &defaultHasuraPort
	}
	return "http://" + c.HasuraHost + ":" + *port
}

func (c HasuraCacheConfig) RedisAddress() string {
	if c.RedisAddr != nil {
		return *c.RedisAddr
	}
	return ":6379"
}

// CacheTTL returns the cache's expiration time.
func (c HasuraCacheConfig) CacheTTL() time.Duration {
	if c.TTLInSeconds == nil {
		return time.Second * 60
	}
	if *c.TTLInSeconds <= 0 {
		panic(fmt.Errorf("HasuraCacheConfig.TTLInSeconds must be positive (current value: %d)", *c.TTLInSeconds))
	}
	return time.Second * (time.Duration(*c.TTLInSeconds))
}
