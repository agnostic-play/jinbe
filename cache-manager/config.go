package cache_manager

import `time`

const (
	MaxCacheCapacity        = 1_000 // safety cap against runaway memory/concurrency
	DefaultCacheCapacity    = 1
	DefaultStaleInSec       = 5 * time.Minute
	DefaultMaxAgeInSec      = 10 * time.Hour
	DefaultRefreshPeriod    = 5 * time.Minute
	DefaultMaxRefreshWorker = 20
	DefaultMaxRetries       = 1
	DefaultFetcherTimeout   = 2 * time.Second
	DefaultRetryBackoff     = 500 * time.Millisecond
	DefaultMaxRetryBackoff  = 2 * time.Second
)

// Config controls caching behavior.
//
// Zero/negative semantics:
//   - CacheCapacity <= 0         : unbounded cache size
//   - StaleInSec <= 0            : entries never considered stale (no refresh trigger)
//   - MaxAgeInSec <= 0           : entries never expire due to age
//   - RefreshPeriod <= 0         : disable periodic/background refresh
//   - MaxRetries <= 0            : no retries (single attempt only)
//   - FetcherTimeout <= 0        : no timeout applied to the fetcher
//   - RetryBackoff <= 0          : no backoff delay between retries
//
// Notes:
//   - MaxRetries is the number of *additional* attempts after the initial try.
//   - Durations are time.Duration; names ending with InSec are historical.
//   - CacheCapacity is limited to MaxCacheCapacity to prevent goroutine floods when many entries are stale.
type Config struct {
	CacheCapacity     int           // Max number of items kept in memory; <= 0 means unlimited.
	StaleInSec        time.Duration // Age after which an item is considered stale; <= 0 means never stale.
	MaxAgeInSec       time.Duration // Hard TTL; <= 0 means never evict by age.
	RefreshPeriod     time.Duration // Cadence for background refresh; <= 0 disables periodic refresh.
	MaxRefreshWorkers int           // Upper bound for concurrent refresh workers; <= 0 uses default.
	MaxRetries        int           // Number of retries after the initial attempt; <= 0 means no retries.
	FetcherTimeout    time.Duration // Per-fetch timeout; <= 0 means no timeout.
	RetryBackoff      time.Duration // Delay between retry attempts; <= 0 means no backoff.
	MaxRetryBackoff   time.Duration // Delay between retry attempts; <= 0 means no backoff.
}

// DefaultConfig returns a fully-populated Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		CacheCapacity:     DefaultCacheCapacity,
		StaleInSec:        DefaultStaleInSec,
		MaxAgeInSec:       DefaultMaxAgeInSec,
		RefreshPeriod:     DefaultRefreshPeriod,
		MaxRefreshWorkers: DefaultMaxRefreshWorker,
		MaxRetries:        DefaultMaxRetries,
		FetcherTimeout:    DefaultFetcherTimeout,
		RetryBackoff:      DefaultRetryBackoff,
		MaxRetryBackoff:   DefaultMaxRetryBackoff,
	}
}

// WithDefaults applies default values where fields are zero/negative,
// enforces safety caps, and normalizes interdependent fields.
//
// Behavior:
//   - If StaleInSec <= 0, periodic refresh is disabled.
//   - If both StaleInSec and MaxAgeInSec are positive and stale > maxAge,
//     stale is clamped to maxAge/2 (MUST-fix guard).
//   - CacheCapacity is capped at MaxCacheCapacity (unless <= 0 which means unbounded).
func (c Config) WithDefaults() Config {
	def := DefaultConfig()

	// Simple field defaults
	if c.CacheCapacity == 0 {
		c.CacheCapacity = def.CacheCapacity
	}
	if c.StaleInSec <= 0 {
		c.StaleInSec = 0 // explicitly keep "never stale" semantic
	}
	if c.MaxAgeInSec == 0 {
		c.MaxAgeInSec = def.MaxAgeInSec
	}
	if c.RefreshPeriod == 0 {
		c.RefreshPeriod = def.RefreshPeriod
	}
	if c.MaxRefreshWorkers <= 0 {
		c.MaxRefreshWorkers = def.MaxRefreshWorkers
	}
	if c.FetcherTimeout <= 0 {
		c.FetcherTimeout = def.FetcherTimeout
	}
	if c.MaxRetries <= 0 {
		c.MaxRetries = def.MaxRetries
	}
	if c.RetryBackoff <= 0 {
		c.RetryBackoff = def.RetryBackoff
	}

	// If entries never become stale, disable periodic refresh explicitly.
	if c.StaleInSec <= 0 {
		c.RefreshPeriod = -1
	}

	// MUST-fix: stale should not exceed max age (when both are positive).
	if c.StaleInSec > 0 && c.MaxAgeInSec > 0 && c.StaleInSec > c.MaxAgeInSec {
		c.StaleInSec = c.MaxAgeInSec / 2
	}

	// Cap capacity (unless unbounded by design).
	if c.CacheCapacity > MaxCacheCapacity {
		c.CacheCapacity = MaxCacheCapacity
	}

	return c
}
