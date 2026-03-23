package cache_manager

import `sync`

// sync.Pool provides a pool of reusable objects to reduce allocations and improve performance.
var objectPool = sync.Pool{New: func() any { return new(cacheItem) }}

func getEntryToObjectPool() *cacheItem {
	return objectPool.Get().(*cacheItem)
}

func putEntryToObjectPool(e *cacheItem) {
	e.reset()
	objectPool.Put(e)
}
