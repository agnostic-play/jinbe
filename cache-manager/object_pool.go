package cache_manager

import "sync"

// objectPool reduces allocations by reusing cacheItem instances.
var objectPool = sync.Pool{New: func() any { return new(cacheItem) }}

func getEntryFromObjectPool() *cacheItem {
	return objectPool.Get().(*cacheItem)
}

func putEntryToObjectPool(e *cacheItem) {
	e.reset()
	objectPool.Put(e)
}
