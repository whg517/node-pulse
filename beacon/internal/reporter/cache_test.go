package reporter

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/whg517/node-pulse/beacon/internal/logger"
)

// TestMain initializes the logger for all tests
func TestMain(m *testing.M) {
	// Initialize a simple test logger
	logger.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	os.Exit(m.Run())
}

func TestNewPriorityCache(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false) // 1MB
	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}
	if cache.MaxSize() != 1024*1024 {
		t.Errorf("Expected max size 1MB, got %d", cache.MaxSize())
	}
}

func TestPriorityCache_AddP0Entry(t *testing.T) {
	cache := NewPriorityCache(1024, "", false) // 1KB

	entry := &CacheEntry{
		ID:       "test-p0",
		Priority: CacheP0,
		Data:     []byte("test data for p0 entry"),
	}

	err := cache.Add(entry)
	if err != nil {
		t.Fatalf("Failed to add P0 entry: %v", err)
	}

	if cache.Count() != 1 {
		t.Errorf("Expected count 1, got %d", cache.Count())
	}

	retrieved, exists := cache.Get("test-p0")
	if !exists {
		t.Fatal("Expected to retrieve P0 entry")
	}
	if retrieved.ID != "test-p0" {
		t.Errorf("Expected ID test-p0, got %s", retrieved.ID)
	}
}

func TestPriorityCache_AddP2Entry(t *testing.T) {
	cache := NewPriorityCache(1024, "", false) // 1KB

	entry := &CacheEntry{
		ID:       "test-p2",
		Priority: CacheP2,
		Data:     []byte("test data for p2 entry"),
	}

	err := cache.Add(entry)
	if err != nil {
		t.Fatalf("Failed to add P2 entry: %v", err)
	}

	if cache.Count() != 1 {
		t.Errorf("Expected count 1, got %d", cache.Count())
	}
}

func TestPriorityCache_P1Bypass(t *testing.T) {
	cache := NewPriorityCache(1024, "", false)

	entry := &CacheEntry{
		ID:       "test-p1",
		Priority: CacheP1,
		Data:     []byte("test data"),
	}

	err := cache.Add(entry)
	if err == nil {
		t.Error("Expected error when adding P1 entry (should bypass cache)")
	}
}

func TestPriorityCache_FIFOEviction(t *testing.T) {
	// Small cache to trigger eviction
	cache := NewPriorityCache(200, "", false)

	// Add first P2 entry
	entry1 := &CacheEntry{
		ID:       "p2-first",
		Priority: CacheP2,
		Data:     make([]byte, 50),
	}
	_ = cache.Add(entry1)

	// Add second P2 entry
	entry2 := &CacheEntry{
		ID:       "p2-second",
		Priority: CacheP2,
		Data:     make([]byte, 50),
	}
	_ = cache.Add(entry2)

	// Add a large entry to trigger eviction
	entry3 := &CacheEntry{
		ID:       "p2-large",
		Priority: CacheP2,
		Data:     make([]byte, 100),
	}
	_ = cache.Add(entry3)

	// First entry should have been evicted
	_, exists := cache.Get("p2-first")
	if exists {
		t.Error("Expected first P2 entry to be evicted (FIFO)")
	}

	// Large entry should exist
	_, exists = cache.Get("p2-large")
	if !exists {
		t.Error("Expected large P2 entry to exist")
	}

	// Should have recorded eviction
	if cache.Evictions() == 0 {
		t.Error("Expected evictions to be recorded")
	}
}

func TestPriorityCache_P0NeverEvicted(t *testing.T) {
	// Very small cache
	cache := NewPriorityCache(200, "", false)

	// Add P0 entry
	p0Entry := &CacheEntry{
		ID:       "p0-critical",
		Priority: CacheP0,
		Data:     make([]byte, 50),
	}
	_ = cache.Add(p0Entry)

	// Try to add many P2 entries to force eviction
	for i := 0; i < 10; i++ {
		entry := &CacheEntry{
			ID:       string(rune('a' + i)),
			Priority: CacheP2,
			Data:     make([]byte, 100),
		}
		_ = cache.Add(entry)
	}

	// P0 entry should still exist
	_, exists := cache.Get("p0-critical")
	if !exists {
		t.Error("P0 entry should never be evicted")
	}
}

func TestPriorityCache_Remove(t *testing.T) {
	cache := NewPriorityCache(1024, "", false)

	entry := &CacheEntry{
		ID:       "to-remove",
		Priority: CacheP2,
		Data:     []byte("data"),
	}
	_ = cache.Add(entry)

	removed := cache.Remove("to-remove")
	if !removed {
		t.Error("Expected entry to be removed")
	}

	if cache.Count() != 0 {
		t.Errorf("Expected count 0 after removal, got %d", cache.Count())
	}

	// Try to remove non-existent entry
	removed = cache.Remove("non-existent")
	if removed {
		t.Error("Expected false when removing non-existent entry")
	}
}

func TestPriorityCache_GetAllEntriesForUpload(t *testing.T) {
	cache := NewPriorityCache(10*1024, "", false) // 10KB

	// Add P0 and P2 entries
	_ = cache.Add(&CacheEntry{ID: "p2-1", Priority: CacheP2, Data: []byte("data1")})
	_ = cache.Add(&CacheEntry{ID: "p0-1", Priority: CacheP0, Data: []byte("data2")})
	_ = cache.Add(&CacheEntry{ID: "p2-2", Priority: CacheP2, Data: []byte("data3")})

	entries := cache.GetAllEntriesForUpload()

	// P0 entries should come first
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	if entries[0].Priority != CacheP0 {
		t.Error("Expected first entry to be P0 (priority order)")
	}
}

func TestPriorityCache_Clear(t *testing.T) {
	cache := NewPriorityCache(1024, "", false)

	_ = cache.Add(&CacheEntry{ID: "p0-1", Priority: CacheP0, Data: []byte("data")})
	_ = cache.Add(&CacheEntry{ID: "p2-1", Priority: CacheP2, Data: []byte("data")})

	cache.Clear()

	if cache.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", cache.Count())
	}

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func TestPriorityCache_IncrementRetryCount(t *testing.T) {
	cache := NewPriorityCache(1024, "", false)

	entry := &CacheEntry{
		ID:       "retry-test",
		Priority: CacheP2,
		Data:     []byte("data"),
	}
	_ = cache.Add(entry)

	count := cache.IncrementRetryCount("retry-test")
	if count != 1 {
		t.Errorf("Expected retry count 1, got %d", count)
	}

	count = cache.IncrementRetryCount("retry-test")
	if count != 2 {
		t.Errorf("Expected retry count 2, got %d", count)
	}
}

func TestPriorityCache_Persist(t *testing.T) {
	// Create temp file
	tmpDir := os.TempDir()
	cachePath := filepath.Join(tmpDir, "test-cache.dat")
	defer func() { _ = os.Remove(cachePath) }()

	cache := NewPriorityCache(1024, cachePath, true)

	// Add entries
	_ = cache.Add(&CacheEntry{
		ID:        "persist-test",
		Priority:  CacheP0,
		Data:      []byte("persist data"),
		Checksum:  12345,
		Timestamp: time.Now(),
	})

	// Persist
	err := cache.Persist()
	if err != nil {
		t.Fatalf("Failed to persist cache: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Fatal("Cache file should exist after persist")
	}

	// Load into new cache
	cache2 := NewPriorityCache(1024, cachePath, true)

	// Verify entry loaded
	entry, exists := cache2.Get("persist-test")
	if !exists {
		t.Fatal("Expected persisted entry to be loaded")
	}
	if entry.Checksum != 12345 {
		t.Errorf("Expected checksum 12345, got %d", entry.Checksum)
	}
}

func TestPriorityCache_GetStats(t *testing.T) {
	cache := NewPriorityCache(1024, "", false)

	_ = cache.Add(&CacheEntry{ID: "p0-1", Priority: CacheP0, Data: make([]byte, 100)})
	_ = cache.Add(&CacheEntry{ID: "p2-1", Priority: CacheP2, Data: make([]byte, 100)})

	stats := cache.GetStats()

	if stats.P0Count != 1 {
		t.Errorf("Expected P0 count 1, got %d", stats.P0Count)
	}
	if stats.P2Count != 1 {
		t.Errorf("Expected P2 count 1, got %d", stats.P2Count)
	}
	if stats.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", stats.TotalCount)
	}
}
