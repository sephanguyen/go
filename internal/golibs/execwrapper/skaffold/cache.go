package skaffoldwrapper

import "sync"

type renderResult struct {
	res []interface{}
	err error
}

type renderCacheResult struct {
	cache map[Command]renderResult
	mu    sync.Mutex
}

func (r *renderCacheResult) Get(c Command) (renderResult, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res, exists := r.cache[c]
	return res, exists
}

func (r *renderCacheResult) Set(c Command, rs renderResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[c] = rs
}

var globalRenderCache = renderCacheResult{
	cache: map[Command]renderResult{},
}

var globalRenderCacheV2 = renderCacheResult{
	cache: map[Command]renderResult{},
}
