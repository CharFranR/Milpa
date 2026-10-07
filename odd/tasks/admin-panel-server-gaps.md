# Feature: Admin Panel — server-side gaps (epic #25)

Status: WU1-WU5 done, `API.md` updated, WU6 partially done. Nothing committed.
Owner: el Gentleman (orchestrator) + delegated writers
Scope: `server/` only
Tracker: https://github.com/CharFranR/Milpa/issues/25 — `[Sprint 3] Epic: Admin Panel` (RF-17)

## Context

Epic #25 is still open while its implementation is mostly shipped. A read-only audit
(2026-10-07) mapped the six acceptance criteria against the code:

| Criterion | Status before this feature | Evidence |
| --- | --- | --- |
| List/suspend/activate users | done | `server/infrastructure/adapters/primary/api/router.go:179-181`, `server/aplication/use-cases/report.go:289` |
| Category (product type) management | done | `server/aplication/use-cases/category.go:44,76,117`, `router.go:184-186` |
| Units of measure management | **missing** | only `server/aplication/use-cases/unit_of_measure.go:19` and public `GET /units-of-measure` (`router.go:62`) |
| Reports inbox with moderation actions | done | `server/aplication/use-cases/report.go:94,141`, `router.go:120-123` |
| Admin-role-only access | done but fragile | role check duplicated inside each use case (`report.go:99,146,293…`, `category.go:51,83,124`); no route-level guard |
| Tests for admin endpoints | partial | categories and moderation covered; `/admin/users`, `/admin/users/{id}/role`, `/admin/audit-logs`, `DELETE /admin/offerings` and router-level `/reports` have none |

This feature closes those four gaps. The frontend is out of scope by explicit user
decision ("nosotros únicamente nos encargamos del servidor").

## Authorization

User instruction: "Nosotros únicamente nos encargamos del servidor, completemos los gaps."
The user cancelled the scope questionnaire, so the orchestrator fixed the decisions below
and recorded them here for review before any source write.

## Decisions

1. **Scope = the four server gaps.** Units CRUD, admin guard, missing tests, activity endpoint.
   WU4 (activity endpoint) is last and independently droppable: it is not one of the six
   acceptance criteria, it comes from the epic's "Alcance" section.
2. **Unit deactivation is a soft delete, mirroring `Category.SetStatus`.** No reference
   counting against `categories.default_unit_of_measure_id` or `offerings.unit_of_measure_id`.
   Deactivated units leave the public list; existing references stay intact.
   Rationale: both FKs are `ON DELETE SET NULL`, so a hard delete would silently blank live
   references, and the repo already has an established soft-delete convention for catalogue rows.
3. **`GET /admin/units-of-measure` (admin, includes inactive) is added.** Without it a
   deactivated unit becomes unreachable, which would make "manage units" unfulfillable.
4. **The admin guard applies to mutating routes.** Every POST/PATCH/DELETE under `/admin` and
   `PATCH /reports/{id}/action` gains `adminMW.RequireAdmin`. The `GET` routes keep
   `auditorMW.CheckReadOnly` because that chain is keyed to the read verb, not to a grant.
   Corrected during WU3: `ListUsers`, `ListAuditLogs` and `ReportUseCase.List` gate on
   `RoleAdmin` inside the use case, so an auditor's `GET` still ends in 403 no matter what the
   route chain says. The epic asks for admin-only endpoints, so the effective behaviour is
   compliant; but the auditor role has no server-side read path over the admin surface at all,
   which contradicts `frontend/src/App.jsx:172` (`RequireRole(['admin', 'auditor'])`). That is a
   privilege-model change, out of scope here, escalated to the user instead of fixed silently.
5. **Use-case admin checks stay.** The new middleware is defense in depth, not a replacement;
   the existing use-case tests keep passing unchanged.
6. **The guard middleware is constructed inside `NewRouter`**, exactly like `auditorMW`
   (`router.go:113`). This avoids touching the 12 call sites of the 22-arg positional
   `NewRouter` constructor for a stateless decorator.
7. **Commits.** ODD step 6 calls for a work-unit commit per task. The harness safety rule is
   "never commit unless the user explicitly asks", and the user did not ask. Work stays
   uncommitted; commit identities below read "uncommitted" until the user authorizes.
   Documented deviation, not an oversight.

## Non-goals

- Any `frontend/` change (auditor UI gating, dead buttons, hardcoded badges).
- `API.md` update: it is repository-root documentation, outside `server/`. Pending user decision.
- KPI caching. Counts are read straight from Postgres.
- Audit logging for catalogue mutations: `CategoryUseCase` does not audit, so units will not either.

## Work units

### WU1 — Units of measure CRUD (admin)
- [ ] Domain: `NewUnitOfMeasure(code, name)`, `Activate()`, `Deactivate()` in `server/domain/entities/unit_of_measure.go`; new `domain.ErrCodeRequired` in `errors.go` plus its `httpx.IsValidationError` entry.
- [ ] Ports: `FindByID` + `Save` on `UnitOfMeasureRepository` (`server/domain/port/secondary/repository.go:55`); `ListAll` + `Create` + `Update` + `SetStatus` on `UnitOfMeasureUseCase` (`server/domain/port/primary/unit_of_measure.go`).
- [ ] Adapter: `FindByID` (map `pgx.ErrNoRows` to `domain.ErrNotFound`), `Save` (upsert, `isUniqueViolation` to `domain.ErrDuplicate`), `FindAll` (no `is_active` filter) in `server/infrastructure/adapters/secondary/repository/unit_repo.go`. Public `List` contract unchanged.
- [ ] DTOs in `server/aplication/dto/category.go`: `is_active` on `UnitOfMeasureDTO`, plus `CreateUnitOfMeasureRequest`, `UpdateUnitOfMeasureRequest`, `UnitOfMeasureStatusRequest`.
- [ ] Use case `server/aplication/use-cases/unit_of_measure.go`: four new methods with the same admin gate as `CategoryUseCase`.
- [ ] Handler `server/infrastructure/adapters/primary/api/handler/unit_of_measure.go`: `ListAll`, `Create`, `Update`, `SetStatus`, mirroring `handler/category.go`.
- [ ] Router `/admin` block: `GET`, `POST /units-of-measure`, `PATCH /units-of-measure/{id}`, `PATCH /units-of-measure/{id}/status`.
- [ ] Tests: use-case unit tests, router authz tests, repository integration tests; fix `stubUnitRepo` in `router_unit_test.go` for the extended port.
- Evidence: uncommitted. Focused checks: `go build ./...`, `go test ./tests/unitary/usecases/ -run UnitOfMeasure`, `go test ./infrastructure/adapters/primary/api/ -run Unit`.

Allowed edit surfaces:
`server/domain/entities/unit_of_measure.go`, `server/domain/entities/errors.go`, `server/domain/port/primary/unit_of_measure.go`, `server/domain/port/secondary/repository.go`, `server/infrastructure/adapters/secondary/repository/unit_repo.go`, `server/aplication/dto/category.go`, `server/aplication/use-cases/unit_of_measure.go`, `server/infrastructure/adapters/primary/api/handler/unit_of_measure.go`, `server/infrastructure/adapters/primary/api/router.go`, `server/infrastructure/adapters/primary/api/httpx/httpx.go`, `server/infrastructure/adapters/primary/api/router_unit_test.go`, `server/tests/unitary/usecases/unit_of_measure*_test.go`, `server/infrastructure/adapters/primary/api/router_unit_admin_authz_test.go`, `server/tests/integration/unit_of_measure_repository_test.go`

### WU2 — Route-level admin guard
- [ ] `server/infrastructure/adapters/primary/api/middleware/admin.go`: `AdminMiddleware.RequireAdmin` — no principal to 401, non-admin to 403, admin passes through, mirroring `auditor.go` conventions.
- [ ] Construct `adminMW` inside `NewRouter` and append it to every mutating route under `/admin` and to `PATCH /reports/{id}/action`.
- [ ] Leave every `GET` on the existing `Authenticate + CheckSuspension + CheckReadOnly` chain.
- [ ] Tests: `middleware/admin_test.go` for the four cases; confirm no existing api-package test regresses.
- Evidence: uncommitted. Focused checks: `go test ./infrastructure/adapters/primary/api/...`

Allowed edit surfaces:
`server/infrastructure/adapters/primary/api/middleware/admin.go`, `server/infrastructure/adapters/primary/api/middleware/admin_test.go`, `server/infrastructure/adapters/primary/api/router.go`

### WU3 — Missing admin test coverage
- [ ] `server/tests/unitary/usecases/moderation_admin_test.go`: `ListUsers`, `SetUserRole` (valid role, invalid role, non-admin), `ListAuditLogs`, reusing the fakes in `report_fakes_test.go`.
- [ ] `server/infrastructure/adapters/primary/api/router_admin_authz_test.go`: 401 anonymous, 403 non-admin and 403 auditor-on-mutation for each mutating admin route; auditor allowed on the admin `GET`s; admin happy path where the fakes allow.
- [ ] `/reports` at router level: 401 anonymous, 403 non-admin on `PATCH /reports/{id}/action`, admin resolves.
- [ ] Do not duplicate what `tests/unitary/usecases/report_test.go:481,508,533` already asserts.
- Evidence: uncommitted. Focused checks: `go test ./tests/unitary/usecases/ ./infrastructure/adapters/primary/api/`

Allowed edit surfaces:
`server/tests/unitary/usecases/moderation_admin_test.go`, `server/infrastructure/adapters/primary/api/router_admin_authz_test.go`

### WU4 — `GET /admin/stats` activity counts
- [ ] `server/domain/entities/admin_stats.go`: `AdminStats` with the five counters.
- [ ] `server/domain/port/primary/stats.go` and `AdminStatsRepository` in `server/domain/port/secondary/repository.go`.
- [ ] `server/infrastructure/adapters/secondary/repository/admin_stats.go`: one query of scalar subselects for `total_users`, `suspended_users`, `active_offerings`, `open_liquidations`, `pending_reports`. Column and enum values must be read from the migrations, never guessed.
- [ ] `server/aplication/dto/stats.go`, `server/aplication/use-cases/admin_stats.go` (admin gate), `server/infrastructure/adapters/primary/api/handler/admin_stats.go`.
- [ ] `stats *handler.AdminStatsHandler` appended as the last `NewRouter` parameter; update all 12 call sites (compiler-guided) plus the `main.go` wiring.
- [ ] Tests: use-case unit tests, router authz tests, one integration test asserting the counts against seeded rows.
- Evidence: uncommitted. Focused checks: `go build ./...`, `go test ./...`

Allowed edit surfaces:
`server/domain/entities/admin_stats.go`, `server/domain/port/primary/stats.go`, `server/domain/port/secondary/repository.go`, `server/infrastructure/adapters/secondary/repository/admin_stats.go`, `server/aplication/dto/stats.go`, `server/aplication/use-cases/admin_stats.go`, `server/infrastructure/adapters/primary/api/handler/admin_stats.go`, `server/infrastructure/adapters/primary/api/router.go`, `server/cmd/api/main.go`, `server/infrastructure/adapters/primary/api/router_*_test.go`, `server/tests/unitary/usecases/admin_stats_test.go`, `server/tests/integration/admin_stats_repo_test.go`

### WU5 — Independent verification
- [ ] `gentle-ai-verify`: `gofmt -l .` empty, `go build ./...`, `go vet ./...`, `go test ./...` with the baseline diff, plus a route-level check that each admin mutation is unreachable without the admin role.
- [ ] Report every failed or skipped check.

### WU6 — Tracker and close
- [x] `API.md` updated (user-authorized). The admin table went from 4 to 13 rows: the 5 new
  routes plus `GET /admin/users` and the 3 category routes, which were registered but never
  documented. Added the two-layer refusal contract (route guard `an administrator role is
  required` vs use case `forbidden` vs auditor `an auditor has read-only access`), the
  `AdminStatsDTO` and `UnitOfMeasureDTO` rows, the soft-delete note and the counter semantics.
- [x] Corrected the `GET /api/v1/units-of-measure` row, which described the payload as
  `id`, `code`, `name` and then said "solo `is_active`" — half Spanish, half wrong.
- [x] Fixed two stale checklist claims found while documenting, both verified against source:
  the route count (91 distinct: 86 in `NewRouter` + 5 in `RegisterTransactionRoutes`, measured
  with `chi.Walk` and by counting registrations; the document claimed 81/80 and a double
  registration that does not exist) and the claim that `PATCH /api/v1/offerings/{id}` is
  shadowed by `DeleteOffering` — it is registered once to `offering.Update`, and
  `OfferingHandler.DeleteOffering` is registered nowhere.
- [ ] Update the issue #25 acceptance checkboxes. Still requires explicit user authorization.
- [ ] Decide the auditor privilege model (see the findings below).

Method note: every claim written into `API.md` was read out of the source in this session
(handler status codes, pagination clamps, DTO field lists) or measured with `chi.Walk`, not
copied from the plan or inferred.

## Verification plan

Baseline captured before the first write (gentle-ai-verify, read-only).
Per work unit: focused tests by the writer.
After WU4: whole-suite verification by `gentle-ai-verify`, including integration tests
against testcontainers Postgres. Docker is available; no `docker compose` interaction.

## Verification (WU5)

Independent adversarial verification ran before any commit, with the instruction to falsify
rather than re-run the author's commands. Verdicts on the six claims: units CRUD admin-only
VERIFIED, every mutating admin route guarded VERIFIED, tests not vacuous VERIFIED with a caveat,
activity endpoint VERIFIED, no regression VERIFIED, public contract VERIFIED.

Per work unit:

| Work unit | State | Evidence |
| --- | --- | --- |
| WU1 units CRUD | implemented, uncommitted | integration tests prove real-repo persistence; `gofmt` now clean in the touched files |
| WU2 admin guard | implemented, uncommitted | all ten mutating admin routes enumerated with the guard; no GET carries it |
| WU3 missing tests | implemented, uncommitted | plus the reports route guard the first pass missed |
| WU4 activity endpoint | implemented, uncommitted | SQL runs against the real schema; `active_offerings` reuses `cataloguePredicate` |
| WU5 verification | done | `go build`, `go vet`, `go test -count=1 ./...` and `-race` on touched packages all green; `gofmt -l` down from 10 files to 3, all three untouched by this feature |

The verifier's five findings and their disposition:

| Finding | Severity | Disposition |
| --- | --- | --- |
| The `router.go` comment implied an auditor can read the admin GETs | low | fixed: the comment now states only that the `RoleAdmin` gate lives inside each use case |
| No test pinned the route guard's wiring, because removing `adminMW.RequireAdmin` left every test green on the strength of the use-case gate alone | low, but a real verification hole | fixed: `TestAdminMutationRoutesEnforceTheRouteGuardOverPermissiveUseCases` wires permissive use-case stubs and asserts 403 for a non-admin on all ten mutating routes plus a 2xx for an admin. Proven discriminating by temporarily removing the guard (the farmer got 201 and the test failed), then restoring it |
| Auditor 403 on `GET /admin/units-of-measure` and `GET /admin/stats` was reasoned, never executed | low | fixed: both have explicit auditor-refusal tests |
| Duplicate to 409 and over-long to 400 had no HTTP-level assertion | info | fixed: both asserted at the router level |
| The new units integration test used fixed codes while `cleanupTables` does not truncate `units_of_measure`, so `go test -count=2` failed on the unique constraint | low, real defect | fixed: each test deletes the exact rows it created through `t.Cleanup`, and `-count=2` passes. The shared harness was deliberately left alone, because truncating `units_of_measure` would delete the migration seed rows other tests depend on |

## Findings escalated to the user

These were found while implementing and are deliberately NOT fixed:

1. **The auditor role cannot read anything on the admin surface.** `ListUsers`,
   `ListAuditLogs` and `ReportUseCase.List` require `RoleAdmin` in the use case, so an auditor
   token gets 403 on `GET /admin/users`, `GET /admin/audit-logs` and `GET /reports`, while the
   frontend lets auditors into the whole `/admin` area. Pre-existing, not caused by this feature.
   Either the auditor gets a server-side read path or the frontend stops granting the panel.
2. **`SetUserRole` rejects an invalid role with `domain.ErrInvalidInput`, not
   `domain.ErrValidRoleRequired`** (`server/aplication/use-cases/report.go:344`), so the sentinel
   named for that case is unreachable on this path. Both map to HTTP 400. Cosmetic.
3. **`PATCH /reports/{id}/action` initially shipped without the route guard** in WU2; found by
   the WU3 writer and fixed in the WU3 follow-up. The 403 had been coming from the use case only.

## Risks

- `NewRouter` is a 22-arg positional constructor and needs a 23rd argument in WU4, touching
  10 files. The compiler finds every call site; the mechanical edit is contained in one work unit.
- WU4 guesses nothing about schema: column and enum values are read from
  `server/infrastructure/adapters/secondary/repository/migrations/`.
- Keeping the use-case admin checks means two enforcement points. Deliberate; documented in decision 5.
