package cache_manager

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ParseCacheToStruct tries to extract a *T from a cached any value.
// It accepts either a T or *T stored in the any.
func ParseCacheToStruct[T any](cacheVal any) (*T, error) {
	if cacheVal == nil {
		return nil, errors.New("nil pointer to any")
	}

	if ptr, ok := cacheVal.(*T); ok {
		if ptr == nil {
			return nil, errors.New("nil *T value")
		}
		return ptr, nil
	}

	if val, ok := cacheVal.(T); ok {
		return &val, nil
	}

	var tZero T
	return nil, fmt.Errorf("type mismatch: have %T, want %T or *%T", cacheVal, tZero, &tZero)
}

func randomID() string {
	return uuid.New().String()
}
