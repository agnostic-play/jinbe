package cache_manager

import (
	`context`
	`errors`
)

type FetcherFn func(ctx context.Context, key string) (any, error)

var (
	ErrNotFound = errors.New("cache: not found")
	ErrExpired  = errors.New("cache: expired")
)
