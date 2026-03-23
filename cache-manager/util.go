package cache_manager

import (
	`errors`
	`fmt`
	`math/rand/v2`
)

// ParseCacheToStruct tries to extract a *T from a *any.
// It accepts either a T or *T stored in the any.
// Returns an error for nils or mismatched types.
func ParseCacheToStruct[T any](cacheVal any) (*T, error) {
	if cacheVal == nil {
		return nil, errors.New("nil pointer to any")
	}

	v := cacheVal
	// Case 1: the underlying value is already *T
	if ptr, ok := v.(*T); ok {
		if ptr == nil {
			return nil, errors.New("nil *T value")
		}
		return ptr, nil
	}

	// Case 2: the underlying value is T (by value)
	if val, ok := v.(T); ok {
		return &val, nil
	}

	// Helpful error message with the concrete type we actually saw.
	var tZero T
	return nil, fmt.Errorf("type mismatch: have %T, want %T or *%T", v, tZero, &tZero)
}

func randomID() string {
	// format as 7-digit string with leading zeros if needed
	return fmt.Sprintf("%07d", rand.IntN(10000000))
}
