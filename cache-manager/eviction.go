package cache_manager

import (
	"math"
	"sort"
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

	// Collect eviction candidates
	candidates := cm.collectEvictionCandidates(plan.triggerSize)
	if len(candidates) < plan.triggerSize {
		return
	}

	// Determine exact number of items to evict
	evictCount := cm.boundEvictCount(plan, len(candidates))

	// Perform eviction
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

	// Base eviction policy
	var evictCount int
	switch {
	case capacity <= 10:
		evictCount = 1
	case capacity <= 50:
		evictCount = int(math.Max(1, math.Min(2, float64(capacity)*0.1)))
	default:
		evictCount = int(math.Max(1, float64(size)*0.05))
	}

	return &evictionPlan{
		totalItems:  size,
		capacity:    capacity,
		triggerSize: triggerSize,
		evictCount:  evictCount,
	}
}

// collectEvictionCandidates returns non-expired entries for eviction consideration.
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

	// Sort by last access time → least recently used first.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].lastAccess < candidates[j].lastAccess
	})

	return candidates
}

// boundEvictCount ensures eviction count stays within safe limits.
func (cm *cacheManager) boundEvictCount(plan *evictionPlan, candidateCount int) int {
	count := plan.evictCount

	if count < 1 {
		count = 1
	}
	if count > candidateCount {
		count = candidateCount
	}

	// For larger caches, enforce min/max bounds
	if plan.capacity > 50 {
		minCount := int(math.Max(1, math.Min(10, float64(plan.capacity)*0.02)))
		if count < minCount {
			count = minCount
			if count > candidateCount {
				count = candidateCount
			}
		}
		maxCount := int(math.Max(1, float64(candidateCount)*0.10))
		if count > maxCount {
			count = maxCount
		}
	}

	return count
}

// removeLeastRecentlyUsed deletes the first N candidates.
func (cm *cacheManager) removeLeastRecentlyUsed(candidates []entryMeta, n int) {
	for i := 0; i < n && i < len(candidates); i++ {
		if entry, ok := cm.data[candidates[i].key]; ok {
			delete(cm.data, candidates[i].key)
			objectPool.Put(entry)
		}
	}
}
