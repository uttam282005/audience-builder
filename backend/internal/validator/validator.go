package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/mable/audience-builder/backend/internal/model"
)

// ValidatePreviewRequest performs strict validation on incoming preview requests.
// It populates req.ParsedAsOf on success.
func ValidatePreviewRequest(req *model.PreviewRequest) []model.FieldError {
	var errs []model.FieldError

	// Validate rule name
	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" {
		errs = append(errs, model.FieldError{
			Field: "name",
			Issue: "name is required and cannot be empty",
		})
	} else if len(trimmedName) > 256 {
		errs = append(errs, model.FieldError{
			Field: "name",
			Issue: "name must not exceed 256 characters",
		})
	}

	// Validate asOf timestamp
	trimmedAsOf := strings.TrimSpace(req.AsOf)
	if trimmedAsOf == "" {
		errs = append(errs, model.FieldError{
			Field: "asOf",
			Issue: "asOf is required and must be an ISO 8601 / RFC 3339 timestamp",
		})
	} else {
		parsed, err := time.Parse(time.RFC3339, trimmedAsOf)
		if err != nil {
			// Try RFC3339Nano in case sub-second precision is provided
			parsed, err = time.Parse(time.RFC3339Nano, trimmedAsOf)
		}
		if err != nil {
			// Also accept date-only ISO 8601 (YYYY-MM-DD)
			parsed, err = time.Parse("2006-01-02", trimmedAsOf)
		}
		if err != nil {
			errs = append(errs, model.FieldError{
				Field: "asOf",
				Issue: fmt.Sprintf("invalid timestamp format '%s', expected RFC 3339 or ISO date (e.g. 2026-09-29T00:00:00.000Z or 2026-09-29)", trimmedAsOf),
			})
		} else {
			req.ParsedAsOf = parsed.UTC()
		}
	}

	// Validate conditions array
	if len(req.Conditions) == 0 {
		errs = append(errs, model.FieldError{
			Field: "conditions",
			Issue: "at least one condition must be specified",
		})
	} else if len(req.Conditions) > 25 {
		errs = append(errs, model.FieldError{
			Field: "conditions",
			Issue: "cannot specify more than 25 conditions",
		})
	} else {
		for i, cond := range req.Conditions {
			prefix := fmt.Sprintf("conditions[%d]", i)

			if !model.ValidEventTypes[cond.EventType] {
				errs = append(errs, model.FieldError{
					Field: prefix + ".eventType",
					Issue: fmt.Sprintf("invalid eventType '%s', must be one of: page_view, product_view, add_to_cart, checkout_started, purchase", cond.EventType),
				})
			}

			if !model.ValidOperators[cond.Operator] {
				errs = append(errs, model.FieldError{
					Field: prefix + ".operator",
					Issue: fmt.Sprintf("invalid operator '%s', must be 'at_least' or 'exactly'", cond.Operator),
				})
			}

			if cond.Count < 0 {
				errs = append(errs, model.FieldError{
					Field: prefix + ".count",
					Issue: fmt.Sprintf("count must be a non-negative integer, got %d", cond.Count),
				})
			}

			if cond.WithinDays < 1 {
				errs = append(errs, model.FieldError{
					Field: prefix + ".withinDays",
					Issue: fmt.Sprintf("withinDays must be at least 1 day, got %d", cond.WithinDays),
				})
			} else if cond.WithinDays > 365 {
				errs = append(errs, model.FieldError{
					Field: prefix + ".withinDays",
					Issue: fmt.Sprintf("withinDays cannot exceed 365 days, got %d", cond.WithinDays),
				})
			}
		}
	}

	return errs
}
