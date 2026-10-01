package store_test

import (
	"testing"

	"github.com/mable/audience-builder/backend/internal/store"
)

func TestOpenAndSeed(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	if err := db.SeedData(); err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	var userCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if userCount != 9 {
		t.Errorf("expected 9 seeded users, got %d", userCount)
	}

	var eventCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount); err != nil {
		t.Fatalf("failed to count events: %v", err)
	}
	if eventCount == 0 {
		t.Errorf("expected > 0 seeded events, got 0")
	}

	// Calling SeedData again should be idempotent
	if err := db.SeedData(); err != nil {
		t.Fatalf("failed calling SeedData a second time: %v", err)
	}

	var userCountAfter int
	_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCountAfter)
	if userCountAfter != 9 {
		t.Errorf("expected still 9 users after second seed, got %d", userCountAfter)
	}
}
