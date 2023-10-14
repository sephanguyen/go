package repositories

import (
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/jackc/pgtype"
)

type OnlineUserCacheRepo struct {
	simplelru.LRUCache
}

const OnlineUser = "online-user:"

type CacheUserOnlineEntity struct {
	Nodes     pgtype.TextArray
	CreatedAt time.Time
}

func (r *OnlineUserCacheRepo) Find(userIDs pgtype.TextArray, since pgtype.Timestamptz) (mapNodeUserIDs map[pgtype.Text][]string, pgOfflineUserIDs pgtype.TextArray) {
	mapNodeUserIDs = make(map[pgtype.Text][]string)
	var offlineUserIDs []string

	for _, userID := range userIDs.Elements {
		v, ok := r.LRUCache.Get(OnlineUser + userID.String)
		if !ok {
			offlineUserIDs = append(offlineUserIDs, userID.String)
			continue
		}

		cacheEntity, ok := v.(*CacheUserOnlineEntity)
		if !ok {
			offlineUserIDs = append(offlineUserIDs, userID.String)
			continue
		}

		if cacheEntity.CreatedAt.Before(since.Time) {
			r.LRUCache.Remove(OnlineUser + userID.String)
			offlineUserIDs = append(offlineUserIDs, userID.String)

			continue
		}

		for _, node := range cacheEntity.Nodes.Elements {
			mapNodeUserIDs[node] = append(mapNodeUserIDs[node], userID.String)
		}
	}

	pgOfflineUserIDs.Set(offlineUserIDs)

	return mapNodeUserIDs, pgOfflineUserIDs
}

func (r *OnlineUserCacheRepo) Add(userID pgtype.Text, nodes pgtype.TextArray) bool {
	return r.LRUCache.Add(OnlineUser+userID.String, &CacheUserOnlineEntity{
		Nodes:     nodes,
		CreatedAt: time.Now(),
	})
}

func (r *OnlineUserCacheRepo) InvalidateCache(ttl time.Duration) {
	since := time.Now().Add(-ttl)
	for _, k := range r.LRUCache.Keys() {
		v, ok := r.LRUCache.Peek(k)
		if !ok {
			continue
		}

		cacheEntity, ok := v.(*CacheUserOnlineEntity)
		if !ok {
			continue
		}

		if cacheEntity.CreatedAt.After(since) {
			continue
		}

		r.LRUCache.Remove(k)
	}
}
