# Architectural Design Document

## 1. Data Model
The persistence layer uses SQLite via a pure-Go driver (`modernc.org/sqlite`), eliminating CGO/GCC toolchain dependencies while providing deterministic local evaluation across environments.

The schema separates users from events:
```sql
CREATE TABLE users (
    anonymous_id TEXT PRIMARY KEY,
    created_at   TEXT NOT NULL
);

CREATE TABLE events (
    id           TEXT PRIMARY KEY,
    anonymous_id TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    timestamp    TEXT NOT NULL,
    FOREIGN KEY(anonymous_id) REFERENCES users(anonymous_id)
);

CREATE INDEX idx_events_type_time ON events(event_type, timestamp, anonymous_id);
CREATE INDEX idx_events_user ON events(anonymous_id);
```

**Key Rationale:**
- Separating `users` from `events` provides a stable user universe. A visitor who performed zero events within a lookback window is still evaluated.
- The composite index `(event_type, timestamp, anonymous_id)` covers the exact lookup pattern used by windowed condition filters, allowing index range scans without full table scans.

---

## 2. Rule-Evaluation Choice
Audience conditions are combined with logical `AND`, supporting `at_least` and `exactly` operators.

### Dynamic SQL CTE Aggregation
Rather than pulling unaggregated event streams into application memory (which causes high GC pressure and OOM risks on large cohorts), evaluation is pushed entirely into SQLite via dynamically constructed Common Table Expressions (CTEs):

```sql
WITH user_universe AS (
  SELECT anonymous_id FROM users
),
c0 AS (
  SELECT anonymous_id, COUNT(*) AS cnt
  FROM events
  WHERE event_type = ? AND timestamp >= ? AND timestamp <= ?
  GROUP BY anonymous_id
)
SELECT u.anonymous_id, COALESCE(c0.cnt, 0)
FROM user_universe u
LEFT JOIN c0 ON u.anonymous_id = c0.anonymous_id
WHERE COALESCE(c0.cnt, 0) >= ?;
```

**Why CTEs over alternatives:**
- **Handling Non-Occurrence (`count == 0`):** Filtering directly (`WHERE event_type = 'purchase' GROUP BY anonymous_id HAVING count = 0`) fails because non-purchasers have no purchase rows to group. Left-joining each condition CTE to `user_universe` produces `NULL` for inactive users, which `COALESCE(..., 0)` turns into `0`.
- **Injection Safety:** All column aliases and CTE structures are static/programmatic, while event types, timestamp strings, and count thresholds are bound strictly as parameterized placeholders (`?`).
- **Granular Evidence:** Observed counts for each condition are projected directly in the `SELECT` clause, generating per-user evidence in a single database round-trip.

---

## 3. Time-Window Decision
The lookback period is defined as:
$$\text{Window} = [\text{asOf} - (\text{withinDays} \times 24\text{ hours}),\, \text{asOf}]$$

- **Inclusive Bounds:** Evaluated as `timestamp >= window_start AND timestamp <= asOf`. A user who viewed a product exactly at `asOf` or at `asOf - withinDays` is included.
- **UTC RFC 3339 Normalization:** All timestamps are parsed into Go `time.Time` and formatted as standard UTC RFC 3339 strings (`YYYY-MM-DDTHH:MM:SSZ`). Because ISO 8601 strings in UTC sort lexicographically identical to chronological order, SQLite string comparisons are exact.
- **Future Event Isolation:** Events with `timestamp > asOf` are strictly excluded. The engine evaluates historical state as it existed at `asOf`, guaranteeing reproducibility regardless of the current server clock.

---

## 4. Scaling and Product Trade-Offs

### Analytical Scaling vs. Relational Simplicity
SQLite is optimal for local development, reproducible tests, and intern assignment evaluation. However, in production with billions of behavioral events:
- **Write/Read Contention:** SQLite's file-level locking limits concurrent writes during real-time event streaming.
- **OLAP vs. OLTP:** Row-oriented relational databases suffer when scanning billions of rows to compute counts across multiple time windows.

### Recommended Production Architecture
1. **Columnar Event Storage (e.g., ClickHouse or DuckDB):** Store append-only events partitioned by month and hashed by `anonymous_id`. Columnar compression achieves 5–10x space reduction.
2. **Roaring Bitmaps for Pre-Aggregation:** Instead of scanning raw events at query time, pre-aggregate daily active user bitmaps by `(event_type, date)`. Lookback evaluation becomes bitwise `AND`/`OR` operations across daily bitmaps, completing audience previews across 100M+ users in tens of milliseconds.
3. **Async Streaming Architecture:** For interactive ad-hoc exploration, return fast count estimates using HyperLogLog; queue exact cohort extraction as asynchronous jobs dispatched to worker pools.
