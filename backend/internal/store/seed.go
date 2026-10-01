package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mable/audience-builder/backend/internal/model"
)

// DefaultSeedAsOf is the canonical reference timestamp used in the assignment prompt.
const DefaultSeedAsOf = "2026-09-29T00:00:00.000Z"

// SeedData contains synthetic personas and events for reproducible evaluation.
func (db *DB) SeedData() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return nil // Already seeded
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	userStmt, err := tx.Prepare("INSERT INTO users (anonymous_id, created_at) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare user insert: %w", err)
	}
	defer userStmt.Close()

	eventStmt, err := tx.Prepare("INSERT INTO events (id, anonymous_id, event_type, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare event insert: %w", err)
	}
	defer eventStmt.Close()

	addUser := func(anonID, createdAt string) error {
		_, err := userStmt.Exec(anonID, createdAt)
		return err
	}

	addEvent := func(anonID string, eventType model.EventType, timestamp string) error {
		id := generateID()
		_, err := eventStmt.Exec(id, anonID, string(eventType), timestamp)
		return err
	}

	// 1. anon_target_1: Active browser (3 views in window, 0 purchases) -> Matches sample rule
	if err := addUser("anon_target_1", "2026-09-01T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_target_1", model.EventTypeProductView, "2026-09-25T10:00:00Z")
	_ = addEvent("anon_target_1", model.EventTypeProductView, "2026-09-26T14:30:00Z")
	_ = addEvent("anon_target_1", model.EventTypeProductView, "2026-09-28T09:15:00Z")
	_ = addEvent("anon_target_1", model.EventTypeAddToCart, "2026-09-28T09:20:00Z")

	// 2. anon_target_2: Exact threshold browser (2 views in window, 0 purchases) -> Matches sample rule
	if err := addUser("anon_target_2", "2026-09-10T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_target_2", model.EventTypeProductView, "2026-09-23T11:00:00Z")
	_ = addEvent("anon_target_2", model.EventTypeProductView, "2026-09-27T16:00:00Z")

	// 3. anon_buyer: Converted shopper (2 views, 1 purchase in window) -> Does NOT match (purchase != 0)
	if err := addUser("anon_buyer", "2026-09-05T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_buyer", model.EventTypeProductView, "2026-09-24T12:00:00Z")
	_ = addEvent("anon_buyer", model.EventTypeProductView, "2026-09-25T15:00:00Z")
	_ = addEvent("anon_buyer", model.EventTypeCheckoutStarted, "2026-09-25T15:10:00Z")
	_ = addEvent("anon_buyer", model.EventTypePurchase, "2026-09-25T15:15:00Z")

	// 4. anon_insufficient: Casual visitor (1 view, 0 purchases) -> Does NOT match (views < 2)
	if err := addUser("anon_insufficient", "2026-09-15T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_insufficient", model.EventTypePageView, "2026-09-24T08:00:00Z")
	_ = addEvent("anon_insufficient", model.EventTypeProductView, "2026-09-24T08:05:00Z")

	// 5. anon_stale_views: Stale browsing (views happened >7 days before asOf) -> Does NOT match
	if err := addUser("anon_stale_views", "2026-08-01T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_stale_views", model.EventTypeProductView, "2026-09-20T10:00:00Z")
	_ = addEvent("anon_stale_views", model.EventTypeProductView, "2026-09-21T18:00:00Z")

	// 6. anon_future_events: Activity occurring after asOf -> Does NOT match (reproducibility guarantee)
	if err := addUser("anon_future_events", "2026-09-20T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_future_events", model.EventTypeProductView, "2026-09-29T02:00:00Z")
	_ = addEvent("anon_future_events", model.EventTypeProductView, "2026-09-30T10:00:00Z")

	// 7. anon_exact_boundary: Views at exact window boundaries (2 views, 0 purchases) -> Matches sample rule
	if err := addUser("anon_exact_boundary", "2026-09-01T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_exact_boundary", model.EventTypeProductView, "2026-09-22T00:00:00Z") // asOf - 7d
	_ = addEvent("anon_exact_boundary", model.EventTypeProductView, "2026-09-29T00:00:00Z") // exact asOf

	// 8. anon_old_buyer: Historical buyer, but purchase occurred outside 7d window -> Matches sample rule
	if err := addUser("anon_old_buyer", "2026-08-15T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_old_buyer", model.EventTypePurchase, "2026-09-14T10:00:00Z") // 15 days ago
	_ = addEvent("anon_old_buyer", model.EventTypeProductView, "2026-09-26T09:00:00Z")
	_ = addEvent("anon_old_buyer", model.EventTypeProductView, "2026-09-27T14:00:00Z")

	// 9. anon_other_only: Visited other pages, no product views -> Does NOT match
	if err := addUser("anon_other_only", "2026-09-12T00:00:00Z"); err != nil {
		return err
	}
	_ = addEvent("anon_other_only", model.EventTypePageView, "2026-09-25T11:00:00Z")
	_ = addEvent("anon_other_only", model.EventTypeAddToCart, "2026-09-25T11:05:00Z")

	return tx.Commit()
}

// ResetAndSeed drops existing data and seeds fresh dataset.
func (db *DB) ResetAndSeed() error {
	if _, err := db.Exec("DELETE FROM events; DELETE FROM users;"); err != nil {
		return err
	}
	return db.SeedData()
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// EnsureValidTimestamp formats a time to standard RFC 3339 UTC string.
func FormatRFC3339UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
