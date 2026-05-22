package cache_manager

import (
	"cmp"
	"math"
	"slices"
	"time"
)

// entryMeta holds metadata used for eviction decisions.
type entryMeta struct {
	key        string
	lastAccess int64
}

// evictionPlan defines when and how many items to evict.
type evictionPlan struct {
	totalItems  int
	capacity    int
	triggerSize int
	evictCount  int
}

// evictIfNeededLocked removes least recently used, non-expired items.
// Caller must hold the lock.
func (cm *cacheManager) evictIfNeededLocked() {
	plan := cm.computeEvictionPlan()
	if plan == nil {
		return
	}

	candidates := cm.collectEvictionCandidates(plan.triggerSize)
	if len(candidates) == 0 {
		// All entries are expired; autoRefresh will clean them shortly.
		return
	}

	evictCount := cm.boundEvictCount(plan, len(candidates))
	cm.removeLeastRecentlyUsed(candidates, evictCount)
}

// computeEvictionPlan decides whether eviction is needed and how many items to evict.
func (cm *cacheManager) computeEvictionPlan() *evictionPlan {
	size := len(cm.data)
	if size == 0 {
		return nil
	}

	capacity := cm.cfg.CacheCapacity
	if capacity <= 0 {
		capacity = 1 // defensive fallback
	}

	// Evict when reaching capacity (small caches) or 90% full (larger caches).
	triggerSize := capacity
	if capacity >= 50 {
		triggerSize = int(math.Ceil(float64(capacity) * 0.9))
	}

	if size < triggerSize {
		return nil
	}

	var evictCount int
	switch {
	case capacity <= 10:
		evictCount = 1
	case capacity <= 50:
		evictCount = max(1, min(2, int(float64(capacity)*0.1)))
	default:
		evictCount = max(1, int(float64(size)*0.05))
	}

	return &evictionPlan{
		totalItems:  size,
		capacity:    capacity,
		triggerSize: triggerSize,
		evictCount:  evictCount,
	}
}

// collectEvictionCandidates returns non-expired entries sorted by last access (LRU first).
func (cm *cacheManager) collectEvictionCandidates(triggerSize int) []entryMeta {
	now := time.Now().UnixNano()
	candidates := make([]entryMeta, 0, triggerSize)

	for key, entry := range cm.data {
		if entry.expired(now) {
			continue
		}
		candidates = append(candidates, entryMeta{
			key:        key,
			lastAccess: entry.lastAccess.Load(),
		})
	}

	slices.SortFunc(candidates, func(a, b entryMeta) int {
		return cmp.Compare(a.lastAccess, b.lastAccess)
	})

	return candidates
}

// boundEvictCount ensures eviction count stays within safe limits.
func (cm *cacheManager) boundEvictCount(plan *evictionPlan, candidateCount int) int {
	count := max(1, min(plan.evictCount, candidateCount))

	if plan.capacity > 50 {
		minCount := max(1, min(10, int(float64(plan.capacity)*0.02)))
		count = min(max(count, minCount), candidateCount)
		maxCount := max(1, int(float64(candidateCount)*0.10))
		count = min(count, maxCount)
	}

	return count
}

// removeLeastRecentlyUsed deletes the first n candidates (already sorted LRU first).
func (cm *cacheManager) removeLeastRecentlyUsed(candidates []entryMeta, n int) {
	for i := range min(n, len(candidates)) {
		if entry, ok := cm.data[candidates[i].key]; ok {
			delete(cm.data, candidates[i].key)
			putEntryToObjectPool(entry)
		}
	}
}
