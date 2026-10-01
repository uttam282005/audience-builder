package evaluator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mable/audience-builder/backend/internal/model"
	"github.com/mable/audience-builder/backend/internal/store"
)

// Evaluator evaluates audience conditions against the event database.
type Evaluator struct {
	db *store.DB
}

// New creates a new Evaluator.
func New(db *store.DB) *Evaluator {
	return &Evaluator{db: db}
}

// Evaluate processes an audience definition against the database and returns matching members with evidence.
func (e *Evaluator) Evaluate(ctx context.Context, req *model.PreviewRequest) (*model.PreviewResponse, error) {
	if len(req.Conditions) == 0 {
		return &model.PreviewResponse{
			Name:    req.Name,
			AsOf:    req.AsOf,
			Total:   0,
			Members: []model.Member{},
		}, nil
	}

	asOfUTC := req.ParsedAsOf.UTC()

	var cteParts []string
	var selectCols []string
	var leftJoins []string
	var whereClauses []string
	var args []any

	// 1. User universe CTE: start with known users
	cteParts = append(cteParts, "user_universe AS (SELECT anonymous_id FROM users)")

	// 2. Condition CTEs
	for i, cond := range req.Conditions {
		cteAlias := fmt.Sprintf("c%d", i)
		windowEnd := asOfUTC
		windowStart := asOfUTC.AddDate(0, 0, -cond.WithinDays)

		cteParts = append(cteParts, fmt.Sprintf(
			"%s AS (SELECT anonymous_id, COUNT(*) AS cnt FROM events WHERE event_type = ? AND timestamp >= ? AND timestamp <= ? GROUP BY anonymous_id)",
			cteAlias,
		))

		args = append(args, string(cond.EventType), windowStart.Format(time.RFC3339), windowEnd.Format(time.RFC3339))

		selectCols = append(selectCols, fmt.Sprintf("COALESCE(%s.cnt, 0)", cteAlias))
		leftJoins = append(leftJoins, fmt.Sprintf("LEFT JOIN %s ON u.anonymous_id = %s.anonymous_id", cteAlias, cteAlias))

		switch cond.Operator {
		case model.OperatorAtLeast:
			whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(%s.cnt, 0) >= ?", cteAlias))
		case model.OperatorExactly:
			whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(%s.cnt, 0) = ?", cteAlias))
		default:
			return nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
		}
	}

	// 3. Append comparison threshold arguments
	for _, cond := range req.Conditions {
		args = append(args, cond.Count)
	}

	// 4. Assemble complete SQL query
	query := fmt.Sprintf(
		"WITH %s\nSELECT u.anonymous_id, %s\nFROM user_universe u\n%s\nWHERE %s\nORDER BY u.anonymous_id ASC",
		strings.Join(cteParts, ",\n"),
		strings.Join(selectCols, ", "),
		strings.Join(leftJoins, "\n"),
		strings.Join(whereClauses, " AND "),
	)

	rows, err := e.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("evaluator query execution failed: %w", err)
	}
	defer rows.Close()

	members := make([]model.Member, 0)
	numConditions := len(req.Conditions)

	for rows.Next() {
		var anonID string
		counts := make([]int, numConditions)
		destinations := make([]any, 1+numConditions)
		destinations[0] = &anonID
		for j := 0; j < numConditions; j++ {
			destinations[j+1] = &counts[j]
		}

		if err := rows.Scan(destinations...); err != nil {
			return nil, fmt.Errorf("failed scanning member row: %w", err)
		}

		evidence := make([]model.EvidenceItem, numConditions)
		for j := 0; j < numConditions; j++ {
			evidence[j] = model.EvidenceItem{
				EventType:     req.Conditions[j].EventType,
				ObservedCount: counts[j],
			}
		}

		members = append(members, model.Member{
			AnonymousID: anonID,
			Evidence:    evidence,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading member rows: %w", err)
	}

	return &model.PreviewResponse{
		Name:    req.Name,
		AsOf:    req.AsOf,
		Total:   len(members),
		Members: members,
	}, nil
}
