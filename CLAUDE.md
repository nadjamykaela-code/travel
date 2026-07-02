# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Travel Bot — automated flight-deal search running entirely within the GCP Always Free tier. Users create search filters via the web dashboard; a scheduled worker polls the Skyscanner API and notifies via email (SendGrid) and push (FCM) when matching flights appear.

## Commands

### Go (API + Worker + shared packages)

```bash
go run ./apps/api/cmd/main.go       # run API locally (port 8080)
go run ./apps/worker/cmd/main.go    # run worker locally (HTTP server, POST /run triggers a cycle)

go vet ./...
go test -race -count=1 ./...                       # full suite, as run in CI
go test -race -count=1 ./pkg/filters/...            # single package
go test -race -run TestApplyFilters ./pkg/filters/  # single test

E2E_TEST=1 go test -race -count=1 ./test/e2e/       # requires docker-compose stack running

go build -o /dev/null ./apps/api/cmd     # CI build check
go build -o /dev/null ./apps/worker/cmd
```

CI (`.github/workflows/ci.yml`) runs `go vet`, `go test -race`, `golangci-lint` (no repo-local config, uses defaults), and builds both binaries separately — always keep `apps/api` and `apps/worker` independently buildable.

### Web (`apps/web/`) — requires `bun`, npm is not supported

```bash
cd apps/web
bun install
bun run dev          # Vite dev server, :5173, proxies /api -> :8080
bun run build         # tsc -b && vite build
bun run test          # vitest run
bun run test:watch
```

### Docker (full stack)

```bash
cd infra/docker
cp .env.example .env   # fill in API keys
docker compose up --build
```

## Architecture

Three deployables sharing Go packages via a single module (`github.com/nadjamykaela-code/travel`):

- `apps/api/` — Gin HTTP API. Serves filter CRUD, auth verification, and place autocomplete to the web frontend. Entry point `apps/api/cmd/main.go`.
- `apps/worker/` — HTTP server triggered by Cloud Scheduler (`POST /run`) every 15 min; not a long-running cron loop itself. Entry point `apps/worker/cmd/main.go`.
- `apps/web/` — React 18 + TypeScript + Vite SPA, Firebase Auth for login, Axios for the API.

Shared packages:
- `pkg/clients/` — Skyscanner integrations: `SkyscannerSearcher` (Indicative v3, cheaper/cached), `LivePricingClient` (Live Pricing v3, real-time), and `FailoverSearcher`, which tries Live Pricing first and falls back to Indicative on error (see `pkg/clients/failover.go`). Also `PlaceSearchClient` for airport/place autocomplete.
- `pkg/filters/engine.go` — pure filtering logic (`ApplyFilters`/`FilterResults`) matching itineraries against a `models.Filter` on price, stops, duration, airline allow/deny lists, and departure/arrival time windows (±1h margin). No side effects — this is the easiest package to unit test in isolation.
- `pkg/models/` — `Filter` and `Itinerary`/flight types shared by API, worker, and clients, including `Filter.Validate()` and storage-estimate helpers.
- `pkg/notifications/` — SendGrid (email) and FCM (push) senders, each behind a small interface.
- `internal/config/` — env-var config loading + `Validate()`; both API and worker call this the same way (see `apps/api/cmd/main.go` and `apps/worker/cmd/main.go`).
- `internal/firestore/` — `firestore.New(ctx, projectID, credentialsPath)` returns a client wrapper (dependency-injected, not a package global) used by both apps.

### Worker execution flow (`apps/worker/job/runner.go`)

`Runner.Run` is driven entirely by small interfaces (`FilterLister`, `FlightSearcher`, `ResultFilter`, `Notifier`, `HistoryRecorder`, `APILimiter`) so it can be tested without Firestore or the network. Per filter: check the rate limiter → search via the failover searcher → apply `pkg/filters` → send notifications → record history. Errors from individual filters are collected with `errors.Join` and don't stop processing of the remaining filters.

### API composition (`apps/api/cmd/main.go`, `apps/api/routes/routes.go`)

`main.go` wires concrete implementations (Firestore-backed filter store, Firebase auth service, Skyscanner place client) into a `routes.Dependencies` struct, then `RegisterRoutes` constructs handlers/middleware from it. If Firebase auth credentials aren't available, `main.go` falls back to `service.NewNoopAuthService()`, which rejects all requests — auth endpoints degrade closed, not open. Middleware order matters: trace → metrics → CORS → (per-group) rate limit → auth.

### Frontend (`apps/web/src/`)

- `services/api.ts` — Axios instance with a Bearer-token interceptor reading the Firebase JWT from `localStorage`.
- `context/AuthContext.tsx` + `hooks/useAuth.ts` — Firebase Auth session state.
- `hooks/useFilters.ts` — filter CRUD against the API.
- `components/ProtectedRoute.tsx` — route guard redirecting unauthenticated users to `/login`.
- Routes: `/login`, `/dashboard` (protected), `*` → redirect to `/dashboard`.

### Resilience knobs (env vars, see `internal/config/config.go`)

Skyscanner calls are wrapped with timeout, retry, and circuit-breaker settings (`SKYSCANNER_TIMEOUT`, `SKYSCANNER_RETRIES`, `CB_MAX_REQUESTS`, `CB_INTERVAL`, `CB_TIMEOUT`) using `sony/gobreaker`. `Config.Validate()` requires `GCP_PROJECT_ID`, `SKYSCANNER_API_KEY`, and `PORT` — both API and worker exit at startup if these are missing.

## Deploy

Terraform (`infra/terraform/`) provisions Cloud Run (api + worker) and Cloud Scheduler. `.github/workflows/deploy.yml` authenticates via Workload Identity Federation (no static SA keys) — see the README's CI/CD section for required GitHub secrets/variables (`GCP_PROJECT_ID`, `WIF_PROVIDER`, `DEPLOY_SA`, `SKYSCANNER_API_KEY`, `SENDGRID_API_KEY`, `FCM_CREDENTIALS`).

Everything is designed to stay within GCP Always Free limits (Cloud Run 2M req/mo, Cloud Scheduler 3 free jobs, Firestore 1GB/50K reads per day) — keep this in mind when changing worker frequency, Firestore read/write patterns, or result history retention (`pkg/filters.TruncateResults` exists specifically to bound Firestore writes per run).
