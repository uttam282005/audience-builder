package model

import "time"

// EventType represents the supported anonymous event categories.
type EventType string

const (
	EventTypePageView        EventType = "page_view"
	EventTypeProductView     EventType = "product_view"
	EventTypeAddToCart       EventType = "add_to_cart"
	EventTypeCheckoutStarted EventType = "checkout_started"
	EventTypePurchase        EventType = "purchase"
)

// ValidEventTypes contains the whitelist of supported event types.
var ValidEventTypes = map[EventType]bool{
	EventTypePageView:        true,
	EventTypeProductView:     true,
	EventTypeAddToCart:       true,
	EventTypeCheckoutStarted: true,
	EventTypePurchase:        true,
}

// Operator defines comparison logic against the condition count.
type Operator string

const (
	OperatorAtLeast Operator = "at_least"
	OperatorExactly Operator = "exactly"
)

// ValidOperators contains the whitelist of supported comparison operators.
var ValidOperators = map[Operator]bool{
	OperatorAtLeast: true,
	OperatorExactly: true,
}

// Condition defines a single event-based filter for an audience.
type Condition struct {
	EventType  EventType `json:"eventType"`
	Operator   Operator  `json:"operator"`
	Count      int       `json:"count"`
	WithinDays int       `json:"withinDays"`
}

// PreviewRequest is the incoming payload for POST /v1/audiences/preview.
type PreviewRequest struct {
	Name       string      `json:"name"`
	AsOf       string      `json:"asOf"`
	Conditions []Condition `json:"conditions"`

	// ParsedAsOf holds the validated time parsed from AsOf.
	ParsedAsOf time.Time `json:"-"`
}

// EvidenceItem documents the observed event count for a user against a condition.
type EvidenceItem struct {
	EventType     EventType `json:"eventType"`
	ObservedCount int       `json:"observedCount"`
}

// Member represents an anonymous user who qualified for the audience.
type Member struct {
	AnonymousID string         `json:"anonymousId"`
	Evidence    []EvidenceItem `json:"evidence"`
}

// PreviewResponse is the response payload for POST /v1/audiences/preview.
type PreviewResponse struct {
	Name    string   `json:"name"`
	AsOf    string   `json:"asOf"`
	Total   int      `json:"total"`
	Members []Member `json:"members"`
}

// FieldError provides structured validation issue details.
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// ErrorResponse defines the standard error envelope returned by the API.
type ErrorResponse struct {
	Error   string       `json:"error"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// Event represents a row in the database events table.
type Event struct {
	ID          string    `json:"id"`
	AnonymousID string    `json:"anonymousId"`
	EventType   EventType `json:"eventType"`
	Timestamp   time.Time `json:"timestamp"`
}

// User represents a row in the database users table.
type User struct {
	AnonymousID string    `json:"anonymousId"`
	CreatedAt   time.Time `json:"createdAt"`
}
