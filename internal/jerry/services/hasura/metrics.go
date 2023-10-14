package hasura

import (
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	QueryNameKey    = tag.MustNewKey("hasura_query")
	ProxyServiceKey = tag.MustNewKey("hasura_proxy_service")
	CacheStatusKey  = tag.MustNewKey("cache_status")

	cacheHitMissCounter = stats.Int64("hasura/cache", "cache hit counter, with status hit/miss", stats.UnitDimensionless)
	cacheHitCounter     = stats.Int64("manabie.io/hasura/cache/hit", "Total number of Hasura cache hit", stats.UnitDimensionless)
	cacheMissCounter    = stats.Int64("manabie.io/hasura/cache/miss", "Total number of Hasura cache miss", stats.UnitDimensionless)
	cacheStaleCounter   = stats.Int64("manabie.io/hasura/cache/stale", "Total number of Hasura cache stale", stats.UnitDimensionless)
)

func Views() []*view.View {
	return []*view.View{
		{
			Name:        "hasura/cache",
			Measure:     cacheHitMissCounter,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{QueryNameKey, ProxyServiceKey, CacheStatusKey},
			Description: "cache hits counter, by call",
		},
		{
			Measure:     cacheHitCounter,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{QueryNameKey, ProxyServiceKey},
			Description: "Total number of Hasura cache hit",
		},
		{
			Measure:     cacheMissCounter,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{QueryNameKey, ProxyServiceKey},
			Description: "Total number of Hasura cache miss",
		},
		{
			Measure:     cacheStaleCounter,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{QueryNameKey, ProxyServiceKey},
			Description: "Total number of Hasura cache stale (cache out-of-date compare to real value)",
		},
		{
			Name:        "opencensus.io/http/client/completed_count",
			Measure:     ochttp.ClientRoundtripLatency,
			Aggregation: view.Count(),
			TagKeys:     []tag.Key{ochttp.KeyClientStatus, QueryNameKey, ProxyServiceKey},
			Description: "Count of completed requests, by HTTP method and response status",
		},
		{
			Name:        "opencensus.io/http/client/roundtrip_latency",
			Measure:     ochttp.ClientRoundtripLatency,
			Aggregation: interceptors.MillisecondsDistribution,
			Description: "End-to-end latency, by HTTP method and response status",
			TagKeys:     []tag.Key{ochttp.KeyClientStatus, QueryNameKey, ProxyServiceKey},
		},
	}
}
