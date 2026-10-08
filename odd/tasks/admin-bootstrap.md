# Feature: Admin bootstrap command (first administrator)

Status: WU1–WU3 done, verification green. Native review closed without receipt: RDD disabled for this clone by operator decision (see Progress).
Owner: el Gentleman (orchestrator) + delegated writer
Scope: `server/` + root config (`docker-compose.yml`, `README.md`)
Tracker: follow-up of the RBAC audit (2026-10-08). No GitHub issue.

## Context

The RBAC audit found that roles 5 (admin) and 6 (auditor) exist and are enforced, but a fresh
database has no admin and no code path can create one:

- Registration rejects roles 5–6 (`server/aplication/use-cases/user.go:39`, `IsRegistrationRole`).
- `cmd/seed` only creates a demo farmer and buyer; no admin or auditor.
- `SetUserRole` requires an existing admin and cannot assign `RoleAdmin`
  (`server/aplication/use-cases/report.go:335-346`).

Decision (user, 2026-10-08): implement **option 1** — an idempotent CLI bootstrap driven by
environment secrets. No HTTP surface.

## Decisions

1. **CLI only** (`server/cmd/bootstrap`), no endpoint. Secrets via `ADMIN_EMAIL` /
   `ADMIN_PASSWORD` env vars (`server/.env` locally, Render `sync: false` in deploy).
2. **Idempotent**: missing email → create; existing non-admin → promote; existing admin → no-op.
   **Never resets an existing password.**
3. **Password policy for the admin account**: minimum 8 characters (stricter than registration,
   which only checks non-empty).
4. **No DB audit entry in v1**: there is no "system actor" semantics in `AuditLog`; stdout plus
   the commit history are the evidence. Follow-up only if needed.
5. **No "only when zero admins" guard**: the command ensures the given email is an admin;
   running it later with another email is an explicit operator action.
6. **Placeholder phone on creation**: `UserRepositoryImpl.Save` requires a non-empty
   `PhoneNumber` (`userRepo.go:137`), so the create path stores a clearly fake placeholder
   (`+505 0000 0000`, documented in the use case). The operator can replace it later through
   the profile update endpoint. The promote path is untouched (real users always carry a phone).

## Tasks

- [x] WU1 — `BootstrapAdmin` use case + unit tests (test-first, RED → GREEN).
      Commit `31394be feat(admin): add idempotent admin bootstrap use case`
- [x] WU2 — `cmd/bootstrap` + Dockerfile + compose service + Makefile target + `.env.example`
      + README docs. Commit `5102e1e feat(admin): add bootstrap command for the first administrator`
- [x] WU3 — phone gap fix flagged by the writer (placeholder on create + test assertion).
      Commit `5333e8f fix(admin): set a placeholder phone on bootstrap-created admins`

## Acceptance criteria

- `make bootstrap` with env vars creates an admin; re-running is a no-op; promoting an existing
  non-admin works; missing email or short password fails with a non-zero exit.
- `go build ./...`, focused tests and `go vet ./...` pass.
- No password ever logged, committed or echoed.

## Checks

- `cd server && go build ./...`
- `cd server && go test ./aplication/use-cases/ -run 'TestBootstrap' -count=1 -v`
- `cd server && go test ./domain/... -count=1`
- `cd server && go vet ./...`

## Route declaration

| Task | Route | Trigger evidence |
| --- | --- | --- |
| WU1–WU3 | delegated direct (one writer) | 2 new non-trivial Go files + 1 test + 4 config/docs edits → writer trigger |

## Delivery strategy

`ask-on-risk` (default). Forecast: ~350 authored lines, below the ~400 budget; no chained PR
expected. Branch: `feat/admin-bootstrap` (from `develop`).

## Progress

- 2026-10-08: WU1–WU3 committed on `feat/admin-bootstrap` (4 commits: WU1–WU3 + feature doc,
  branch point `8a59ace`).
  Verification evidence (writer + parent spot check): `go build ./...` exit 0;
  `go test ./aplication/use-cases/ -run 'TestBootstrap' -count=1` → ok;
  `go test ./domain/... -count=1` → ok; `go vet ./...` exit 0;
  `docker compose --profile tools config -q` exit 0; `gofmt -l` clean.
- Native review attempt (2026-10-08): START created lineage `review-ca2611f7ba5ba1a6` (risk
  medium, one lens `review-reliability`, correction budget 200) after explicit consent. The
  lens could not complete: the reviewer model exhausted its reasoning budget (32k reasoning /
  0 output tokens, `reason: length`) on every attempt — 5/5 sessions across 2026-10-07/08.
  Native eligibility then reported `stop` only (`forbidden_manual_intervention_required`).
- Operator decision (2026-10-08): RDD disabled for this clone only
  (`gentle-ai review mode disable --scope clone`). Delivery follows ordinary repository policy;
  no receipt, no fabricated approval. The open lineage remains non-terminal and is not retried.
- Pending: end-to-end run of `make bootstrap` against a live database (needs the Docker stack
  and `ADMIN_EMAIL`/`ADMIN_PASSWORD` in `server/.env`). Not executed; unit + build evidence only.
