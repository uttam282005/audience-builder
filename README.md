# Mable Audience Builder

A lightweight, production-grade application for defining behavioral audience cohorts from anonymous customer telemetry and inspecting matching evidence.

Built for the **Mable Software Engineering Intern Assignment**.

---

## Live Deployments

- **Frontend Application (Vercel)**: [https://audience-builder-two.vercel.app](https://audience-builder-two.vercel.app)
- **Backend API (Render)**: [https://audience-builder-5fci.onrender.com](https://audience-builder-5fci.onrender.com)
  - Health check: `https://audience-builder-5fci.onrender.com/health`
  - Audience preview: `POST https://audience-builder-5fci.onrender.com/v1/audiences/preview`

---

## Architecture Overview

The system consists of two independently runnable applications:
- **Backend (`backend/`)**: A Go HTTP service powered by the Go standard library and SQLite (`modernc.org/sqlite`, 100% pure Go, no CGO required). Evaluates audience conditions using dynamic Common Table Expression (CTE) queries, correctly accounting for non-occurrence (`count == 0`) and temporal reproducibility via an `asOf` timestamp.
- **Frontend (`frontend/`)**: A React + TypeScript operator interface built with Vite. Features keyboard-operable condition builders, accessible semantic HTML, real-time backend health monitoring, and comprehensive states (idle, loading, validation, empty results, and error with retry).

```
audience-builder/
├── SPEC.md                    # Master engineering specification
├── README.md                  # Prerequisites, run/test commands, and user guide
├── docs/
│   ├── DESIGN.md              # Architecture design document (≤750 words)
│   └── AI_USAGE.md            # AI tools usage log and verification audit
├── backend/                   # Go HTTP backend & SQLite evaluator
│   ├── cmd/server/main.go     # Server entry point
│   ├── internal/api/          # HTTP handlers, router & middleware
│   ├── internal/evaluator/    # Dynamic CTE audience evaluation engine
│   ├── internal/model/        # Domain structs & request/response types
│   ├── internal/store/        # SQLite migrations & deterministic seeder
│   └── internal/validator/    # Input boundary validator
└── frontend/                  # React + TypeScript Vite frontend
    ├── src/api/               # Typed API client with retry support
    ├── src/components/        # Accessible UI components (Form, Results, Alerts)
    └── src/types/             # TypeScript domain interfaces
```

---

## Prerequisites

- **Go**: 1.22 or higher (`go version`)
- **Node.js**: 18 or higher (`node -v` and `npm -v`)
- **No external database required**: Uses self-contained SQLite with pure Go drivers.

---

## Quickstart (Running Locally)

### 1. Start the Backend Server

From the repository root:

```bash
cd backend
go run ./cmd/server
```

By default, the backend:
- Starts on `http://localhost:8080`.
- Automatically initializes and seeds `audience.db` with synthetic personas if the database is empty.
- Responds to `GET /health` and `POST /v1/audiences/preview`.

> **Optional flags:**
> - `-port=8080`: HTTP port (or `PORT` environment variable).
> - `-db=audience.db`: SQLite database file path (or `:memory:`).
> - `-reset-seed=true`: Drop existing data and reseed fresh synthetic data.

### 2. Start the Frontend Application

In a new terminal window:

```bash
cd frontend
npm install
npm run dev
```

The frontend will start at `http://localhost:5173`. Open this URL in your web browser.

---

## Running Automated Tests

### Backend Test Suite
Runs unit and integration tests across domain validation, store seeding, CTE query evaluation, and HTTP endpoints:

```bash
cd backend
go test -v ./...
```

To run with race detection:
```bash
cd backend
go test -v -race ./...
```

### Frontend Type Check & Build
Ensures strict TypeScript compilation and production bundle build:

```bash
cd frontend
npm run build
```

---

## How to Preview an Audience

### Option A: Using the Operator Web Interface
1. Open `http://localhost:5173` in your browser.
2. Confirm the header shows **Backend: Connected** (green indicator).
3. Under the **Presets** bar, click **"Viewed but not purchased (Sample)"**.
   - This automatically populates:
     - **Name**: `Viewed but not purchased`
     - **asOf**: `2026-09-29T00:00:00.000Z`
     - **Condition 1**: `product_view` $\ge$ 2 within 7 days.
     - **Condition 2**: `purchase` $=$ 0 within 7 days.
4. Click the **"Preview Audience"** button.
5. Inspect the **Audience Preview** panel on the right:
   - Total matched members (`4` matching users).
   - Each matched anonymous user (e.g. `anon_target_1`, `anon_target_2`, `anon_exact_boundary`, `anon_old_buyer`).
   - Per-user evidence badges showing observed counts for both conditions (e.g. `Product View: 3`, `Purchase: 0`).
6. Try modifying conditions or clicking **"Converted Buyers"** preset to observe instant re-evaluation.
7. Test error resilience: Stop the backend process and click **"Preview Audience"** — an alert banner with a visible **"Retry Request"** button will appear. Restart the backend and click **Retry Request** to verify immediate recovery.

---

### Option B: Using `curl` (API Direct)

To evaluate the canonical scenario against the backend:

```bash
curl -s -X POST http://localhost:8080/v1/audiences/preview \
  -H "Content-Type: application/json" \
  -d '{
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
  }' | jq .
```

#### Example Response:
```json
{
  "name": "Viewed but not purchased",
  "asOf": "2026-09-29T00:00:00.000Z",
  "total": 4,
  "members": [
    {
      "anonymousId": "anon_exact_boundary",
      "evidence": [
        { "eventType": "product_view", "observedCount": 2 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    },
    {
      "anonymousId": "anon_old_buyer",
      "evidence": [
        { "eventType": "product_view", "observedCount": 2 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    },
    {
      "anonymousId": "anon_target_1",
      "evidence": [
        { "eventType": "product_view", "observedCount": 3 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    },
    {
      "anonymousId": "anon_target_2",
      "evidence": [
        { "eventType": "product_view", "observedCount": 2 },
        { "eventType": "purchase", "observedCount": 0 }
      ]
    }
  ]
}
```

#### Health Check:
```bash
curl -i http://localhost:8080/health
```
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2026-10-01T21:15:00Z"
}
```

---

## Seed Data & Personas

The seed dataset contains deterministic test personas calibrated against reference date `2026-09-29T00:00:00Z`:

| Persona ID | Behavioral Pattern in Lookback Window | Qualifying for Sample Rule? |
| :--- | :--- | :--- |
| `anon_target_1` | 3 product views, 0 purchases | **Yes** (Active browser) |
| `anon_target_2` | 2 product views, 0 purchases | **Yes** (Exact threshold) |
| `anon_exact_boundary` | 1 view at $t-7d$, 1 view at $t$, 0 purchases | **Yes** (Boundary verification) |
| `anon_old_buyer` | 2 views in window, purchase was 15 days ago | **Yes** (Purchase outside window) |
| `anon_buyer` | 2 product views, 1 purchase in window | **No** (Purchase count violated) |
| `anon_insufficient` | 1 product view in window | **No** (View threshold not met) |
| `anon_stale_views` | 2 product views 8+ days ago | **No** (Events outside lookback) |
| `anon_future_events` | 2 product views occurring after `asOf` | **No** (Future events excluded) |
| `anon_other_only` | Page views and cart adds only | **No** (No product views) |

---

## Documentation Links

- [Master Engineering Specification (SPEC.md)](SPEC.md)
- [Architectural Design Document (docs/DESIGN.md)](docs/DESIGN.md)
- [AI Usage Log (docs/AI_USAGE.md)](docs/AI_USAGE.md)
