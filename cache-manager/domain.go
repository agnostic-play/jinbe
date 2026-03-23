package cache_manager

import (
	`context`
	`fmt`
	`sync/atomic`
)

type cacheItem struct {
	val        atomic.Pointer[any]
	fetcherFn  FetcherFn
	created    atomic.Int64
	staleAt    int64 // unix-nano
	destroyAt  int64 // unix-nano
	lastAccess atomic.Int64
}

func (e *cacheItem) expired(now int64) bool {
	return e.destroyAt > 0 && now > e.destroyAt
}

func (e *cacheItem) stale(now int64) bool {
	return e.staleAt > 0 && now > e.staleAt
}

func (e *cacheItem) setFetcher(fn FetcherFn) {
	if fn == nil {
		e.fetcherFn = func(ctx context.Context, key string) (any, error) {
			return nil, fmt.Errorf(`no fetcher for key "%s"`, key)
		}
	}

	e.fetcherFn = fn
}

func (e *cacheItem) setVal(v *any) {
	e.val.Store(v)
}

func (e *cacheItem) value() (any, bool) {
	p := e.val.Load()
	if p == nil {
		return nil, false
	}
	return *p, true
}

func (e *cacheItem) reset() {
	e.fetcherFn = nil
	e.staleAt = 0
	e.destroyAt = 0
	e.created.Store(0)
	e.lastAccess.Store(0)
	e.val.Store(nil)
}
