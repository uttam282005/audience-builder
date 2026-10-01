# Mable Audience Builder — Master Engineering Specification

## 1. Overview & Objective

Mable helps teams turn raw product telemetry into actionable decisions. A foundational workflow is segmenting anonymous visitors into cohorts ("audiences") based on behavioral patterns across time, and providing clear, auditable evidence for why any given member qualified.

This project implements the **Mable Audience Builder**, consisting of:
1. **An independently runnable Go backend**: Evaluates rule conditions against synthetic event data stored in SQLite relative to a reproducible `asOf` timestamp, returning matching anonymous users with granular matching evidence.
2. **An independently runnable TypeScript/React frontend**: Provides an accessible, responsive operator interface to define rules, preview audiences, inspect membership evidence, and handle loading/validation/error states with retry.
3. **Architecture documentation**: Includes a concise design document (`docs/DESIGN.md`, ≤750 words) defending data modeling, evaluation mechanics, time boundary semantics, and scaling trade-offs, alongside an AI usage log (`docs/AI_USAGE.md`) and comprehensive `README.md`.

---

## 2. Core Functional Requirements

### 2.1 Supported Event Types
All activity is synthetic and anonymous. The engine supports five distinct event types:
- `page_view`: Page visits.
- `product_view`: Individual product inspections.
- `add_to_cart`: Items added to shopping carts.
- `checkout_started`: Initiation of checkout flow.
- `purchase`: Successful transaction completion.

### 2.2 Rule Specification & Condition Semantics
An audience rule comprises:
- `name` (string): Human-readable title (e.g. `"Viewed but not purchased"`).
- `asOf` (ISO 8601 string): Reference point in time making evaluation completely reproducible. Events occurring after `asOf` are disregarded.
- `conditions` (array): One or more behavioral criteria, **combined strictly using logical AND**.
  - `eventType` (string): Must match one of the 5 allowed event types.
  - `operator` (string): Supported operators:
    - `at_least`: Observed count $\ge$ specified count.
    - `exactly`: Observed count $=$ specified count (crucially supporting `exactly 0` for non-occurrence).
  - `count` (integer): Non-negative integer ($\ge 0$).
  - `withinDays` (integer): Rolling lookback window before `asOf` in days ($\ge 1$).

### 2.3 Time Window Semantics
For any condition with lookback `withinDays` and reference time `asOf`:
$$\text{Window} = [\text{asOf} - (\text{withinDays} \times 24\text{ hours}),\, \text{asOf}]$$
- The boundary is inclusive on both ends: `timestamp >= window_start AND timestamp <= asOf`.
- All timestamps are stored and evaluated in UTC RFC 3339 format to ensure exact chronological ordering.
- Events with `timestamp > asOf` are excluded to maintain reproducible evaluation regardless of system clock.

---

## 3. API Specification

### 3.1 `POST /v1/audiences/preview`

#### Request Payload
```json
{
  "name": "Viewed but not purchased",
  "asOf": "2026-09-29T00:00:00.000Z",
  "conditions": [
    {
      "eventType": "product_view",
      "operator": "at_least",
      "count": 2,
      "withinDays": 7
    },
    {
      "eventType": "purchase",
      "operator": "exactly",
      "count": 0,
      "withinDays": 7
    }
  ]
}
```

#### Success Response (`200 OK`)
```json
{
  "name": "Viewed but not purchased",
  "asOf": "2026-09-29T00:00:00.000Z",
  "total": 2,
  "members": [
    {
      "anonymousId": "anon_123",
      "evidence": [
        { "eventType": "product_view", "observedCount": 3 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    },
    {
      "anonymousId": "anon_456",
      "evidence": [
        { "eventType": "product_view", "observedCount": 2 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    }
  ]
}
```

#### Validation Error Response (`400 Bad Request`)
```json
{
  "error": "validation_error",
  "message": "Invalid audience rule definition",
  "details": [
    {
      "field": "conditions[0].count",
      "issue": "count must be greater than or equal to 0"
    }
  ]
}
```

### 3.2 `GET /health`

#### Success Response (`200 OK`)
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2026-10-01T21:00:00Z"
}
```

---

## 4. Backend Architecture & Evaluation Engine

### 4.1 Database Design (SQLite)
SQLite provides lightweight, self-contained persistence without requiring external infrastructure.
```sql
CREATE TABLE IF NOT EXISTS users (
    anonymous_id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    anonymous_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    FOREIGN KEY(anonymous_id) REFERENCES users(anonymous_id)
);

CREATE INDEX IF NOT EXISTS idx_events_type_time ON events(event_type, timestamp, anonymous_id);
CREATE INDEX IF NOT EXISTS idx_events_user ON events(anonymous_id);
```

### 4.2 Dynamic CTE Evaluation Query
To avoid loading all events into application memory or miscounting non-occurring events (`count == 0`), the evaluator dynamically builds a parameterized Common Table Expression (CTE) query:

```sql
WITH user_universe AS (
  SELECT anonymous_id FROM users
),
c0 AS (
  SELECT anonymous_id, COUNT(*) AS cnt
  FROM events
  WHERE event_type = ? AND timestamp >= ? AND timestamp <= ?
  GROUP BY anonymous_id
),
c1 AS (
  SELECT anonymous_id, COUNT(*) AS cnt
  FROM events
  WHERE event_type = ? AND timestamp >= ? AND timestamp <= ?
  GROUP BY anonymous_id
)
SELECT
  u.anonymous_id,
  COALESCE(c0.cnt, 0) AS count_0,
  COALESCE(c1.cnt, 0) AS count_1
FROM user_universe u
LEFT JOIN c0 ON u.anonymous_id = c0.anonymous_id
LEFT JOIN c1 ON u.anonymous_id = c1.anonymous_id
WHERE
  COALESCE(c0.cnt, 0) >= ?
  AND COALESCE(c1.cnt, 0) = ?
ORDER BY u.anonymous_id ASC;
```

#### Key Guarantees:
1. **Non-occurrence correctness**: A user with 0 events of a given type produces `NULL` in the `LEFT JOIN`, coalesced safely to `0`. If the rule requires `exactly 0`, `0 = 0` evaluates to true.
2. **SQL Injection Safety**: All variables (event types, ISO timestamp strings, comparison integers) are bound through positional parameters (`?`). Table and CTE aliases are generated internally.
3. **Evidence Extraction**: The SELECT clause returns the exact count observed for each condition, allowing direct population of the per-user evidence list.

---

## 5. Synthetic Persona Matrix (Seed Data)
Seeded with reference date `2026-09-29T00:00:00Z` to support verification of the canonical scenario:
*"Users who viewed a product at least twice in the previous 7 days but did not purchase in that period."*

| Persona ID | Description | 7-day Views | 7-day Purchases | Other Events / Notes | Matches Sample Rule? |
| :--- | :--- | :---: | :---: | :--- | :---: |
| `anon_target_1` | Canonical active shopper | 3 | 0 | 1 `add_to_cart` | **YES** |
| `anon_target_2` | Exact threshold shopper | 2 | 0 | None | **YES** |
| `anon_buyer` | Converted buyer | 3 | 1 | 1 `checkout_started`, 1 `purchase` | **NO** (purchase $\ne$ 0) |
| `anon_insufficient` | Casual single view | 1 | 0 | 1 `page_view` | **NO** (views $< 2$) |
| `anon_stale_views` | Past shopper outside window | 0 | 0 | 3 views 8-10 days ago | **NO** (stale) |
| `anon_future_events` | Activity post-asOf | 0 | 0 | 2 views dated 2026-09-30 | **NO** (future filtered) |
| `anon_exact_boundary` | Exact window boundaries | 2 | 0 | 1 view at $t-7d$, 1 view at $t$ | **YES** (boundary hit) |
| `anon_old_buyer` | Buyer outside window | 2 | 0 | 1 purchase 14 days ago | **YES** (purchase out of window) |
| `anon_other_only` | Non-product visitors | 0 | 0 | 4 `page_view`, 1 `add_to_cart` | **NO** |

---

## 6. Frontend Operator Interface

- **Rule Definition Form**:
  - Audience name input with instant feedback.
  - Reference `asOf` date picker with one-click presets:
    - *Default Dataset As-Of* (`2026-09-29T00:00:00Z`)
    - *Current Time*
  - Condition Builder:
    - Add Condition / Remove Condition buttons.
    - Accessible `<select>` for Event Type and Operator (`at_least`, `exactly`).
    - Numeric inputs for Count ($\ge 0$) and Lookback Days ($\ge 1$).
    - "Load Preset: Viewed But Not Purchased" button for instant 1-click test flow.
- **Results View**:
  - Audience size badge (`total`).
  - Tabular / card display of matched `anonymousId`s.
  - Per-user evidence badges showing observed count for each condition.
- **UI State Machine**:
  - `idle`: Friendly empty-state prompt before submission.
  - `loading`: Accessible spinner with `aria-live="polite"`.
  - `success`: Render total members and evidence breakdown.
  - `validation_error`: Clear inline messages for invalid inputs.
  - `api_error`: Distinct alert banner displaying backend error details with an interactive **Retry** button.

---

## 7. Quality & Verification Standards

1. **Automated Backend Tests**:
   - `validator_test.go`: Boundary, type, and edge-case validation.
   - `evaluator_test.go`: SQL query evaluation across all synthetic personas, zero-case handling, and boundary timestamps.
   - `handlers_test.go`: HTTP handler contract tests for `/health` and `/v1/audiences/preview`.
2. **Frontend Verification**:
   - Strict TypeScript compilation (`tsc --noEmit`).
   - Clean production build (`npm run build`).
3. **Commit Progression**:
   - Commit 1: `chore: initialize repository and master specification`
   - Commit 2: `feat(backend): implement sqlite storage, seeder, cte evaluator, and api handlers`
   - Commit 3: `feat(frontend): build operator audience builder interface with evidence preview`
   - Commit 4: `docs: add README, DESIGN.md, and AI_USAGE.md`
