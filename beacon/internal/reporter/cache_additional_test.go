package reporter

import (
	"testing"
)

// TestPriorityCache_GetP0EntriesForUpload tests GetP0EntriesForUpload
func TestPriorityCache_GetP0EntriesForUpload(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	// Initially empty
	entries := cache.GetP0EntriesForUpload()
	if entries == nil {
		t.Fatal("Expected non-nil entries slice")
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 P0 entries, got %d", len(entries))
	}

	// Add P0 entry
	entry := &CacheEntry{
		ID:       "p0-entry-1",
		Priority: CacheP0,
		Data:     []byte("alert data"),
	}
	if err := cache.Add(entry); err != nil {
		t.Fatalf("Failed to add P0 entry: %v", err)
	}

	entries = cache.GetP0EntriesForUpload()
	if len(entries) != 1 {
		t.Errorf("Expected 1 P0 entry, got %d", len(entries))
	}
	if entries[0].ID != "p0-entry-1" {
		t.Errorf("Expected ID 'p0-entry-1', got %s", entries[0].ID)
	}
}

// TestPriorityCache_GetP2EntriesForUpload tests GetP2EntriesForUpload
func TestPriorityCache_GetP2EntriesForUpload(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	// Initially empty
	entries := cache.GetP2EntriesForUpload()
	if entries == nil {
		t.Fatal("Expected non-nil entries slice")
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 P2 entries, got %d", len(entries))
	}

	// Add P2 entries
	for i := 0; i < 3; i++ {
		entry := &CacheEntry{
			ID:       "p2-entry-" + string(rune('0'+i)),
			Priority: CacheP2,
			Data:     []byte("probe data"),
		}
		if err := cache.Add(entry); err != nil {
			t.Fatalf("Failed to add P2 entry: %v", err)
		}
	}

	entries = cache.GetP2EntriesForUpload()
	if len(entries) != 3 {
		t.Errorf("Expected 3 P2 entries, got %d", len(entries))
	}
}

// TestPriorityCache_Remove_P0 tests Remove for P0 entry
func TestPriorityCache_Remove_P0_Additional(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	entry := &CacheEntry{
		ID:       "p0-to-remove-add",
		Priority: CacheP0,
		Data:     []byte("data"),
	}
	if err := cache.Add(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Remove it
	removed := cache.Remove("p0-to-remove-add")
	if !removed {
		t.Error("Expected Remove to return true for existing P0 entry")
	}

	if cache.Count() != 0 {
		t.Errorf("Expected count 0 after removal, got %d", cache.Count())
	}
}

// TestPriorityCache_Remove_P2 tests Remove for P2 entry
func TestPriorityCache_Remove_P2_Additional(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	entry := &CacheEntry{
		ID:       "p2-to-remove-add",
		Priority: CacheP2,
		Data:     []byte("data"),
	}
	if err := cache.Add(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	// Remove it
	removed := cache.Remove("p2-to-remove-add")
	if !removed {
		t.Error("Expected Remove to return true for existing P2 entry")
	}

	if cache.Count() != 0 {
		t.Errorf("Expected count 0 after removal, got %d", cache.Count())
	}
}

// TestPriorityCache_Remove_NonExistent tests Remove for non-existent entry
func TestPriorityCache_Remove_NonExistent_Additional(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	// Remove non-existent entry
	removed := cache.Remove("nonexistent-add")
	if removed {
		t.Error("Expected Remove to return false for non-existent entry")
	}
}

// TestPriorityCache_IncrementRetryCount_P2 tests IncrementRetryCount for P2
func TestPriorityCache_IncrementRetryCount_P2(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	// Add P2 entry
	p2Entry := &CacheEntry{
		ID:       "retry-p2-add",
		Priority: CacheP2,
		Data:     []byte("data"),
	}
	if err := cache.Add(p2Entry); err != nil {
		t.Fatalf("Failed to add P2 entry: %v", err)
	}

	// Increment retry count for P2
	count := cache.IncrementRetryCount("retry-p2-add")
	if count != 1 {
		t.Errorf("Expected retry count 1 for P2, got %d", count)
	}

	count = cache.IncrementRetryCount("retry-p2-add")
	if count != 2 {
		t.Errorf("Expected retry count 2 for P2, got %d", count)
	}
}
func TestPriorityCache_GetP0AndP2Separation(t *testing.T) {
	cache := NewPriorityCache(1024*1024, "", false)

	// Add both P0 and P2 entries
	p0Entry := &CacheEntry{
		ID:       "p0-entry",
		Priority: CacheP0,
		Data:     []byte("alert"),
	}
	p2Entry := &CacheEntry{
		ID:       "p2-entry",
		Priority: CacheP2,
		Data:     []byte("probe"),
	}

	if err := cache.Add(p0Entry); err != nil {
		t.Fatalf("Failed to add P0 entry: %v", err)
	}
	if err := cache.Add(p2Entry); err != nil {
		t.Fatalf("Failed to add P2 entry: %v", err)
	}

	// P0 entries should only have P0
	p0Entries := cache.GetP0EntriesForUpload()
	if len(p0Entries) != 1 || p0Entries[0].ID != "p0-entry" {
		t.Errorf("Expected 1 P0 entry, got %d", len(p0Entries))
	}

	// P2 entries should only have P2
	p2Entries := cache.GetP2EntriesForUpload()
	if len(p2Entries) != 1 || p2Entries[0].ID != "p2-entry" {
		t.Errorf("Expected 1 P2 entry, got %d", len(p2Entries))
	}

	// All entries should have both
	allEntries := cache.GetAllEntriesForUpload()
	if len(allEntries) != 2 {
		t.Errorf("Expected 2 total entries, got %d", len(allEntries))
	}

	// P0 should come first
	if allEntries[0].Priority != CacheP0 {
		t.Error("Expected P0 entries first in GetAllEntriesForUpload")
	}
}

// TestPriorityCache_Persist_Disabled tests Persist when not enabled
func TestPriorityCache_Persist_Disabled(t *testing.T) {
cache := NewPriorityCache(1024*1024, "", false)

if err := cache.Persist(); err != nil {
t.Errorf("Expected nil error when persist disabled, got: %v", err)
}
}

// TestPriorityCache_Persist_WithData tests Persist with actual data
func TestPriorityCache_Persist_WithData(t *testing.T) {
tmpDir := t.TempDir()
persistPath := tmpDir + "/cache.dat"

cache := NewPriorityCache(1024*1024, persistPath, true)

// Add some entries
for i := 0; i < 3; i++ {
entry := &CacheEntry{
ID:       "entry-persist-" + string(rune('0'+i)),
Priority: CacheP2,
Data:     []byte("test data"),
}
if err := cache.Add(entry); err != nil {
t.Fatalf("Failed to add entry: %v", err)
}
}

if err := cache.Persist(); err != nil {
t.Errorf("Persist failed: %v", err)
}
}

// TestPriorityCache_NewWithPersist_Load tests NewPriorityCache with persist load
func TestPriorityCache_NewWithPersist_Load(t *testing.T) {
tmpDir := t.TempDir()
persistPath := tmpDir + "/cache.dat"

// Create initial cache with data and persist
cache1 := NewPriorityCache(1024*1024, persistPath, true)
entry := &CacheEntry{
ID:       "persist-test",
Priority: CacheP2,
Data:     []byte("test data"),
Size:     9,
}
if err := cache1.Add(entry); err != nil {
t.Fatalf("Failed to add entry: %v", err)
}
if err := cache1.Persist(); err != nil {
t.Fatalf("Persist failed: %v", err)
}

// Load from file by creating new cache
cache2 := NewPriorityCache(1024*1024, persistPath, true)
if cache2.Count() == 0 {
t.Error("Expected entries to be loaded from file")
}
}
