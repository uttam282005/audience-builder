# AI Usage Log

## 1. Overview
In accordance with the assignment guidelines, this document transparently records the AI tools leveraged during the development of the **Mable Audience Builder**, their specific contributions, and the human oversight applied to ensure correctness, code quality, and architectural soundess.

---

## 2. Tools Used
- **Google Antigravity CLI (Gemini 3.8 Flash)**: Interactive pair-programming agent used for requirements analysis, specification drafting, code scaffolding, test case generation, and documentation.

---

## 3. Roles and Contributions

### 3.1 Requirements Analysis & Specification
- Extracted and systematized functional constraints from the assignment PDF into a master engineering specification (`SPEC.md`).
- Designed the synthetic persona matrix covering positive matches, threshold boundaries, non-occurrences (`count == 0`), stale events, and future event exclusions.

### 3.2 Backend Engineering & Query Strategy
- Formulated the dynamic Common Table Expression (CTE) query strategy using `LEFT JOIN` and `COALESCE(count, 0)` in SQLite to solve the non-occurrence problem without memory-heavy in-process aggregation.
- Scaffolding Go domain structs, `net/http` route handlers, structured JSON logging (`log/slog`), and CORS middleware.
- Generated unit and integration tests across domain validation (`validator_test.go`), database seeding (`seed_test.go`), evaluator mechanics (`evaluator_test.go`), and HTTP endpoints (`handlers_test.go`).

### 3.3 Frontend Architecture & Accessible UX
- Scaffolded Vite + React + TypeScript application structure.
- Implemented accessible semantic controls (native `<select>`, `<input>`, keyboard navigation, ARIA live regions for loading/results).
- Built comprehensive UI state handling (idle, loading, zero matches, active cohort with evidence chips, and API error state with visible retry).

### 3.4 Documentation
- Drafted the concise design rationale (`docs/DESIGN.md`, kept strictly under the 750-word limit) defending data modeling, SQL CTE mechanics, inclusive time window semantics, and columnar/bitmap scaling strategies.
- Drafted the root `README.md` with explicit prerequisites, build/run instructions, API documentation, and test verification procedures.

---

## 4. Human Oversight and Verification
All AI-generated code, schemas, and queries were subject to direct engineering verification:
1. **Zero CGO Requirement**: Explicitly chose `modernc.org/sqlite` instead of `mattn/go-sqlite3` to guarantee CGO-free, cross-platform reproducibility without requiring external C compilers.
2. **Correctness Auditing**: Verified SQL generation for injection vulnerabilities by ensuring positional parameter binding (`?`) for all dynamic condition values.
3. **Automated Verification**: Ran full Go test suites (`go test -v ./...`) and TypeScript compiler checks (`tsc --noEmit`, `npm run build`) ensuring zero warnings or build errors.
4. **End-to-End Testing**: Performed manual verification across both the HTTP API via `curl` and the browser UI to validate retry paths and evidence rendering.
