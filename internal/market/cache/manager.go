package cache

import (
	"context"
	"log/slog"
	"ntx/internal/market"
	"sync"
	"time"
)

// CacheManager provides sophisticated caching with TTL, invalidation, and statistics
// Implements multi-tier caching strategy for optimal performance during trading hours
type CacheManager struct {
	logger *slog.Logger
	config *Config
	mu     sync.RWMutex

	// Primary cache: Hot data for active trading
	hotCache map[string]*CacheEntry

	// Secondary cache: Warm data for recent lookups
	warmCache map[string]*CacheEntry

	// Statistics tracking
	stats *CacheStats

	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// Config holds cache configuration
type Config struct {
	// Hot cache TTL during trading hours (default: 30 seconds)
	HotCacheTTL time.Duration

	// Warm cache TTL for non-trading hours (default: 5 minutes)
	WarmCacheTTL time.Duration

	// Cold cache TTL for historical data (default: 1 hour)
	ColdCacheTTL time.Duration

	// Maximum entries in hot cache (default: 1000)
	MaxHotEntries int

	// Maximum entries in warm cache (default: 5000)
	MaxWarmEntries int

	// Cleanup interval (default: 1 minute)
	CleanupInterval time.Duration

	// Enable cache statistics (default: true)
	EnableStats bool

	// Enable automatic cache warming (default: true)
	EnableWarming bool
}

// CacheEntry represents a cached market data entry with metadata
type CacheEntry struct {
	Data        *market.ScrapedData `json:"data"`
	CachedAt    time.Time           `json:"cached_at"`
	AccessCount int64               `json:"access_count"`
	LastAccess  time.Time           `json:"last_access"`
	TTL         time.Duration       `json:"ttl"`
	Tier        CacheTier           `json:"tier"`
}

// CacheTier represents the caching tier
type CacheTier string

const (
	TierHot  CacheTier = "hot"  // Active trading data
	TierWarm CacheTier = "warm" // Recently accessed data
	TierCold CacheTier = "cold" // Historical/infrequent data

	// Cache performance constants
	EstimatedEntrySize    = 200 // Approximate bytes per cache entry
	DefaultPromotionCount = 5   // Access count threshold for cache promotion
)

// CacheStats provides cache performance metrics
type CacheStats struct {
	mu sync.RWMutex

	// Hit/Miss statistics
	HotHits  int64 `json:"hot_hits"`
	WarmHits int64 `json:"warm_hits"`
	Misses   int64 `json:"misses"`

	// Entry counts
	HotEntries  int `json:"hot_entries"`
	WarmEntries int `json:"warm_entries"`

	// Performance metrics
	AvgLookupTime time.Duration `json:"avg_lookup_time"`
	EvictionCount int64         `json:"eviction_count"`

	// Memory usage estimation (in bytes)
	EstimatedSize int64 `json:"estimated_size"`

	// Operational metrics
	StartTime    time.Time `json:"start_time"`
	LastCleanup  time.Time `json:"last_cleanup"`
	CleanupCount int64     `json:"cleanup_count"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *Config {
	return &Config{
		HotCacheTTL:     30 * time.Second,
		WarmCacheTTL:    5 * time.Minute,
		ColdCacheTTL:    1 * time.Hour,
		MaxHotEntries:   1000,
		MaxWarmEntries:  5000,
		CleanupInterval: 1 * time.Minute,
		EnableStats:     true,
		EnableWarming:   true,
	}
}

// NewCacheManager creates a new cache manager with multi-tier caching
func NewCacheManager(logger *slog.Logger, config *Config) *CacheManager {
	if config == nil {
		config = DefaultCacheConfig()
	}

	cm := &CacheManager{
		logger:    logger,
		config:    config,
		hotCache:  make(map[string]*CacheEntry),
		warmCache: make(map[string]*CacheEntry),
		stats: &CacheStats{
			StartTime: time.Now(),
		},
		stopChan: make(chan struct{}),
	}

	// Start cleanup routine
	cm.cleanupTicker = time.NewTicker(config.CleanupInterval)
	go cm.cleanupRoutine()

	logger.Info("Cache manager initialized",
		"hot_ttl", config.HotCacheTTL,
		"warm_ttl", config.WarmCacheTTL,
		"max_hot_entries", config.MaxHotEntries)

	return cm
}

// Get retrieves data from cache with multi-tier lookup
func (cm *CacheManager) Get(symbol string, isMarketOpen bool) (*market.ScrapedData, bool) {
	startTime := time.Now()
	defer func() {
		if cm.config.EnableStats {
			cm.updateLookupTime(time.Since(startTime))
		}
	}()

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Try hot cache first (during trading hours)
	if entry, exists := cm.hotCache[symbol]; exists {
		if cm.isEntryValid(entry, isMarketOpen) {
			entry.AccessCount++
			entry.LastAccess = time.Now()
			cm.recordHit(TierHot)
			return entry.Data, true
		}
	}

	// Try warm cache
	if entry, exists := cm.warmCache[symbol]; exists {
		if cm.isEntryValid(entry, isMarketOpen) {
			entry.AccessCount++
			entry.LastAccess = time.Now()
			cm.recordHit(TierWarm)

			// Promote to hot cache if frequently accessed during trading
			if isMarketOpen && entry.AccessCount > DefaultPromotionCount {
				cm.promoteToHot(symbol, entry)
			}

			return entry.Data, true
		}
	}

	// Cache miss
	cm.recordMiss()
	return nil, false
}

// Set stores data in appropriate cache tier based on market conditions
func (cm *CacheManager) Set(symbol string, data *market.ScrapedData, isMarketOpen bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var tier CacheTier
	var ttl time.Duration

	// Determine cache tier and TTL based on market status
	if isMarketOpen {
		tier = TierHot
		ttl = cm.config.HotCacheTTL
	} else {
		tier = TierWarm
		ttl = cm.config.WarmCacheTTL
	}

	entry := &CacheEntry{
		Data:        data,
		CachedAt:    now,
		AccessCount: 1,
		LastAccess:  now,
		TTL:         ttl,
		Tier:        tier,
	}

	if tier == TierHot {
		// Check capacity and evict if necessary
		if len(cm.hotCache) >= cm.config.MaxHotEntries {
			cm.evictLRU(cm.hotCache)
		}
		cm.hotCache[symbol] = entry
	} else {
		// Check capacity and evict if necessary
		if len(cm.warmCache) >= cm.config.MaxWarmEntries {
			cm.evictLRU(cm.warmCache)
		}
		cm.warmCache[symbol] = entry
	}

	cm.updateStats()
}

// SetBatch efficiently stores multiple entries
func (cm *CacheManager) SetBatch(data map[string]*market.ScrapedData, isMarketOpen bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for symbol, marketData := range data {
		now := time.Now()
		var tier CacheTier
		var ttl time.Duration

		if isMarketOpen {
			tier = TierHot
			ttl = cm.config.HotCacheTTL
		} else {
			tier = TierWarm
			ttl = cm.config.WarmCacheTTL
		}

		entry := &CacheEntry{
			Data:        marketData,
			CachedAt:    now,
			AccessCount: 1,
			LastAccess:  now,
			TTL:         ttl,
			Tier:        tier,
		}

		if tier == TierHot {
			if len(cm.hotCache) >= cm.config.MaxHotEntries {
				cm.evictLRU(cm.hotCache)
			}
			cm.hotCache[symbol] = entry
		} else {
			if len(cm.warmCache) >= cm.config.MaxWarmEntries {
				cm.evictLRU(cm.warmCache)
			}
			cm.warmCache[symbol] = entry
		}
	}

	cm.updateStats()
	cm.logger.Debug("Batch cache update completed",
		"symbols_count", len(data),
		"tier", string(data[getFirstKey(data)].ScrapedAt.String())) // Using scraped time as tier indicator
}

// Invalidate removes specific entry from all cache tiers
func (cm *CacheManager) Invalidate(symbol string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.hotCache, symbol)
	delete(cm.warmCache, symbol)

	cm.logger.Debug("Cache invalidated", "symbol", symbol)
}

// InvalidateAll clears all cache tiers
func (cm *CacheManager) InvalidateAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.hotCache = make(map[string]*CacheEntry)
	cm.warmCache = make(map[string]*CacheEntry)

	cm.logger.Info("All cache tiers invalidated")
}

// GetStats returns current cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	// Create a copy without the mutex to avoid race conditions
	statsCopy := CacheStats{
		HotHits:       cm.stats.HotHits,
		WarmHits:      cm.stats.WarmHits,
		Misses:        cm.stats.Misses,
		HotEntries:    cm.stats.HotEntries,
		WarmEntries:   cm.stats.WarmEntries,
		AvgLookupTime: cm.stats.AvgLookupTime,
		EvictionCount: cm.stats.EvictionCount,
		EstimatedSize: cm.stats.EstimatedSize,
		StartTime:     cm.stats.StartTime,
		LastCleanup:   cm.stats.LastCleanup,
		CleanupCount:  cm.stats.CleanupCount,
	}

	// Update current entry counts
	cm.mu.RLock()
	statsCopy.HotEntries = len(cm.hotCache)
	statsCopy.WarmEntries = len(cm.warmCache)
	cm.mu.RUnlock()

	return &statsCopy
}

// GetHitRatio returns cache hit ratio as percentage
func (cm *CacheManager) GetHitRatio() float64 {
	cm.stats.mu.RLock()
	defer cm.stats.mu.RUnlock()

	totalRequests := cm.stats.HotHits + cm.stats.WarmHits + cm.stats.Misses
	if totalRequests == 0 {
		return 0.0
	}

	hits := cm.stats.HotHits + cm.stats.WarmHits
	return float64(hits) / float64(totalRequests) * 100.0
}

// WarmCache preloads frequently accessed symbols
func (cm *CacheManager) WarmCache(ctx context.Context, symbols []string, dataProvider func(string) (*market.ScrapedData, error)) error {
	if !cm.config.EnableWarming {
		return nil
	}

	cm.logger.Info("Starting cache warming", "symbols_count", len(symbols))

	for _, symbol := range symbols {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if already cached
		if _, exists := cm.Get(symbol, false); exists {
			continue
		}

		// Fetch and cache data
		if data, err := dataProvider(symbol); err == nil {
			cm.Set(symbol, data, false)
		} else {
			cm.logger.Warn("Failed to warm cache for symbol", "symbol", symbol, "error", err)
		}
	}

	cm.logger.Info("Cache warming completed", "symbols_warmed", len(symbols))
	return nil
}

// Close stops the cache manager and cleanup routines
func (cm *CacheManager) Close() error {
	cm.logger.Info("Stopping cache manager")

	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}

	close(cm.stopChan)
	return nil
}

// Private methods

// isEntryValid checks if cache entry is still valid based on TTL and market conditions
func (cm *CacheManager) isEntryValid(entry *CacheEntry, isMarketOpen bool) bool {
	age := time.Since(entry.CachedAt)

	// During trading hours, use hot cache TTL for stricter freshness
	if isMarketOpen && entry.Tier == TierHot {
		return age <= cm.config.HotCacheTTL
	}

	// For warm cache or non-trading hours
	return age <= entry.TTL
}

// promoteToHot moves frequently accessed entry from warm to hot cache
func (cm *CacheManager) promoteToHot(symbol string, entry *CacheEntry) {
	// Check hot cache capacity
	if len(cm.hotCache) >= cm.config.MaxHotEntries {
		cm.evictLRU(cm.hotCache)
	}

	// Update entry metadata
	entry.Tier = TierHot
	entry.TTL = cm.config.HotCacheTTL
	entry.CachedAt = time.Now()

	// Move entry
	cm.hotCache[symbol] = entry
	delete(cm.warmCache, symbol)

	cm.logger.Debug("Cache entry promoted to hot tier", "symbol", symbol)
}

// evictLRU removes least recently used entry from cache
func (cm *CacheManager) evictLRU(cache map[string]*CacheEntry) {
	var oldestSymbol string
	oldestTime := time.Now()

	for symbol, entry := range cache {
		if entry.LastAccess.Before(oldestTime) {
			oldestTime = entry.LastAccess
			oldestSymbol = symbol
		}
	}

	if oldestSymbol != "" {
		delete(cache, oldestSymbol)
		cm.recordEviction()
		cm.logger.Debug("Cache entry evicted (LRU)", "symbol", oldestSymbol)
	}
}

// cleanupRoutine periodically removes expired entries
func (cm *CacheManager) cleanupRoutine() {
	for {
		select {
		case <-cm.cleanupTicker.C:
			cm.performCleanup()
		case <-cm.stopChan:
			return
		}
	}
}

// performCleanup removes expired entries from all cache tiers
func (cm *CacheManager) performCleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()

	// Use separate slices to avoid map iteration during deletion (performance optimization)
	var expiredHot, expiredWarm []string

	// Collect expired entries from hot cache
	for symbol, entry := range cm.hotCache {
		if now.Sub(entry.CachedAt) > entry.TTL {
			expiredHot = append(expiredHot, symbol)
		}
	}

	// Collect expired entries from warm cache
	for symbol, entry := range cm.warmCache {
		if now.Sub(entry.CachedAt) > entry.TTL {
			expiredWarm = append(expiredWarm, symbol)
		}
	}

	// Delete expired entries
	for _, symbol := range expiredHot {
		delete(cm.hotCache, symbol)
	}
	for _, symbol := range expiredWarm {
		delete(cm.warmCache, symbol)
	}

	expiredCount := len(expiredHot) + len(expiredWarm)
	if expiredCount > 0 {
		cm.logger.Debug("Cache cleanup completed",
			"expired_entries", expiredCount,
			"hot_entries", len(cm.hotCache),
			"warm_entries", len(cm.warmCache))
	}

	cm.recordCleanup()
	cm.updateStats()
}

// Statistics tracking methods

func (cm *CacheManager) recordHit(tier CacheTier) {
	if !cm.config.EnableStats {
		return
	}

	cm.stats.mu.Lock()
	defer cm.stats.mu.Unlock()

	switch tier {
	case TierHot:
		cm.stats.HotHits++
	case TierWarm:
		cm.stats.WarmHits++
	}
}

func (cm *CacheManager) recordMiss() {
	if !cm.config.EnableStats {
		return
	}

	cm.stats.mu.Lock()
	defer cm.stats.mu.Unlock()

	cm.stats.Misses++
}

func (cm *CacheManager) recordEviction() {
	if !cm.config.EnableStats {
		return
	}

	cm.stats.mu.Lock()
	defer cm.stats.mu.Unlock()

	cm.stats.EvictionCount++
}

func (cm *CacheManager) recordCleanup() {
	if !cm.config.EnableStats {
		return
	}

	cm.stats.mu.Lock()
	defer cm.stats.mu.Unlock()

	cm.stats.LastCleanup = time.Now()
	cm.stats.CleanupCount++
}

func (cm *CacheManager) updateLookupTime(duration time.Duration) {
	cm.stats.mu.Lock()
	defer cm.stats.mu.Unlock()

	// Simple moving average
	if cm.stats.AvgLookupTime == 0 {
		cm.stats.AvgLookupTime = duration
	} else {
		cm.stats.AvgLookupTime = (cm.stats.AvgLookupTime + duration) / 2
	}
}

func (cm *CacheManager) updateStats() {
	if !cm.config.EnableStats {
		return
	}

	// Estimate memory usage (rough calculation)
	entrySize := int64(EstimatedEntrySize)
	totalEntries := int64(len(cm.hotCache) + len(cm.warmCache))

	cm.stats.mu.Lock()
	cm.stats.EstimatedSize = totalEntries * entrySize
	cm.stats.mu.Unlock()
}

// Helper function to get first key from map
func getFirstKey(m map[string]*market.ScrapedData) string {
	for k := range m {
		return k
	}
	return ""
}
