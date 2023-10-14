package hasura

import (
	"context"
	"errors"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type CacheStore struct {
	svcname  string
	rdb      *redis.Client
	cacheTTL time.Duration
	l        *zap.Logger
}

func NewCacheStore(svcname, redisAddress string, cacheTTL time.Duration, logger *zap.Logger) *CacheStore {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	logger.Info("redis address", zap.String("address", redisAddress))
	out := &CacheStore{
		svcname:  svcname,
		rdb:      rdb,
		cacheTTL: cacheTTL,
		l:        logger,
	}
	return out
}

// ServiceName returns the name of the service for this cache store, usually for metric ID.
func (c CacheStore) ServiceName() string {
	return c.svcname
}

// UpsertAnalyze analyzes and reports cache hit/miss/stale status to metrics,
// then upserts the new cache response to the cache store.
//
// Question: do we upsert the cache immediately, or do we wait for the cache to timeout first
// before re-adding it again?
func (c *CacheStore) UpsertAnalyze(ctx context.Context, req *Request, live *Response) error {
	return multierr.Combine(
		c.analyze(ctx, req, live),
		c.set(ctx, req, live),
	)
}

// analyze gets the data from Redis server then compare that response with the live response.
func (c *CacheStore) analyze(ctx context.Context, req *Request, live *Response) error {
	res, err := c.get(ctx, req)
	if errors.Is(err, redis.Nil) || res == nil {
		c.l.Debug("hasura cache miss",
			zap.String("query", req.queryName))
		if err := c.reportMiss(ctx, req); err != nil {
			return fmt.Errorf("failed to report cache miss: %s", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to lookup cache: %s", err)
	}

	match, err := res.Compare(live)
	if err != nil {
		return fmt.Errorf("failed to compare cache with live response: %s", err)
	}
	if match {
		c.l.Debug("hasura cache hit", zap.String("query", req.queryName))
		if err := c.reportHit(ctx, req); err != nil {
			return fmt.Errorf("failed to report cache hit: %s", err)
		}
	} else {
		c.l.Warn("hasura cache stale",
			zap.String("query", req.queryName),
			zap.ByteString("cache", res.Data()),
			zap.ByteString("live", live.Data()),
		)
		if err := c.reportStale(ctx, req); err != nil {
			return fmt.Errorf("failed to report cache staleness: %s", err)
		}
	}
	return nil
}

// get performs a GET operation to Redis server.
func (c *CacheStore) get(ctx context.Context, req *Request) (*Response, error) {
	if req.queryType != QueryTypeQuery {
		return nil, fmt.Errorf("cannot cache query with type %q (only %q is allowed)", req.queryType, QueryTypeQuery)
	}
	s, err := c.rdb.Get(ctx, req.Key()).Result()
	if err != nil {
		return nil, err
	}
	return &Response{data: []byte(s)}, nil
}

const cacheSizeThresholdBytes = 1024 * 1024 * 128 // 128 MB

// set performs a SET operation to Redis server.
func (c *CacheStore) set(ctx context.Context, req *Request, res *Response) error {
	if req.queryType != QueryTypeQuery {
		return fmt.Errorf("cannot cache query with type %q (only %q is allowed)", req.queryType, QueryTypeQuery)
	}

	// warn if the cache size is too large (>128MB)
	if len(res.Data()) >= cacheSizeThresholdBytes {
		c.l.Warn("cache size too large",
			zap.Int("size_in_bytes", len(res.Data())),
			zap.String("query_name", req.queryName),
		)
	}

	return c.rdb.Set(ctx, req.Key(), res.Data(), c.cacheTTL).Err()
}

func (c *CacheStore) reportHit(ctx context.Context, r *Request) error {
	if err := c.reportHitPersistent(ctx, r); err != nil {
		return err
	}
	return stats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(ProxyServiceKey, c.svcname),
			tag.Upsert(QueryNameKey, r.ID()),
			tag.Upsert(CacheStatusKey, "hit"),
		},
		cacheHitMissCounter.M(1),
	)
}

const (
	redisKeyHasuraCacheHitMetric   = "metrics_hasura_cache_hit"
	redisKeyhasuraCacheMissMetric  = "metrics_hasura_cache_miss"
	redisKeyhasuraCacheStaleMetric = "metrics_hasura_cache_stale"
)

func (c *CacheStore) reportHitPersistent(ctx context.Context, r *Request) error {
	v, err := c.rdb.HIncrBy(ctx, redisKeyHasuraCacheHitMetric, r.ID(), 1).Result()
	if err != nil {
		return fmt.Errorf("failed to increment %s/%s", redisKeyHasuraCacheHitMetric, r.ID())
	}
	if err := stats.RecordWithTags(ctx, r.TagMutators(), cacheHitCounter.M(v)); err != nil {
		return fmt.Errorf("failed to record metric for %s", cacheHitCounter.Name())
	}
	return nil
}

func (c *CacheStore) reportMiss(ctx context.Context, r *Request) error {
	if err := c.reportMissPersistent(ctx, r); err != nil {
		return err
	}
	return stats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(ProxyServiceKey, c.svcname),
			tag.Upsert(QueryNameKey, r.ID()),
			tag.Upsert(CacheStatusKey, "miss"),
		},
		cacheHitMissCounter.M(1),
	)
}

func (c *CacheStore) reportMissPersistent(ctx context.Context, r *Request) error {
	v, err := c.rdb.HIncrBy(ctx, redisKeyhasuraCacheMissMetric, r.ID(), 1).Result()
	if err != nil {
		return fmt.Errorf("failed to increment %s/%s", redisKeyhasuraCacheMissMetric, r.ID())
	}
	if err := stats.RecordWithTags(ctx, r.TagMutators(), cacheMissCounter.M(v)); err != nil {
		return fmt.Errorf("failed to record metric for %s", cacheMissCounter.Name())
	}
	return nil
}

func (c *CacheStore) reportStale(ctx context.Context, r *Request) error {
	if err := c.reportStalePersistent(ctx, r); err != nil {
		return err
	}
	return stats.RecordWithTags(ctx,
		[]tag.Mutator{
			tag.Upsert(ProxyServiceKey, c.svcname),
			tag.Upsert(QueryNameKey, r.ID()),
			tag.Upsert(CacheStatusKey, "stale"),
		},
		cacheHitMissCounter.M(1),
	)
}

func (c *CacheStore) reportStalePersistent(ctx context.Context, r *Request) error {
	v, err := c.rdb.HIncrBy(ctx, redisKeyhasuraCacheStaleMetric, r.ID(), 1).Result()
	if err != nil {
		return fmt.Errorf("failed to increment %s/%s", redisKeyhasuraCacheStaleMetric, r.ID())
	}
	if err := stats.RecordWithTags(ctx, r.TagMutators(), cacheStaleCounter.M(v)); err != nil {
		return fmt.Errorf("failed to record metric for %s", cacheStaleCounter.Name())
	}
	return nil
}
