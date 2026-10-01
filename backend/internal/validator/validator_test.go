package validator_test

import (
	"testing"
	"time"

	"github.com/mable/audience-builder/backend/internal/model"
	"github.com/mable/audience-builder/backend/internal/validator"
)

func TestValidatePreviewRequest_Valid(t *testing.T) {
	req := &model.PreviewRequest{
		Name: "Viewed but not purchased",
		AsOf: "2026-09-29T00:00:00.000Z",
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

	errs := validator.ValidatePreviewRequest(req)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d: %+v", len(errs), errs)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2026-09-29T00:00:00.000Z")
	if !req.ParsedAsOf.Equal(expectedTime) {
		t.Errorf("expected ParsedAsOf %v, got %v", expectedTime, req.ParsedAsOf)
	}
}

func TestValidatePreviewRequest_InvalidCases(t *testing.T) {
	tests := []struct {
		name          string
		req           model.PreviewRequest
		expectedField string
	}{
		{
			name: "empty name",
			req: model.PreviewRequest{
				Name: "   ",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: model.OperatorAtLeast, Count: 1, WithinDays: 7},
				},
			},
			expectedField: "name",
		},
		{
			name: "invalid asOf format",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "not-a-timestamp",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: model.OperatorAtLeast, Count: 1, WithinDays: 7},
				},
			},
			expectedField: "asOf",
		},
		{
			name: "empty conditions list",
			req: model.PreviewRequest{
				Name:       "Valid name",
				AsOf:       "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{},
			},
			expectedField: "conditions",
		},
		{
			name: "invalid eventType",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: "random_click", Operator: model.OperatorAtLeast, Count: 1, WithinDays: 7},
				},
			},
			expectedField: "conditions[0].eventType",
		},
		{
			name: "invalid operator",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: "greater_than", Count: 1, WithinDays: 7},
				},
			},
			expectedField: "conditions[0].operator",
		},
		{
			name: "negative count",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: model.OperatorAtLeast, Count: -1, WithinDays: 7},
				},
			},
			expectedField: "conditions[0].count",
		},
		{
			name: "zero withinDays",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: model.OperatorAtLeast, Count: 1, WithinDays: 0},
				},
			},
			expectedField: "conditions[0].withinDays",
		},
		{
			name: "excessive withinDays",
			req: model.PreviewRequest{
				Name: "Valid name",
				AsOf: "2026-09-29T00:00:00.000Z",
				Conditions: []model.Condition{
					{EventType: model.EventTypeProductView, Operator: model.OperatorAtLeast, Count: 1, WithinDays: 400},
				},
			},
			expectedField: "conditions[0].withinDays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidatePreviewRequest(&tt.req)
			if len(errs) == 0 {
				t.Fatalf("expected validation error for %s, got none", tt.name)
			}
			found := false
			for _, e := range errs {
				if e.Field == tt.expectedField {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error field '%s', got %+v", tt.expectedField, errs)
			}
		})
	}
}
