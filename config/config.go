package config

import (
	`context`
	`encoding/json`
	`fmt`

	cache_manager `berlin.allobank.local/common/gommon/cache-manager`
	`berlin.allobank.local/common/gommon/util`
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

	enableCache  bool
	cacheManager cache_manager.CacheManager

	logFn util.LoggerFn
}

func NewClientConfig[T any](config BaseConfig[T], repo Repositories) ClientConfig[T] {
	return &clientConfig[T]{
		configID: config.GetConfigID(),
		repo:     repo,
	}
}

func (c *clientConfig[T]) EnableCache(cache cache_manager.CacheManager) {
	c.enableCache = true
	c.cacheManager = cache
}

func (c *clientConfig[T]) Get(ctx context.Context) (T, error) {
	var conf T
	if c.enableCache {
		cacheItem, err := c.cacheManager.Get(ctx, c.configID)
		if err == nil && cacheItem != nil {
			if val, ok := cacheItem.(T); ok {
				c.log(ctx, fmt.Sprintf("return config %s from cache", c.configID))
				return val, nil
			}
		} else {
			c.log(ctx, fmt.Sprintf("fail to get config %s from cache", c.configID))
			if err != nil {
				c.log(ctx, fmt.Sprintf("fail to get config %s from cache %s", c.configID, err.Error()))
			}
		}
	}

	c.log(ctx, fmt.Sprintf("try to fetch config %s from db", c.configID))
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

	if c.enableCache {
		c.log(ctx, fmt.Sprintf("save config %s to cache", c.configID))
		c.cacheManager.Set(ctx, c.GetConfigEnt, c.configID, conf)
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

func (c *clientConfig[T]) log(ctx context.Context, phase string) {
	if c.logFn != nil {
		c.logFn(ctx, fmt.Sprintf("clientConfig ID:%s", c.configID), phase)
	}
}
