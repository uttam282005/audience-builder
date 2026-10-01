package evaluator_test

import (
	"context"
	"testing"
	"time"

	"github.com/mable/audience-builder/backend/internal/evaluator"
	"github.com/mable/audience-builder/backend/internal/model"
	"github.com/mable/audience-builder/backend/internal/store"
)

func setupTestDB(t *testing.T) *store.DB {
	t.Helper()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := db.SeedData(); err != nil {
		t.Fatalf("failed to seed test db: %v", err)
	}
	return db
}

func TestEvaluator_CanonicalScenario(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eval := evaluator.New(db)

	asOfTime, err := time.Parse(time.RFC3339, "2026-09-29T00:00:00.000Z")
	if err != nil {
		t.Fatalf("failed parsing asOf: %v", err)
	}

	req := &model.PreviewRequest{
		Name:       "Viewed but not purchased",
		AsOf:       "2026-09-29T00:00:00.000Z",
		ParsedAsOf: asOfTime,
		Conditions: []model.Condition{
			{
				EventType:  model.EventTypeProductView,
				Operator:   model.OperatorAtLeast,
				Count:      2,
				WithinDays: 7,
			},
			{
				EventType:  model.EventTypePurchase,
				Operator:   model.OperatorExactly,
				Count:      0,
				WithinDays: 7,
			},
		},
	}

	resp, err := eval.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected evaluate error: %v", err)
	}

	if resp.Name != req.Name {
		t.Errorf("expected response name %q, got %q", req.Name, resp.Name)
	}
	if resp.AsOf != req.AsOf {
		t.Errorf("expected response asOf %q, got %q", req.AsOf, resp.AsOf)
	}

	expectedIDs := map[string]bool{
		"anon_target_1":       true,
		"anon_target_2":       true,
		"anon_exact_boundary": true,
		"anon_old_buyer":      true,
	}

	if resp.Total != len(expectedIDs) {
		t.Fatalf("expected %d matched members, got %d", len(expectedIDs), resp.Total)
	}

	for _, member := range resp.Members {
		if !expectedIDs[member.AnonymousID] {
			t.Errorf("unexpected member in audience: %s", member.AnonymousID)
		}

		if len(member.Evidence) != 2 {
			t.Fatalf("member %s expected 2 evidence items, got %d", member.AnonymousID, len(member.Evidence))
		}

		// Evidence verification
		switch member.AnonymousID {
		case "anon_target_1":
			if member.Evidence[0].ObservedCount != 3 || member.Evidence[1].ObservedCount != 0 {
				t.Errorf("anon_target_1 unexpected evidence: %+v", member.Evidence)
			}
		case "anon_target_2":
			if member.Evidence[0].ObservedCount != 2 || member.Evidence[1].ObservedCount != 0 {
				t.Errorf("anon_target_2 unexpected evidence: %+v", member.Evidence)
			}
		case "anon_exact_boundary":
			if member.Evidence[0].ObservedCount != 2 || member.Evidence[1].ObservedCount != 0 {
				t.Errorf("anon_exact_boundary unexpected evidence: %+v", member.Evidence)
			}
		case "anon_old_buyer":
			if member.Evidence[0].ObservedCount != 2 || member.Evidence[1].ObservedCount != 0 {
				t.Errorf("anon_old_buyer unexpected evidence: %+v", member.Evidence)
			}
		}
	}
}

func TestEvaluator_ConvertedBuyerScenario(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eval := evaluator.New(db)

	asOfTime, _ := time.Parse(time.RFC3339, "2026-09-29T00:00:00.000Z")

	req := &model.PreviewRequest{
		Name:       "Recent purchasers",
		AsOf:       "2026-09-29T00:00:00.000Z",
		ParsedAsOf: asOfTime,
		Conditions: []model.Condition{
			{
				EventType:  model.EventTypePurchase,
				Operator:   model.OperatorAtLeast,
				Count:      1,
				WithinDays: 7,
			},
		},
	}

	resp, err := eval.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 1 {
		t.Fatalf("expected 1 purchaser, got %d", resp.Total)
	}

	if resp.Members[0].AnonymousID != "anon_buyer" {
		t.Errorf("expected anon_buyer, got %s", resp.Members[0].AnonymousID)
	}
	if resp.Members[0].Evidence[0].ObservedCount != 1 {
		t.Errorf("expected observed count 1, got %d", resp.Members[0].Evidence[0].ObservedCount)
	}
}

func TestEvaluator_EmptyMatchScenario(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eval := evaluator.New(db)

	asOfTime, _ := time.Parse(time.RFC3339, "2026-09-29T00:00:00.000Z")

	req := &model.PreviewRequest{
		Name:       "High volume buyers",
		AsOf:       "2026-09-29T00:00:00.000Z",
		ParsedAsOf: asOfTime,
		Conditions: []model.Condition{
			{
				EventType:  model.EventTypePurchase,
				Operator:   model.OperatorAtLeast,
				Count:      10,
				WithinDays: 7,
			},
		},
	}

	resp, err := eval.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("expected 0 matches, got %d", resp.Total)
	}
	if len(resp.Members) != 0 {
		t.Errorf("expected empty members slice, got len %d", len(resp.Members))
	}
}

func TestEvaluator_FutureEventsExcluded(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	eval := evaluator.New(db)

	// asOf before anon_future_events occurred
	asOfTime, _ := time.Parse(time.RFC3339, "2026-09-29T00:00:00.000Z")

	req := &model.PreviewRequest{
		Name:       "Future event check",
		AsOf:       "2026-09-29T00:00:00.000Z",
		ParsedAsOf: asOfTime,
		Conditions: []model.Condition{
			{
				EventType:  model.EventTypeProductView,
				Operator:   model.OperatorAtLeast,
				Count:      1,
				WithinDays: 1,
			},
		},
	}

	resp, err := eval.Evaluate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, member := range resp.Members {
		if member.AnonymousID == "anon_future_events" {
			t.Errorf("anon_future_events should not have matched asOf 2026-09-29T00:00:00Z")
		}
	}
}
