package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	cache_manager `github.com/agnostic-play/jinbe/cache-manager`
	`github.com/agnostic-play/jinbe/util`
)

type BaseConfig[T any] interface {
	GetConfigID() string
	GetConfigValue() T
}

type ClientConfig[T any] interface {
	Get(ctx context.Context) (T, error)
	EnableCache(cache cache_manager.CacheManager)
}

type clientConfig[T any] struct {
	configID string
	repo     Repositories

	mu           sync.RWMutex
	enableCache  bool
	cacheManager cache_manager.CacheManager

	logFn util.LoggerFn
}

// NewClientConfig creates a ClientConfig backed by repo. Pass an optional logFn to enable logging.
func NewClientConfig[T any](config BaseConfig[T], repo Repositories, logFn ...util.LoggerFn) ClientConfig[T] {
	var fn util.LoggerFn
	if len(logFn) > 0 {
		fn = logFn[0]
	}
	return &clientConfig[T]{
		configID: config.GetConfigID(),
		repo:     repo,
		logFn:    fn,
	}
}

func (c *clientConfig[T]) EnableCache(cache cache_manager.CacheManager) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enableCache = true
	c.cacheManager = cache
}

func (c *clientConfig[T]) Get(ctx context.Context) (T, error) {
	var conf T

	c.mu.RLock()
	enableCache := c.enableCache
	cacheMgr := c.cacheManager
	c.mu.RUnlock()

	if enableCache {
		cacheItem, err := cacheMgr.Get(ctx, c.configID)
		if err == nil && cacheItem != nil {
			if val, ok := cacheItem.(T); ok {
				c.log(ctx, fmt.Sprintf("return config %s from cache", c.configID))
				return val, nil
			}
			c.log(ctx, fmt.Sprintf("cache type assertion failed for config %s", c.configID))
		} else if err != nil {
			c.log(ctx, fmt.Sprintf("cache miss for config %s: %s", c.configID, err.Error()))
		}
	}

	c.log(ctx, fmt.Sprintf("fetching config %s from db", c.configID))
	entity, err := c.repo.GetConfigEntity(ctx, c.configID)
	if err != nil {
		return conf, err
	}

	if entity.RawValue == "" {
		return conf, fmt.Errorf("config %s is empty", c.configID)
	}

	if err := json.Unmarshal([]byte(entity.RawValue), &conf); err != nil {
		return conf, err
	}

	if enableCache {
		c.log(ctx, fmt.Sprintf("saving config %s to cache", c.configID))
		cacheMgr.Set(ctx, c.GetConfigEnt, c.configID, conf)
	}

	return conf, nil
}

func (c *clientConfig[T]) GetConfigEnt(ctx context.Context, configID string) (any, error) {
	var conf T

	entity, err := c.repo.GetConfigEntity(ctx, configID)
	if err != nil {
		return conf, err
	}

	if entity.RawValue == "" {
		return nil, fmt.Errorf("config %s is empty", configID)
	}

	if err := json.Unmarshal([]byte(entity.RawValue), &conf); err != nil {
		return conf, err
	}

	return conf, nil
}

func (c *clientConfig[T]) log(ctx context.Context, msg string) {
	if c.logFn != nil {
		c.logFn(ctx, fmt.Sprintf("clientConfig ID:%s", c.configID), msg)
	}
}
