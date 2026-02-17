package reporter

import (
	"container/list"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"beacon/internal/logger"
)

// CachePriority represents the priority level of cached data (FR-4.1.7)
type CachePriority int

const (
	// CacheP0 is the highest priority - Alert data (never drop)
	CacheP0 CachePriority = iota

	// CacheP1 bypasses cache - Heartbeat data (send directly)
	CacheP1

	// CacheP2 is the lowest priority - Regular probe data (FIFO eviction)
	CacheP2
)

// CacheEntry represents a single entry in the priority cache
type CacheEntry struct {
	ID         string        `json:"id"`
	Priority   CachePriority `json:"priority"`
	Data       []byte        `json:"data"`
	Checksum   uint32        `json:"checksum"` // CRC32 checksum
	Size       int64         `json:"size"`
	Timestamp  time.Time     `json:"timestamp"`
	RetryCount int           `json:"retry_count"`
}

// PriorityCache implements a priority-based cache with FIFO eviction for P2 entries
type PriorityCache struct {
	mu sync.RWMutex

	// Maximum cache size in bytes
	maxSize int64

	// Current cache size in bytes
	currentSize int64

	// P0 entries (never evicted)
	p0Entries map[string]*list.Element
	p0List    *list.List

	// P2 entries (FIFO eviction)
	p2Entries map[string]*list.Element
	p2List    *list.List

	// Total evictions counter
	evictions int64

	// Cache file path for persistence
	persistPath string

	// Enable persistence
	persistEnabled bool
}

// NewPriorityCache creates a new priority cache
func NewPriorityCache(maxSizeBytes int64, persistPath string, persistEnabled bool) *PriorityCache {
	c := &PriorityCache{
		maxSize:        maxSizeBytes,
		p0Entries:      make(map[string]*list.Element),
		p0List:         list.New(),
		p2Entries:      make(map[string]*list.Element),
		p2List:         list.New(),
		persistPath:    persistPath,
		persistEnabled: persistEnabled,
	}

	// Load persisted cache if enabled
	if persistEnabled && persistPath != "" {
		if err := c.load(); err != nil {
			logger.WithFields(map[string]interface{}{
				"component": "cache",
				"error":     err.Error(),
			}).Warn("Failed to load persisted cache, starting fresh")
		}
	}

	return c
}

// Add adds an entry to the cache with the specified priority
func (c *PriorityCache) Add(entry *CacheEntry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	// P1 entries bypass the cache (send directly)
	if entry.Priority == CacheP1 {
		return fmt.Errorf("P1 entries bypass cache, send directly")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate size
	entry.Size = int64(len(entry.Data) + 100) // Add overhead estimate
	entry.Timestamp = time.Now()

	// Check if entry already exists
	var existingList *list.List
	var existingMap map[string]*list.Element

	if entry.Priority == CacheP0 {
		existingList = c.p0List
		existingMap = c.p0Entries
	} else {
		existingList = c.p2List
		existingMap = c.p2Entries
	}

	// Remove existing entry if present
	if elem, exists := existingMap[entry.ID]; exists {
		oldEntry := elem.Value.(*CacheEntry)
		c.currentSize -= oldEntry.Size
		existingList.Remove(elem)
	}

	// Evict P2 entries if necessary (FIFO)
	for c.currentSize+entry.Size > c.maxSize && c.p2List.Len() > 0 {
		c.evictOldestP2()
	}

	// Check if we have space now
	if c.currentSize+entry.Size > c.maxSize {
		return fmt.Errorf("cache full, cannot add entry (size: %d, max: %d)",
			c.currentSize+entry.Size, c.maxSize)
	}

	// Add entry
	elem := existingList.PushBack(entry)
	existingMap[entry.ID] = elem
	c.currentSize += entry.Size

	logger.WithFields(map[string]interface{}{
		"component":  "cache",
		"entry_id":   entry.ID,
		"priority":   entry.Priority,
		"size":       entry.Size,
		"cache_size": c.currentSize,
	}).Debug("Cache entry added")

	return nil
}

// Get retrieves an entry from the cache
func (c *PriorityCache) Get(id string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check P0 first
	if elem, exists := c.p0Entries[id]; exists {
		return elem.Value.(*CacheEntry), true
	}

	// Check P2
	if elem, exists := c.p2Entries[id]; exists {
		return elem.Value.(*CacheEntry), true
	}

	return nil, false
}

// Remove removes an entry from the cache
func (c *PriorityCache) Remove(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check P0
	if elem, exists := c.p0Entries[id]; exists {
		entry := elem.Value.(*CacheEntry)
		c.currentSize -= entry.Size
		c.p0List.Remove(elem)
		delete(c.p0Entries, id)
		return true
	}

	// Check P2
	if elem, exists := c.p2Entries[id]; exists {
		entry := elem.Value.(*CacheEntry)
		c.currentSize -= entry.Size
		c.p2List.Remove(elem)
		delete(c.p2Entries, id)
		return true
	}

	return false
}

// evictOldestP2 evicts the oldest P2 entry (FIFO)
func (c *PriorityCache) evictOldestP2() {
	if c.p2List.Len() == 0 {
		return
	}

	// Get oldest entry (front of list)
	elem := c.p2List.Front()
	if elem == nil {
		return
	}

	entry := elem.Value.(*CacheEntry)
	c.currentSize -= entry.Size
	c.p2List.Remove(elem)
	delete(c.p2Entries, entry.ID)
	c.evictions++

	logger.WithFields(map[string]interface{}{
		"component":  "cache",
		"entry_id":   entry.ID,
		"evictions":  c.evictions,
		"cache_size": c.currentSize,
	}).Warn("P2 cache entry evicted (FIFO)")
}

// GetP2EntriesForUpload returns all P2 entries ready for upload
func (c *PriorityCache) GetP2EntriesForUpload() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CacheEntry, 0, c.p2List.Len())
	for elem := c.p2List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	return entries
}

// GetP0EntriesForUpload returns all P0 entries ready for upload
func (c *PriorityCache) GetP0EntriesForUpload() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CacheEntry, 0, c.p0List.Len())
	for elem := c.p0List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	return entries
}

// GetAllEntriesForUpload returns all entries prioritized (P0 first, then P2)
func (c *PriorityCache) GetAllEntriesForUpload() []*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]*CacheEntry, 0, c.p0List.Len()+c.p2List.Len())

	// Add P0 entries first (highest priority)
	for elem := c.p0List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	// Add P2 entries (FIFO order)
	for elem := c.p2List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	return entries
}

// IncrementRetryCount increments the retry count for an entry
func (c *PriorityCache) IncrementRetryCount(id string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check P0
	if elem, exists := c.p0Entries[id]; exists {
		entry := elem.Value.(*CacheEntry)
		entry.RetryCount++
		return entry.RetryCount
	}

	// Check P2
	if elem, exists := c.p2Entries[id]; exists {
		entry := elem.Value.(*CacheEntry)
		entry.RetryCount++
		return entry.RetryCount
	}

	return 0
}

// Size returns the current cache size in bytes
func (c *PriorityCache) Size() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}

// Count returns the total number of entries in the cache
func (c *PriorityCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.p0List.Len() + c.p2List.Len()
}

// Evictions returns the total number of evictions
func (c *PriorityCache) Evictions() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictions
}

// MaxSize returns the maximum cache size
func (c *PriorityCache) MaxSize() int64 {
	return c.maxSize
}

// Clear clears all entries from the cache
func (c *PriorityCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.p0Entries = make(map[string]*list.Element)
	c.p0List = list.New()
	c.p2Entries = make(map[string]*list.Element)
	c.p2List = list.New()
	c.currentSize = 0

	logger.WithField("component", "cache").Info("Cache cleared")
}

// Persist saves the cache to disk
func (c *PriorityCache) Persist() error {
	if !c.persistEnabled || c.persistPath == "" {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create directory if needed
	dir := filepath.Dir(c.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Collect all entries
	entries := make([]*CacheEntry, 0, c.p0List.Len()+c.p2List.Len())

	for elem := c.p0List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	for elem := c.p2List.Front(); elem != nil; elem = elem.Next() {
		entries = append(entries, elem.Value.(*CacheEntry))
	}

	// Marshal to JSON
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to file
	if err := os.WriteFile(c.persistPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	logger.WithFields(map[string]interface{}{
		"component":  "cache",
		"entries":    len(entries),
		"cache_size": c.currentSize,
	}).Info("Cache persisted to disk")

	return nil
}

// load loads the cache from disk
func (c *PriorityCache) load() error {
	data, err := os.ReadFile(c.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file, start fresh
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var entries []*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	// Add entries to cache
	for _, entry := range entries {
		if entry.Priority == CacheP0 {
			elem := c.p0List.PushBack(entry)
			c.p0Entries[entry.ID] = elem
		} else if entry.Priority == CacheP2 {
			elem := c.p2List.PushBack(entry)
			c.p2Entries[entry.ID] = elem
		}
		c.currentSize += entry.Size
	}

	logger.WithFields(map[string]interface{}{
		"component":  "cache",
		"entries":    len(entries),
		"cache_size": c.currentSize,
	}).Info("Cache loaded from disk")

	return nil
}

// CacheStats represents cache statistics
type CacheStats struct {
	CurrentSizeBytes int64 `json:"current_size_bytes"`
	MaxSizeBytes     int64 `json:"max_size_bytes"`
	P0Count          int   `json:"p0_count"`
	P2Count          int   `json:"p2_count"`
	TotalCount       int   `json:"total_count"`
	Evictions        int64 `json:"evictions"`
}

// GetStats returns cache statistics
func (c *PriorityCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		CurrentSizeBytes: c.currentSize,
		MaxSizeBytes:     c.maxSize,
		P0Count:          c.p0List.Len(),
		P2Count:          c.p2List.Len(),
		TotalCount:       c.p0List.Len() + c.p2List.Len(),
		Evictions:        c.evictions,
	}
}
