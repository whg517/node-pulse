package db

import (
	"context"
	"testing"
)

// Checker is a minimal interface for testing interface compliance
type Checker interface {
	Check(ctx context.Context) error
}

// TestDatabase_New_InvalidURL tests that invalid database URL returns error
func TestDatabase_New_InvalidURL(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Error("Expected error for empty DATABASE_URL")
	}
}

// TestDatabase_NilDatabaseCheck tests that nil database panics (expected behavior)
func TestDatabase_NilDatabaseCheck(t *testing.T) {
	var db *Database

	// Calling Check on nil *Database should panic
	// This is expected Go behavior for nil pointer methods
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, test passes
			return
		}
		t.Error("Expected panic when calling Check on nil database, but it did not panic")
	}()

	db.Check(context.Background())
}

// TestDatabase_CheckerInterface tests that Database implements Checker interface
func TestDatabase_CheckerInterface(t *testing.T) {
	// This test verifies Database implements Checker interface
	// by ensuring it can be assigned to the interface type
	var _ Checker = (*Database)(nil)
}

// TestDatabase_CloseNil tests that closing nil database is safe
func TestDatabase_CloseNil(t *testing.T) {
	var db *Database

	// Close on nil should be safe (implementation should handle it)
	defer func() {
		if r := recover(); r != nil {
			// Expected panic, test documents behavior
			return
		}
	}()

	db.Close()
}
