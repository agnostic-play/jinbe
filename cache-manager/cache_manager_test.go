package cache_manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock fetcher function that simulates data fetching
type MockFetcher struct {
	mu         sync.RWMutex
	data       map[string]interface{}
	fetchCount int32
	shouldFail bool
	fetchDelay time.Duration
}

func NewMockFetcher() *MockFetcher {
	return &MockFetcher{
		data: make(map[string]interface{}),
	}
}

func (mf *MockFetcher) SetData(key string, value interface{}) {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	mf.data[key] = value
}

func (mf *MockFetcher) UpdateData(key string, value interface{}) {
	mf.SetData(key, value)
}

func (mf *MockFetcher) SetShouldFail(fail bool) {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	mf.shouldFail = fail
}

func (mf *MockFetcher) SetFetchDelay(delay time.Duration) {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	mf.fetchDelay = delay
}

func (mf *MockFetcher) GetFetchCount() int32 {
	return atomic.LoadInt32(&mf.fetchCount)
}

func (mf *MockFetcher) ResetFetchCount() {
	atomic.StoreInt32(&mf.fetchCount, 0)
}

func (mf *MockFetcher) Fetch(ctx context.Context, key string) (interface{}, error) {
	atomic.AddInt32(&mf.fetchCount, 1)

	mf.mu.RLock()
	delay := mf.fetchDelay
	shouldFail := mf.shouldFail
	value, exists := mf.data[key]
	mf.mu.RUnlock()

	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if shouldFail {
		return nil, fmt.Errorf("mock fetch failed for key: %s", key)
	}

	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// Test comprehensive cache flow with Set/Get operations
func TestCacheManager_ComprehensiveFlow(t *testing.T) {
	// Setup cache with short durations for testing
	cfg := Config{
		CacheCapacity:     100,
		StaleInSec:        2 * time.Second,  // Items become stale after 2 seconds
		MaxAgeInSec:       10 * time.Second, // Items expire after 10 seconds
		RefreshPeriod:     1 * time.Second,  // Check for stale items every 1 second
		MaxRefreshWorkers: 5,
		MaxRetries:        2,
		FetcherTimeout:    5 * time.Second,
		RetryBackoff:      100 * time.Millisecond,
		MaxRetryBackoff:   500 * time.Millisecond,
	}

	mockFetcher := NewMockFetcher()
	cm := NewCacheManager(cfg, func(ctx context.Context, identifier, msg string) {
		t.Logf("%s", msg)
	})
	defer cm.Stop()

	ctx := context.Background()

	t.Run("1. Initial Cache Population", func(t *testing.T) {
		// Set initial data in mock fetcher
		initialData := "initial-value-1"
		mockFetcher.SetData("key1", initialData)

		// Set data in cache
		oldVal := cm.Set(ctx, mockFetcher.Fetch, "key1", initialData)
		assert.Nil(t, oldVal, "First set should return nil")

		// Verify data can be retrieved
		val, err := cm.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, initialData, val, "Retrieved value should match initial data")
	})

	t.Run("2. Get Data From Cache (Cache Hit)", func(t *testing.T) {
		// Reset fetch counter to verify cache hit
		mockFetcher.ResetFetchCount()

		// Get data multiple times - should not trigger fetch
		for i := 0; i < 5; i++ {
			val, err := cm.Get(ctx, "key1")
			require.NoError(t, err)
			assert.Equal(t, "initial-value-1", val)
		}

		// Verify no additional fetches occurred
		assert.Equal(t, int32(0), mockFetcher.GetFetchCount(), "Cache hits should not trigger fetches")
	})

	t.Run("3. Update Data and Auto Refetch", func(t *testing.T) {
		// Update data in mock fetcher (simulating external data change)
		updatedData := "updated-value-1"
		mockFetcher.UpdateData("key1", updatedData)

		// Wait for data to become stale (2 seconds)
		t.Log("Waiting for data to become stale...")
		time.Sleep(3 * time.Second)

		// Reset fetch counter
		initialFetchCount := mockFetcher.GetFetchCount()

		// Wait for auto refresh to kick in
		time.Sleep(2 * time.Second)

		// Verify that auto refresh happened
		assert.Greater(t, mockFetcher.GetFetchCount(), initialFetchCount, "Auto refresh should have triggered")

		// Get data - should return updated value
		val, err := cm.Get(ctx, "key1")
		require.NoError(t, err)
		assert.Equal(t, updatedData, val, "Should get updated data after auto refresh")
	})

	t.Run("4. Compare Data Equality", func(t *testing.T) {
		// Get current cached value
		cachedVal, err := cm.Get(ctx, "key1")
		require.NoError(t, err)

		// Get expected value from mock fetcher
		expectedVal, exists := mockFetcher.data["key1"]
		require.True(t, exists, "Key should exist in mock fetcher")

		// Compare values
		assert.Equal(t, expectedVal, cachedVal, "Cached value should equal latest data from fetcher")
	})
}

// Test cache with 100 entries, eviction, and auto removal
func TestCacheManager_LargeDatasetEviction(t *testing.T) {
	// Configure cache with small capacity to test eviction
	cfg := Config{
		CacheCapacity:     50, // Smaller than our test data to force eviction
		StaleInSec:        3 * time.Second,
		MaxAgeInSec:       8 * time.Second, // Shorter max age for testing expiration
		RefreshPeriod:     1 * time.Second,
		MaxRefreshWorkers: 10,
		MaxRetries:        1,
		FetcherTimeout:    2 * time.Second,
		RetryBackoff:      50 * time.Millisecond,
		MaxRetryBackoff:   200 * time.Millisecond,
	}

	mockFetcher := NewMockFetcher()
	cm := NewCacheManager(cfg, func(ctx context.Context, identifer, msg string) {
		t.Logf(msg)
	})
	defer cm.Stop()

	ctx := context.Background()

	t.Run("5. Populate Cache with 100 Items", func(t *testing.T) {
		// Add 100 items to cache
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key_%03d", i)
			value := fmt.Sprintf("value_%03d", i)

			// Set data in mock fetcher
			mockFetcher.SetData(key, value)

			// Set data in cache
			cm.Set(ctx, mockFetcher.Fetch, key, value)
		}

		// Verify some items are in cache (not all due to capacity limit)
		cacheSize := cm.GetAvailableCacheSlot()
		t.Logf("Cache size after adding 100 items: %d", cacheSize)
		assert.LessOrEqual(t, cacheSize, cfg.CacheCapacity, "Cache size should not exceed capacity")
	})

	t.Run("6. Test Eviction Mechanism", func(t *testing.T) {
		// Access some items to update their last access time
		accessedKeys := []string{"key_090", "key_091", "key_092", "key_093", "key_094"}

		for _, key := range accessedKeys {
			val, err := cm.Get(ctx, key)
			if err == nil {
				t.Logf("Accessed key %s with value %v", key, val)
			}
		}

		// Add more items to trigger eviction
		for i := 100; i < 120; i++ {
			key := fmt.Sprintf("key_%03d", i)
			value := fmt.Sprintf("value_%03d", i)

			mockFetcher.SetData(key, value)
			cm.Set(ctx, mockFetcher.Fetch, key, value)
		}

		// Verify cache size is still within limits
		cacheSize := cm.GetAvailableCacheSlot()
		t.Logf("Cache size after adding more items: %d", cacheSize)
		assert.LessOrEqual(t, cacheSize, cfg.CacheCapacity, "Cache size should remain within capacity")

		// Verify some recently accessed items are still in cache
		stillCachedCount := 0
		for _, key := range accessedKeys {
			if _, err := cm.Get(ctx, key); err == nil {
				stillCachedCount++
			}
		}
		t.Logf("Recently accessed items still in cache: %d/%d", stillCachedCount, len(accessedKeys))
	})

	t.Run("7. Test Auto Removal of Expired Items", func(t *testing.T) {
		// Wait for items to expire
		t.Log("Waiting for items to expire...")
		time.Sleep(cfg.MaxAgeInSec + 2*time.Second)

		// Wait for cleanup to occur
		time.Sleep(2 * time.Second)

		// Most items should be expired and removed
		cacheSize := cm.GetAvailableCacheSlot()
		t.Logf("Cache size after expiration cleanup: %d", cacheSize)

		// Try to get some items - most should return ErrExpired or ErrNotFound
		expiredCount := 0
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("key_%03d", i)
			_, err := cm.Get(ctx, key)
			if err != nil {
				expiredCount++
			}
		}
		t.Logf("Expired/removed items found: %d/20", expiredCount)
		assert.Greater(t, expiredCount, 10, "Most items should be expired/removed")
	})
}

// Test edge cases and error conditions
func TestCacheManager_EdgeCases(t *testing.T) {
	cfg := Config{
		CacheCapacity:     10,
		StaleInSec:        1 * time.Second,
		MaxAgeInSec:       3 * time.Second,
		RefreshPeriod:     500 * time.Millisecond,
		MaxRefreshWorkers: 2,
		MaxRetries:        1,
		FetcherTimeout:    1 * time.Second,
		RetryBackoff:      100 * time.Millisecond,
		MaxRetryBackoff:   300 * time.Millisecond,
	}

	mockFetcher := NewMockFetcher()
	cm := NewCacheManager(cfg, func(ctx context.Context, identifier, msg string) {
		t.Logf(msg)
	})
	defer cm.Stop()

	ctx := context.Background()

	t.Run("8. Test Fetch Failure and Recovery", func(t *testing.T) {
		key := "failing_key"
		value := "test_value"

		// Initially set fetcher to fail
		mockFetcher.SetShouldFail(true)
		mockFetcher.SetData(key, value)

		// Set initial value in cache
		cm.Set(ctx, mockFetcher.Fetch, key, value)

		// Get initial value (should work)
		val, err := cm.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, val)

		// Wait for item to become stale
		time.Sleep(2 * time.Second)

		// Allow fetcher to work again
		mockFetcher.SetShouldFail(false)
		updatedValue := "updated_test_value"
		mockFetcher.UpdateData(key, updatedValue)

		// Wait for refresh
		time.Sleep(2 * time.Second)

		// Should eventually get updated value
		val, err = cm.Get(ctx, key)
		if err == nil {
			t.Logf("Got value after recovery: %v", val)
		}
	})

	t.Run("9. Test Non-existent Key", func(t *testing.T) {
		// Try to get non-existent key
		_, err := cm.Get(ctx, "non_existent_key")
		assert.Error(t, err, "Should return error for non-existent key")
		assert.Equal(t, ErrNotFound, err, "Should return ErrNotFound")
	})

	t.Run("10. Test Cache Update (Replace Existing)", func(t *testing.T) {
		key := "update_key"
		initialValue := "initial"
		updatedValue := "updated"

		mockFetcher.SetData(key, initialValue)

		// Set initial value
		oldVal := cm.Set(ctx, mockFetcher.Fetch, key, initialValue)
		assert.Nil(t, oldVal)

		// Update with new value
		oldVal = cm.Set(ctx, mockFetcher.Fetch, key, updatedValue)
		assert.Equal(t, initialValue, oldVal, "Should return previous value")

		// Get updated value
		val, err := cm.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, updatedValue, val, "Should return updated value")
	})
}

// Helper function to create a simple cache manager for testing
func createTestCacheManager(t *testing.T, cfg Config) (*cacheManager, *MockFetcher) {
	mockFetcher := NewMockFetcher()
	cm := NewCacheManager(cfg, func(ctx context.Context, identifier, msg string) {
		t.Logf(msg)
	})
	return cm.(*cacheManager), mockFetcher
}

// Benchmark test for performance
func BenchmarkCacheManager_SetGet(b *testing.B) {
	cfg := DefaultConfig()
	cfg.CacheCapacity = 1000
	cfg.RefreshPeriod = -1 // Disable auto refresh for benchmark

	mockFetcher := NewMockFetcher()
	cm := NewCacheManager(cfg, nil)
	defer cm.Stop()

	ctx := context.Background()

	// Pre-populate some data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)
		mockFetcher.SetData(key, value)
		cm.Set(ctx, mockFetcher.Fetch, key, value)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i%100)
			_, _ = cm.Get(ctx, key)
			i++
		}
	})
}
