package cache_manager

import (
	"context"
	`fmt`
	"sync"

	"golang.org/x/sync/singleflight"

	`github.com/agnostic-play/jinbe/util`
)

type CacheManager interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, fetcherFn FetcherFn, key string, val any) any
	Remove(ctx context.Context, key string)
	GetAvailableCacheSlot() int

	Stop()
}

type cacheManager struct {
	id  string
	cfg Config

	mu   sync.RWMutex
	data map[string]*cacheItem

	refreshGrp singleflight.Group
	refreshSem chan struct{}

	loggerFn util.LoggerFn

	stopChan chan struct{}
	wgDone   sync.WaitGroup
}

func NewCacheManager(cfg Config, fn util.LoggerFn) CacheManager {
	cfg = cfg.WithDefaults()
	cm := &cacheManager{
		id:         randomID(),
		cfg:        cfg,
		loggerFn:   fn,
		data:       make(map[string]*cacheItem),
		refreshSem: make(chan struct{}, cfg.MaxRefreshWorkers),
		stopChan:   make(chan struct{}),
	}

	cm.log(context.Background(), fmt.Sprintf("config: %+v", cfg))

	cm.wgDone.Add(1)
	go func() {
		defer cm.wgDone.Done()
		cm.autoRefresh()
	}()

	return cm
}

func (cm *cacheManager) Stop() {
	close(cm.stopChan)
	cm.wgDone.Wait()
}

func (cm *cacheManager) GetAvailableCacheSlot() int {
	return cm.cfg.CacheCapacity - len(cm.data)
}
