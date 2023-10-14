package caching

import (
	"context"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestIsNoCache(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	assert.False(t, IsNoCache(ctx), "expecting IsNoCache return false when no special header added")

	ctx = metadata.NewIncomingContext(ctx, metadata.MD{"no-cache": []string{"1"}})
	assert.True(t, IsNoCache(ctx), "expecting IsNoCache return true when no-cache added")
}

func TestRistrettoWrapper(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e5,
		MaxCost:     1e4,
		BufferItems: 64,
		Cost: func(value interface{}) int64 { // need to find a better way to handle
			switch value.(type) {
			case string:
				return 1
			default:
				return 1
			}
		},
	})

	if err != nil {
		t.Error("init error", err)
	}

	w := &RistrettoWrapper{
		RistrettoCacher: cache,
	}
	v := int(69)
	w.Set(ctx, "test-group", "test-key", v, 100*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	n, ok := w.Get(ctx, "test-group", "test-key")
	if !ok {
		t.Error("unexpected issue when write")
	}

	if n.(int) != v {
		t.Error("unexpected error returned")
	}
	time.Sleep(110 * time.Millisecond)

	_, ok = w.Get(ctx, "test-group", "test-key")
	if ok {
		t.Error("unexpected value return")
	}

	w.Set(ctx, "test-group", "test-key-2", v, 100*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	w.Del(ctx, "test-group", "test-key-2")
	n, ok = w.Get(ctx, "test-group", "test-key-2")
	if ok {
		t.Error("expecting no error return after delete call")
	}
}
