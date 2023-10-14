package caching

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"google.golang.org/grpc/metadata"
)

var (
	cacheGroupTag       = tag.MustNewKey("cache_group")
	cacheStatusTag      = tag.MustNewKey("cache_status")
	cacheHitMissCounter = stats.Int64("cache/call", "cache hit counter, with status hit/miss", stats.UnitDimensionless)
)

// CacheCounterView should be export if caching is being use
var CacheCounterView = &view.View{
	Name:        "cache/local/call",
	Description: "cache hits counter, by call",
	TagKeys:     []tag.Key{cacheGroupTag, cacheStatusTag},
	Measure:     cacheHitMissCounter,
	Aggregation: view.Count(),
}

// LocalCacher interface
type LocalCacher interface {
	Set(ctx context.Context, group, key string, value interface{}, ttl time.Duration) bool
	Get(ctx context.Context, group, key string) (interface{}, bool)
	Del(ctx context.Context, group, key string) bool
}

// RistrettoCacher mostly for testing
type RistrettoCacher interface {
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
	Get(key interface{}) (interface{}, bool)
	Del(key interface{})
}

// RistrettoWrapper implements LocalCacher with some monitoring
type RistrettoWrapper struct {
	RistrettoCacher
}

// IsNoCache check "no-cache" metadata
func IsNoCache(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	_, ok = md["no-cache"]
	return ok
}

// Set set value with cost 0
func (c *RistrettoWrapper) Set(ctx context.Context, group, key string, value interface{}, ttl time.Duration) bool {
	ctx, span := interceptors.StartSpan(ctx, "RistrettoWrapper.Set")
	defer span.End()

	return c.RistrettoCacher.SetWithTTL(group+key, value, 0, ttl)
}

// Get returns values if found
func (c *RistrettoWrapper) Get(ctx context.Context, group, key string) (interface{}, bool) {
	ctx, span := interceptors.StartSpan(ctx, "RistrettoWrapper.Get")
	defer span.End()

	v, hit := c.RistrettoCacher.Get(group + key)
	if hit {
		stats.RecordWithTags(ctx, []tag.Mutator{
			tag.Upsert(cacheGroupTag, group),
			tag.Upsert(cacheStatusTag, "hit"),
		}, cacheHitMissCounter.M(1))
	} else {
		stats.RecordWithTags(ctx, []tag.Mutator{
			tag.Upsert(cacheGroupTag, group),
			tag.Upsert(cacheStatusTag, "miss"),
		}, cacheHitMissCounter.M(1))
	}

	return v, hit
}

// Del always returns true
func (c *RistrettoWrapper) Del(ctx context.Context, group, key string) bool {
	ctx, span := interceptors.StartSpan(ctx, "RistrettoWrapper.Del")
	defer span.End()

	c.RistrettoCacher.Del(group + key)
	return true
}
