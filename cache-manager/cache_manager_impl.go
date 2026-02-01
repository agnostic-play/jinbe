package cache_manager

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

func (cm *cacheManager) Get(ctx context.Context, key string) (any, error) {
	cm.log(ctx, fmt.Sprint("Get cache-item ", key))

	now := time.Now().UnixNano()
	item := cm.getCacheItem(key, now)
	if item == nil {
		return nil, ErrNotFound
	}

	if item.expired(now) {
		go cm.Remove(ctx, key)
		return nil, ErrExpired
	}

	cacheVal, ok := item.value()
	if !ok {
		return nil, ErrNotFound
	}

	return cacheVal, nil
}

func (cm *cacheManager) Set(ctx context.Context, fetcherFn FetcherFn, key string, val any) any {
	var item any

	now := time.Now()
	nowNano := now.UnixNano()
	staleAt := nowNano + cm.cfg.StaleInSec.Nanoseconds()
	destroyAt := nowNano + cm.cfg.MaxAgeInSec.Nanoseconds()

	cm.mu.Lock()
	defer cm.mu.Unlock()

	prev, isUpdate := cm.data[key]
	if isUpdate {
		item, _ = prev.value()
	}

	if !isUpdate && cm.cfg.CacheCapacity > 0 && len(cm.data) >= cm.cfg.CacheCapacity {
		cm.log(ctx, "cache capacity exceeded | start to evict")
		cm.evictIfNeededLocked()
	}

	e := getEntryToObjectPool()
	e = &cacheItem{
		fetcherFn: fetcherFn,
		staleAt:   staleAt,
		destroyAt: destroyAt,
	}

	e.setVal(&val)
	e.created.Store(nowNano)
	e.lastAccess.Store(nowNano)

	cm.log(ctx, fmt.Sprintf("save %s to cache", key))
	cm.data[key] = e
	return item
}

func (cm *cacheManager) Remove(ctx context.Context, key string) {
	cm.mu.Lock()
	e, ok := cm.data[key]
	delete(cm.data, key)
	cm.mu.Unlock()
	if ok {
		putEntryToObjectPool(e)
	}
}

func (cm *cacheManager) getCacheItem(key string, nowNano int64) *cacheItem {
	cm.mu.RLock()
	e := cm.data[key]
	cm.mu.RUnlock()
	if e != nil {
		e.lastAccess.Store(nowNano)
	}
	return e
}

func (cm *cacheManager) autoRefresh() {
	tick := time.NewTicker(cm.cfg.RefreshPeriod)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			cm.log(context.Background(), fmt.Sprintf("autoRefresh is start at %s", time.Now().Format(time.RFC3339)))
			cm.checkStaleAndExpired()
		case <-cm.stopChan:
			cm.log(context.Background(), fmt.Sprintf("autoRefresh is stopped at %s", time.Now().Format(time.RFC3339)))
			return
		}
	}
}

func (cm *cacheManager) checkStaleAndExpired() {
	now := time.Now().UnixNano()

	cm.mu.RLock()
	staleKeys := make([]string, 0, 64)
	expiredKeys := make([]string, 0, 32)
	for k, e := range cm.data {
		if e.expired(now) {
			expiredKeys = append(expiredKeys, k)
		} else if e.stale(now) {
			staleKeys = append(staleKeys, k)
		}
	}
	cm.mu.RUnlock()

	if len(expiredKeys) > 0 {
		cm.mu.Lock()
		for _, k := range expiredKeys {
			delete(cm.data, k)
		}
		cm.mu.Unlock()
		cm.log(context.Background(), fmt.Sprintf("cleaned %d expired", len(expiredKeys)))
	}

	for _, k := range staleKeys {
		select {
		case <-cm.stopChan:
			return
		default:
			go cm.refreshOne(k)
		}
	}
}

func (cm *cacheManager) refreshOne(key string) {
	// this will wait till go routine slot available
	select {
	case cm.refreshSem <- struct{}{}:
		defer func() { <-cm.refreshSem }()
	case <-cm.stopChan:
		return
	}

	_, err, _ := cm.refreshGrp.Do(fmt.Sprintf("%s|fetch|%s", cm.id, key), func() (any, error) {
		select {
		case <-cm.stopChan:
			return nil, context.Canceled
		default:
		}

		ctx := context.Background()
		now := time.Now().UnixNano()
		e := cm.getCacheItem(key, now)
		if e == nil || e.expired(now) {
			cm.Remove(ctx, key)
			cm.log(ctx, key+" expired, removed")
			return nil, nil
		}

		cm.log(ctx, fmt.Sprintf("%s stale, refreshing", key))

		val, err := cm.fetchWithRetry(ctx, e.fetcherFn, key)
		if err != nil {
			cm.log(ctx, key+" refresh failed: "+err.Error())
			return nil, err
		}

		cm.logWithObject(ctx, fmt.Sprintf("%s save to cache", key), val)

		cm.mu.Lock()
		if cur := cm.data[key]; cur != nil {
			cur.setVal(&val)
		}
		cm.mu.Unlock()

		cm.log(ctx, fmt.Sprintf("%s refreshed", key))

		return nil, nil
	})
	_ = err
}

func (cm *cacheManager) fetchWithRetry(ctx context.Context, fn FetcherFn, key string) (any, error) {
	backoff := cm.cfg.RetryBackoff
	maxBackoff := cm.cfg.MaxRetryBackoff

	for attempt := 0; attempt <= cm.cfg.MaxRetries; attempt++ {
		cm.log(ctx, fmt.Sprintf("fetch data %s attempt #%d", key, attempt))
		if attempt > 0 {
			// apply delay on second attempts
			backoff = time.Duration(float64(backoff) * (1.5 + 0.5*math.Min(1, float64(attempt)/3)))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

		}

		fetchCtx, cancel := context.WithTimeout(ctx, cm.cfg.FetcherTimeout)
		val, err := fn(fetchCtx, key)
		cancel()
		if err == nil {
			return val, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("fetch failed after %d retries", cm.cfg.MaxRetries)
}

func (cm *cacheManager) log(ctx context.Context, msg string) {
	if cm.loggerFn != nil {
		cm.loggerFn(ctx, fmt.Sprintf("cacheManager ID:%s", cm.id), msg)
	}
}

func (cm *cacheManager) logWithObject(ctx context.Context, msg string, object any) {
	if cm.loggerFn != nil {
		cm.loggerFn(ctx, fmt.Sprintf("cacheManager ID:%s", cm.id), msg, object)
	}
}
